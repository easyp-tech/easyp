package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"os"

	"golang.org/x/mod/semver"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

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
func (c *Cache) FetchMigrationWithRoots(ctx context.Context, source, version, legacyHash string, roots []string) (fetched modules.Fetched, err error) {
	if (version != "" || legacyHash != "") && !v1.IsCommitRef(version) && !semver.IsValid(version) {
		return modules.Fetched{}, fmt.Errorf("migration version %q must be a full Git commit or SemVer tag", version)
	}
	if err := v1.ValidateModuleRoots(roots); err != nil {
		return modules.Fetched{}, fmt.Errorf("ValidateModuleRoots: %w", err)
	}
	checkout, err := checkoutV1ModuleWithRoots(ctx, source, version, c.root, nil, true)
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
	if err := moduleconfig.ValidateLegacyMajor(checkout.snapshot, source, version); err != nil {
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
	native, err := hasNativeMigrationModule(checkout.snapshot, files, source)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hasNativeMigrationModule: %w", err)
	}
	var layout migrationLegacyLayout
	if roots != nil {
		layout, err = proveMigrationLegacyLayout(ctx, checkout, source, legacyHash, tracked, native, true)
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("proveMigrationLegacyLayout: %w", err)
		}
		if len(roots) == 0 {
			inferred, err := modules.ResolveIntrinsicImportRoots(ctx, checkout.module, checkout.inspection)
			if err != nil {
				return modules.Fetched{}, fmt.Errorf("ResolveIntrinsicImportRoots: %w", err)
			}
			roots = inferred
		}
	}
	checkout.module, err = applyV1ModuleRoots(checkout.module, roots)
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
	if layout.files == nil {
		if !native || legacyHash != "" {
			if err := verifyMigrationLegacyHash(ctx, checkout, source, legacyHash, tracked); err != nil {
				return modules.Fetched{}, fmt.Errorf("verifyMigrationLegacyHash: %w", err)
			}
		}
		layout, err = proveMigrationLegacyLayout(ctx, checkout, source, legacyHash, tracked, native, false)
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
	if version == "" {
		version = checkout.commit
	}
	module, bindings, err := c.resolveBSRDependencies(ctx, checkout.module)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("resolveBSRDependencies: %w", err)
	}
	return modules.Fetched{Module: module, Lock: v1.LockedModule{
		Source: source, Version: version, Commit: checkout.commit, Hash: hash, Roots: lockedV1ModuleRoots(module, roots), BSR: bindings,
	}, Inspection: checkout.inspection}, nil
}
