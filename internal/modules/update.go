package modules

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

// Update refreshes tagged requirements within their major versions and
// versionless requirements to Git HEAD; explicit commit pins remain fixed.
func Update(ctx context.Context, root string, repository VersionedRepository) error {
	original, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	if len(module.Replaces) > 0 {
		return fmt.Errorf("module %s: remove local replacements before updating a reproducible lock", module.Name)
	}
	updatedVersions := make(map[string]string, len(module.Requires))
	for _, requirement := range module.Requires {
		version, err := latestCompatibleV1Tag(ctx, requirement.Module, requirement.Version, repository)
		if err != nil {
			return fmt.Errorf("latestCompatibleV1Tag: %w", err)
		}
		updatedVersions[requirement.Module] = version
	}
	updated := rewriteV1RequiredVersions(original, updatedVersions)
	updatedModule, err := v1.ParseModule(bytes.NewReader(updated))
	if err != nil {
		return fmt.Errorf("ParseModule: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, updatedModule, v1.Lock{}, repository)
	if err != nil {
		return fmt.Errorf("resolveV1LockWithPins: %w", err)
	}
	updated, err = augmentV1ManifestRequirements(updated, root, updatedModule, lock, repository)
	if err != nil {
		return fmt.Errorf("augmentV1ManifestRequirements: %w", err)
	}
	return writeV1ResolvedFiles(root, original, updated, lock)
}

func latestCompatibleV1Tag(ctx context.Context, source, current string, repository VersionedRepository) (string, error) {
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
	versions, err := repository.Versions(ctx, source)
	if err != nil {
		return "", fmt.Errorf("Versions: %w", err)
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

func rewriteV1RequiredVersions(original []byte, updates map[string]string) []byte {
	lines := strings.Split(string(original), "\n")
	for _, line := range parseV1RequirementLines(lines) {
		version, ok := updates[line.module]
		if !ok || line.version == "" {
			continue
		}
		lines[line.index] = line.withVersion(version).String()
	}
	return []byte(strings.Join(lines, "\n"))
}

func writeV1Manifest(root string, raw []byte) error {
	if err := disk.WriteAtomicFile(filepath.Join(root, v1.ModuleFile), raw, 0o600); err != nil {
		return fmt.Errorf("WriteAtomicFile: %w", err)
	}
	return nil
}
