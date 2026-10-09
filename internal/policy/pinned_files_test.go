package policy

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestPinnedPolicyFilesPreserveRelativeChainsWithoutMaterialization(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	raw := []byte("version: v1\nlinters:\n  extends: example.test/policies#config/base.rules\n")
	filename := filepath.Join(root, "easyp.yaml")
	require.NoError(t, os.WriteFile(filename, raw, 0o644))
	cfg, err := v1.ParsePolicyLiteral(bytes.NewReader(raw))
	require.NoError(t, err)
	presence, err := v1.ParsePolicyPresence(raw)
	require.NoError(t, err)
	provider := policyFileMap{
		"config/base.rules":  "version: v1\nlinters:\n  extends: ./strict.conf\n",
		"config/strict.conf": "version: v1\nlinters:\n  default: MINIMAL\n",
	}
	boundary := filepath.Join(t.TempDir(), "not-materialized")
	resolver := NewResolver(root, func(context.Context, string) (modules.PolicyGraph, error) {
		return modules.PolicyGraph{"example.test/policies": {Name: "example.test/policies", Directory: boundary, Files: provider}}, nil
	})
	result, err := resolver.ResolveLint(t.Context(), LintInput{PolicyPath: filename, Policy: cfg, Presence: presence, ModuleDir: root})
	require.NoError(t, err)
	require.Equal(t, "MINIMAL", result.Policy.Linters.Default)
	_, err = os.Stat(boundary)
	require.ErrorIs(t, err, os.ErrNotExist)
}

type policyFileMap map[string]string

func (f policyFileMap) Read(_ context.Context, name string) (modules.PolicyFile, error) {
	data, ok := f[name]
	if !ok {
		return modules.PolicyFile{}, os.ErrNotExist
	}
	return modules.PolicyFile{Path: name, Canonical: "pinned@commit:" + name, Content: []byte(data)}, nil
}

func TestPinnedPolicySourceDoesNotLeakIntoAnotherConsumerReplacement(t *testing.T) {
	t.Parallel()
	provider := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(provider, "base.rules"), []byte("version: v1\nlinters:\n  default: STANDARD\n"), 0o644))
	first, second := t.TempDir(), t.TempDir()
	raw := []byte("version: v1\nlinters:\n  extends: example.test/policies#base.rules\n")
	for _, root := range []string{first, second} {
		require.NoError(t, os.WriteFile(filepath.Join(root, "easyp.yaml"), raw, 0o644))
	}
	resolver := NewResolver(filepath.Dir(first), func(_ context.Context, consumer string) (modules.PolicyGraph, error) {
		module := modules.PolicyModule{Name: "example.test/policies", Directory: provider, Replaced: consumer == second}
		if consumer == first {
			module.Files = policyFileMap{"base.rules": "version: v1\nlinters:\n  default: MINIMAL\n"}
		}
		return modules.PolicyGraph{"example.test/policies": module}, nil
	})
	cfg, err := v1.ParsePolicyLiteral(bytes.NewReader(raw))
	require.NoError(t, err)
	presence, err := v1.ParsePolicyPresence(raw)
	require.NoError(t, err)
	for _, tc := range []struct{ name, root, want string }{{"pinned", first, "MINIMAL"}, {"local replacement", second, "STANDARD"}} {
		resolver.WorkspaceRoot = tc.root
		result, err := resolver.ResolveLint(t.Context(), LintInput{PolicyPath: filepath.Join(tc.root, "easyp.yaml"), Policy: cfg, Presence: presence, ModuleDir: tc.root})
		require.NoError(t, err)
		require.Equal(t, tc.want, result.Policy.Linters.Default)
	}
}
