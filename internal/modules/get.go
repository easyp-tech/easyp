package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Get adds or promotes one direct requirement and records its resolved transitive requirements.
func Get(ctx context.Context, root string, requirement v1.Requirement, repository Repository) error {
	original, module, err := ReadManifest(root)
	if err != nil {
		return fmt.Errorf("ReadManifest: %w", err)
	}
	if module.Name == requirement.Module {
		return fmt.Errorf("module %s cannot require itself", module.Name)
	}
	updated, err := addDirectV1Requirement(original, requirement)
	if err != nil {
		return fmt.Errorf("addDirectV1Requirement: %w", err)
	}
	updatedModule, err := v1.ParseModule(bytes.NewReader(updated))
	if err != nil {
		return fmt.Errorf("ParseModule: %w", err)
	}
	existing, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("ReadLock: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, updatedModule, existing, repository)
	if err != nil {
		return fmt.Errorf("resolveV1LockWithPins: %w", err)
	}
	updated = appendV1IndirectRequirements(updated, updatedModule, lock)
	return writeV1ResolvedFiles(root, original, updated, lock)
}

// addDirectV1Requirement changes one requirement while preserving the rest of
// the manifest. A repeated get leaves an existing version in place unless the
// user explicitly requests another one.
func addDirectV1Requirement(original []byte, target v1.Requirement) ([]byte, error) {
	lines := strings.Split(string(original), "\n")
	found := false
	for _, line := range parseV1RequirementLines(lines) {
		if line.module != target.Module {
			continue
		}
		if found {
			return nil, fmt.Errorf("protobuf.mod: duplicate require %s", target.Module)
		}
		found = true
		if target.Version != "" {
			line = line.withVersion(target.Version)
		}
		lines[line.index] = line.direct().String()
	}
	if found {
		return []byte(strings.Join(lines, "\n")), nil
	}
	return appendV1Requirements(original, []v1ManifestRequirement{{Requirement: target}}), nil
}

func appendV1IndirectRequirements(original []byte, module v1.Module, lock v1.Lock) []byte {
	existing := make(map[string]bool, len(module.Requires))
	for _, requirement := range module.Requires {
		existing[requirement.Module] = true
	}
	var additions []v1ManifestRequirement
	for _, entry := range lock.Modules {
		if existing[entry.Source] {
			continue
		}
		additions = append(additions, v1ManifestRequirement{Requirement: v1.Requirement{Module: entry.Source, Version: entry.Version}, indirect: true})
		existing[entry.Source] = true
	}
	return appendV1Requirements(original, additions)
}
