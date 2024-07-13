package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*EnumFirstValueZero)(nil)

// EnumFirstValueZero this rule enforces that the first enum value is the zero value,
// which is a proto3 requirement on build,
// but isn't required in proto2 on build. The rule enforces that the requirement is also followed in proto2.
type EnumFirstValueZero struct{}

// Message implements lint.Rule.
func (c *EnumFirstValueZero) Message() string {
	return "enum first value must be zero"
}

// Validate implements lint.Rule.
func (c *EnumFirstValueZero) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue
	for _, enum := range protoInfo.Info.ProtoBody.Enums {
		if val := enum.EnumBody.EnumFields[0]; val.Number != "0" {
			res = lint.AppendIssue(res, c, val.Meta.Pos, val.Number, val.Comments)
		}

	}

	return res, nil
}
