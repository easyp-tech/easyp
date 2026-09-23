package moduleconfig

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const dependencyManifestFile = "protobuf.mod"

type gitDependencyMode string

const (
	gitDependencyManifest     gitDependencyMode = dependencyManifestFile
	gitDependencyBufWorkspace gitDependencyMode = bufWorkConfigFile
	gitDependencyBufModule    gitDependencyMode = bufModuleConfigFile
	gitDependencyLegacyEasyP  gitDependencyMode = legacyEasyPConfigFile
)

// ReadGitDependency adapts metadata in a checked-out Git repository to
// the roots and requirements needed by the v1 resolver. Its identity is the
// source named by the requiring module for pre-v1 repositories.
func ReadGitDependency(dir, source string) (v1.Module, error) {
	modes, err := detectGitDependencyModes(dir)
	if err != nil {
		return v1.Module{}, fmt.Errorf("detectGitDependencyModes: %w", err)
	}
	module := v1.Module{Name: source}
	var bufRoots, legacyRoots []string
	foundBuf := false
	for _, mode := range modes {
		path := filepath.Join(dir, string(mode))
		switch mode {
		case gitDependencyManifest:
			parsed, isV1, err := readGitDependencyManifest(path, source)
			if err != nil {
				return v1.Module{}, fmt.Errorf("readGitDependencyManifest: %w", err)
			}
			if isV1 {
				return parsed, nil
			}
			module.Requires = append(module.Requires, parsed.Requires...)
		case gitDependencyBufWorkspace:
			bufRoots, err = readBufDependencyWorkspace(path)
			if err != nil {
				return v1.Module{}, fmt.Errorf("readBufDependencyWorkspace: %w", err)
			}
			foundBuf = true
		case gitDependencyBufModule:
			if foundBuf {
				continue
			}
			bufRoots, err = readBufDependencyModule(path)
			if err != nil {
				return v1.Module{}, fmt.Errorf("readBufDependencyModule: %w", err)
			}
			foundBuf = true
		case gitDependencyLegacyEasyP:
			roots, requires, err := readLegacyEasyPRootsAndRequires(path)
			if err != nil {
				return v1.Module{}, fmt.Errorf("readLegacyEasyPRootsAndRequires: %w", err)
			}
			legacyRoots = roots
			module.Requires = append(module.Requires, requires...)
		default:
			return v1.Module{}, fmt.Errorf("unsupported dependency config mode %q", mode)
		}
	}
	if foundBuf {
		module.Roots = bufRoots
	} else {
		module.Roots = legacyRoots
	}
	if len(module.Roots) == 0 {
		module.Roots = []string{"."}
	}
	for _, root := range module.Roots {
		if !filepath.IsLocal(root) && root != "." {
			return v1.Module{}, fmt.Errorf("dependency %s root %q leaves the repository", source, root)
		}
	}
	return module, nil
}

// detectGitDependencyModes only looks for root-level config files. Parsing and
// format-specific precedence belong to ReadGitDependency.
func detectGitDependencyModes(dir string) ([]gitDependencyMode, error) {
	var modes []gitDependencyMode
	for _, mode := range []gitDependencyMode{
		gitDependencyManifest,
		gitDependencyBufWorkspace,
		gitDependencyBufModule,
		gitDependencyLegacyEasyP,
	} {
		_, err := os.Lstat(filepath.Join(dir, string(mode)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("Lstat: %w", err)
		}
		modes = append(modes, mode)
	}
	return modes, nil
}

func readGitDependencyManifest(path, source string) (v1.Module, bool, error) {
	module := v1.Module{Name: source}
	manifest, err := os.ReadFile(path)
	if err != nil {
		return v1.Module{}, false, fmt.Errorf("ReadFile: %w", err)
	}
	if v1.IsModuleManifest(manifest) {
		parsed, err := v1.ParseModule(bytes.NewReader(manifest))
		if err != nil {
			return v1.Module{}, false, fmt.Errorf("ParseModule: %w", err)
		}
		if parsed.Name != source {
			return v1.Module{}, false, fmt.Errorf("%s declares module %s, want %s", path, parsed.Name, source)
		}
		return parsed, true, nil
	}
	legacy, err := modfile.Parse(manifest)
	if err != nil {
		return v1.Module{}, false, fmt.Errorf("Parse: %s: %w", path, err)
	}
	for _, raw := range legacy.Direct {
		requirement, err := parseLegacyV1Requirement(raw)
		if err != nil {
			return v1.Module{}, false, fmt.Errorf("parseLegacyV1Requirement: %w", err)
		}
		module.Requires = append(module.Requires, requirement)
	}
	return module, false, nil
}
