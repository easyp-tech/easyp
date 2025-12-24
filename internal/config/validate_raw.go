package config

import (
	"fmt"
	"os"

	v "github.com/Yakwilik/go-yamlvalidator"
	"github.com/a8m/envsubst"
)

// ValidationIssue describes a single validation problem with the config.
type ValidationIssue struct {
	Code     string `json:"code" yaml:"code"`
	Message  string `json:"message" yaml:"message"`
	Line     int    `json:"line,omitempty" yaml:"line,omitempty"`
	Column   int    `json:"column,omitempty" yaml:"column,omitempty"`
	Severity string `json:"severity,omitempty" yaml:"severity,omitempty"` // "error" or "warn"
}

// ValidateRaw validates config bytes using go-yamlvalidator schema.
// Env vars are expanded before validation. Unknown keys produce warnings; type/structure errors produce errors.
func ValidateRaw(buf []byte) ([]ValidationIssue, error) {
	issues := make([]ValidationIssue, 0)

	expanded, err := envsubst.String(string(buf))
	if err != nil {
		issues = append(issues, newIssue("envsubst_error", err.Error(), "error"))
		return issues, nil
	}
	buf = []byte(expanded)

	validator := v.NewValidator(buildSchema())
	ctx := v.ValidationContext{
		StrictKeys:     true,
		YAML11Booleans: true,
	}
	result := validator.ValidateWithOptions(buf, ctx)

	for _, e := range result.Collector.All() {
		severity := "warn"
		if e.Level == v.LevelError {
			severity = "error"
		}

		msg := e.Message
		if e.Expected != "" || e.Got != "" {
			msg = fmt.Sprintf("%s (expected %s, got %s)", e.Message, e.Expected, e.Got)
		}
		if e.Path != "" {
			msg = fmt.Sprintf("%s (path: %s)", msg, e.Path)
		}

		issues = append(issues, ValidationIssue{
			Code:     "yaml_validation",
			Message:  msg,
			Line:     e.Line,
			Column:   e.Column,
			Severity: severity,
		})
	}

	return issues, nil
}

// ValidateFile validates config file on disk.
func ValidateFile(path string) ([]ValidationIssue, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}
	return ValidateRaw(buf)
}

// buildSchema builds the YAML validation schema matching easyp.yaml structure.
func buildSchema() *v.FieldSchema {
	stringSeq := &v.FieldSchema{Type: v.TypeSequence, ItemSchema: &v.FieldSchema{Type: v.TypeString}}
	anyMap := &v.FieldSchema{Type: v.TypeMap, AdditionalProperties: &v.FieldSchema{Type: v.TypeAny}}

	lintSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"use":                    stringSeq,
			"enum_zero_value_suffix": {Type: v.TypeString},
			"service_suffix":         {Type: v.TypeString},
			"ignore":                 stringSeq,
			"except":                 stringSeq,
			"allow_comment_ignores":  {Type: v.TypeBool},
			"ignore_only": {
				Type:                 v.TypeMap,
				AdditionalProperties: stringSeq,
			},
		},
		UnknownKeyPolicy: v.UnknownKeyWarn,
	}

	depsSchema := &v.FieldSchema{Type: v.TypeSequence, ItemSchema: &v.FieldSchema{Type: v.TypeString}, Nullable: true}

	inputDirSchema := &v.FieldSchema{
		Type: v.TypeAny, // string or map
		AllowedKeys: map[string]*v.FieldSchema{
			"path": {Type: v.TypeString},
			"root": {Type: v.TypeString},
		},
		UnknownKeyPolicy: v.UnknownKeyWarn,
		Validators:       []v.ValueValidator{directoryValidator{}},
	}

	inputGitSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"url":           {Type: v.TypeString, Required: true},
			"sub_directory": {Type: v.TypeString},
			"out":           {Type: v.TypeString},
			"root":          {Type: v.TypeString},
		},
		UnknownKeyPolicy: v.UnknownKeyWarn,
	}

	inputSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"directory": inputDirSchema,
			"git_repo":  inputGitSchema,
		},
		AnyOf:             [][]string{{"directory"}, {"git_repo"}},
		MutuallyExclusive: []string{"directory", "git_repo"},
		UnknownKeyPolicy:  v.UnknownKeyWarn,
	}

	pluginSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"name":         {Type: v.TypeString},
			"remote":       {Type: v.TypeString},
			"path":         {Type: v.TypeString},
			"command":      {Type: v.TypeSequence, ItemSchema: &v.FieldSchema{Type: v.TypeString}},
			"out":          {Type: v.TypeString},
			"opts":         anyMap,
			"with_imports": {Type: v.TypeBool},
		},
		AnyOf:             [][]string{{"name"}, {"remote"}, {"path"}, {"command"}},
		MutuallyExclusive: []string{"name", "remote", "path", "command"},
		UnknownKeyPolicy:  v.UnknownKeyWarn,
		Validators:        []v.ValueValidator{pluginSourceValidator{}},
	}

	managedDisableSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"module":       {Type: v.TypeString},
			"path":         {Type: v.TypeString},
			"file_option":  {Type: v.TypeString},
			"field_option": {Type: v.TypeString},
			"field":        {Type: v.TypeString},
		},
		AnyOf:             [][]string{{"module"}, {"path"}, {"file_option"}, {"field_option"}, {"field"}},
		MutuallyExclusive: []string{"file_option", "field_option"},
		UnknownKeyPolicy:  v.UnknownKeyWarn,
		Validators:        []v.ValueValidator{managedDisableValidator{}},
	}

	managedOverrideSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"file_option":  {Type: v.TypeString},
			"field_option": {Type: v.TypeString},
			"value":        {Type: v.TypeAny, Required: true},
			"module":       {Type: v.TypeString},
			"path":         {Type: v.TypeString},
			"field":        {Type: v.TypeString},
		},
		AnyOf:             [][]string{{"file_option"}, {"field_option"}},
		MutuallyExclusive: []string{"file_option", "field_option"},
		UnknownKeyPolicy:  v.UnknownKeyWarn,
		Validators:        []v.ValueValidator{managedOverrideValidator{}},
	}

	managedSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"enabled":  {Type: v.TypeBool},
			"disable":  {Type: v.TypeSequence, ItemSchema: managedDisableSchema},
			"override": {Type: v.TypeSequence, ItemSchema: managedOverrideSchema},
		},
		UnknownKeyPolicy: v.UnknownKeyWarn,
	}

	generateSchema := &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"inputs":  {Type: v.TypeSequence, ItemSchema: inputSchema, Required: true, MinItems: v.Ptr[int](1)},
			"plugins": {Type: v.TypeSequence, ItemSchema: pluginSchema, Required: true, MinItems: v.Ptr[int](1)},
			"managed": managedSchema,
		},
		UnknownKeyPolicy: v.UnknownKeyWarn,
	}

	breakingSchema := &v.FieldSchema{Type: v.TypeMap, UnknownKeyPolicy: v.UnknownKeyIgnore}

	return &v.FieldSchema{
		Type: v.TypeMap,
		AllowedKeys: map[string]*v.FieldSchema{
			"lint":     lintSchema,
			"deps":     depsSchema,
			"generate": generateSchema,
			"breaking": breakingSchema,
			"version":  {Type: v.TypeString},
		},
		Required:         true,
		UnknownKeyPolicy: v.UnknownKeyWarn,
	}
}

// newIssue constructs ValidationIssue with a severity.
func newIssue(code, msg, severity string) ValidationIssue {
	return ValidationIssue{
		Code:     code,
		Message:  msg,
		Severity: severity,
	}
}
