package v1

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		var module Module
		module, err = ParseModule(bytes.NewReader(raw))
		if err == nil {
			err = validateReplacementDirectories(filepath.Dir(path), module)
		}
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
	if LegacyPolicy(raw) {
		return []config.ValidationIssue{{Code: "yaml_validation", Message: ErrLegacyConfiguration.Error(), Severity: config.SeverityError, Line: 1, Column: 1}}
	}
	issues := ValidatePolicyYAML(raw)
	if config.HasErrors(issues) {
		return issues
	}
	policy, err := ParsePolicy(bytes.NewReader(raw))
	if err != nil {
		return append(issues, validationError(name, err))
	}
	if policy.Linters.Extends == "" {
		if _, err := policy.LintConfig(); err != nil {
			return append(issues, validationError(name, err))
		}
	}
	if policy.Breaking.Extends == "" {
		if _, err := policy.BreakingConfig(""); err != nil {
			return append(issues, validationError(name, err))
		}
	}
	return issues
}

func validationError(name string, err error) config.ValidationIssue {
	return config.ValidationIssue{
		Code: "v1_validation", Message: validationMessage(name, err), Severity: config.SeverityError,
	}
}

func validationMessage(name string, err error) string {
	message := err.Error()
	if strings.HasPrefix(message, name+":") {
		return message
	}
	return fmt.Sprintf("%s: %s", name, message)
}

func validateReplacementDirectories(directory string, module Module) error {
	for _, replacement := range module.Replaces {
		target := replacement.Target
		if !filepath.IsAbs(target) {
			target = filepath.Join(directory, target)
		}
		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("replace %s => %s: %w", replacement.Module, replacement.Target, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("replace %s => %s: target is not a directory", replacement.Module, replacement.Target)
		}
	}
	return nil
}
