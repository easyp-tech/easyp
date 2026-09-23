package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/prompter"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type initialConfigFile struct {
	name string
	data []byte
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
	targets, err := initialConfigFiles(identity)
	if err != nil {
		return fmt.Errorf("initialConfigFiles: %w", err)
	}
	selected, err := selectInitialConfigFiles(ctx, root, targets, prompt)
	if err != nil {
		return fmt.Errorf("selectInitialConfigFiles: %w", err)
	}
	for _, target := range selected {
		if err := writeAtomicFile(filepath.Join(root, target.name), target.data, 0o600); err != nil {
			return fmt.Errorf("writeAtomicFile: %w", err)
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
	if identity := gitV1ModuleIdentity(ctx, root); identity != "" {
		return identity, nil
	}
	value, err := prompt.Input(ctx, "Protobuf module identity (for example github.com/acme/service)", "")
	if err != nil {
		return "", fmt.Errorf("module identity is required; pass --module: %w", err)
	}
	return strings.TrimSpace(value), nil
}

func initialConfigFiles(identity string) ([]initialConfigFile, error) {
	moduleContents := []byte("module " + identity + "\n")
	if _, err := v1.ParseModule(bytes.NewReader(moduleContents)); err != nil {
		return nil, fmt.Errorf("module identity %q: %w", identity, err)
	}
	policyContents := []byte("version: v1\nlinters:\n  default: STANDARD\nbreaking:\n  baseline: git:main\n")
	if _, err := v1.ParsePolicy(bytes.NewReader(policyContents)); err != nil {
		return nil, fmt.Errorf("ParsePolicy: %w", err)
	}
	genContents := []byte("version: v1\nplugins: []\n")
	if _, err := v1.ParseGenerate(bytes.NewReader(genContents)); err != nil {
		return nil, fmt.Errorf("ParseGenerate: %w", err)
	}
	return []initialConfigFile{
		{name: v1.ModuleFile, data: moduleContents},
		{name: v1.PolicyFile, data: policyContents},
		{name: v1.GenerateFile, data: genContents},
	}, nil
}

func selectInitialConfigFiles(ctx context.Context, root string, targets []initialConfigFile, prompt prompter.Prompter) ([]initialConfigFile, error) {
	var selected []initialConfigFile
	for _, target := range targets {
		path := filepath.Join(root, target.name)
		current, found, err := readOptionalFile(path)
		if err != nil {
			return nil, fmt.Errorf("readOptionalFile: %w", err)
		}
		if !found {
			selected = append(selected, target)
			continue
		}
		if bytes.Equal(current, target.data) {
			continue
		}
		overwrite, err := prompt.Confirm(ctx, target.name+" already exists. Overwrite?", false)
		if err != nil {
			return nil, fmt.Errorf("Confirm: %w", err)
		}
		if !overwrite {
			continue
		}
		selected = append(selected, target)
	}
	return selected, nil
}

func gitV1ModuleIdentity(ctx context.Context, root string) string {
	repoRoot, err := gitV1(ctx, root, "rev-parse", "--show-toplevel")
	if err != nil || filepath.Clean(strings.TrimSpace(repoRoot)) != filepath.Clean(root) {
		return ""
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "remote", "get-url", "origin")
	raw, err := cmd.Output()
	if err != nil {
		return ""
	}
	remote := strings.TrimSpace(string(raw))
	if strings.Contains(remote, "://") {
		parsed, err := url.Parse(remote)
		if err != nil || parsed.Host == "" {
			return ""
		}
		return strings.TrimSuffix(parsed.Host+"/"+strings.TrimPrefix(parsed.Path, "/"), ".git")
	}
	if _, rest, ok := strings.Cut(remote, "@"); ok {
		host, path, ok := strings.Cut(rest, ":")
		if ok && host != "" && path != "" {
			return strings.TrimSuffix(host+"/"+path, ".git")
		}
	}
	return ""
}
