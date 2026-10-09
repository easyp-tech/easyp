// Package gitmodules retrieves and verifies Git module contents in the v1 cache.
package gitmodules

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/adapters/bsr"
	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Cache owns Git checkouts, installed contents, and their layout below EASYPPATH.
type Cache struct {
	root        string
	bsrResolver modules.BSRResolver
	logger      logger.Logger
}

// WithLogger returns an isolated cache handle with dependency progress logging.
// Historical-pin scopes preserve this logger through their existing shallow copy.
func (c *Cache) WithLogger(log logger.Logger) *Cache {
	scoped := *c
	scoped.logger = log
	return &scoped
}

func (c *Cache) operationContext(ctx context.Context, source, version string) context.Context {
	if c.logger == nil {
		return ctx
	}
	return gitcommand.WithLogger(ctx, c.logger.With(slog.String("source", source), slog.String("version", version)))
}

// New places the v1 Git cache below the supplied EasyP storage directory.
func New(storageDir string) *Cache {
	return NewWithBSRResolver(storageDir, bsr.StaticResolver{})
}

// NewWithBSRResolver supplies the backend used when fetching Buf dependency metadata.
// Cached and frozen operations replay the lock's bindings without calling this backend.
func NewWithBSRResolver(storageDir string, resolver modules.BSRResolver) *Cache {
	return &Cache{root: filepath.Join(storageDir, "v1", "git"), bsrResolver: resolver}
}

// SourceCacheDirectories returns repository-owned storage that consumer source
// traversal and writes must exclude, including objects and unused snapshots.
func (c *Cache) SourceCacheDirectories() []string {
	return []string{c.root}
}

// Cached reads installed metadata without downloading or changing files.
// Call Install first to verify contents against the lock.
func (c *Cache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	if err := v1.ValidateModuleVersion(entry.Source, entry.Version); err != nil {
		return "", v1.Module{}, fmt.Errorf("ValidateModuleVersion: %w", err)
	}
	if err := v1.ValidateModuleRoots(entry.Roots); err != nil {
		return "", v1.Module{}, fmt.Errorf("ValidateModuleRoots: %w", err)
	}
	directory := v1ModuleCachePath(c.root, entry)
	module, err := readCachedV1Module(directory, entry.Source)
	if err != nil {
		return "", v1.Module{}, fmt.Errorf("readCachedV1Module: %w", err)
	}
	module, err = applyV1ModuleRoots(module, entry.Roots)
	if err != nil {
		return "", v1.Module{}, fmt.Errorf("applyV1ModuleRoots: %w", err)
	}
	if err := moduleconfig.ValidateLegacyMajor(directory, entry.Source, entry.Version); err != nil {
		return "", v1.Module{}, fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	module, err = restoreBSRDependencies(module, entry.BSR)
	if err != nil {
		return "", v1.Module{}, fmt.Errorf("restoreBSRDependencies: %w", err)
	}
	return directory, module, nil
}

// Versions lists semantic versions available for a module, including nested modules.
func (c *Cache) Versions(ctx context.Context, source string) (versions []string, err error) {
	ctx = c.operationContext(ctx, source, "")
	finish := gitcommand.Start(ctx, "list module versions")
	defer func() { finish(err) }()
	return listV1ModuleTags(ctx, source)
}

// ValidateSource checks whether a module identity can identify a Git repository.
func ValidateSource(source string) error {
	_, err := v1GitModuleCandidates(source)
	return err
}

// readCachedV1Module reuses the verified major-layout rules without requesting
// network metadata or letting an unrelated native module poison cache lookup.
func readCachedV1Module(directory, source string) (v1.Module, error) {
	major, err := v1.ModulePathMajor(source)
	if err != nil {
		return v1.Module{}, fmt.Errorf("ModulePathMajor: %w", err)
	}
	if major == "" {
		return moduleconfig.ReadGitDependency(directory, source)
	}
	candidates, err := v1GitModuleCandidates(source)
	if err != nil {
		return v1.Module{}, fmt.Errorf("v1GitModuleCandidates: %w", err)
	}
	var firstErr error
	for _, candidate := range candidates {
		module, err := moduleconfig.ReadGitDependencyAt(directory, source, candidate.subdir)
		if err == nil {
			return module, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return v1.Module{}, fmt.Errorf("ReadGitDependencyAt: %w", firstErr)
}
