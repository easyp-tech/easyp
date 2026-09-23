package moduleconfig

import (
	"fmt"
	"os"

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
}

func readBufDependencyWorkspace(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %s: %w", path, err)
	}
	var workspace bufDependencyWorkspaceConfig
	err = yaml.Unmarshal(raw, &workspace)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	if workspace.Version != "v1" || len(workspace.Directories) == 0 {
		return nil, fmt.Errorf("%s: expected v1 with nonempty directories", path)
	}
	return workspace.Directories, nil
}

func readBufDependencyModule(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %s: %w", path, err)
	}
	var buf bufDependencyConfig
	err = yaml.Unmarshal(raw, &buf)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	switch buf.Version {
	case "v2":
		if len(buf.Modules) == 0 {
			return nil, fmt.Errorf("%s v2: missing modules", path)
		}
		roots := make([]string, 0, len(buf.Modules))
		for _, item := range buf.Modules {
			if item.Path == "" {
				return nil, fmt.Errorf("%s v2: module path is empty", path)
			}
			if len(item.Includes) > 0 || len(item.Excludes) > 0 {
				return nil, fmt.Errorf("%s v2: modules.includes/excludes are not supported as Git dependency import roots", path)
			}
			roots = append(roots, item.Path)
		}
		return roots, nil
	case "v1":
		return []string{"."}, nil
	case "v1beta1":
		if len(buf.Build.Excludes) > 0 {
			return nil, fmt.Errorf("%s v1beta1: build.excludes is not supported as Git dependency import roots", path)
		}
		if len(buf.Build.Roots) > 0 {
			return buf.Build.Roots, nil
		}
		return []string{"."}, nil
	default:
		return nil, fmt.Errorf("%s: unsupported version %q", path, buf.Version)
	}
}
