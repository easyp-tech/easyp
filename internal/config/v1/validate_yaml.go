package v1

import (
	"fmt"
	"strings"
	"sync"

	yamlvalidator "github.com/Yakwilik/go-yamlvalidator"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/rules"
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
	return validateExpandedV1YAML(expanded, loadSchema)
}

// validateExpandedV1YAML also serves parsing, which has already substituted the
// environment. Expanding again here would interpret escaped $$ expressions.
func validateExpandedV1YAML(expanded []byte, loadSchema func() (*yamlvalidator.FieldSchema, error)) []config.ValidationIssue {
	schema, err := loadSchema()
	if err != nil {
		return []config.ValidationIssue{{Code: "v1_validation", Message: err.Error(), Severity: config.SeverityError}}
	}
	result := yamlvalidator.NewValidator(schema).ValidateWithOptions(expanded, yamlvalidator.ValidationContext{
		StrictKeys:  true,
		StrictTypes: true,
	})
	result.SortByPosition()
	allIssues := result.Collector.All()
	values := scalarValuesByPosition(expanded)
	compositePaths := make(map[string]string)
	for _, issue := range allIssues {
		if issue.Code == "oneOf" || issue.Code == "anyOf" {
			compositePaths[issue.Path] = issue.SchemaPath
		}
	}
	issues := make([]config.ValidationIssue, 0, len(allIssues))
	for _, issue := range allIssues {
		if issue.Code == "group" || issue.Code == "allOf" {
			// The schema engine can emit a summary together with precise child
			// errors. Keep the actionable diagnostics, not the duplicate summary.
			detailed := false
			for _, child := range allIssues {
				if child.Code != "group" && child.Level == issue.Level &&
					(child.Path == issue.Path || strings.HasPrefix(child.Path, issue.Path+".") || strings.HasPrefix(child.Path, issue.Path+"[")) &&
					strings.HasPrefix(child.SchemaPath, issue.SchemaPath+"/") {
					detailed = true
					break
				}
			}
			if detailed {
				continue
			}
		}
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
		if issue.Code == "enum" && isLintRuleValuePath(issue.Path) {
			if value, ok := values[[2]int{issue.Line, issue.Column}]; ok {
				if err := rules.ValidateNames([]string{value}); err != nil {
					message = err.Error()
				}
			}
		}
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

func isLintRuleValuePath(path string) bool {
	return strings.HasPrefix(path, "linters.enable[") || strings.HasPrefix(path, "linters.disable[") ||
		(strings.HasPrefix(path, "issues[\"exclude-rules\"][") && strings.Contains(path, "].linters["))
}

// scalarValuesByPosition supplements schema errors which omit the rejected
// value. Syntax errors are already reported by the schema validator.
func scalarValuesByPosition(raw []byte) map[[2]int]string {
	values := make(map[[2]int]string)
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return values
	}
	var visit func(*yaml.Node)
	visit = func(node *yaml.Node) {
		if node.Kind == yaml.ScalarNode {
			values[[2]int{node.Line, node.Column}] = node.Value
		}
		for _, child := range node.Content {
			visit(child)
		}
	}
	visit(&root)
	return values
}
