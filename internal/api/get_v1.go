package api

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Get adds a direct module requirement and resolves its dependency graph.
type Get struct{}

var _ Handler = Get{}

func (g Get) Command() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "add a Git module and its transitive dependencies",
		ArgsUsage: "<module>[@version|@commit]",
		Action:    g.Action,
	}
}

func (g Get) Action(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return errors.New("get expects one module: easyp get <module>[@version|@commit]")
	}
	requirement, err := parseV1GetRequirement(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("parseV1GetRequirement: %w", err)
	}
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	original, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
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
	existing, err := readV1Lock(filepath.Join(root, v1.LockFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("readV1Lock: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, updatedModule, existing)
	if err != nil {
		return fmt.Errorf("resolveV1LockWithPins: %w", err)
	}
	updated = appendV1IndirectRequirements(updated, updatedModule, lock)
	return writeV1ResolvedFiles(root, original, updated, lock)
}

func parseV1GetRequirement(spec string) (v1.Requirement, error) {
	version := ""
	source := spec
	if at := strings.LastIndex(spec, "@"); at > strings.LastIndex(spec, "/") {
		source, version = spec[:at], spec[at+1:]
		if !semver.IsValid(version) && !v1.IsCommitRef(version) {
			return v1.Requirement{}, fmt.Errorf("get %q: expected a semantic version or full Git commit after @", spec)
		}
	}
	if strings.ContainsAny(source, " \t\r\n") {
		return v1.Requirement{}, fmt.Errorf("get %q: invalid module identity", spec)
	}
	if _, err := v1GitModuleCandidates(source); err != nil {
		return v1.Requirement{}, fmt.Errorf("get %q: %w", spec, err)
	}
	return v1.Requirement{Module: source, Version: version}, nil
}

// addDirectV1Requirement changes one requirement while preserving the rest of
// the manifest. A repeated get leaves an existing version in place unless the
// user explicitly requests another one.
func addDirectV1Requirement(original []byte, target v1.Requirement) ([]byte, error) {
	found := false
	updated, err := editV1RequirementLines(original, func(line v1RequirementLine) (string, error) {
		if line.module != target.Module {
			return line.String(), nil
		}
		if found {
			return "", fmt.Errorf("protobuf.mod: duplicate require %s", target.Module)
		}
		found = true
		if target.Version != "" {
			line = line.withVersion(target.Version)
		}
		return line.direct().String(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("editV1RequirementLines: %w", err)
	}
	if found {
		return updated, nil
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
