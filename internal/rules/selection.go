package rules

import (
	"errors"
	"fmt"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"slices"
)

// ValidateNames checks every rule/group, including names that will be disabled.
// The executable registry is also the source for schemas and completion values.
func ValidateNames(names []string) error {
	allowed := AllLintUseValues()
	var issues []error
	for _, name := range names {
		if name == "PACKAGE_NO_IMPORT_CYCLE" {
			issues = append(issues, fmt.Errorf("%w: %s is not implemented", core.ErrInvalidRule, name))
			continue
		}
		if !slices.Contains(allowed, name) {
			issues = append(issues, fmt.Errorf("%w: %s", core.ErrInvalidRule, name))
		}
	}
	return errors.Join(issues...)
}

func validateConfigRuleNames(cfg config.LintConfig) error {
	var issues []error
	for _, selection := range []struct {
		path  string
		names []string
	}{
		{path: "use", names: cfg.Use}, {path: "except", names: cfg.Except},
	} {
		err := ValidateNames(selection.names)
		if err != nil {
			issues = append(issues, fmt.Errorf("%s: %w", selection.path, err))
		}
	}
	keys := make([]string, 0, len(cfg.IgnoreOnly))
	for name := range cfg.IgnoreOnly {
		keys = append(keys, name)
	}
	slices.Sort(keys)
	err := ValidateNames(keys)
	if err != nil {
		issues = append(issues, fmt.Errorf("ignore_only: %w", err))
	}
	return errors.Join(issues...)
}
