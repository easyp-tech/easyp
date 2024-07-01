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

// Validate implements lint.Rule.
func (c *EnumFirstValueZero) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if val := enum.EnumBody.EnumFields[0]; val.Number != "0" {
			res = append(res, BuildError(protoInfo.Path, val.Meta.Pos, val.Number, lint.ErrEnumFirstValueZero))
		}

	}

	if len(res) == 0 {
		return nil
	}
	return res
}
