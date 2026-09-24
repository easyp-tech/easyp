package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	"github.com/easyp-tech/easyp/internal/adapters/prompter"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

type initialConfigFile struct {
	name     string
	contents []byte
}

// initializeV1 writes the three independent v1 files. All overwrite choices
// are collected before changing any file, so declining one cannot truncate it.
func initializeV1(ctx context.Context, root, identity string, prompt prompter.Prompter) error {
	if err := checkV1Initialization(root); err != nil {
		return fmt.Errorf("checkV1Initialization: %w", err)
	}
	identity, err := initialModuleIdentity(ctx, root, identity, prompt)
	if err != nil {
		return fmt.Errorf("initialModuleIdentity: %w", err)
	}
	manifest := []byte("module " + identity + "\n")
	if _, err := v1.ParseModule(bytes.NewReader(manifest)); err != nil {
		return fmt.Errorf("ParseModule: %w", err)
	}
	files := []initialConfigFile{
		{name: v1.ModuleFile, contents: manifest},
		{name: v1.PolicyFile, contents: []byte("version: v1\nlinters:\n  default: STANDARD\nbreaking:\n  baseline: git:main\n")},
		{name: v1.GenerateFile, contents: []byte("version: v1\nplugins: []\n")},
	}
	selected, err := confirmInitialConfigFiles(ctx, root, files, prompt)
	if err != nil {
		return fmt.Errorf("confirmInitialConfigFiles: %w", err)
	}
	for _, file := range selected {
		if err := disk.WriteAtomicFile(filepath.Join(root, file.name), file.contents, 0o600); err != nil {
			return fmt.Errorf("WriteAtomicFile: %w", err)
		}
	}
	return nil
}

func checkV1Initialization(root string) error {
	for _, name := range []string{"buf.yaml", "buf.yml"} {
		_, err := os.Stat(filepath.Join(root, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("Stat: %w", err)
		}
		return fmt.Errorf("%s exists: Buf dependency migration needs explicit Git module mappings and is not available yet", name)
	}
	return nil
}

func initialModuleIdentity(ctx context.Context, root, identity string, prompt prompter.Prompter) (string, error) {
	if identity != "" {
		return identity, nil
	}
	raw, found, err := readOptionalFile(filepath.Join(root, v1.ModuleFile))
	if err != nil {
		return "", fmt.Errorf("readOptionalFile: %w", err)
	}
	if found {
		module, err := v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return "", fmt.Errorf("ParseModule: %w", err)
		}
		return module.Name, nil
	}
	if identity := gitmodules.WorkspaceIdentity(ctx, root); identity != "" {
		return identity, nil
	}
	value, err := prompt.Input(ctx, "Protobuf module identity (for example github.com/acme/service)", "")
	if err != nil {
		return "", fmt.Errorf("module identity is required; pass --module: %w", err)
	}
	return strings.TrimSpace(value), nil
}

func confirmInitialConfigFiles(ctx context.Context, root string, files []initialConfigFile, prompt prompter.Prompter) ([]initialConfigFile, error) {
	var selected []initialConfigFile
	for _, file := range files {
		path := filepath.Join(root, file.name)
		current, found, err := readOptionalFile(path)
		if err != nil {
			return nil, fmt.Errorf("readOptionalFile: %w", err)
		}
		if !found {
			selected = append(selected, file)
			continue
		}
		if bytes.Equal(current, file.contents) {
			continue
		}
		overwrite, err := prompt.Confirm(ctx, file.name+" already exists. Overwrite?", false)
		if err != nil {
			return nil, fmt.Errorf("Confirm: %w", err)
		}
		if !overwrite {
			continue
		}
		selected = append(selected, file)
	}
	return selected, nil
}
