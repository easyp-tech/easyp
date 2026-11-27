package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*EnumZeroValueSuffix)(nil)

// EnumZeroValueSuffix this rule checks that every enum value numbered zero ends with the configured suffix.
type EnumZeroValueSuffix struct {
	// Suffix is the required ending, with or without its leading underscore.
	Suffix string
}

// Message implements lint.Rule.
func (e *EnumZeroValueSuffix) Message() string {
	return "enum zero value suffix is not valid"
}

// Validate implements lint.Rule.
func (e *EnumZeroValueSuffix) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	suffix := e.Suffix
	if !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}
	protoInfo.File.Walk(protoast.Visitor{EnumValue: func(_ *protoast.Enum, value *protoast.EnumValue) {
		if n, ok := enumNumber(value.Number); ok && n == 0 && !strings.HasSuffix(value.Name, suffix) {
			res = core.AppendIssue(res, e, value.NamePos, value.Name)
		}
	}})

	return res, nil
}
