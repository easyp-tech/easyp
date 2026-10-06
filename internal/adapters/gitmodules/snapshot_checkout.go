package gitmodules

import (
	"context"
	"fmt"
	"os"
	"strings"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Fetch resolves a revision to the current immutable logical snapshot policy.
func (c *Cache) Fetch(ctx context.Context, source, version string) (modules.Fetched, error) {
	checkout, err := checkoutV1Module(ctx, source, version, c.root)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("checkoutV1Module: %w", err)
	}
	defer func() { _ = os.RemoveAll(checkout.dir); _ = os.RemoveAll(checkout.snapshot) }()
	if err := moduleconfig.ValidateLegacyMajor(checkout.snapshot, source, version); err != nil {
		return modules.Fetched{}, fmt.Errorf("ValidateLegacyMajor: %w", err)
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
	return modules.Fetched{Module: module, Lock: v1.LockedModule{Source: source, Version: version, Commit: checkout.commit, Hash: hash, BSR: bindings}}, nil
}

// checkoutV1Module retains Git objects and an index for historical proofs and a
// bounded regular logical snapshot for all current metadata and hashing.
func checkoutV1Module(ctx context.Context, source, version, cacheRoot string) (checkout v1ModuleCheckout, err error) {
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
		checkout, cloned, candidateErr := checkoutSnapshotCandidate(ctx, dir, source, version, candidate)
		if candidateErr == nil {
			return checkout, nil
		}
		_ = os.RemoveAll(dir)
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

func checkoutSnapshotCandidate(ctx context.Context, dir, source, version string, candidate v1GitModuleCandidate) (v1ModuleCheckout, bool, error) {
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
	stage, module, err := prepareV1Snapshot(ctx, dir, commit, source, candidate.subdir)
	if err != nil {
		return v1ModuleCheckout{}, true, fmt.Errorf("prepareV1Snapshot: %w", err)
	}
	return v1ModuleCheckout{dir: dir, module: module, commit: commit, snapshot: stage}, true, nil
}

func fetchPinnedV1Module(ctx context.Context, entry v1.LockedModule, cacheRoot, installed string) error {
	checkout, err := checkoutV1Module(ctx, entry.Source, entry.Commit, cacheRoot)
	if err != nil {
		return fmt.Errorf("checkoutV1Module: %w", err)
	}
	defer func() { _ = os.RemoveAll(checkout.dir); _ = os.RemoveAll(checkout.snapshot) }()
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
