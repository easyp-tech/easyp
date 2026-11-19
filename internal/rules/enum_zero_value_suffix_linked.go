package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*EnumZeroValueSuffixLinked)(nil)

// EnumZeroValueSuffixLinked is the linked-engine port of EnumZeroValueSuffix
// (research inventory #26, main-diverged; threat T-04-15): main checks ONLY
// the first value and requires EXACT equality with the carried helper of the
// enum name plus "_" and the configured suffix — a later value with a wrong
// suffix is invisible, and a first value that merely ends with the suffix
// still issues. The lab's every-zero-value numbered HasSuffix reading is
// not ported. The first value's position is its first token (main's
// Meta.Pos ≡ Decl.Pos), not the lab's name token.
//
// The Suffix struct field carries main's tags byte-verbatim, INCLUDING the
// ENV tag the lab dropped (config surface parity, T-04-15). The suffix
// defaults to "UNSPECIFIED" through NewLinked's defaultIfEmpty, mirroring
// builder.go.
//
// The len(Values)==0 guard is the same documented crash-safety deviation as
// EnumFirstValueZero (research A7): main indexes the first value unguarded
// and panics on an empty enum; protoc rejects empty enums, so the engine
// path never sees one.
type EnumZeroValueSuffixLinked struct {
	Suffix string `json:"suffix" yaml:"suffix" ENV:"ENUM_ZERO_VALUE_SUFFIX"`
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (e *EnumZeroValueSuffixLinked) LintRuleName() string {
	return "ENUM_ZERO_VALUE_SUFFIX"
}

// Message implements lint.LintRule.
func (e *EnumZeroValueSuffixLinked) Message() string {
	return "enum zero value suffix is not valid"
}

// Validate implements lint.LintRule.
func (e *EnumZeroValueSuffixLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range info.File.Enums {
		// Documented deviation (research A7): main panics on an empty enum
		// (unguarded Values[0]); the guard keeps the port total — same
		// crash-safety note as EnumFirstValueZeroLinked.
		if len(enum.Values) == 0 {
			continue
		}
		zeroValue := enum.Values[0]
		if zeroValue.Name != pascalToUpperSnakeLinked(enum.Name)+"_"+e.Suffix {
			res = core.AppendLinkedIssue(res, e, zeroValue.Pos, zeroValue.Name)
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			if len(enum.Values) == 0 {
				continue
			}
			zeroValue := enum.Values[0]
			if zeroValue.Name != pascalToUpperSnakeLinked(enum.Name)+"_"+e.Suffix {
				res = core.AppendLinkedIssue(res, e, zeroValue.Pos, zeroValue.Name)
			}
		}
	}

	return res, nil
}
