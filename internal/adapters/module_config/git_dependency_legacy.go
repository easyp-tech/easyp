package moduleconfig

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const legacyEasyPConfigFile = v1.PolicyFile

type legacyDependencyGitRepo struct {
	URL string `yaml:"url"`
}

type legacyDependencyInput struct {
	Directory config.InputFilesDir    `yaml:"directory"`
	GitRepo   legacyDependencyGitRepo `yaml:"git_repo"`
}

type legacyDependencyGenerate struct {
	Inputs []legacyDependencyInput `yaml:"inputs"`
}

// legacyDependencyList rejects null and non-string entries instead of letting
// YAML decoding silently remove them from the dependency graph.
type legacyDependencyList []string

func (list *legacyDependencyList) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("deps must be a list of module strings")
	}
	values := make([]string, 0, len(node.Content))
	for i, entry := range node.Content {
		if entry.Kind == yaml.AliasNode {
			entry = entry.Alias
		}
		if entry == nil || entry.Kind != yaml.ScalarNode || entry.Tag != "!!str" {
			return fmt.Errorf("deps[%d] must be a module string", i)
		}
		values = append(values, entry.Value)
	}
	*list = values
	return nil
}

type legacyDependencyConfig struct {
	Deps     legacyDependencyList     `yaml:"deps"`
	Generate legacyDependencyGenerate `yaml:"generate"`
}

func readLegacyEasyPRootsAndRequires(path string) ([]string, []v1.Requirement, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("ReadFile: %s: %w", path, err)
	}
	var old legacyDependencyConfig
	err = yaml.Unmarshal(raw, &old)
	if err != nil {
		return nil, nil, fmt.Errorf("Unmarshal: %s: %w", path, err)
	}
	var roots []string
	var requires []v1.Requirement
	seen := make(map[v1.Requirement]bool)
	appendRequirement := func(raw string) error {
		requirement, err := parseLegacyV1Requirement(raw)
		if err != nil {
			return err
		}
		if !seen[requirement] {
			requires = append(requires, requirement)
			seen[requirement] = true
		}
		return nil
	}
	for i, raw := range old.Deps {
		if err := appendRequirement(raw); err != nil {
			return nil, nil, fmt.Errorf("parseLegacyV1Requirement: %s deps[%d]: %w", path, i, err)
		}
	}
	for i, input := range old.Generate.Inputs {
		if input.Directory.Path != "" || input.Directory.Root != "" {
			roots = append(roots, input.Directory.Root)
		}
		if input.GitRepo.URL != "" {
			if err := appendRequirement(input.GitRepo.URL); err != nil {
				return nil, nil, fmt.Errorf("parseLegacyV1Requirement: %s generate.inputs[%d].git_repo.url: %w", path, i, err)
			}
		}
	}
	return roots, requires, nil
}

func parseLegacyV1Requirement(raw string) (v1.Requirement, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return v1.Requirement{}, fmt.Errorf("empty legacy dependency")
	}
	source, version := raw, ""
	if at := strings.LastIndex(raw, "@"); at > strings.LastIndex(raw, "/") {
		source, version = raw[:at], raw[at+1:]
		if !semver.IsValid(version) && !v1.IsCommitRef(version) {
			return v1.Requirement{}, fmt.Errorf("legacy dependency %q needs a SemVer tag or full Git commit", raw)
		}
	}
	if source == "" {
		return v1.Requirement{}, fmt.Errorf("legacy dependency %q has no Git source", raw)
	}
	return v1.Requirement{Module: source, Version: version}, nil
}
