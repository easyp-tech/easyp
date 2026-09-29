package v1

import (
	"errors"
	"fmt"

	"github.com/easyp-tech/easyp/internal/rules"
)

func (p Policy) validateLintSelections() error {
	var issues []error
	for _, selection := range []struct {
		path  string
		names []string
	}{
		{path: "linters.enable", names: p.Linters.Enable},
		{path: "linters.disable", names: p.Linters.Disable},
	} {
		err := rules.ValidateNames(selection.names)
		if err != nil {
			issues = append(issues, fmt.Errorf("%s: %w", selection.path, err))
		}
	}
	for i, exclusion := range p.Issues.ExcludeRules {
		err := rules.ValidateNames(exclusion.Linters)
		if err != nil {
			issues = append(issues, fmt.Errorf("issues.exclude-rules[%d].linters: %w", i, err))
		}
	}
	return errors.Join(issues...)
}
