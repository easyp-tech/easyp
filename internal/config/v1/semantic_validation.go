package v1

import (
	"errors"
	"fmt"

	"github.com/easyp-tech/easyp/internal/config"
)

// yamlValidationError keeps the same schema checks at the parsing boundary,
// including value types which yaml.v3 can otherwise coerce into Go strings.
func yamlValidationError(name string, issues []config.ValidationIssue) error {
	var errs []error
	for _, issue := range issues {
		if issue.Severity != config.SeverityError {
			continue
		}
		if issue.Line > 0 {
			errs = append(errs, fmt.Errorf("%s:%d:%d: %s", name, issue.Line, issue.Column, issue.Message))
		} else {
			errs = append(errs, fmt.Errorf("%s: %s", name, issue.Message))
		}
	}
	return errors.Join(errs...)
}
