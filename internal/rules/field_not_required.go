package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*FieldNotRequired)(nil)

// FieldNotRequired this rule checks that fields are not required.
type FieldNotRequired struct{}

// Message implements lint.Rule.
func (f *FieldNotRequired) Message() string {
	return "field should not be required"
}

// Validate implements lint.Rule.
func (f *FieldNotRequired) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Field: func(field *protoast.Field) {
		if field.Label == "required" {
			res = core.AppendIssue(res, f, field.NamePos, field.Name)
		}
	}})

	return res, nil
}
