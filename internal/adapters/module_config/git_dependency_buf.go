package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	bufWorkConfigFile   = "buf.work.yaml"
	bufModuleConfigFile = "buf.yaml"
)

type bufDependencyWorkspaceConfig struct {
	Version     string   `yaml:"version"`
	Directories []string `yaml:"directories"`
}

type bufDependencyModule struct {
	Path     string   `yaml:"path"`
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes"`
}

type bufDependencyBuild struct {
	Roots    []string `yaml:"roots"`
	Excludes []string `yaml:"excludes"`
}

type bufDependencyConfig struct {
	Version string                `yaml:"version"`
	Modules []bufDependencyModule `yaml:"modules"`
	Build   bufDependencyBuild    `yaml:"build"`
	Deps    []string              `yaml:"deps"`
}

// BufRegistryDependency preserves one BSR reference together with the config
// that declared it. Mapping the reference to a Git module is a separate policy.
type BufRegistryDependency struct {
	Config    string
	Reference string
}

// UnsupportedBufRegistryDependenciesError reports BSR metadata that EasyP can
// read but cannot safely translate into Git requirements yet.
type UnsupportedBufRegistryDependenciesError struct {
	Dependencies []BufRegistryDependency
}

func (e UnsupportedBufRegistryDependenciesError) Error() string {
	parts := make([]string, 0, len(e.Dependencies))
	for _, dependency := range e.Dependencies {
		parts = append(parts, fmt.Sprintf("%s (from %s)", dependency.Reference, dependency.Config))
	}
	return fmt.Sprintf(
		"buf registry dependencies %s require BSR-to-Git mapping; automatic BSR-to-Git mapping is not implemented; provide Git-native dependency metadata in protobuf.mod before using this module",
		strings.Join(parts, ", "),
	)
}

type bufDependencyMetadata struct {
	Roots                []string
	RegistryDependencies []BufRegistryDependency
}

func readBufDependencyWorkspace(path string) (bufDependencyMetadata, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("ReadFile: %s: %w", path, err)
	}
	var workspace bufDependencyWorkspaceConfig
	err = yaml.Unmarshal(raw, &workspace)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	if workspace.Version != "v1" || len(workspace.Directories) == 0 {
		return bufDependencyMetadata{}, fmt.Errorf("%s: expected v1 with nonempty directories", path)
	}
	metadata := bufDependencyMetadata{Roots: append([]string(nil), workspace.Directories...)}
	base := filepath.Dir(path)
	for _, directory := range workspace.Directories {
		if !filepath.IsLocal(directory) && directory != "." {
			return bufDependencyMetadata{}, fmt.Errorf("%s: directory %q leaves the repository", path, directory)
		}
		modulePath := filepath.Join(base, directory, bufModuleConfigFile)
		_, err := os.Stat(modulePath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return bufDependencyMetadata{}, fmt.Errorf("Stat: %s: %w", modulePath, err)
		}
		module, err := readBufDependencyModule(modulePath)
		if err != nil {
			return bufDependencyMetadata{}, fmt.Errorf("readBufDependencyModule: %w", err)
		}
		metadata.RegistryDependencies = append(metadata.RegistryDependencies, module.RegistryDependencies...)
	}
	return metadata, nil
}

func readBufDependencyModule(path string) (bufDependencyMetadata, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("ReadFile: %s: %w", path, err)
	}
	var buf bufDependencyConfig
	err = yaml.Unmarshal(raw, &buf)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	metadata := bufDependencyMetadata{}
	for _, dependency := range buf.Deps {
		dependency = strings.TrimSpace(dependency)
		if dependency == "" {
			return bufDependencyMetadata{}, fmt.Errorf("%s: deps entry is empty", path)
		}
		metadata.RegistryDependencies = append(metadata.RegistryDependencies, BufRegistryDependency{
			Config:    path,
			Reference: dependency,
		})
	}
	switch buf.Version {
	case "v2":
		if len(buf.Modules) == 0 {
			return bufDependencyMetadata{}, fmt.Errorf("%s v2: missing modules", path)
		}
		metadata.Roots = make([]string, 0, len(buf.Modules))
		for _, item := range buf.Modules {
			if item.Path == "" {
				return bufDependencyMetadata{}, fmt.Errorf("%s v2: module path is empty", path)
			}
			if len(item.Includes) > 0 || len(item.Excludes) > 0 {
				return bufDependencyMetadata{}, fmt.Errorf("%s v2: modules.includes/excludes are not supported as Git dependency import roots", path)
			}
			metadata.Roots = append(metadata.Roots, item.Path)
		}
	case "v1":
		metadata.Roots = []string{"."}
	case "v1beta1":
		if len(buf.Build.Excludes) > 0 {
			return bufDependencyMetadata{}, fmt.Errorf("%s v1beta1: build.excludes is not supported as Git dependency import roots", path)
		}
		if len(buf.Build.Roots) > 0 {
			metadata.Roots = append([]string(nil), buf.Build.Roots...)
		} else {
			metadata.Roots = []string{"."}
		}
	default:
		return bufDependencyMetadata{}, fmt.Errorf("%s: unsupported version %q", path, buf.Version)
	}
	return metadata, nil
}
