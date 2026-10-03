package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

// Update refreshes tagged requirements within their major versions and
// versionless requirements to Git HEAD; explicit commit pins remain fixed.
// Local replacements skip remote version lookup and suppress lock/indirect edits.
func Update(ctx context.Context, root string, repository VersionedRepository) error {
	original, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	module = directV1Module(original, module)
	updatedVersions := make(map[string]string, len(module.Requires))
	replaced := replacedModules(module)
	for _, requirement := range module.Requires {
		if replaced[requirement.Module] {
			continue
		}
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
	if len(updatedModule.Replaces) > 0 {
		if err := validateLocalOverlay(ctx, root, updatedModule, repository, true); err != nil {
			return err
		}
		if bytes.Equal(original, updated) {
			return nil
		}
		return writeV1Manifest(root, updated)
	}
	updatedModule = directV1Module(updated, updatedModule)
	existing, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("ReadLock: %w", err)
	}
	lock, err := resolveV1Lock(ctx, root, updatedModule, existing, repository, false)
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
