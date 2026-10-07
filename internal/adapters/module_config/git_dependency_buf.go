package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
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

type bufDependencyMetadata struct {
	Roots           []string
	ProtoFilters    []v1.ProtoFileFilter
	BSRDependencies []v1.BSRDependency
}

func readBufDependencyWorkspace(path string) (bufDependencyMetadata, error) {
	raw, err := readGitDependencyConfig(path)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("readGitDependencyConfig: %w", err)
	}
	if err := validateGitDependencyRootYAML(raw, bufWorkConfigFile); err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("validateGitDependencyRootYAML: %s: %w", path, err)
	}
	var workspace bufDependencyWorkspaceConfig
	err = yaml.Unmarshal(raw, &workspace)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	if workspace.Version != "v1" || len(workspace.Directories) == 0 {
		return bufDependencyMetadata{}, fmt.Errorf("%s: expected v1 with nonempty directories", path)
	}
	var metadata bufDependencyMetadata
	base := filepath.Dir(path)
	for _, directory := range workspace.Directories {
		if !filepath.IsLocal(directory) && directory != "." {
			return bufDependencyMetadata{}, fmt.Errorf("%s: directory %q leaves the repository", path, directory)
		}
		if err := validateGitDependencyDirectory(base, directory); err != nil {
			return bufDependencyMetadata{}, fmt.Errorf("validateGitDependencyDirectory: %w", err)
		}
		modulePath := filepath.Join(base, directory, bufModuleConfigFile)
		_, err := os.Lstat(modulePath)
		if os.IsNotExist(err) {
			metadata.Roots = append(metadata.Roots, filepath.Clean(directory))
			metadata.ProtoFilters = append(metadata.ProtoFilters, v1.ProtoFileFilter{Root: filepath.Clean(directory)})
			continue
		}
		if err != nil {
			return bufDependencyMetadata{}, fmt.Errorf("Lstat: %s: %w", modulePath, err)
		}
		module, err := readBufDependencyModule(modulePath)
		if err != nil {
			return bufDependencyMetadata{}, fmt.Errorf("readBufDependencyModule: %w", err)
		}
		for _, root := range module.Roots {
			metadata.Roots = append(metadata.Roots, filepath.Join(directory, root))
		}
		for _, filter := range module.ProtoFilters {
			filter.Root = filepath.Join(directory, filter.Root)
			for i, include := range filter.Includes {
				filter.Includes[i] = filepath.Join(directory, include)
			}
			for i, exclude := range filter.Excludes {
				filter.Excludes[i] = filepath.Join(directory, exclude)
			}
			metadata.ProtoFilters = append(metadata.ProtoFilters, filter)
		}
		for _, dependency := range module.BSRDependencies {
			dependency.Config = filepath.ToSlash(filepath.Join(directory, dependency.Config))
			if !slices.Contains(metadata.BSRDependencies, dependency) {
				metadata.BSRDependencies = append(metadata.BSRDependencies, dependency)
			}
		}
	}
	return metadata, nil
}

func readBufDependencyModule(path string) (bufDependencyMetadata, error) {
	raw, err := readGitDependencyConfig(path)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("readGitDependencyConfig: %w", err)
	}
	if err := validateGitDependencyRootYAML(raw, bufModuleConfigFile); err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("validateGitDependencyRootYAML: %s: %w", path, err)
	}
	var buf bufDependencyConfig
	err = yaml.Unmarshal(raw, &buf)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	var metadata bufDependencyMetadata
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
			root := filepath.Clean(item.Path)
			if !slices.Contains(metadata.Roots, root) {
				metadata.Roots = append(metadata.Roots, root)
			}
			metadata.ProtoFilters = append(metadata.ProtoFilters, v1.ProtoFileFilter{
				Root: root, Includes: item.Includes, Excludes: item.Excludes,
			})
		}
	case "v1":
		metadata.Roots = []string{"."}
		metadata.ProtoFilters = []v1.ProtoFileFilter{{Root: ".", Excludes: buf.Build.Excludes}}
	case "v1beta1":
		if len(buf.Build.Roots) > 0 {
			metadata.Roots = append([]string(nil), buf.Build.Roots...)
		} else {
			metadata.Roots = []string{"."}
		}
		for i, root := range metadata.Roots {
			if !filepath.IsLocal(root) {
				return bufDependencyMetadata{}, fmt.Errorf("%s: root %q leaves the repository", path, root)
			}
			metadata.Roots[i] = filepath.Clean(root)
			var excludes []string
			for _, exclude := range buf.Build.Excludes {
				if bufPathWithin(exclude, root) {
					excludes = append(excludes, exclude)
				}
			}
			metadata.ProtoFilters = append(metadata.ProtoFilters, v1.ProtoFileFilter{Root: filepath.Clean(root), Excludes: excludes})
		}
		for _, exclude := range buf.Build.Excludes {
			if !slices.ContainsFunc(metadata.Roots, func(root string) bool { return bufPathWithin(exclude, root) }) {
				return bufDependencyMetadata{}, fmt.Errorf("%s: exclude %q is outside build.roots", path, exclude)
			}
		}
	default:
		return bufDependencyMetadata{}, fmt.Errorf("%s: unsupported version %q", path, buf.Version)
	}
	for i := range metadata.ProtoFilters {
		if err := validateBufProtoFilter(&metadata.ProtoFilters[i]); err != nil {
			return bufDependencyMetadata{}, fmt.Errorf("validateBufProtoFilter: %s: %w", path, err)
		}
	}
	metadata.BSRDependencies, err = readBufBSRDependencies(path, buf.Deps)
	if err != nil {
		return bufDependencyMetadata{}, fmt.Errorf("readBufBSRDependencies: %w", err)
	}
	return metadata, nil
}

func validateBufProtoFilter(filter *v1.ProtoFileFilter) error {
	if !filepath.IsLocal(filter.Root) {
		return fmt.Errorf("root %q leaves the repository", filter.Root)
	}
	for i, include := range filter.Includes {
		if !bufPathWithin(include, filter.Root) {
			return fmt.Errorf("include %q is outside module root %q", include, filter.Root)
		}
		filter.Includes[i] = filepath.Clean(include)
	}
	for i, exclude := range filter.Excludes {
		if !bufPathWithin(exclude, filter.Root) {
			return fmt.Errorf("exclude %q is outside module root %q", exclude, filter.Root)
		}
		if len(filter.Includes) > 0 && !slices.ContainsFunc(filter.Includes, func(include string) bool { return bufPathWithin(exclude, include) }) {
			return fmt.Errorf("exclude %q is outside module includes", exclude)
		}
		filter.Excludes[i] = filepath.Clean(exclude)
	}
	return nil
}

func bufPathWithin(path, directory string) bool {
	if !filepath.IsLocal(path) {
		return false
	}
	path, directory = filepath.Clean(path), filepath.Clean(directory)
	return directory == "." || path == directory || strings.HasPrefix(path, directory+string(filepath.Separator))
}
