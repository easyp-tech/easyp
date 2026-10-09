package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Fetch resolves a revision to the current immutable logical snapshot policy.
func (c *Cache) Fetch(ctx context.Context, source, version string) (modules.Fetched, error) {
	fetched, err := c.FetchWithRoots(ctx, source, version, nil)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("FetchWithRoots: %w", err)
	}
	return fetched, nil
}

// FetchWithRoots resolves a revision with bounded roots when metadata declares none.
// Authoritative native, Buf, and legacy EasyP roots may only be repeated unchanged.
func (c *Cache) FetchWithRoots(ctx context.Context, source, version string, roots []string) (modules.Fetched, error) {
	fetched, err := c.fetchWithRootSelection(ctx, source, version, roots, len(roots) > 0)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("fetchWithRootSelection: %w", err)
	}
	return fetched, nil
}

// FetchForRootResolution inspects pinned logical sources before roots are inferred.
// Metadata-free inspections have an incomplete lock hash and cannot be installed.
func (c *Cache) FetchForRootResolution(ctx context.Context, source, version string) (modules.Fetched, error) {
	fetched, err := c.fetchWithRootSelection(ctx, source, version, nil, true)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("fetchWithRootSelection: %w", err)
	}
	return fetched, nil
}

func (c *Cache) fetchWithRootSelection(ctx context.Context, source, version string, roots []string, inspect bool) (fetched modules.Fetched, err error) {
	ctx = c.operationContext(ctx, source, version)
	finish := gitcommand.Start(ctx, "resolve module")
	defer func() { finish(err) }()
	checkout, err := checkoutV1ModuleWithRoots(ctx, source, version, c.root, roots, inspect)
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
	var hash string
	if checkout.inspection == nil || !checkout.inspection.Provisional {
		hash, err = hashSnapshotV1Files(checkout.snapshot)
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("hashSnapshotV1Files: %w", err)
		}
	}
	if version == "" {
		version = checkout.commit
	}
	module, bindings, err := c.resolveBSRDependencies(ctx, checkout.module)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("resolveBSRDependencies: %w", err)
	}
	return modules.Fetched{Module: module, Lock: v1.LockedModule{Source: source, Version: version, Commit: checkout.commit, Hash: hash, Roots: lockedV1ModuleRoots(module, roots), BSR: bindings}, Inspection: checkout.inspection}, nil
}

// checkoutV1ModuleWithRoots retains Git objects and an index for historical
// proofs, plus a bounded logical snapshot or deferred source inspection.
func checkoutV1ModuleWithRoots(ctx context.Context, source, version, cacheRoot string, roots []string, inspect bool) (checkout v1ModuleCheckout, err error) {
	if _, err := gitcommand.Timeout(); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("Timeout: %w", err)
	}
	if err := v1.ValidateModuleRoots(roots); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("ValidateModuleRoots: %w", err)
	}
	if err := v1.ValidateModuleVersion(source, version); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("ValidateModuleVersion: %w", err)
	}
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("MkdirAll: %w", err)
	}
	var candidates []v1GitModuleCandidate
	if version != "" && !v1.IsCommitRef(version) {
		candidate, err := findV1GitModuleTag(ctx, source, version)
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("findV1GitModuleTag: %w", err)
		}
		candidates = []v1GitModuleCandidate{candidate}
	} else {
		candidates, err = v1GitModuleCandidates(source)
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("v1GitModuleCandidates: %w", err)
		}
	}
	var firstErr, moduleErr error
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("Err: %w", err)
		}
		dir, err := os.MkdirTemp(cacheRoot, "git-*")
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("MkdirTemp: %w", err)
		}
		candidateContext := gitcommand.WithAttributes(ctx, slog.String("repository", candidate.remote), slog.String("module_directory", candidate.subdir))
		checkout, cloned, candidateErr := checkoutSnapshotCandidate(candidateContext, dir, source, version, candidate, roots, inspect)
		if candidateErr == nil {
			return checkout, nil
		}
		removeErr := os.RemoveAll(dir)
		if removeErr != nil {
			return v1ModuleCheckout{}, errors.Join(fmt.Errorf("checkoutSnapshotCandidate: %w", candidateErr), fmt.Errorf("RemoveAll: %w", removeErr))
		}
		if gitcommand.Interrupted(candidateErr) {
			return v1ModuleCheckout{}, fmt.Errorf("checkoutSnapshotCandidate: %w", candidateErr)
		}
		if firstErr == nil {
			firstErr = candidateErr
		}
		if cloned {
			moduleErr = candidateErr
		}
	}
	if moduleErr != nil {
		return v1ModuleCheckout{}, fmt.Errorf("%s: %w", source, moduleErr)
	}
	if v1.IsCommitRef(version) {
		return v1ModuleCheckout{}, fmt.Errorf("%s: could not fetch locked commit %s; check repository access before changing the lock: %w", source, version, firstErr)
	}
	return v1ModuleCheckout{}, fmt.Errorf("could not fetch Git module %s@%s: %w", source, version, firstErr)
}

func checkoutSnapshotCandidate(ctx context.Context, dir, source, version string, candidate v1GitModuleCandidate, roots []string, inspect bool) (v1ModuleCheckout, bool, error) {
	if v1.IsCommitRef(version) {
		cloned, err := checkoutCachedCommit(ctx, dir, v1.LockedModule{Source: source, Commit: version}, candidate)
		if err != nil {
			return v1ModuleCheckout{}, cloned, err
		}
	} else {
		args := []string{"clone", "--quiet", "--depth=1", "--no-checkout"}
		if version != "" {
			args = append(args, "--branch", candidate.tag(strings.TrimSuffix(version, "+incompatible")))
		}
		args = append(args, "--", candidate.remote, dir)
		if _, err := gitV1(ctx, "", args...); err != nil {
			return v1ModuleCheckout{}, false, fmt.Errorf("gitV1: %w", err)
		}
	}
	revision := "HEAD"
	if version != "" && !v1.IsCommitRef(version) {
		revision = "refs/tags/" + candidate.tag(strings.TrimSuffix(version, "+incompatible")) + "^{commit}"
	}
	commit, err := gitV1(ctx, dir, "rev-parse", "--verify", revision)
	if err != nil {
		return v1ModuleCheckout{}, true, fmt.Errorf("gitV1: %w", err)
	}
	commit = strings.TrimSpace(commit)
	if _, err := gitV1(ctx, dir, "read-tree", commit); err != nil {
		return v1ModuleCheckout{}, true, fmt.Errorf("gitV1: %w", err)
	}
	stage, module, inspection, err := prepareV1Snapshot(ctx, dir, commit, source, candidate.subdir, roots, inspect)
	if err != nil {
		return v1ModuleCheckout{}, true, fmt.Errorf("prepareV1Snapshot: %w", err)
	}
	return v1ModuleCheckout{dir: dir, module: module, commit: commit, snapshot: stage, inspection: inspection}, true, nil
}

func fetchPinnedV1Module(ctx context.Context, entry v1.LockedModule, cacheRoot, installed string) (err error) {
	checkout, err := checkoutV1ModuleWithRoots(ctx, entry.Source, entry.Commit, cacheRoot, entry.Roots, false)
	if err != nil {
		return fmt.Errorf("checkoutV1ModuleWithRoots: %w", err)
	}
	defer func() {
		for _, directory := range []string{checkout.snapshot, checkout.dir} {
			removeErr := os.RemoveAll(directory)
			if removeErr != nil {
				err = errors.Join(err, fmt.Errorf("RemoveAll: %w", removeErr))
			}
		}
	}()
	actual, err := hashSnapshotV1Files(checkout.snapshot)
	if err != nil {
		return fmt.Errorf("hashSnapshotV1Files: %w", err)
	}
	if !strings.EqualFold(checkout.commit, entry.Commit) {
		return fmt.Errorf("%s: snapshot commit does not match lock", entry.Source)
	}
	if actual != entry.Hash {
		return fmt.Errorf("%s@%s hash mismatch: got %s, want %s; downloaded contents do not match the pinned hash; investigate repository integrity and keep protobuf.lock unchanged", entry.Source, entry.Commit, actual, entry.Hash)
	}
	if err := moduleconfig.ValidateLegacyMajor(checkout.snapshot, entry.Source, entry.Version); err != nil {
		return fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	if err := os.Rename(checkout.snapshot, installed); err != nil {
		return fmt.Errorf("Rename: %w", err)
	}
	return nil
}
