package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*EnumFirstValueZero)(nil)

// EnumFirstValueZero this rule checks that the first value of every enum is numerically zero.
type EnumFirstValueZero struct{}

// Message implements lint.Rule.
func (c *EnumFirstValueZero) Message() string {
	return "enum first value must be zero"
}

// Validate implements lint.Rule.
func (c *EnumFirstValueZero) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Enum: func(enum *protoast.Enum) {
		if len(enum.Values) == 0 {
			return
		}
		first := enum.Values[0]
		if n, ok := enumNumber(first.Number); !ok || n != 0 {
			res = core.AppendIssue(res, c, first.NumberPos, first.Number)
		}
	}})

	return res, nil
}
