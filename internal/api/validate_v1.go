package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// validateConfigFile chooses the v1 parser for the named configuration file.
func validateConfigFile(path string) ([]config.ValidationIssue, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	var validationErr error
	switch name {
	case "easyp.gen.yaml":
		gen, err := v1.ParseGenerate(strings.NewReader(string(raw)))
		if err != nil {
			validationErr = err
		} else {
			validationErr = gen.Generate.Managed.Validate()
		}
	case "protobuf.mod":
		_, validationErr = v1.ParseModule(strings.NewReader(string(raw)))
	case "protobuf.lock":
		_, validationErr = v1.ParseLock(strings.NewReader(string(raw)))
	default:
		policy, err := v1.ParsePolicy(strings.NewReader(string(raw)))
		if err != nil {
			validationErr = err
		} else if _, err := policy.LintConfig(); err != nil {
			validationErr = err
		} else if _, err := policy.BreakingConfig(""); err != nil {
			validationErr = err
		}
	}
	if validationErr != nil {
		return []config.ValidationIssue{{
			Code: "v1_validation", Message: fmt.Sprintf("%s: %v", name, validationErr), Severity: config.SeverityError,
		}}, nil
	}
	return nil, nil
}
