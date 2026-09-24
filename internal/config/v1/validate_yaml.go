package v1

import (
	"fmt"
	"strings"
	"sync"

	yamlvalidator "github.com/Yakwilik/go-yamlvalidator"

	"github.com/easyp-tech/easyp/internal/config"
)

var (
	policySchema = sync.OnceValues(func() (*yamlvalidator.FieldSchema, error) {
		return compileSchema("easyp")
	})
	generateSchema = sync.OnceValues(func() (*yamlvalidator.FieldSchema, error) {
		return compileSchema("easyp.gen")
	})
)

// ValidatePolicyYAML reports JSON Schema violations with YAML source coordinates.
func ValidatePolicyYAML(raw []byte) []config.ValidationIssue {
	return validateV1YAML(raw, policySchema)
}

// ValidateGenerateYAML reports JSON Schema violations with YAML source coordinates.
func ValidateGenerateYAML(raw []byte) []config.ValidationIssue {
	return validateV1YAML(raw, generateSchema)
}

func compileSchema(name string) (*yamlvalidator.FieldSchema, error) {
	raw, err := SchemaJSON(name)
	if err != nil {
		return nil, fmt.Errorf("SchemaJSON: %w", err)
	}
	schema, err := yamlvalidator.CompileJSONSchema(raw)
	if err != nil {
		return nil, fmt.Errorf("CompileJSONSchema: %w", err)
	}
	return schema, nil
}

func validateV1YAML(raw []byte, loadSchema func() (*yamlvalidator.FieldSchema, error)) []config.ValidationIssue {
	expanded, err := expandConfigBytes(raw)
	if err != nil {
		return []config.ValidationIssue{{Code: "v1_validation", Message: err.Error(), Severity: config.SeverityError}}
	}
	schema, err := loadSchema()
	if err != nil {
		return []config.ValidationIssue{{Code: "v1_validation", Message: err.Error(), Severity: config.SeverityError}}
	}
	result := yamlvalidator.NewValidator(schema).ValidateWithOptions(expanded, yamlvalidator.ValidationContext{
		StrictKeys:  true,
		StrictTypes: true,
	})
	allIssues := result.Collector.All()
	compositePaths := make(map[string]string)
	for _, issue := range allIssues {
		if issue.Code == "oneOf" || issue.Code == "anyOf" {
			compositePaths[issue.Path] = issue.SchemaPath
		}
	}
	issues := make([]config.ValidationIssue, 0, len(allIssues))
	for _, issue := range allIssues {
		if parentPath, ok := compositePaths[issue.Path]; ok && strings.HasPrefix(issue.SchemaPath, parentPath+"/") {
			continue
		}
		if issue.Code == "oneOf" || issue.Code == "anyOf" {
			// A nested alternative already points to the offending YAML value.
			nested := false
			for path := range compositePaths {
				if strings.HasPrefix(path, issue.Path+".") || strings.HasPrefix(path, issue.Path+"[") {
					nested = true
					break
				}
			}
			if nested {
				continue
			}
		}
		severity := config.SeverityWarn
		if issue.Level == yamlvalidator.LevelError {
			severity = config.SeverityError
		}
		message := issue.Message
		if issue.Expected != "" {
			message += fmt.Sprintf(" (expected %s)", issue.Expected)
		}
		if issue.Got != "" {
			message += fmt.Sprintf(" (got %s)", issue.Got)
		}
		if issue.Path != "" {
			message += fmt.Sprintf(" (path: %s)", issue.Path)
		}
		issues = append(issues, config.ValidationIssue{
			Code: "yaml_validation", Message: message,
			Line: issue.Line, Column: issue.Column, Severity: severity,
		})
	}
	return issues
}
