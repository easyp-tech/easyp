package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/migration"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestMigrationInteractiveMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                       string
		module, flagSet, requested, terminal, want bool
		wantError                                  bool
	}{
		{name: "plain terminal starts wizard", terminal: true, want: true},
		{name: "module preserves terminal preview", module: true, terminal: true},
		{name: "pipe is never interactive"},
		{name: "explicit module in pipe", module: true},
		{name: "explicit wizard with module", module: true, flagSet: true, requested: true, terminal: true, want: true},
		{name: "explicit wizard without module", flagSet: true, requested: true, terminal: true, want: true},
		{name: "wizard refuses pipe", flagSet: true, requested: true, wantError: true},
		{name: "wizard refuses redirected output even with module", module: true, flagSet: true, requested: true, wantError: true},
		{name: "explicit false disables wizard", flagSet: true, terminal: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := migrationInteractiveMode(tt.module, tt.flagSet, tt.requested, tt.terminal)
			if tt.wantError {
				require.ErrorContains(t, err, "requires terminal input and output")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationPromptConfirmation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input string
		want        bool
		wantError   error
		invalid     bool
	}{
		{name: "yes", input: "yes\n", want: true},
		{name: "uppercase yes and CRLF", input: "Y\r\n", want: true},
		{name: "no", input: "no\n"},
		{name: "Enter defaults to no", input: "\n"},
		{name: "bad answer must be corrected", input: "maybe\nyes\n", want: true, invalid: true},
		{name: "EOF does not accept default", wantError: io.EOF},
		{name: "unterminated yes is not consent", input: "yes", wantError: io.EOF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			got, err := newMigrationPrompt(strings.NewReader(tt.input), &out).confirm(t.Context(), "Apply?")
			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
				assert.False(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, out.String(), "[y/N]")
			assert.Equal(t, tt.invalid, strings.Contains(out.String(), "Please answer yes or no"))
		})
	}
}

func TestMigrationWizardAppliesOnlyReviewedConsent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input string
		apply       bool
		wantError   bool
	}{
		{name: "explicit yes", input: "yes\n", apply: true},
		{name: "explicit no", input: "no\n"},
		{name: "empty answer", input: "\n"},
		{name: "EOF", wantError: true},
		{name: "incomplete yes", input: "yes", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			original := []byte("lint:\n  use: [MINIMAL]\n")
			path := filepath.Join(root, v1.PolicyFile)
			require.NoError(t, os.WriteFile(path, original, 0o640))
			var out bytes.Buffer
			wizard := migrationWizard{
				prompt: newMigrationPrompt(strings.NewReader(tt.input), &out), writer: &out,
				repository: func() (migration.Repository, error) { t.Fatal("unneeded dependency access"); return nil, nil },
			}
			err := wizard.run(t.Context(), migration.Options{Dir: root, Module: "example.com/app", ResolveLock: true}, true, true)
			if tt.wantError {
				require.ErrorIs(t, err, io.EOF)
			} else {
				require.NoError(t, err)
			}
			assert.Contains(t, out.String(), "easyp.gen.yaml (candidate")
			assert.Contains(t, out.String(), "Apply these changes")
			assert.NotContains(t, out.String(), "Allow dependency")
			after, err := os.ReadFile(path)
			require.NoError(t, err)
			if !tt.apply {
				assert.Equal(t, original, after)
				assert.NoFileExists(t, path+".v0.bak")
				assert.NoFileExists(t, filepath.Join(root, v1.ModuleFile))
				return
			}
			assert.NotEqual(t, original, after)
			backup, err := os.ReadFile(path + ".v0.bak")
			require.NoError(t, err)
			assert.Equal(t, original, backup)
			out.Reset()
			wizard.prompt = newMigrationPrompt(strings.NewReader(""), &out)
			require.NoError(t, wizard.run(t.Context(), migration.Options{Dir: root, Module: "example.com/app"}, true, true))
			assert.Contains(t, out.String(), "Already v1")
			assert.NotContains(t, out.String(), "Apply these changes")
		})
	}
}

type interactiveMigrationRepository struct {
	calls int
	err   error
}

func (r *interactiveMigrationRepository) FetchMigration(_ context.Context, source, version, _ string) (modules.Fetched, error) {
	r.calls++
	return modules.Fetched{Module: v1.Module{Name: source, Roots: []string{"."}}, Lock: v1.LockedModule{
		Source: source, Version: version, Commit: strings.Repeat("a", 40), Hash: "h1:" + strings.Repeat("A", 43) + "=",
	}}, r.err
}

func TestMigrationWizardSeparatesDependencyAndWriteConsent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, answers     string
		access, apply     bool
		resolutionFailure bool
	}{
		{name: "decline access", answers: "no\n"},
		{name: "empty access confirmation", answers: "\n"},
		{name: "access is not write consent", answers: "yes\nno\n", access: true},
		{name: "both permissions", answers: "yes\nyes\n", access: true, apply: true},
		{name: "verification failure never asks to apply", answers: "yes\nyes\n", access: true, resolutionFailure: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			original := []byte("lint: {}\ndeps: [github.com/acme/dep@v1.0.0]\n")
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.PolicyFile), original, 0o644))
			repo := &interactiveMigrationRepository{}
			failure := errors.New("legacy integrity mismatch")
			if tt.resolutionFailure {
				repo.err = failure
			}
			opened := 0
			var out bytes.Buffer
			wizard := migrationWizard{
				prompt: newMigrationPrompt(strings.NewReader(tt.answers), &out), writer: &out,
				repository: func() (migration.Repository, error) { opened++; return repo, nil },
			}
			err := wizard.run(t.Context(), migration.Options{Dir: root, Module: "example.com/app", ResolveLock: true}, true, true)
			if tt.resolutionFailure {
				require.ErrorIs(t, err, failure)
				assert.NotContains(t, out.String(), "Apply these changes")
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.access, opened != 0)
			assert.Equal(t, tt.access, repo.calls != 0)
			if !tt.apply {
				after, err := os.ReadFile(filepath.Join(root, v1.PolicyFile))
				require.NoError(t, err)
				assert.Equal(t, original, after)
				assert.NoFileExists(t, filepath.Join(root, v1.LockFile))
				return
			}
			assert.FileExists(t, filepath.Join(root, v1.LockFile))
			assert.Less(t, strings.Index(out.String(), "Verified migration preview:"), strings.Index(out.String(), "Apply these changes"))
		})
	}
}

func TestMigrationWizardInputCorrection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, v1.PolicyFile), []byte("lint: {}\n"), 0o644))
	var out bytes.Buffer
	wizard := migrationWizard{
		prompt: newMigrationPrompt(strings.NewReader(filepath.Join(root, "missing")+"\n"+root+"\nbad identity\nexample.com/app\nyes\n"), &out), writer: &out,
	}
	require.NoError(t, wizard.run(t.Context(), migration.Options{Dir: root}, false, false))
	assert.Contains(t, out.String(), "Invalid directory")
	assert.Contains(t, out.String(), "Invalid module identity")
	assert.FileExists(t, filepath.Join(root, v1.ModuleFile))
}

func TestMigrateFlagModeNeverPromptsOnPipes(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		args      []string
		wantError string
	}{
		{name: "missing module", wantError: "--module is required"},
		{name: "forced interactive", args: []string{"--interactive"}, wantError: "requires terminal"},
		{name: "explicit false", args: []string{"--interactive=false"}, wantError: "--module is required"},
		{name: "flag preview", args: []string{"--module", "example.com/app"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.PolicyFile), []byte("lint: {}\n"), 0o644))
			var out bytes.Buffer
			// The CLI library shares its default help flag; do not share it across isolated app tests.
			command := (Migrate{}).Command()
			command.HideHelp = true
			app := &cli.App{HideHelp: true, HideHelpCommand: true, Commands: []*cli.Command{command}, Reader: strings.NewReader("yes\nyes\n"), Writer: &out, ErrWriter: &out}
			args := append([]string{"easyp", "migrate", "--dir", root}, tt.args...)
			err := app.RunContext(t.Context(), args)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				assert.Contains(t, out.String(), "Preview only")
			}
			assert.NotContains(t, out.String(), "[y/N]")
			assert.NoFileExists(t, filepath.Join(root, v1.ModuleFile))
		})
	}
}

type migrationTestWriter func([]byte) (int, error)

func (f migrationTestWriter) Write(p []byte) (int, error) { return f(p) }

func TestMigrationWizardRejectsSourceChangeDuringConfirmation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, v1.PolicyFile)
	require.NoError(t, os.WriteFile(path, []byte("lint: {}\n"), 0o644))
	changed := []byte("lint: {}\n# user's concurrent edit\n")
	var out bytes.Buffer
	writer := migrationTestWriter(func(p []byte) (int, error) {
		if bytes.Contains(p, []byte("Apply these changes")) {
			require.NoError(t, os.WriteFile(path, changed, 0o644))
		}
		return out.Write(p)
	})
	wizard := migrationWizard{prompt: newMigrationPrompt(strings.NewReader("yes\n"), writer), writer: writer}
	err := wizard.run(t.Context(), migration.Options{Dir: root, Module: "example.com/app"}, true, true)
	require.ErrorContains(t, err, "changed")
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, changed, after)
	assert.NoFileExists(t, path+".v0.bak")
	assert.NoFileExists(t, filepath.Join(root, v1.ModuleFile))
}

func TestMigrationPromptFailuresAreNotConsent(t *testing.T) {
	t.Parallel()
	t.Run("cancelled context", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		var out bytes.Buffer
		ok, err := newMigrationPrompt(strings.NewReader("yes\n"), &out).confirm(ctx, "Apply?")
		require.ErrorIs(t, err, context.Canceled)
		assert.False(t, ok)
		assert.Empty(t, out.String())
	})
	t.Run("output unavailable", func(t *testing.T) {
		t.Parallel()
		failed := errors.New("output unavailable")
		writer := migrationTestWriter(func([]byte) (int, error) { return 0, failed })
		ok, err := newMigrationPrompt(strings.NewReader("yes\n"), writer).confirm(t.Context(), "Apply?")
		require.ErrorIs(t, err, failed)
		assert.False(t, ok)
	})
	t.Run("oversized input", func(t *testing.T) {
		t.Parallel()
		ok, err := newMigrationPrompt(strings.NewReader(strings.Repeat("y", 32*1024)+"\n"), io.Discard).confirm(t.Context(), "Apply?")
		require.Error(t, err)
		assert.False(t, ok)
	})
}

func TestMigrationSuggestedIdentityIsLocalAndCredentialFree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	runTestGit(t, root, "init", "-q")
	runTestGit(t, root, "remote", "add", "origin", "https://test-user:do-not-show@example.test/acme/contracts.git?token=hidden")
	assert.Equal(t, "example.test/acme/contracts", migrationSuggestedIdentity(t.Context(), root))
	require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), []byte("module example.test/native\n"), 0o644))
	assert.Equal(t, "example.test/native", migrationSuggestedIdentity(t.Context(), root))
}

func TestMigrationWizardDoesNotResolveChangedDependencies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, v1.PolicyFile)
	require.NoError(t, os.WriteFile(path, []byte("lint: {}\ndeps: [github.com/acme/dep@v1.0.0]\n"), 0o644))
	var out bytes.Buffer
	writer := migrationTestWriter(func(p []byte) (int, error) {
		if bytes.Contains(p, []byte("Allow dependency resolution")) {
			require.NoError(t, os.WriteFile(path, []byte("lint: {}\ndeps: [github.com/acme/different@v1.0.0]\n"), 0o644))
		}
		return out.Write(p)
	})
	wizard := migrationWizard{
		prompt: newMigrationPrompt(strings.NewReader("yes\nyes\n"), writer), writer: writer,
		repository: func() (migration.Repository, error) { t.Fatal("unreviewed dependencies accessed"); return nil, nil },
	}
	err := wizard.run(t.Context(), migration.Options{Dir: root, Module: "example.com/app"}, true, true)
	require.ErrorContains(t, err, "changed")
	assert.NoFileExists(t, filepath.Join(root, v1.ModuleFile))
}
