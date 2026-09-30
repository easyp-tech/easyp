package api

import (
	"fmt"
	"os"

	"github.com/easyp-tech/easyp/internal/workspace"
)

func moduleWorkingDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Getwd: %w", err)
	}
	root, err := workspace.Module(cwd)
	if err != nil {
		return "", fmt.Errorf("Module: %w", err)
	}
	return root, nil
}
