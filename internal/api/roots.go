package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/flags"
)

// resolveRoots computes configPath, projectRoot, and the operation root.
func resolveRoots(ctx *cli.Context, rootFlagName string) (string, string, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("Getwd: %w", err)
	}

	root := ctx.String(rootFlagName)
	configPath := ctx.String(flags.Config.Name)
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(workingDir, configPath)
	}
	projectRoot := filepath.Dir(configPath)

	opRoot := projectRoot
	if root != "" {
		opRoot = root
		if !filepath.IsAbs(root) {
			opRoot = filepath.Join(workingDir, root)
		}
	}

	opRoot, err = filepath.Abs(opRoot)
	if err != nil {
		return "", "", "", fmt.Errorf("Abs: %w", err)
	}

	return configPath, projectRoot, opRoot, nil
}
