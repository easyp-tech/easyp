package api

import (
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

// initializeV1 writes the three independent v1 files. All overwrite choices
// are collected before changing any file, so declining one cannot truncate it.
func initializeV1(ctx context.Context, root, identity string, prompt prompter.Prompter) error {
	for _, name := range []string{"buf.yaml", "buf.yml"} {
		_, err := os.Stat(filepath.Join(root, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("%s exists: Buf dependency migration needs explicit Git module mappings and is not available yet", name)
	}
	if identity == "" {
		raw, found, err := readOptionalFile(filepath.Join(root, "protobuf.mod"))
		if err != nil {
			return err
		}
		if found {
			module, err := v1.ParseModule(strings.NewReader(string(raw)))
			if err != nil {
				return err
			}
			identity = module.Name
		}
	}
	if identity == "" {
		identity = gitV1ModuleIdentity(ctx, root)
	}
	if identity == "" {
		value, err := prompt.Input(ctx, "Protobuf module identity (for example github.com/acme/service)", "")
		if err != nil {
			return fmt.Errorf("module identity is required; pass --module: %w", err)
		}
		identity = strings.TrimSpace(value)
	}
	moduleContents := []byte("module " + identity + "\n")
	if _, err := v1.ParseModule(strings.NewReader(string(moduleContents))); err != nil {
		return fmt.Errorf("module identity %q: %w", identity, err)
	}
	policyContents := []byte("version: v1\nlinters:\n  default: STANDARD\nbreaking:\n  baseline: git:main\n")
	if _, err := v1.ParsePolicy(strings.NewReader(string(policyContents))); err != nil {
		return err
	}
	genContents := []byte("version: v1\nplugins: []\n")
	if _, err := v1.ParseGenerate(strings.NewReader(string(genContents))); err != nil {
		return err
	}

	targets := []struct {
		name string
		data []byte
	}{
		{"protobuf.mod", moduleContents},
		{"easyp.yaml", policyContents},
		{"easyp.gen.yaml", genContents},
	}
	var selected []int
	for i, target := range targets {
		path := filepath.Join(root, target.name)
		current, found, err := readOptionalFile(path)
		if err != nil {
			return err
		}
		if found {
			if string(current) == string(target.data) {
				continue
			}
			overwrite, err := prompt.Confirm(ctx, target.name+" already exists. Overwrite?", false)
			if err != nil {
				return err
			}
			if !overwrite {
				continue
			}
		}
		selected = append(selected, i)
	}
	for _, i := range selected {
		target := targets[i]
		path := filepath.Join(root, target.name)
		tmp, err := os.CreateTemp(root, ".easyp-init-*")
		if err != nil {
			return err
		}
		if _, err := tmp.Write(target.data); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			return err
		}
		if err := tmp.Close(); err != nil {
			_ = os.Remove(tmp.Name())
			return err
		}
		if err := os.Rename(tmp.Name(), path); err != nil {
			_ = os.Remove(tmp.Name())
			return err
		}
	}
	return nil
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
