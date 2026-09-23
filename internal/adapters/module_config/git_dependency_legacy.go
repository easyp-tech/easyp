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

type legacyDependencyConfig struct {
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
	for _, input := range old.Generate.Inputs {
		if input.Directory.Path != "" || input.Directory.Root != "" {
			roots = append(roots, input.Directory.Root)
		}
		if input.GitRepo.URL != "" {
			requirement, err := parseLegacyV1Requirement(input.GitRepo.URL)
			if err != nil {
				return nil, nil, err
			}
			requires = append(requires, requirement)
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
