package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/workspace"
)

func selectGenerateConfigs(request Request) ([]string, error) {
	projects := slices.Clone(request.Projects)
	if request.Project != "" {
		projects = append(projects, request.Project)
	}
	if request.AllProjects && len(projects) != 0 {
		return nil, fmt.Errorf("--all and --project are mutually exclusive")
	}
	if request.AllProjects {
		return discoverV1GenerateConfigs(request.WorkDir, "")
	}
	if len(projects) > 0 {
		var configs []string
		for _, project := range projects {
			if project == "" {
				return nil, fmt.Errorf("--project must not be empty")
			}
			paths, err := discoverV1GenerateConfigs(request.WorkDir, project)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				if !slices.Contains(configs, path) {
					configs = append(configs, path)
				}
			}
		}
		return configs, nil
	}
	path, err := workspace.FindUp(request.WorkDir, request.WorkspaceRoot, "easyp.gen.yaml")
	if err != nil {
		return nil, err
	}
	if path == "" {
		legacy, lookupErr := workspace.FindUp(request.WorkDir, request.WorkspaceRoot, v1.PolicyFile)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if legacy != "" {
			raw, readErr := os.ReadFile(legacy)
			if readErr != nil {
				return nil, readErr
			}
			if v1.LegacyPolicy(raw) {
				return nil, v1.ErrLegacyConfiguration
			}
		}
		return nil, fmt.Errorf("no selected easyp.gen.yaml; choose --project <directory> (repeatable), or --all for recursive generation")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("automatic generation selection does not follow symlink config %s; select its project explicitly", path)
	}
	return []string{filepath.Clean(path)}, nil
}
