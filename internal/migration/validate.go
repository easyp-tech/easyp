package migration

import (
	"bytes"
	"fmt"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func validatePolicy(raw []byte) error {
	var policy v1.Policy
	if err := decodeStrict(raw, &policy); err != nil {
		return fmt.Errorf("decodeStrict: %w", err)
	}
	if policy.Version != "" && policy.Version != "v1" {
		return fmt.Errorf("unsupported policy version %q", policy.Version)
	}
	switch policy.Linters.Default {
	case "", "MINIMAL", "BASIC", "STANDARD", "COMMENTS":
	default:
		return fmt.Errorf("unknown v1 linter preset %q", policy.Linters.Default)
	}
	if _, err := policy.LintConfig(); err != nil {
		return fmt.Errorf("LintConfig: %w", err)
	}
	if _, err := policy.BreakingConfig(""); err != nil {
		return fmt.Errorf("BreakingConfig: %w", err)
	}
	if policy.Breaking.Baseline != "" && (!strings.HasPrefix(policy.Breaking.Baseline, "git:") || policy.Breaking.Baseline == "git:") {
		return fmt.Errorf("breaking.baseline requires git:<ref>")
	}
	for _, settings := range policy.LinterSettings {
		for key := range settings {
			if key != "suffix" {
				return fmt.Errorf("unknown linters-settings field %q", key)
			}
		}
	}
	return nil
}

func validateGenerate(raw []byte) error {
	var gen v1.Generate
	if err := decodeStrict(raw, &gen); err != nil {
		return fmt.Errorf("decodeStrict: %w", err)
	}
	if gen.Version != "" && gen.Version != "v1" {
		return fmt.Errorf("unsupported generation version %q", gen.Version)
	}
	for i, module := range gen.Generate.Modules {
		if err := module.Validate(); err != nil {
			return fmt.Errorf("generate.modules[%d]: %w", i, err)
		}
	}
	if err := v1.ValidatePackageSelectors(gen.Generate.Packages); err != nil {
		return fmt.Errorf("ValidatePackageSelectors: %w", err)
	}
	if err := v1.ValidatePathSelectors(gen.Generate.Paths); err != nil {
		return fmt.Errorf("ValidatePathSelectors: %w", err)
	}
	if err := gen.Generate.Managed.Validate(); err != nil {
		return fmt.Errorf("Validate: %w", err)
	}
	for _, plugin := range gen.Plugins {
		if err := plugin.Validate(); err != nil {
			return fmt.Errorf("Validate: %w", err)
		}
	}
	return nil
}

func validateNativeLock(raw []byte) (v1.Lock, error) {
	lock, err := v1.ParseLock(bytes.NewReader(raw))
	if err != nil {
		return v1.Lock{}, fmt.Errorf("ParseLock: %w", err)
	}
	return lock, nil
}

func checkManagedLocal(cfg legacyConfig, identity string, hasLocal bool) error {
	if !hasLocal || !cfg.Generate.Managed.Enabled {
		return nil
	}
	var selectors []string
	for _, rule := range cfg.Generate.Managed.Disable {
		selectors = append(selectors, rule.Module)
	}
	for _, rule := range cfg.Generate.Managed.Override {
		selectors = append(selectors, rule.Module)
	}
	for _, selector := range selectors {
		if selector == identity || strings.Contains(selector, "$") {
			return fmt.Errorf("managed module selector %q may newly match local files named %s in v1 (v0 used an empty local identity); migrate this selector manually", selector, identity)
		}
	}
	return nil
}
