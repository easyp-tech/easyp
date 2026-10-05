package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/migration"
)

type migrationWizard struct {
	prompt     *migrationPrompt
	writer     io.Writer
	repository func() (migration.Repository, error)
}

func (w migrationWizard) run(ctx context.Context, options migration.Options, directorySet, moduleSet bool) error {
	if _, err := io.WriteString(w.writer, "EasyP v0 to v1 migration wizard\nNo plugins will run. Dependency access and file writes require separate confirmations.\n"); err != nil {
		return fmt.Errorf("WriteString: %w", err)
	}
	directory, err := w.directory(ctx, options.Dir, directorySet)
	if err != nil {
		return fmt.Errorf("directory: %w", err)
	}
	identity, err := w.identity(ctx, directory, options.Module, moduleSet)
	if err != nil {
		return fmt.Errorf("identity: %w", err)
	}
	if _, err := fmt.Fprintf(w.writer, "\nProject: %q\nModule: %q\n\nMigration preview (no dependency access):\n", directory, identity); err != nil {
		return fmt.Errorf("Fprintf: %w", err)
	}
	// Flags may prefill the form, but never bypass consent in the wizard.
	options = migration.Options{Dir: directory, Module: identity}
	plan, err := migration.Build(ctx, options)
	if err != nil {
		return fmt.Errorf("Build: %w", err)
	}
	if err := printMigrationPlan(w.writer, plan); err != nil {
		return err
	}
	if plan.AlreadyV1() {
		return nil
	}
	if plan.NeedsLockResolution() {
		allowed, err := w.prompt.confirm(ctx, "Allow dependency resolution and cache writes (may access the network)?")
		if err != nil {
			return fmt.Errorf("confirm: %w", err)
		}
		if !allowed {
			return w.cancelled("Dependencies were not resolved.")
		}
		if err := plan.CheckUnchanged(); err != nil {
			return fmt.Errorf("CheckUnchanged: %w", err)
		}
		originalPlan := plan
		options.Repository, err = w.repository()
		if err != nil {
			return fmt.Errorf("repository: %w", err)
		}
		options.ResolveLock = true
		plan, err = migration.Build(ctx, options)
		if err != nil {
			return fmt.Errorf("Build: %w", err)
		}
		if err := originalPlan.CheckUnchanged(); err != nil {
			return fmt.Errorf("CheckUnchanged: %w", err)
		}
		if _, err := io.WriteString(w.writer, "\nVerified migration preview:\n"); err != nil {
			return fmt.Errorf("WriteString: %w", err)
		}
		if err := printMigrationPlan(w.writer, plan); err != nil {
			return err
		}
		if plan.AlreadyV1() {
			return nil
		}
	}
	allowed, err := w.prompt.confirm(ctx, "Apply these changes and create legacy backups?")
	if err != nil {
		return fmt.Errorf("confirm: %w", err)
	}
	if !allowed {
		return w.cancelled("Any explicitly authorized dependency cache entries are retained.")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Apply exactly the plan that was shown; it rechecks observed input state.
	if err := plan.Apply(); err != nil {
		return fmt.Errorf("Apply: %w", err)
	}
	if _, err := io.WriteString(w.writer, "\nMigration applied. Legacy backups were preserved.\n"); err != nil {
		return fmt.Errorf("WriteString: %w", err)
	}
	return nil
}

func (w migrationWizard) directory(ctx context.Context, value string, explicit bool) (string, error) {
	for {
		if !explicit {
			entered, err := w.prompt.input(ctx, "Migration directory", value)
			if err != nil {
				return "", err
			}
			value = entered
		}
		root, err := migrationDirectory(value)
		if err == nil || explicit {
			return root, err
		}
		if _, err := fmt.Fprintf(w.writer, "Invalid directory: %v\n", err); err != nil {
			return "", fmt.Errorf("Fprintf: %w", err)
		}
	}
}

func migrationDirectory(value string) (string, error) {
	root, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("Abs: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("Stat: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", root)
	}
	info, err = os.Lstat(filepath.Join(root, v1.PolicyFile))
	if err != nil {
		return "", fmt.Errorf("Lstat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("easyp.yaml must be a regular file")
	}
	return root, nil
}

func (w migrationWizard) identity(ctx context.Context, root, value string, explicit bool) (string, error) {
	fallback := value
	if !explicit && fallback == "" {
		fallback = migrationSuggestedIdentity(ctx, root)
	}
	for {
		if !explicit {
			entered, err := w.prompt.input(ctx, "Module identity (for example github.com/acme/contracts)", fallback)
			if err != nil {
				return "", err
			}
			value = entered
		}
		err := migration.ValidateModuleIdentity(value)
		if err == nil || explicit {
			return value, err
		}
		if _, err := fmt.Fprintf(w.writer, "Invalid module identity: %v\n", err); err != nil {
			return "", fmt.Errorf("Fprintf: %w", err)
		}
	}
}

func migrationSuggestedIdentity(ctx context.Context, root string) string {
	var identity string
	raw, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
	if err == nil {
		module, err := v1.ParseModule(bytes.NewReader(raw))
		if err == nil {
			identity = module.Name
		}
	}
	if identity == "" {
		// Reads local Git configuration only; credentials are removed by the helper.
		canonical, err := filepath.EvalSymlinks(root)
		if err == nil {
			identity = gitmodules.WorkspaceIdentity(ctx, canonical)
		}
	}
	if migration.ValidateModuleIdentity(identity) != nil || strings.IndexFunc(identity, unicode.IsControl) >= 0 {
		return ""
	}
	return identity
}

func (w migrationWizard) cancelled(detail string) error {
	if _, err := fmt.Fprintf(w.writer, "\nMigration cancelled. Project files were not changed. %s\n", detail); err != nil {
		return fmt.Errorf("Fprintf: %w", err)
	}
	return nil
}
