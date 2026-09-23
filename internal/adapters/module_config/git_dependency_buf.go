package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func readBufDependencyRoots(dir string) ([]string, bool, error) {
	if raw, err := os.ReadFile(filepath.Join(dir, "buf.work.yaml")); err == nil {
		var workspace struct {
			Version     string   `yaml:"version"`
			Directories []string `yaml:"directories"`
		}
		if err := yaml.Unmarshal(raw, &workspace); err != nil {
			return nil, true, fmt.Errorf("buf.work.yaml: %w", err)
		}
		if workspace.Version != "v1" || len(workspace.Directories) == 0 {
			return nil, true, fmt.Errorf("buf.work.yaml: expected v1 with nonempty directories")
		}
		return workspace.Directories, true, nil
	} else if !os.IsNotExist(err) {
		return nil, true, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "buf.yaml"))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	var buf struct {
		Version string `yaml:"version"`
		Modules []struct {
			Path     string   `yaml:"path"`
			Includes []string `yaml:"includes"`
			Excludes []string `yaml:"excludes"`
		} `yaml:"modules"`
		Build struct {
			Roots    []string `yaml:"roots"`
			Excludes []string `yaml:"excludes"`
		} `yaml:"build"`
	}
	if err := yaml.Unmarshal(raw, &buf); err != nil {
		return nil, true, fmt.Errorf("buf.yaml: %w", err)
	}
	switch buf.Version {
	case "v2":
		if len(buf.Modules) == 0 {
			return nil, true, fmt.Errorf("buf.yaml v2: missing modules")
		}
		roots := make([]string, 0, len(buf.Modules))
		for _, item := range buf.Modules {
			if item.Path == "" {
				return nil, true, fmt.Errorf("buf.yaml v2: module path is empty")
			}
			if len(item.Includes) > 0 || len(item.Excludes) > 0 {
				return nil, true, fmt.Errorf("buf.yaml v2: modules.includes/excludes are not supported as Git dependency import roots")
			}
			roots = append(roots, item.Path)
		}
		return roots, true, nil
	case "v1":
		return []string{"."}, true, nil
	case "v1beta1":
		if len(buf.Build.Excludes) > 0 {
			return nil, true, fmt.Errorf("buf.yaml v1beta1: build.excludes is not supported as Git dependency import roots")
		}
		if len(buf.Build.Roots) > 0 {
			return buf.Build.Roots, true, nil
		}
		return []string{"."}, true, nil
	default:
		return nil, true, fmt.Errorf("buf.yaml: unsupported version %q", buf.Version)
	}
}
