package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPolicyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantV1  bool
		wantErr bool
	}{
		{name: "explicit v1", raw: "version: v1\n", wantV1: true},
		{name: "explicit legacy", raw: "version: v0\nlinters: {}\n"},
		{name: "explicit empty version", raw: "version: ''\nlinters: {}\n"},
		{name: "legacy lint", raw: "lint:\n  use: [DEFAULT]\n"},
		{name: "legacy generate", raw: "generate:\n  plugins: []\n"},
		{name: "legacy breaking", raw: "breaking:\n  against_git_ref: main\n"},
		{name: "empty", raw: "{}\n"},
		{name: "comments only", raw: "# policy\n"},
		{name: "ambiguous breaking", raw: "breaking:\n  ignore: [generated]\n"},
		{name: "linters", raw: "linters:\n  default: MINIMAL\n", wantV1: true},
		{name: "empty linters", raw: "linters: {}\n", wantV1: true},
		{name: "null linters", raw: "linters:\n", wantV1: true},
		{name: "linter settings", raw: "linters-settings: {}\n", wantV1: true},
		{name: "issues", raw: "issues: {}\n", wantV1: true},
		{name: "breaking baseline", raw: "breaking:\n  baseline: git:main\n", wantV1: true},
		{name: "breaking categories", raw: "breaking:\n  categories: []\n", wantV1: true},
		{name: "breaking ignore unstable", raw: "breaking:\n  ignore_unstable: false\n", wantV1: true},
		{name: "breaking extends", raw: "breaking:\n  extends: ''\n", wantV1: true},
		{name: "malformed yaml", raw: "linters: [\n", wantErr: true},
		{name: "duplicate keys", raw: "version: v1\nversion: v0\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := IsPolicyConfig([]byte(tt.raw))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantV1, got)
		})
	}
}

func TestParsePolicy(t *testing.T) {
	policy, err := ParsePolicy(strings.NewReader(`version: v1
linters:
  default: MINIMAL
  enable: [FILE_LOWER_SNAKE_CASE]
  disable: [PACKAGE_DIRECTORY_MATCH]
`))
	if err != nil {
		t.Fatal(err)
	}
	if policy.Linters.Default != "MINIMAL" || policy.Linters.Enable[0] != "FILE_LOWER_SNAKE_CASE" {
		t.Fatalf("unexpected policy: %#v", policy)
	}
}

func TestParsePolicyRejectsLegacyLint(t *testing.T) {
	if _, err := ParsePolicy(strings.NewReader("version: v1\nlint:\n  use: [DEFAULT]\n")); err == nil {
		t.Fatal("v1 policy accepted legacy lint key")
	}
}

func TestPathlessIssueExclusion(t *testing.T) {
	policy, err := ParsePolicy(strings.NewReader(`version: v1
issues:
  exclude-rules:
    - linters: [FILE_LOWER_SNAKE_CASE]
`))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := policy.LegacyLint()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Except) != 1 || cfg.Except[0] != "FILE_LOWER_SNAKE_CASE" || policy.ExcludesAllIssues() {
		t.Fatalf("unexpected exclusion: %#v", cfg.Except)
	}
	all, err := ParsePolicy(strings.NewReader("version: v1\nissues:\n  exclude-rules:\n    - {}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !all.ExcludesAllIssues() {
		t.Fatal("pathless empty rule did not suppress all issues")
	}
	withPath, err := ParsePolicy(strings.NewReader("version: v1\nissues:\n  exclude-rules:\n    - path: 'legacy/**'\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := withPath.LegacyLint(); err == nil {
		t.Fatal("path glob unexpectedly accepted without specified base semantics")
	}
}

func TestLegacyBreakingUsesV1Baseline(t *testing.T) {
	policy, err := ParsePolicy(strings.NewReader("version: v1\nbreaking:\n  baseline: git:main\n  ignore: [generated]\n"))
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := policy.LegacyBreaking("master")
	if err != nil {
		t.Fatal(err)
	}
	if legacy.AgainstGitRef != "main" || len(legacy.Ignore) != 1 || legacy.Ignore[0] != "generated" {
		t.Fatalf("unexpected breaking config: %#v", legacy)
	}
	policy.Breaking.Categories = []string{"WIRE"}
	if _, err := policy.LegacyBreaking("master"); err == nil {
		t.Fatal("category selection was silently accepted")
	}
}
