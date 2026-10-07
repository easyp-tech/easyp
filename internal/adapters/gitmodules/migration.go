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

// FetchMigrationWithRoots preserves legacy integrity proofs while selecting roots
// for a dependency that has no authoritative source metadata.
func (c *Cache) FetchMigrationWithRoots(ctx context.Context, source, version, legacyHash string, roots []string) (fetched modules.Fetched, err error) {
	if (version != "" || legacyHash != "") && !v1.IsCommitRef(version) && !semver.IsValid(version) {
		return modules.Fetched{}, fmt.Errorf("migration version %q must be a full Git commit or SemVer tag", version)
	}
	checkout, err := checkoutV1ModuleWithRoots(ctx, source, version, c.root, roots, false)
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
	_, err = migrationSelectedFiles(checkout.snapshot, checkout.module)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("migrationSelectedFiles: %w", err)
	}
	if !native || legacyHash != "" {
		if err := verifyMigrationLegacyHash(ctx, checkout, source, legacyHash, tracked); err != nil {
			return modules.Fetched{}, fmt.Errorf("verifyMigrationLegacyHash: %w", err)
		}
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
	}}, nil
}
