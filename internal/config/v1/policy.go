package v1

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
)

// Policy is the producer-side v1 easyp.yaml configuration.
type Policy struct {
	Version        string                       `yaml:"version"`
	Linters        LinterPolicy                 `yaml:"linters"`
	LinterSettings map[string]map[string]string `yaml:"linters-settings"`
	Issues         IssuePolicy                  `yaml:"issues"`
	Breaking       BreakingPolicy               `yaml:"breaking"`
}

// LinterPolicy selects the default preset and individual linter rules.
type LinterPolicy struct {
	Default string   `yaml:"default"`
	Enable  []string `yaml:"enable"`
	Disable []string `yaml:"disable"`
	Extends string   `yaml:"extends"`
}

// IssuePolicy controls suppression of lint findings.
type IssuePolicy struct {
	ExcludeRules []IssueExcludeRule `yaml:"exclude-rules"`
}

// IssueExcludeRule selects the linters and paths to exclude from reporting.
type IssueExcludeRule struct {
	Path    string   `yaml:"path"`
	Linters []string `yaml:"linters"`
}

// BreakingPolicy selects a baseline and compatibility checks.
type BreakingPolicy struct {
	Baseline       string   `yaml:"baseline"`
	Categories     []string `yaml:"categories"`
	IgnoreUnstable bool     `yaml:"ignore_unstable"`
	Extends        string   `yaml:"extends"`
	Ignore         []string `yaml:"ignore"`
}

// ParsePolicy reads and validates a v1 lint and breaking policy.
func ParsePolicy(r io.Reader) (Policy, error) {
	var result Policy
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return Policy{}, fmt.Errorf("decode easyp.yaml: %w", err)
	}
	if result.Version == "" {
		result.Version = "v1"
	}
	if result.Version != "v1" {
		return Policy{}, fmt.Errorf("easyp.yaml version must be v1")
	}
	if result.Linters.Default == "" {
		result.Linters.Default = "STANDARD"
	}
	switch result.Linters.Default {
	case "MINIMAL", "BASIC", "STANDARD", "COMMENTS":
	default:
		return Policy{}, fmt.Errorf("unknown v1 linter preset %q", result.Linters.Default)
	}
	return result, nil
}

// LintConfig translates the policy into the lint engine configuration.
func (p Policy) LintConfig() (config.LintConfig, error) {
	if p.Linters.Extends != "" {
		return config.LintConfig{}, fmt.Errorf("linters.extends policy loading is not implemented")
	}
	use := []string{"MINIMAL"}
	switch p.Linters.Default {
	case "BASIC":
		use = append(use, "BASIC")
	case "STANDARD":
		use = append(use, "BASIC", "DEFAULT")
	case "COMMENTS":
		use = append(use, "BASIC", "DEFAULT", "COMMENTS")
	}
	use = append(use, p.Linters.Enable...)
	cfg := config.LintConfig{Use: use, Except: p.Linters.Disable}
	for _, rule := range p.Issues.ExcludeRules {
		if rule.Path != "" {
			return config.LintConfig{}, fmt.Errorf("issues.exclude-rules.path glob matching is not implemented")
		}
		cfg.Except = append(cfg.Except, rule.Linters...)
	}
	for rule, settings := range p.LinterSettings {
		switch rule {
		case "ENUM_ZERO_VALUE_SUFFIX":
			cfg.EnumZeroValueSuffix = settings["suffix"]
		case "SERVICE_SUFFIX":
			cfg.ServiceSuffix = settings["suffix"]
		default:
			return config.LintConfig{}, fmt.Errorf("unsupported linters-settings rule %q", rule)
		}
	}
	return cfg, nil
}

// ExcludesAllIssues reports whether a pathless exclusion suppresses every
// linter for all files covered by this policy.
func (p Policy) ExcludesAllIssues() bool {
	for _, rule := range p.Issues.ExcludeRules {
		if rule.Path == "" && len(rule.Linters) == 0 {
			return true
		}
	}
	return false
}

// BreakingConfig maps the v1 baseline onto the existing check set.
// Category filtering requires a rule mapping of its own.
func (p Policy) BreakingConfig(fallbackRef string) (config.BreakingCheck, error) {
	if p.Breaking.Extends != "" {
		return config.BreakingCheck{}, fmt.Errorf("breaking.extends policy loading is not implemented")
	}
	if len(p.Breaking.Categories) > 0 {
		return config.BreakingCheck{}, fmt.Errorf("breaking.categories are not supported by the existing checker")
	}
	if p.Breaking.IgnoreUnstable {
		return config.BreakingCheck{}, fmt.Errorf("breaking.ignore_unstable is not supported by the existing checker")
	}
	baseline := strings.TrimPrefix(p.Breaking.Baseline, "git:")
	if baseline == "" {
		baseline = fallbackRef
	}
	return config.BreakingCheck{AgainstGitRef: baseline, Ignore: p.Breaking.Ignore}, nil
}
