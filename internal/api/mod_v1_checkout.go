package api

import (
	"context"
	"fmt"
	"os"
	"strings"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func fetchV1Module(ctx context.Context, source, version, cacheRoot string) (fetchedV1Module, error) {
	checkout, err := checkoutV1Module(ctx, source, version, cacheRoot)
	if err != nil {
		return fetchedV1Module{}, fmt.Errorf("checkoutV1Module: %w", err)
	}
	defer func() { _ = os.RemoveAll(checkout.dir) }()
	files, err := trackedV1Files(ctx, checkout.dir)
	if err != nil {
		return fetchedV1Module{}, fmt.Errorf("trackedV1Files: %w", err)
	}
	hash, err := hashV1Files(checkout.dir, files)
	if err != nil {
		return fetchedV1Module{}, fmt.Errorf("hashV1Files: %w", err)
	}
	if version == "" {
		version = checkout.commit
	}
	return fetchedV1Module{config: checkout.module, lock: v1.LockedModule{
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
func checkoutV1Module(ctx context.Context, source, version, cacheRoot string) (v1ModuleCheckout, error) {
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
	if v1.IsCommitRef(version) {
		return checkoutCommittedV1Module(ctx, source, version, cacheRoot)
	}
	return checkoutTaggedV1Module(ctx, source, version, cacheRoot)
}

func checkoutCommittedV1Module(ctx context.Context, source, commit, cacheRoot string) (v1ModuleCheckout, error) {
	dir, err := clonePinnedV1GitModule(ctx, v1.LockedModule{Source: source, Commit: commit}, cacheRoot)
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("clonePinnedV1GitModule: %w", err)
	}
	selected, err := readCheckedOutV1Module(ctx, dir, source)
	if err != nil {
		_ = os.RemoveAll(dir)
		return v1ModuleCheckout{}, fmt.Errorf("readCheckedOutV1Module: %w", err)
	}
	return selected, nil
}

func checkoutTaggedV1Module(ctx context.Context, source, version, cacheRoot string) (v1ModuleCheckout, error) {
	candidate, err := findV1GitModuleTag(ctx, source, version)
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("findV1GitModuleTag: %w", err)
	}
	dir, err := os.MkdirTemp(cacheRoot, "git-*")
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("MkdirTemp: %w", err)
	}
	selected, err := checkoutV1Tag(ctx, dir, source, candidate, version)
	if err != nil {
		_ = os.RemoveAll(dir)
		return v1ModuleCheckout{}, fmt.Errorf("checkoutV1Tag: %w", err)
	}
	return selected, nil
}

func checkoutV1Tag(ctx context.Context, dir, source string, candidate v1GitModuleCandidate, version string) (v1ModuleCheckout, error) {
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
	module, err := readV1GitModuleCandidate(dir, source, candidate)
	if err != nil {
		return v1ModuleCheckout{}, fmt.Errorf("readV1GitModuleCandidate: %w", err)
	}
	return v1ModuleCheckout{dir: dir, module: module, commit: commit}, nil
}

func readCheckedOutV1Module(ctx context.Context, dir, source string) (v1ModuleCheckout, error) {
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
