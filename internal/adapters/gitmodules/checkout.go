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

// Fetch reads one revision and hashes its tracked contents. Temporary checkouts are removed before returning.
func (c *Cache) Fetch(ctx context.Context, source, version string) (modules.Fetched, error) {
	cacheRoot := c.root
	checkout, err := checkoutV1Module(ctx, source, version, cacheRoot)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("checkoutV1Module: %w", err)
	}
	defer func() { _ = os.RemoveAll(checkout.dir) }()
	files, err := trackedV1Files(ctx, checkout.dir)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("trackedV1Files: %w", err)
	}
	hash, err := hashV1Files(checkout.dir, files)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hashV1Files: %w", err)
	}
	if version == "" {
		version = checkout.commit
	}
	return modules.Fetched{Module: checkout.module, Lock: v1.LockedModule{
		Source: source, Version: version, Commit: checkout.commit, Hash: hash,
	}}, nil
}

type v1ModuleCheckout struct {
	dir    string
	module v1.Module
	commit string
}

// checkoutV1Module selects a repository revision. The caller owns the returned
// directory and removes it after reading its metadata and tracked contents.
func checkoutV1Module(ctx context.Context, source, version, cacheRoot string) (checkout v1ModuleCheckout, err error) {
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("MkdirAll: %w", err)
	}
	if version == "" {
		dir, module, commit, err := cloneHeadV1GitModule(ctx, source, cacheRoot)
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("cloneHeadV1GitModule: %w", err)
		}
		return v1ModuleCheckout{dir: dir, module: module, commit: strings.TrimSpace(commit)}, nil
	}

	var dir string
	// Ownership transfers to the caller only after checkout and metadata agree.
	defer func() {
		if err != nil && dir != "" {
			_ = os.RemoveAll(dir)
		}
	}()
	if v1.IsCommitRef(version) {
		dir, err = clonePinnedV1GitModule(ctx, v1.LockedModule{Source: source, Commit: version}, cacheRoot)
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("clonePinnedV1GitModule: %w", err)
		}
		commit, err := gitV1(ctx, dir, "rev-parse", "HEAD")
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("gitV1: %w", err)
		}
		module, err := moduleconfig.ReadGitDependency(dir, source)
		if err != nil {
			return v1ModuleCheckout{}, fmt.Errorf("ReadGitDependency: %w", err)
		}
		return v1ModuleCheckout{dir: dir, module: module, commit: strings.TrimSpace(commit)}, nil
	}

	candidate, err := findV1GitModuleTag(ctx, source, version)
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("findV1GitModuleTag: %w", err)
	}
	dir, err = os.MkdirTemp(cacheRoot, "git-*")
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("MkdirTemp: %w", err)
	}
	tag := candidate.tag(version)
	if _, err := gitV1(ctx, "", "clone", "--quiet", "--depth=1", "--branch", tag, "--no-checkout", "--", candidate.remote, dir); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("gitV1: %w", err)
	}
	commit, err := gitV1(ctx, dir, "rev-parse", "--verify", "refs/tags/"+tag+"^{commit}")
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("gitV1: %w", err)
	}
	commit = strings.TrimSpace(commit)
	if _, err := gitV1(ctx, dir, "checkout", "--quiet", "--detach", commit); err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("gitV1: %w", err)
	}
	module, err := moduleconfig.ReadGitDependencyAt(dir, source, candidate.subdir)
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("ReadGitDependencyAt: %w", err)
	}
	return v1ModuleCheckout{dir: dir, module: module, commit: commit}, nil
}
