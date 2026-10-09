package generation

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
	"github.com/easyp-tech/easyp/internal/workspace"
)

func selectGenerateConfigs(request Request) ([]string, error) {
	projects := slices.Clone(request.Projects)
	if request.Project != "" {
		projects = append(projects, request.Project)
	}
	if request.GenConfig != "" {
		if request.AllProjects || len(projects) > 0 {
			return nil, fmt.Errorf("--gen-config cannot be combined with --project or --all")
		}
		path := request.GenConfig
		if !filepath.IsAbs(path) {
			path = filepath.Join(request.WorkDir, path)
		}
		relative, err := filepath.Rel(request.WorkspaceRoot, path)
		if err != nil {
			return nil, fmt.Errorf("Rel: %w", err)
		}
		_, err = sourceview.ResolveLocal(context.Background(), request.WorkspaceRoot, relative)
		if err != nil {
			return nil, fmt.Errorf("ResolveLocal: %w", err)
		}
		return []string{filepath.Clean(path)}, nil
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
			raw, readErr := workspace.ReadFile(request.WorkspaceRoot, legacy)
			if readErr != nil {
				return nil, readErr
			}
			if v1.LegacyPolicy(raw) {
				return nil, v1.ErrLegacyConfiguration
			}
		}
		return nil, fmt.Errorf("no selected easyp.gen.yaml; choose --project <directory> (repeatable), or --all for recursive generation")
	}
	return []string{filepath.Clean(path)}, nil
}
