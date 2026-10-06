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
// v1 tracked-checkout hash. Legacy repositories must preserve their proto source
// selection and import names. An empty version requires an empty legacyHash and
// permits initial resolution; native v1 repositories need no legacy comparison.
func (c *Cache) FetchMigration(ctx context.Context, source, version, legacyHash string) (fetched modules.Fetched, err error) {
	if (version != "" || legacyHash != "") && !v1.IsCommitRef(version) && !semver.IsValid(version) {
		return modules.Fetched{}, fmt.Errorf("migration version %q must be a full Git commit or SemVer tag", version)
	}
	checkout, err := checkoutV1Module(ctx, source, version, c.root)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("checkoutV1Module: %w", err)
	}
	defer func() {
		removeErr := os.RemoveAll(checkout.dir)
		if removeErr != nil {
			fetched = modules.Fetched{}
			err = errors.Join(err, fmt.Errorf("RemoveAll: %w", removeErr))
		}
	}()
	if err := moduleconfig.ValidateLegacyMajor(checkout.dir, source, version); err != nil {
		return modules.Fetched{}, fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	tracked, err := migrationTrackedFiles(ctx, checkout.dir, checkout.module.Roots...)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("migrationTrackedFiles: %w", err)
	}
	if err := validateMigrationLegacyReplacements(checkout.dir, tracked.regularFiles); err != nil {
		return modules.Fetched{}, fmt.Errorf("validateMigrationLegacyReplacements: %w", err)
	}
	native, err := hasNativeMigrationModule(checkout.dir, tracked.regularFiles, source)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hasNativeMigrationModule: %w", err)
	}
	if !native || legacyHash != "" {
		if err := verifyMigrationLegacyHash(ctx, checkout, source, legacyHash, tracked); err != nil {
			return modules.Fetched{}, fmt.Errorf("verifyMigrationLegacyHash: %w", err)
		}
		if err := validateMigrationSelection(checkout.dir, tracked.regularFiles, checkout.module); err != nil {
			return modules.Fetched{}, fmt.Errorf("validateMigrationSelection: %w", err)
		}
	}
	files := selectV1ProtoFiles(tracked.regularFiles, checkout.module.ProtoFilters)
	hash, err := hashV1Files(checkout.dir, files)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hashV1Files: %w", err)
	}
	if version == "" {
		version = checkout.commit
	}
	module, bindings, err := c.resolveBSRDependencies(ctx, checkout.module)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("resolveBSRDependencies: %w", err)
	}
	return modules.Fetched{Module: module, Lock: v1.LockedModule{
		Source: source, Version: version, Commit: checkout.commit, Hash: hash, BSR: bindings,
	}}, nil
}
