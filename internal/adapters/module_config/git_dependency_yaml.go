package moduleconfig

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Check root declarations before YAML's permissive string decoding can coerce
// scalars or erase null values and accidentally turn malformed roots into fallback.
func validateGitDependencyRootYAML(raw []byte, name string) error {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("Decode: %w", err)
	}
	var extra yaml.Node
	err := decoder.Decode(&extra)
	if !errors.Is(err, io.EOF) {
		if err != nil {
			return fmt.Errorf("Decode: %w", err)
		}
		return fmt.Errorf("dependency config must contain one YAML document")
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("dependency config must contain a mapping")
	}
	root := document.Content[0]
	switch name {
	case bufWorkConfigFile:
		return validateDependencyPathList(dependencyYAMLField(root, "directories"))
	case bufModuleConfigFile:
		modules := dependencyYAMLField(root, "modules")
		if modules != nil {
			if modules.Kind != yaml.SequenceNode {
				return fmt.Errorf("modules must be a list")
			}
			for _, module := range modules.Content {
				module = dependencyYAMLValue(module)
				if module.Kind != yaml.MappingNode {
					return fmt.Errorf("module must contain a mapping")
				}
				if err := validateDependencyDirectory(dependencyYAMLField(module, "path")); err != nil {
					return fmt.Errorf("validateDependencyDirectory: %w", err)
				}
				for _, field := range []string{"includes", "excludes"} {
					if err := validateDependencyPathList(dependencyYAMLField(module, field)); err != nil {
						return fmt.Errorf("validateDependencyPathList: %w", err)
					}
				}
			}
		}
		build := dependencyYAMLField(root, "build")
		if build != nil {
			if build.Kind != yaml.MappingNode {
				return fmt.Errorf("build must contain a mapping")
			}
			for _, field := range []string{"roots", "excludes"} {
				if err := validateDependencyPathList(dependencyYAMLField(build, field)); err != nil {
					return fmt.Errorf("validateDependencyPathList: %w", err)
				}
			}
		}
	case legacyEasyPConfigFile:
		generate := dependencyYAMLField(root, "generate")
		if generate == nil {
			return nil
		}
		if generate.Kind != yaml.MappingNode {
			return fmt.Errorf("generate must contain a mapping")
		}
		inputs := dependencyYAMLField(generate, "inputs")
		if inputs == nil {
			return nil
		}
		if inputs.Kind != yaml.SequenceNode {
			return fmt.Errorf("inputs must be a list")
		}
		for _, input := range inputs.Content {
			input = dependencyYAMLValue(input)
			if input.Kind != yaml.MappingNode {
				return fmt.Errorf("input must contain a mapping")
			}
			directory := dependencyYAMLField(input, "directory")
			if directory == nil {
				continue
			}
			if directory.Kind != yaml.MappingNode {
				if err := validateDependencyDirectory(directory); err != nil {
					return fmt.Errorf("validateDependencyDirectory: %w", err)
				}
				continue
			}
			for _, field := range []string{"path", "root"} {
				if err := validateDependencyDirectory(dependencyYAMLField(directory, field)); err != nil {
					return fmt.Errorf("validateDependencyDirectory: %w", err)
				}
			}
		}
	}
	return nil
}

func dependencyYAMLValue(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node
}

func dependencyYAMLField(mapping *yaml.Node, field string) *yaml.Node {
	return dependencyYAMLFieldAt(mapping, field, make(map[*yaml.Node]bool), make(map[*yaml.Node]*yaml.Node))
}

func dependencyYAMLFieldAt(mapping *yaml.Node, field string, active map[*yaml.Node]bool, memo map[*yaml.Node]*yaml.Node) (result *yaml.Node) {
	mapping = dependencyYAMLValue(mapping)
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	if value, known := memo[mapping]; known {
		return value
	}
	if active[mapping] {
		return nil
	}
	active[mapping] = true
	// Each search has one fixed field. Cache absent results as well so repeated
	// merge aliases cannot expand an acyclic mapping graph exponentially.
	defer func() {
		delete(active, mapping)
		memo[mapping] = result
	}()
	var merges []*yaml.Node
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == field {
			return dependencyYAMLValue(mapping.Content[index+1])
		}
		if mapping.Content[index].Tag == "!!merge" {
			value := dependencyYAMLValue(mapping.Content[index+1])
			if value.Kind == yaml.SequenceNode {
				merges = append(merges, value.Content...)
			} else {
				merges = append(merges, value)
			}
		}
	}
	for _, merge := range merges {
		if value := dependencyYAMLFieldAt(merge, field, active, memo); value != nil {
			return value
		}
	}
	return nil
}

func validateDependencyDirectory(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return fmt.Errorf("directory path must be a string")
	}
	return nil
}

func validateDependencyPathList(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("directories must be a list of strings")
	}
	for _, directory := range node.Content {
		if err := validateDependencyDirectory(dependencyYAMLValue(directory)); err != nil {
			return fmt.Errorf("validateDependencyDirectory: %w", err)
		}
	}
	return nil
}
