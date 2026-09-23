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
	Version string `yaml:"version"`
	Linters struct {
		Default string   `yaml:"default"`
		Enable  []string `yaml:"enable"`
		Disable []string `yaml:"disable"`
		Extends string   `yaml:"extends"`
	} `yaml:"linters"`
	LinterSettings map[string]map[string]string `yaml:"linters-settings"`
	Issues         struct {
		ExcludeRules []struct {
			Path    string   `yaml:"path"`
			Linters []string `yaml:"linters"`
		} `yaml:"exclude-rules"`
	} `yaml:"issues"`
	Breaking struct {
		Baseline       string   `yaml:"baseline"`
		Categories     []string `yaml:"categories"`
		IgnoreUnstable bool     `yaml:"ignore_unstable"`
		Extends        string   `yaml:"extends"`
		Ignore         []string `yaml:"ignore"`
	} `yaml:"breaking"`
}

// IsPolicyConfig identifies v1 policies by an explicit version or fields unique
// to v1. Ambiguous versionless configurations retain legacy interpretation.
func IsPolicyConfig(raw []byte) (bool, error) {
	var fields map[string]yaml.Node
	err := yaml.Unmarshal(raw, &fields)
	if err != nil {
		return false, fmt.Errorf("Unmarshal: %w", err)
	}
	if version, ok := fields["version"]; ok {
		var value string
		err := version.Decode(&value)
		if err != nil {
			return false, fmt.Errorf("Decode: %w", err)
		}
		return value == "v1", nil
	}
	for _, name := range []string{"linters", "linters-settings", "issues"} {
		if _, ok := fields[name]; ok {
			return true, nil
		}
	}
	if breaking, ok := fields["breaking"]; ok {
		var breakingFields map[string]yaml.Node
		err := breaking.Decode(&breakingFields)
		if err != nil {
			return false, fmt.Errorf("Decode: %w", err)
		}
		for _, name := range []string{"baseline", "categories", "ignore_unstable", "extends"} {
			if _, ok := breakingFields[name]; ok {
				return true, nil
			}
		}
	}
	return false, nil
}

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

func (p Policy) LegacyLint() (config.LintConfig, error) {
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

// LegacyBreaking retains the existing check set while accepting the v1
// baseline spelling. Category filtering requires a rule mapping of its own.
func (p Policy) LegacyBreaking(fallbackRef string) (config.BreakingCheck, error) {
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
