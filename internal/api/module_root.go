package api

import (
	"fmt"
	"os"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/workspace"
)

func moduleWorkingDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Getwd: %w", err)
	}
	root, err := workspace.Module(cwd)
	if err != nil {
		policy, lookupErr := workspace.Policy(cwd)
		if lookupErr == nil {
			raw, readErr := workspace.ReadFileAt(policy)
			if readErr == nil && v1.LegacyPolicy(raw) {
				return "", v1.ErrLegacyConfiguration
			}
		}
		return "", fmt.Errorf("Module: %w", err)
	}
	return root, nil
}
