package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*OneofLowerSnakeCase)(nil)

// OneofLowerSnakeCase this rule checks that oneof names are lower_snake_case.
type OneofLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *OneofLowerSnakeCase) Message() string {
	return "oneof name should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *OneofLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Oneof: func(oneof *protoast.Oneof) {
		if oneof.Name != toLowerSnakeCase(oneof.Name) {
			res = core.AppendIssue(res, c, oneof.NamePos, oneof.Name)
		}
	}})

	return res, nil
}
