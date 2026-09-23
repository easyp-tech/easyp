package api

import (
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
		return err
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
		return err
	}
	updatedModule, err := v1.ParseModule(strings.NewReader(string(updated)))
	if err != nil {
		return fmt.Errorf("updated protobuf.mod: %w", err)
	}
	existing, err := readV1Lock(filepath.Join(root, "protobuf.lock"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing protobuf.lock: %w", err)
	}
	lock, err := resolveV1LockWithPins(ctx, root, updatedModule, existing)
	if err != nil {
		return err
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
	lines := strings.Split(string(original), "\n")
	inRequire := false
	found := false
	for i, line := range lines {
		body, comment, hasComment := splitV1ManifestComment(line)
		trimmed := strings.TrimSpace(body)
		if strings.HasSuffix(trimmed, "(") && strings.TrimSpace(strings.TrimSuffix(trimmed, "(")) == "require" {
			inRequire = true
			continue
		}
		if trimmed == ")" && inRequire {
			inRequire = false
			continue
		}
		fields := strings.Fields(body)
		entry := inRequire && len(fields) >= 1 && len(fields) <= 2 && fields[0] == target.Module
		directive := !inRequire && len(fields) >= 2 && len(fields) <= 3 && fields[0] == "require" && fields[1] == target.Module
		if !entry && !directive {
			continue
		}
		if found {
			return nil, fmt.Errorf("protobuf.mod: duplicate require %s", target.Module)
		}
		found = true
		version := target.Version
		if version == "" {
			version = fields[len(fields)-1]
			if version == target.Module {
				version = ""
			}
		}
		indent := body[:len(body)-len(strings.TrimLeft(body, " \t"))]
		statement := target.Module
		if directive {
			statement = "require " + statement
		}
		if version != "" {
			statement += " " + version
		}
		if hasComment {
			words := strings.Fields(comment)
			if len(words) > 0 && words[0] == "indirect" {
				comment = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comment), "indirect"))
			}
			if strings.TrimSpace(comment) != "" {
				statement += " // " + strings.TrimSpace(comment)
			}
		}
		lines[i] = indent + statement
	}
	if found {
		return []byte(strings.Join(lines, "\n")), nil
	}
	updated := append([]byte(nil), original...)
	if len(updated) > 0 && updated[len(updated)-1] != '\n' {
		updated = append(updated, '\n')
	}
	line := "require " + target.Module
	if target.Version != "" {
		line += " " + target.Version
	}
	return append(updated, []byte(line+"\n")...), nil
}

func appendV1IndirectRequirements(original []byte, module v1.Module, lock v1.Lock) []byte {
	existing := make(map[string]bool, len(module.Requires))
	for _, requirement := range module.Requires {
		existing[requirement.Module] = true
	}
	updated := append([]byte(nil), original...)
	for _, entry := range lock.Modules {
		if existing[entry.Source] {
			continue
		}
		if len(updated) > 0 && updated[len(updated)-1] != '\n' {
			updated = append(updated, '\n')
		}
		updated = append(updated, []byte(fmt.Sprintf("require %s %s // indirect\n", entry.Source, entry.Version))...)
		existing[entry.Source] = true
	}
	return updated
}
