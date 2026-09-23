package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/prompter"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestInitializeV1WritesSelectedConfigs(t *testing.T) {
	t.Parallel()

	const (
		module          = "module example.com/service\n"
		policy          = "version: v1\nlinters:\n  default: STANDARD\nbreaking:\n  baseline: git:main\n"
		generator       = "version: v1\nplugins: []\n"
		customPolicy    = "version: v1\nlinters:\n  default: MINIMAL\n"
		customGenerator = "version: v1\nplugins:\n  - name: go\n    out: gen\n"
	)
	tests := []struct {
		name                  string
		existing              map[string]string
		overwrite             bool
		expectedPolicy        string
		expectedGenerator     string
		expectedConfirmations int
	}{
		{
			name:              "create defaults without prompting",
			expectedPolicy:    policy,
			expectedGenerator: generator,
		},
		{
			name:              "identical files need no confirmation",
			existing:          map[string]string{v1.ModuleFile: module, v1.PolicyFile: policy, v1.GenerateFile: generator},
			expectedPolicy:    policy,
			expectedGenerator: generator,
		},
		{
			name:                  "declined files keep their contents",
			existing:              map[string]string{v1.PolicyFile: customPolicy, v1.GenerateFile: customGenerator},
			expectedPolicy:        customPolicy,
			expectedGenerator:     customGenerator,
			expectedConfirmations: 2,
		},
		{
			name:                  "confirmed files are replaced",
			existing:              map[string]string{v1.PolicyFile: customPolicy, v1.GenerateFile: customGenerator},
			overwrite:             true,
			expectedPolicy:        policy,
			expectedGenerator:     generator,
			expectedConfirmations: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for name, contents := range tt.existing {
				require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600))
			}
			prompt := &initPrompter{overwrite: tt.overwrite}

			err := initializeV1(context.Background(), root, "example.com/service", prompt)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfirmations, prompt.confirmations)
			for name, contents := range map[string]string{
				v1.ModuleFile: module, v1.PolicyFile: tt.expectedPolicy, v1.GenerateFile: tt.expectedGenerator,
			} {
				path := filepath.Join(root, name)
				written, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.Equal(t, contents, string(written), name)
				info, err := os.Stat(path)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), name)
				issues, err := v1.ValidateFile(path)
				require.NoError(t, err)
				assert.Empty(t, issues, name)
			}
		})
	}
}

func TestInitializeV1RejectsBeforeWriting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		identity        string
		existing        map[string]string
		expectedMessage string
	}{
		{
			name:            "invalid identity",
			identity:        "invalid identity",
			expectedMessage: "expected one module identity",
		},
		{
			name:            "buf yaml requires explicit migration",
			identity:        "example.com/service",
			existing:        map[string]string{"buf.yaml": "version: v2\n"},
			expectedMessage: "buf.yaml exists",
		},
		{
			name:            "buf yml requires explicit migration",
			identity:        "example.com/service",
			existing:        map[string]string{"buf.yml": "version: v1\n"},
			expectedMessage: "buf.yml exists",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for name, contents := range tt.existing {
				require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600))
			}
			prompt := &initPrompter{}

			err := initializeV1(context.Background(), root, tt.identity, prompt)

			require.ErrorContains(t, err, tt.expectedMessage)
			assert.Zero(t, prompt.confirmations)
			files, err := os.ReadDir(root)
			require.NoError(t, err)
			assert.Len(t, files, len(tt.existing))
			for name, contents := range tt.existing {
				written, err := os.ReadFile(filepath.Join(root, name))
				require.NoError(t, err)
				assert.Equal(t, contents, string(written))
			}
		})
	}
}

func TestInitializeV1PromptErrorKeepsFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		existing    string
		promptError error
	}{
		{name: "cancel policy overwrite", existing: v1.PolicyFile, promptError: context.Canceled},
		{name: "cancel generator overwrite", existing: v1.GenerateFile, promptError: context.Canceled},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			path := filepath.Join(root, tt.existing)
			require.NoError(t, os.WriteFile(path, []byte("keep this file\n"), 0o600))
			prompt := &initPrompter{confirmErr: tt.promptError}

			err := initializeV1(context.Background(), root, "example.com/service", prompt)

			require.ErrorIs(t, err, tt.promptError)
			assert.Equal(t, 1, prompt.confirmations)
			files, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Len(t, files, 1)
			assert.Equal(t, tt.existing, files[0].Name())
			written, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, "keep this file\n", string(written))
		})
	}
}

// Explicit identities avoid interactive input; only overwrite prompts are used.
type initPrompter struct {
	prompter.Prompter
	overwrite     bool
	confirmErr    error
	confirmations int
}

func (p *initPrompter) Confirm(context.Context, string, bool) (bool, error) {
	p.confirmations++
	return p.overwrite, p.confirmErr
}
