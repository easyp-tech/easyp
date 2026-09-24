package v1

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/config"
)

// ValidateFile checks the structure and semantics of one configuration file.
func ValidateFile(path string) ([]config.ValidationIssue, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}
	name := filepath.Base(path)
	switch name {
	case GenerateFile:
		return validateGenerate(raw), nil
	case ModuleFile:
		_, err = ParseModule(bytes.NewReader(raw))
	case LockFile:
		_, err = ParseLock(bytes.NewReader(raw))
	default:
		return validatePolicy(raw, name), nil
	}
	if err != nil {
		return []config.ValidationIssue{validationError(name, err)}, nil
	}
	return nil, nil
}

func validateGenerate(raw []byte) []config.ValidationIssue {
	issues := ValidateGenerateYAML(raw)
	if config.HasErrors(issues) {
		return issues
	}
	if _, err := ParseGenerate(bytes.NewReader(raw)); err != nil {
		return append(issues, validationError(GenerateFile, err))
	}
	return issues
}

func validatePolicy(raw []byte, name string) []config.ValidationIssue {
	issues := ValidatePolicyYAML(raw)
	if config.HasErrors(issues) {
		return issues
	}
	policy, err := ParsePolicy(bytes.NewReader(raw))
	if err != nil {
		return append(issues, validationError(name, err))
	}
	if _, err := policy.LintConfig(); err != nil {
		return append(issues, validationError(name, err))
	}
	if _, err := policy.BreakingConfig(""); err != nil {
		return append(issues, validationError(name, err))
	}
	return issues
}

func validationError(name string, err error) config.ValidationIssue {
	return config.ValidationIssue{
		Code: "v1_validation", Message: fmt.Sprintf("%s: %v", name, err), Severity: config.SeverityError,
	}
}
