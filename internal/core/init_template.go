package core

import (
	_ "embed"
	"fmt"
	"io"
	"text/template"
)

//go:embed templates/easyp.yaml.tmpl
var initTemplate string

// InitTemplateData contains data for rendering the easyp.yaml template.
type InitTemplateData struct {
	LintGroups          []LintGroup
	EnumZeroValueSuffix string
	ServiceSuffix       string
	AgainstGitRef       string
}

// LintGroup is a group of lint rules with a name and a list of rules.
type LintGroup struct {
	Name  string   // "Minimal", "Basic", "Default", "Comments", "Unary RPC"
	Rules []string // ["DIRECTORY_SAME_PACKAGE", "PACKAGE_DEFINED", ...]
}

// renderInitConfig renders the easyp.yaml template to the given writer.
func renderInitConfig(w io.Writer, data InitTemplateData) error {
	tmpl, err := template.New("easyp.yaml").Parse(initTemplate)
	if err != nil {
		return fmt.Errorf("template.Parse: %w", err)
	}

	return tmpl.Execute(w, data)
}
