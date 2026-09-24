package v1

import (
	"strings"
	"testing"
)

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
	cfg, err := policy.LintConfig()
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
	if _, err := withPath.LintConfig(); err == nil {
		t.Fatal("path glob unexpectedly accepted without specified base semantics")
	}
}

func TestBreakingConfigUsesV1Baseline(t *testing.T) {
	policy, err := ParsePolicy(strings.NewReader("version: v1\nbreaking:\n  baseline: git:main\n  ignore: [generated]\n"))
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := policy.BreakingConfig("master")
	if err != nil {
		t.Fatal(err)
	}
	if legacy.AgainstGitRef != "main" || len(legacy.Ignore) != 1 || legacy.Ignore[0] != "generated" {
		t.Fatalf("unexpected breaking config: %#v", legacy)
	}
	policy.Breaking.Categories = []string{"WIRE"}
	if _, err := policy.BreakingConfig("master"); err == nil {
		t.Fatal("category selection was silently accepted")
	}
}
