package policy

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestResolverEnvironmentIsolation(t *testing.T) {
	const (
		environmentName = "EASY_SECRET_PROBE"
		secret          = "consumer-secret-value"
		moduleName      = "example.test/policies"
	)
	t.Setenv(environmentName, secret)

	t.Run("remote module stays literal even inside workspace", func(t *testing.T) {
		workspace := t.TempDir()
		dependency := filepath.Join(workspace, ".deps", "policies")
		base := filepath.Join(dependency, "base.yaml")
		writeResolverPolicy(t, base, `version: v1
linters:
  default: MINIMAL
linters-settings:
  SERVICE_SUFFIX:
    suffix: ${EASY_SECRET_PROBE}
`)
		result := resolveLintPolicy(t, workspace, moduleName+"#base.yaml", moduleName, dependency)

		assert.Equal(t, "${EASY_SECRET_PROBE}", result.Policy.LinterSettings["SERVICE_SUFFIX"]["suffix"])
		assert.NotContains(t, result.Policy.LinterSettings["SERVICE_SUFFIX"]["suffix"], secret)
	})

	t.Run("consumer local base still expands", func(t *testing.T) {
		workspace := t.TempDir()
		writeResolverPolicy(t, filepath.Join(workspace, "base.yaml"), `version: v1
linters:
  default: MINIMAL
linters-settings:
  SERVICE_SUFFIX:
    suffix: ${EASY_SECRET_PROBE}
`)
		result := resolveLintPolicy(t, workspace, "./base.yaml", "", "")

		assert.Equal(t, secret, result.Policy.LinterSettings["SERVICE_SUFFIX"]["suffix"])
	})

	t.Run("dependency local chain stays literal", func(t *testing.T) {
		workspace := t.TempDir()
		dependency := filepath.Join(workspace, ".deps", "policies")
		writeResolverPolicy(t, filepath.Join(dependency, "base.yaml"), `version: v1
linters:
  extends: ./nested.yaml
`)
		writeResolverPolicy(t, filepath.Join(dependency, "nested.yaml"), `version: v1
linters:
  default: MINIMAL
linters-settings:
  SERVICE_SUFFIX:
    suffix: ${EASY_SECRET_PROBE}
`)
		result := resolveLintPolicy(t, workspace, moduleName+"#base.yaml", moduleName, dependency)

		assert.Equal(t, "${EASY_SECRET_PROBE}", result.Policy.LinterSettings["SERVICE_SUFFIX"]["suffix"])
	})

	t.Run("invalid remote placeholder never leaks value", func(t *testing.T) {
		workspace := t.TempDir()
		dependency := filepath.Join(workspace, ".deps", "policies")
		writeResolverPolicy(t, filepath.Join(dependency, "base.yaml"), `version: v1
linters:
  default: ${EASY_SECRET_PROBE}
`)
		_, err := resolveLintPolicyResult(t, workspace, moduleName+"#base.yaml", moduleName, dependency)

		require.Error(t, err)
		assert.NotContains(t, err.Error(), secret)
		assert.Contains(t, err.Error(), environmentName)
	})
}

func resolveLintPolicy(t *testing.T, workspace, reference, moduleName, dependency string) LintResult {
	t.Helper()
	result, err := resolveLintPolicyResult(t, workspace, reference, moduleName, dependency)
	require.NoError(t, err)
	return result
}

func resolveLintPolicyResult(t *testing.T, workspace, reference, moduleName, dependency string) (LintResult, error) {
	t.Helper()
	policyPath := filepath.Join(workspace, v1.PolicyFile)
	raw := []byte("version: v1\nlinters:\n  extends: " + reference + "\n")
	writeResolverPolicy(t, policyPath, string(raw))
	parsed, err := v1.ParsePolicy(bytes.NewReader(raw))
	require.NoError(t, err)
	presence, err := v1.ParsePolicyPresence(raw)
	require.NoError(t, err)

	var graph GraphProvider
	if moduleName != "" {
		graph = func(context.Context, string) (modules.PolicyGraph, error) {
			return modules.PolicyGraph{
				moduleName: {Name: moduleName, Directory: dependency},
			}, nil
		}
	}
	resolver := NewResolver(workspace, graph)
	return resolver.ResolveLint(t.Context(), LintInput{
		PolicyPath:     policyPath,
		Policy:         parsed,
		Presence:       presence,
		Settings:       parsed.LinterSettings,
		SettingsSource: presence,
		ModuleDir:      workspace,
	})
}

func writeResolverPolicy(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}
