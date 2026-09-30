package v1

import (
	"fmt"
	"slices"
	"strings"
)

// validateSemantics checks the entire policy even when the caller only uses
// lint or breaking settings. It also covers policies assembled without YAML.
func (p Policy) validateSemantics() error {
	if p.Version != "" && p.Version != "v1" {
		return fmt.Errorf("easyp.yaml version must be v1")
	}
	switch p.Linters.Default {
	case "", "MINIMAL", "BASIC", "STANDARD", "COMMENTS":
	default:
		return fmt.Errorf("unknown v1 linter preset %q", p.Linters.Default)
	}
	if err := p.validateLintSelections(); err != nil {
		return fmt.Errorf("validateLintSelections: %w", err)
	}
	if err := p.Issues.validatePaths(); err != nil {
		return fmt.Errorf("validatePaths: %w", err)
	}
	if p.Linters.Extends != "" {
		return fmt.Errorf("linters.extends policy loading is not implemented")
	}
	if p.Breaking.Extends != "" {
		return fmt.Errorf("breaking.extends policy loading is not implemented")
	}
	for _, category := range p.Breaking.Categories {
		if category != breakingCategoryFile {
			return fmt.Errorf("breaking.categories: unsupported category %q; supported: FILE", category)
		}
	}
	baseline := p.Breaking.Baseline
	if baseline != "" && (!strings.HasPrefix(baseline, "git:") || len(baseline) == len("git:") || strings.ContainsAny(baseline, "\r\n")) {
		return fmt.Errorf("breaking.baseline must be empty or git:<ref>")
	}
	// Sort map keys so multiple invalid settings produce a deterministic error.
	rules := make([]string, 0, len(p.LinterSettings))
	for rule := range p.LinterSettings {
		rules = append(rules, rule)
	}
	slices.Sort(rules)
	for _, rule := range rules {
		switch rule {
		case "ENUM_ZERO_VALUE_SUFFIX", "SERVICE_SUFFIX":
		default:
			return fmt.Errorf("unsupported linters-settings rule %q", rule)
		}
		fields := make([]string, 0, len(p.LinterSettings[rule]))
		for field := range p.LinterSettings[rule] {
			fields = append(fields, field)
		}
		slices.Sort(fields)
		for _, field := range fields {
			if field != "suffix" {
				return fmt.Errorf("linters-settings.%s: unsupported setting %q", rule, field)
			}
		}
	}
	return nil
}
