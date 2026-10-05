package v1

import (
	"bytes"
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
	Default             string   `yaml:"default"`
	Enable              []string `yaml:"enable"`
	Disable             []string `yaml:"disable"`
	Extends             string   `yaml:"extends"`
	AllowCommentIgnores *bool    `yaml:"allow_comment_ignores"`
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

// ParsePolicy reads and validates a consumer-owned v1 lint and breaking policy.
// Environment placeholders are expanded before validation.
func ParsePolicy(r io.Reader) (Policy, error) {
	raw, err := expandConfigYAML(r)
	if err != nil {
		return Policy{}, fmt.Errorf("expandConfigYAML: %w", err)
	}
	return parsePolicyBytes(raw)
}

// ParsePolicyLiteral reads and validates a v1 policy without expanding
// environment placeholders. Dependency-owned shared policies use this form so
// their contents cannot read values from the consumer process environment.
func ParsePolicyLiteral(r io.Reader) (Policy, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return Policy{}, fmt.Errorf("ReadAll: %w", err)
	}
	return parsePolicyBytes(raw)
}

func parsePolicyBytes(raw []byte) (Policy, error) {
	if LegacyPolicy(raw) {
		return Policy{}, ErrLegacyConfiguration
	}
	var result Policy
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return Policy{}, fmt.Errorf("decode easyp.yaml: %w", err)
	}
	if err := requireSingleYAMLDocument(decoder, PolicyFile); err != nil {
		return Policy{}, err
	}
	if result.Version == "" {
		result.Version = "v1"
	}
	if result.Linters.Default == "" {
		result.Linters.Default = "STANDARD"
	}
	if err := result.validateSemantics(); err != nil {
		return Policy{}, err
	}
	if err := yamlValidationError(PolicyFile, validateExpandedV1YAML(raw, policySchema)); err != nil {
		return Policy{}, err
	}
	return result, nil
}

// LintConfig translates the policy into the lint engine configuration.
func (p Policy) LintConfig() (config.LintConfig, error) {
	if p.Linters.Extends != "" {
		return config.LintConfig{}, fmt.Errorf("linters.extends %q was not resolved", p.Linters.Extends)
	}
	if err := p.validateSemantics(); err != nil {
		return config.LintConfig{}, err
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
	cfg := config.LintConfig{Use: use, Except: append([]string(nil), p.Linters.Disable...), AllowCommentIgnores: true}
	if p.Linters.AllowCommentIgnores != nil {
		cfg.AllowCommentIgnores = *p.Linters.AllowCommentIgnores
	}
	for _, rule := range p.Issues.ExcludeRules {
		if rule.Path == "" {
			cfg.Except = append(cfg.Except, rule.Linters...)
		}
	}
	for rule, settings := range p.LinterSettings {
		switch rule {
		case "ENUM_ZERO_VALUE_SUFFIX":
			cfg.EnumZeroValueSuffix = settings["suffix"]
		case "SERVICE_SUFFIX":
			cfg.ServiceSuffix = settings["suffix"]
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

// BreakingConfig maps the v1 baseline and selected compatibility profiles onto the checker.
func (p Policy) BreakingConfig(fallbackRef string) (config.BreakingCheck, error) {
	if p.Breaking.Extends != "" {
		return config.BreakingCheck{}, fmt.Errorf("breaking.extends %q was not resolved", p.Breaking.Extends)
	}
	if err := p.validateSemantics(); err != nil {
		return config.BreakingCheck{}, err
	}
	baseline := strings.TrimPrefix(p.Breaking.Baseline, "git:")
	if baseline == "" {
		baseline = fallbackRef
	}
	categories := append([]string(nil), p.Breaking.Categories...)
	if len(categories) == 0 {
		categories = []string{"FILE"}
	}
	return config.BreakingCheck{
		AgainstGitRef:  baseline,
		Ignore:         p.Breaking.Ignore,
		Use:            append([]string(nil), categories...),
		Categories:     categories,
		IgnoreUnstable: p.Breaking.IgnoreUnstable,
	}, nil
}
