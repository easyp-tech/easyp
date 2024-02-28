package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ProtoValidate)(nil)

// ProtoValidate this rule requires that all protovalidate constraints specified are valid.
type ProtoValidate struct {
}

// Validate checks that all protovalidate constraints specified are valid.
func (p ProtoValidate) Validate(protoInfo lint.ProtoInfo) []error {
	//TODO implement me
	panic("implement me")
}
