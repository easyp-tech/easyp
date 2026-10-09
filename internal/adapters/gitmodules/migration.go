package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"golang.org/x/mod/semver"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// migrationProofIntent retains the public nil/empty roots distinction even
// after a selection's roots have been inferred.
type migrationProofIntent uint8

const (
	_ migrationProofIntent = iota
	migrationWholeNamespaceProof
	migrationRetainedRootsProof
	migrationExplicitRootsProof
)

type migrationRequest struct {
	source, version, legacyHash string
	roots                       []string
	proof                       migrationProofIntent
	nativeModule                bool
}

// migrationRequestForRoots receives roots that have passed ValidateModuleRoots.
// Empty selection roots retain producer authority or intrinsic root inference.
func migrationRequestForRoots(source, version, legacyHash string, checkedRoots []string) migrationRequest {
	request := migrationRequest{
		source: source, version: version, legacyHash: legacyHash,
		roots: slices.Clone(checkedRoots),
	}
	switch {
	case checkedRoots == nil:
		request.proof = migrationWholeNamespaceProof
	case len(checkedRoots) == 0:
		request.proof = migrationRetainedRootsProof
	default:
		request.proof = migrationExplicitRootsProof
	}
	return request
}

// FetchMigration verifies a legacy installed-tree hash before calculating the
// v1 logical snapshot hash. Legacy repositories must preserve their proto source
// selection and import names. An empty version requires an empty legacyHash and
// permits initial resolution; native v1 repositories need no legacy comparison.
func (c *Cache) FetchMigration(ctx context.Context, source, version, legacyHash string) (fetched modules.Fetched, err error) {
	fetched, err = c.FetchMigrationWithRoots(ctx, source, version, legacyHash, nil)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("FetchMigrationWithRoots: %w", err)
	}
	return fetched, nil
}

// FetchMigrationWithRoots verifies the historical installed layout before root
// selection. Non-nil roots request a consumer selection proof and return its
// immutable legacy-to-repository mapping; nil retains whole-namespace checks.
func (c *Cache) FetchMigrationWithRoots(ctx context.Context, source, version, legacyHash string, roots []string) (modules.Fetched, error) {
	fetched, err := c.fetchMigration(ctx, source, version, legacyHash, roots, false)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("fetchMigration: %w", err)
	}
	return fetched, nil
}

// FetchMigrationImports verifies the complete legacy archive before adapting a
// dependency's Buf source filters. The returned archive mapping retains excluded
// files as unavailable sources; callers must prove their consumer import bindings.
// Repositories without producer filters retain the whole-namespace guarantee.
func (c *Cache) FetchMigrationImports(ctx context.Context, source, version, legacyHash string) (modules.Fetched, error) {
	fetched, err := c.fetchMigration(ctx, source, version, legacyHash, nil, true)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("fetchMigration: %w", err)
	}
	return fetched, nil
}

func (c *Cache) fetchMigration(ctx context.Context, source, version, legacyHash string, roots []string, imports bool) (fetched modules.Fetched, err error) {
	if (version != "" || legacyHash != "") && !v1.IsCommitRef(version) && !semver.IsValid(version) {
		return modules.Fetched{}, fmt.Errorf("migration version %q must be a full Git commit or SemVer tag", version)
	}
	if err := v1.ValidateModuleRoots(roots); err != nil {
		return modules.Fetched{}, fmt.Errorf("ValidateModuleRoots: %w", err)
	}
	request := migrationRequestForRoots(source, version, legacyHash, roots)
	checkout, err := checkoutV1ModuleWithRoots(ctx, request.source, request.version, c.root, nil, true)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("checkoutV1ModuleWithRoots: %w", err)
	}
	defer func() {
		for _, directory := range []string{checkout.snapshot, checkout.dir} {
			removeErr := os.RemoveAll(directory)
			if removeErr != nil {
				fetched = modules.Fetched{}
				err = errors.Join(err, fmt.Errorf("RemoveAll: %w", removeErr))
			}
		}
	}()
	if err := moduleconfig.ValidateLegacyMajor(checkout.snapshot, request.source, request.version); err != nil {
		return modules.Fetched{}, fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	tracked, err := migrationTrackedFiles(ctx, checkout.dir)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("migrationTrackedFiles: %w", err)
	}
	files, err := snapshotV1Files(checkout.snapshot)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("snapshotV1Files: %w", err)
	}
	if err := validateMigrationLegacyReplacements(checkout.snapshot, files); err != nil {
		return modules.Fetched{}, fmt.Errorf("validateMigrationLegacyReplacements: %w", err)
	}
	request.nativeModule, err = hasNativeMigrationModule(checkout.snapshot, files, request.source)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hasNativeMigrationModule: %w", err)
	}
	// Buf-filtered snapshots omit excluded source bytes. Reproduce their old
	// archive from the immutable Git view, never from the filtered v1 snapshot.
	// A consumer proof decides whether any omitted file is actually required.
	if imports && !request.nativeModule && len(checkout.module.ProtoFilters) > 0 {
		request.proof = migrationRetainedRootsProof
	}
	var layout migrationLegacyLayout
	if request.proof != migrationWholeNamespaceProof {
		layout, err = proveMigrationLegacyLayout(ctx, checkout, tracked, request)
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("proveMigrationLegacyLayout: %w", err)
		}
		if request.proof == migrationRetainedRootsProof {
			inferred, err := modules.ResolveIntrinsicImportRoots(ctx, checkout.module, checkout.inspection)
			if err != nil {
				return modules.Fetched{}, fmt.Errorf("ResolveIntrinsicImportRoots: %w", err)
			}
			request.roots = inferred
		}
	}
	checkout.module, err = applyV1ModuleRoots(checkout.module, request.roots)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("applyV1ModuleRoots: %w", err)
	}
	view, err := sourceV1SnapshotView(ctx, checkout.dir, checkout.commit)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("sourceV1SnapshotView: %w", err)
	}
	if err := materializeSelectedSnapshotAliases(ctx, view, checkout.snapshot, checkout.module, nil); err != nil {
		return modules.Fetched{}, fmt.Errorf("materializeSelectedSnapshotAliases: %w", err)
	}
	checkout.inspection, err = inspectV1Snapshot(ctx, view, checkout.snapshot, checkout.module, checkout.commit)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("inspectV1Snapshot: %w", err)
	}
	if request.proof == migrationWholeNamespaceProof {
		if !request.nativeModule || request.legacyHash != "" {
			if err := verifyMigrationLegacyHash(ctx, checkout, tracked, request); err != nil {
				return modules.Fetched{}, fmt.Errorf("verifyMigrationLegacyHash: %w", err)
			}
		}
		layout, err = proveMigrationLegacyLayout(ctx, checkout, tracked, request)
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("proveMigrationLegacyLayout: %w", err)
		}
	}
	checkout.inspection.LegacyFiles = layout.files
	if _, err := migrationSelectedFiles(checkout.snapshot, checkout.module); err != nil {
		return modules.Fetched{}, fmt.Errorf("migrationSelectedFiles: %w", err)
	}
	hash, err := hashSnapshotV1Files(checkout.snapshot)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hashSnapshotV1Files: %w", err)
	}
	if request.version == "" {
		request.version = checkout.commit
	}
	module, bindings, err := c.resolveBSRDependencies(ctx, checkout.module)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("resolveBSRDependencies: %w", err)
	}
	return modules.Fetched{Module: module, Lock: v1.LockedModule{
		Source: request.source, Version: request.version, Commit: checkout.commit, Hash: hash, Roots: lockedV1ModuleRoots(module, request.roots), BSR: bindings,
	}, Inspection: checkout.inspection}, nil
}
