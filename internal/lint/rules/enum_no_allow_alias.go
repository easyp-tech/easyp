package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

// EnumNoAllowAlias this rule checks that enums are PascalCase.
type EnumNoAllowAlias struct{}

// Name implements lint.Rule.
func (e *EnumNoAllowAlias) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(e).Elem().Name())
}

// Message implements lint.Rule.
func (e *EnumNoAllowAlias) Message() string {
	return "enum must not allow alias"
}

// Validate implements lint.Rule.
func (e *EnumNoAllowAlias) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		for _, opt := range enum.EnumBody.Options {
			if opt.OptionName == "allow_alias" {
				res = append(res, lint.BuildError(enum.Meta.Pos, enum.EnumName, e.Message()))
			}
		}
	}

	return res, nil
}
