package api

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Update refreshes tagged requirements within their major versions and
// versionless requirements to Git HEAD; explicit commit pins remain fixed.
func (m Mod) Update(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	original, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
	}
	if len(module.Replaces) > 0 {
		return fmt.Errorf("module %s: remove local replacements before updating a reproducible lock", module.Name)
	}
	updatedVersions := make(map[string]string, len(module.Requires))
	for _, requirement := range module.Requires {
		version, err := latestCompatibleV1Tag(ctx.Context, requirement.Module, requirement.Version)
		if err != nil {
			return fmt.Errorf("latestCompatibleV1Tag: %w", err)
		}
		updatedVersions[requirement.Module] = version
	}
	updated, err := rewriteV1RequiredVersions(original, updatedVersions)
	if err != nil {
		return fmt.Errorf("rewriteV1RequiredVersions: %w", err)
	}
	updatedModule, err := v1.ParseModule(bytes.NewReader(updated))
	if err != nil {
		return fmt.Errorf("ParseModule: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, updatedModule, v1.Lock{})
	if err != nil {
		return fmt.Errorf("resolveV1LockWithPins: %w", err)
	}
	cacheRoot, err := gitCachePath(getLogger(ctx))
	if err != nil {
		return fmt.Errorf("gitCachePath: %w", err)
	}
	updated, err = augmentV1ManifestRequirements(updated, root, updatedModule, lock, cacheRoot)
	if err != nil {
		return fmt.Errorf("augmentV1ManifestRequirements: %w", err)
	}
	return writeV1ResolvedFiles(root, original, updated, lock)
}

func latestCompatibleV1Tag(ctx context.Context, source, current string) (string, error) {
	if current == "" {
		// Versionless requirements resolve the current Git HEAD.
		return "", nil
	}
	if v1.IsCommitRef(current) {
		return current, nil
	}
	if !semver.IsValid(current) {
		return "", fmt.Errorf("%s: invalid required version %q", source, current)
	}
	versions, err := listV1ModuleTags(ctx, source)
	if err != nil {
		return "", fmt.Errorf("listV1ModuleTags: %w", err)
	}
	best := current
	for _, version := range versions {
		if semver.Major(version) != semver.Major(current) {
			continue
		}
		if semver.Prerelease(current) == "" && semver.Prerelease(version) != "" {
			continue
		}
		if semver.Compare(version, best) > 0 {
			best = version
		}
	}
	return best, nil
}

func rewriteV1RequiredVersions(original []byte, updates map[string]string) ([]byte, error) {
	return editV1RequirementLines(original, func(line v1RequirementLine) (string, error) {
		version, ok := updates[line.module]
		if !ok || line.version == "" {
			return line.String(), nil
		}
		return line.withVersion(version).String(), nil
	})
}

func writeV1Manifest(root string, raw []byte) error {
	if err := writeAtomicFile(filepath.Join(root, v1.ModuleFile), raw, 0o600); err != nil {
		return fmt.Errorf("writeAtomicFile: %w", err)
	}
	return nil
}
