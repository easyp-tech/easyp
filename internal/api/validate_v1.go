package api

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

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
	switch name {
	case "easyp.gen.yaml":
		issues := v1.ValidateGenerateYAML(raw)
		if config.HasErrors(issues) {
			return issues, nil
		}
		_, err := v1.ParseGenerate(bytes.NewReader(raw))
		if err != nil {
			return append(issues, v1ValidationError(name, err)), nil
		}
		return issues, nil
	case "protobuf.mod":
		_, err = v1.ParseModule(bytes.NewReader(raw))
	case "protobuf.lock":
		_, err = v1.ParseLock(bytes.NewReader(raw))
	default:
		issues := v1.ValidatePolicyYAML(raw)
		if config.HasErrors(issues) {
			return issues, nil
		}
		policy, err := v1.ParsePolicy(bytes.NewReader(raw))
		if err != nil {
			return append(issues, v1ValidationError(name, err)), nil
		}
		if _, err := policy.LintConfig(); err != nil {
			return append(issues, v1ValidationError(name, err)), nil
		}
		if _, err := policy.BreakingConfig(""); err != nil {
			return append(issues, v1ValidationError(name, err)), nil
		}
		return issues, nil
	}
	if err != nil {
		return []config.ValidationIssue{v1ValidationError(name, err)}, nil
	}
	return nil, nil
}

func v1ValidationError(name string, err error) config.ValidationIssue {
	return config.ValidationIssue{
		Code: "v1_validation", Message: fmt.Sprintf("%s: %v", name, err), Severity: config.SeverityError,
	}
}
