package rules

import (
	"reflect"
	"regexp"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumPascalCase)(nil)

// EnumPascalCase this rule checks that enums are PascalCase.
type EnumPascalCase struct{}

// Name implements lint.Rule.
func (c *EnumPascalCase) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Validate implements lint.Rule.
func (c *EnumPascalCase) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	pascalCase := regexp.MustCompile("^[A-Z][a-z]+(?:[A-Z][a-z]+)*$")
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if !pascalCase.MatchString(enum.EnumName) {
			res = append(res, BuildError(protoInfo.Path, enum.Meta.Pos, enum.EnumName, lint.ErrEnumPascalCase))
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
