package policy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func writeBreakingResolverPolicy(t *testing.T, root, name, body string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func parseResolverPolicy(t *testing.T, raw string) (v1.Policy, v1.PolicyPresence) {
	t.Helper()
	policy, err := v1.ParsePolicy(strings.NewReader(raw))
	require.NoError(t, err)
	presence, err := v1.ParsePolicyPresence([]byte(raw))
	require.NoError(t, err)
	return policy, presence
}

func TestResolveBreakingLocalInheritance(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	base := writeBreakingResolverPolicy(t, root, "base.yaml", "version: v1\nbreaking:\n  baseline: git:base\n  categories: [WIRE]\n  ignore_unstable: true\n  ignore: [base]\n")
	localRaw := "version: v1\nbreaking:\n  extends: ./base.yaml\n  categories: [FILE]\n  ignore_unstable: false\n  ignore: []\n"
	local := writeBreakingResolverPolicy(t, root, "easyp.yaml", localRaw)
	policy, presence := parseResolverPolicy(t, localRaw)

	result, err := NewResolver(root, nil).ResolveBreaking(t.Context(), BreakingInput{
		PolicyPath: local,
		Policy:     policy,
		Presence:   presence,
	})

	require.NoError(t, err)
	assert.Equal(t, "git:base", result.Policy.Breaking.Baseline)
	assert.Equal(t, []string{"FILE"}, result.Policy.Breaking.Categories)
	assert.False(t, result.Policy.Breaking.IgnoreUnstable)
	assert.Empty(t, result.Policy.Breaking.Ignore)
	assert.Empty(t, result.Policy.Breaking.Extends)
	canonicalBase, err := filepath.EvalSymlinks(base)
	require.NoError(t, err)
	assert.Equal(t, []string{canonicalBase, local}, result.Chain)
}

func TestResolveBreakingRemoteReferencesUseConsumerGraph(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	consumer := filepath.Join(root, "consumer")
	require.NoError(t, os.MkdirAll(consumer, 0o755))

	baseRoot := filepath.Join(root, "remote-base")
	teamRoot := filepath.Join(root, "remote-team")
	writeBreakingResolverPolicy(t, baseRoot, "team/breaking.yaml", "version: v1\nbreaking:\n  baseline: git:wrong\n")
	teamPolicy := writeBreakingResolverPolicy(t, teamRoot, "breaking.yaml", "version: v1\nbreaking:\n  baseline: git:team\n")
	fragmentPolicy := writeBreakingResolverPolicy(t, baseRoot, "nested/base.yaml", "version: v1\nbreaking:\n  baseline: git:fragment\n")

	graph := modules.PolicyGraph{
		"example.test/policies":      {Name: "example.test/policies", Directory: baseRoot},
		"example.test/policies/team": {Name: "example.test/policies/team", Directory: teamRoot},
	}
	calls := 0
	resolver := NewResolver(root, func(_ context.Context, moduleDir string) (modules.PolicyGraph, error) {
		calls++
		assert.Equal(t, consumer, moduleDir)
		return graph, nil
	})

	tests := []struct {
		name      string
		reference string
		want      string
		wantPath  string
	}{
		{
			name:      "longest module prefix",
			reference: "example.test/policies/team/breaking.yaml",
			want:      "git:team",
			wantPath:  teamPolicy,
		},
		{
			name:      "explicit fragment",
			reference: "example.test/policies#nested/base.yaml",
			want:      "git:fragment",
			wantPath:  fragmentPolicy,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			localRaw := "version: v1\nbreaking:\n  extends: " + tt.reference + "\n"
			local := writeBreakingResolverPolicy(t, root, filepath.Join("consumer", strings.ReplaceAll(tt.name, " ", "-")+".yaml"), localRaw)
			policy, presence := parseResolverPolicy(t, localRaw)

			result, err := resolver.ResolveBreaking(t.Context(), BreakingInput{
				PolicyPath: local,
				Policy:     policy,
				Presence:   presence,
				ModuleDir:  consumer,
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Policy.Breaking.Baseline)
			canonicalBase, err := filepath.EvalSymlinks(tt.wantPath)
			require.NoError(t, err)
			assert.Equal(t, canonicalBase, result.Chain[0])
			assert.Equal(t, local, result.Chain[len(result.Chain)-1])
		})
	}
	assert.Equal(t, 1, calls, "consumer graph should be cached across references")
}

func TestResolveBreakingRejectsInvalidInheritance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, root string) (BreakingInput, GraphProvider)
		want  string
	}{
		{
			name: "cycle",
			setup: func(t *testing.T, root string) (BreakingInput, GraphProvider) {
				aRaw := "version: v1\nbreaking:\n  extends: ./b.yaml\n"
				a := writeBreakingResolverPolicy(t, root, "a.yaml", aRaw)
				writeBreakingResolverPolicy(t, root, "b.yaml", "version: v1\nbreaking:\n  extends: ./a.yaml\n")
				policy, presence := parseResolverPolicy(t, aRaw)
				return BreakingInput{PolicyPath: a, Policy: policy, Presence: presence}, nil
			},
			want: "policy extends cycle",
		},
		{
			name: "base lacks breaking",
			setup: func(t *testing.T, root string) (BreakingInput, GraphProvider) {
				raw := "version: v1\nbreaking:\n  extends: ./base.yaml\n"
				path := writeBreakingResolverPolicy(t, root, "easyp.yaml", raw)
				writeBreakingResolverPolicy(t, root, "base.yaml", "version: v1\nlinters:\n  default: MINIMAL\n")
				policy, presence := parseResolverPolicy(t, raw)
				return BreakingInput{PolicyPath: path, Policy: policy, Presence: presence}, nil
			},
			want: "does not define breaking",
		},
		{
			name: "local escape",
			setup: func(t *testing.T, root string) (BreakingInput, GraphProvider) {
				workspace := filepath.Join(root, "workspace")
				require.NoError(t, os.MkdirAll(workspace, 0o755))
				writeBreakingResolverPolicy(t, root, "outside.yaml", "version: v1\nbreaking:\n  baseline: git:outside\n")
				raw := "version: v1\nbreaking:\n  extends: ../outside.yaml\n"
				path := writeBreakingResolverPolicy(t, workspace, "easyp.yaml", raw)
				policy, presence := parseResolverPolicy(t, raw)
				return BreakingInput{PolicyPath: path, Policy: policy, Presence: presence}, nil
			},
			want: "outside allowed policy root",
		},
		{
			name: "undeclared remote module",
			setup: func(t *testing.T, root string) (BreakingInput, GraphProvider) {
				raw := "version: v1\nbreaking:\n  extends: example.test/missing#base.yaml\n"
				path := writeBreakingResolverPolicy(t, root, "easyp.yaml", raw)
				policy, presence := parseResolverPolicy(t, raw)
				provider := func(context.Context, string) (modules.PolicyGraph, error) {
					return modules.PolicyGraph{}, nil
				}
				return BreakingInput{PolicyPath: path, Policy: policy, Presence: presence, ModuleDir: root}, provider
			},
			want: "not declared and resolved",
		},
		{
			name: "remote needs module context",
			setup: func(t *testing.T, root string) (BreakingInput, GraphProvider) {
				raw := "version: v1\nbreaking:\n  extends: example.test/policies#base.yaml\n"
				path := writeBreakingResolverPolicy(t, root, "easyp.yaml", raw)
				policy, presence := parseResolverPolicy(t, raw)
				return BreakingInput{PolicyPath: path, Policy: policy, Presence: presence}, nil
			},
			want: "require the consuming protobuf.mod",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			input, provider := tt.setup(t, root)
			workspace := root
			if tt.name == "local escape" {
				workspace = filepath.Join(root, "workspace")
			}

			_, err := NewResolver(workspace, provider).ResolveBreaking(t.Context(), input)

			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestMatchModuleReference(t *testing.T) {
	t.Parallel()

	graph := modules.PolicyGraph{
		"example.test/policies":      {Name: "example.test/policies", Directory: "/base"},
		"example.test/policies/team": {Name: "example.test/policies/team", Directory: "/team"},
	}

	tests := []struct {
		name         string
		raw          string
		wantModule   string
		wantRelative string
		wantError    string
	}{
		{name: "exact", raw: "example.test/policies", wantModule: "example.test/policies"},
		{name: "longest prefix", raw: "example.test/policies/team/base.yaml", wantModule: "example.test/policies/team", wantRelative: "base.yaml"},
		{name: "fragment", raw: "example.test/policies#nested/base.yaml", wantModule: "example.test/policies", wantRelative: "nested/base.yaml"},
		{name: "unmatched", raw: "example.test/other/base.yaml", wantError: "does not match"},
		{name: "fragment undeclared", raw: "example.test/other#base.yaml", wantError: "not declared and resolved"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reference, err := v1.ParsePolicyReference(tt.raw)
			require.NoError(t, err)

			module, relative, err := matchModuleReference(reference, graph)

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantModule, module)
			assert.Equal(t, tt.wantRelative, relative)
		})
	}
}
