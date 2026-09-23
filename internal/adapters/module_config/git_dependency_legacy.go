package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func readLegacyEasyPRootsAndRequires(dir string) ([]string, []v1.Requirement, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "easyp.yaml"))
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var old struct {
		Generate struct {
			Inputs []struct {
				Directory config.InputFilesDir `yaml:"directory"`
				GitRepo   struct {
					URL string `yaml:"url"`
				} `yaml:"git_repo"`
			} `yaml:"inputs"`
		} `yaml:"generate"`
	}
	if err := yaml.Unmarshal(raw, &old); err != nil {
		return nil, nil, fmt.Errorf("legacy easyp.yaml: %w", err)
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
	if at := strings.LastIndex(raw, "@v"); at >= 0 {
		source, version = raw[:at], raw[at+1:]
		if !semver.IsValid(version) {
			return v1.Requirement{}, fmt.Errorf("legacy dependency %q needs a valid semver tag", raw)
		}
	}
	if at := strings.LastIndex(source, "@"); at > strings.LastIndex(source, "/") {
		return v1.Requirement{}, fmt.Errorf("legacy dependency %q uses a non-SemVer Git ref; use a SemVer tag for v1 resolution", raw)
	}
	if source == "" {
		return v1.Requirement{}, fmt.Errorf("legacy dependency %q has no Git source", raw)
	}
	return v1.Requirement{Module: source, Version: version}, nil
}
