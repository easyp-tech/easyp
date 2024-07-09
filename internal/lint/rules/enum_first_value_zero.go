package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumFirstValueZero)(nil)

// EnumFirstValueZero this rule enforces that the first enum value is the zero value,
// which is a proto3 requirement on build,
// but isn't required in proto2 on build. The rule enforces that the requirement is also followed in proto2.
type EnumFirstValueZero struct{}

// Name implements lint.Rule.
func (c *EnumFirstValueZero) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Message implements lint.Rule.
func (c *EnumFirstValueZero) Message() string {
	return "enum first value must be zero"
}

// Validate implements lint.Rule.
func (c *EnumFirstValueZero) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if val := enum.EnumBody.EnumFields[0]; val.Number != "0" {
			res = append(res, lint.BuildError(val.Meta.Pos, val.Number, c.Message()))
		}

	}

	return res, nil
}
