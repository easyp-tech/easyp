package moduleconfig

import (
	"fmt"

	"gopkg.in/yaml.v3"
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

func readBufDependencyRoots(dir string) ([]string, bool, error) {
	raw, found, err := readOptionalDependencyConfig(dir, bufV1ConfigFile)
	if err != nil {
		return nil, false, fmt.Errorf("readOptionalDependencyConfig: %w", err)
	}
	if found {
		var workspace bufDependencyWorkspaceConfig
		err = yaml.Unmarshal(raw, &workspace)
		if err != nil {
			return nil, true, fmt.Errorf("Unmarshal: %s: %w", bufV1ConfigFile, err)
		}
		if workspace.Version != "v1" || len(workspace.Directories) == 0 {
			return nil, true, fmt.Errorf("%s: expected v1 with nonempty directories", bufV1ConfigFile)
		}
		return workspace.Directories, true, nil
	}
	raw, found, err = readOptionalDependencyConfig(dir, bufV2ConfigFile)
	if err != nil {
		return nil, false, fmt.Errorf("readOptionalDependencyConfig: %w", err)
	}
	if !found {
		return nil, false, nil
	}
	var buf bufDependencyConfig
	err = yaml.Unmarshal(raw, &buf)
	if err != nil {
		return nil, true, fmt.Errorf("Unmarshal: %s: %w", bufV2ConfigFile, err)
	}
	switch buf.Version {
	case "v2":
		if len(buf.Modules) == 0 {
			return nil, true, fmt.Errorf("%s v2: missing modules", bufV2ConfigFile)
		}
		roots := make([]string, 0, len(buf.Modules))
		for _, item := range buf.Modules {
			if item.Path == "" {
				return nil, true, fmt.Errorf("%s v2: module path is empty", bufV2ConfigFile)
			}
			if len(item.Includes) > 0 || len(item.Excludes) > 0 {
				return nil, true, fmt.Errorf("%s v2: modules.includes/excludes are not supported as Git dependency import roots", bufV2ConfigFile)
			}
			roots = append(roots, item.Path)
		}
		return roots, true, nil
	case "v1":
		return []string{"."}, true, nil
	case "v1beta1":
		if len(buf.Build.Excludes) > 0 {
			return nil, true, fmt.Errorf("%s v1beta1: build.excludes is not supported as Git dependency import roots", bufV2ConfigFile)
		}
		if len(buf.Build.Roots) > 0 {
			return buf.Build.Roots, true, nil
		}
		return []string{"."}, true, nil
	default:
		return nil, true, fmt.Errorf("%s: unsupported version %q", bufV2ConfigFile, buf.Version)
	}
}
