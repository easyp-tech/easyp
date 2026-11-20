package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumFirstValueZeroLinked)(nil)

// EnumFirstValueZeroLinked is the linked-engine port of EnumFirstValueZero
// (research inventory #5, main-diverged with one documented deviation): main's
// STRING compare (`Number != "0"`) wins over the lab's numeric
// parse-and-compare — the octal "00" and hex "0x0" spellings of zero issue
// under main and would pass the lab's enumNumber read. The position is the
// value's first token (main's Meta.Pos ≡ Decl.Pos), not the lab's NumberPos.
// The walk is main's two-level shape: top-level enums plus enums of TOP-LEVEL
// messages; an enum inside a nested message is outside main's predicate.
//
// Documented deviation (research A7, threat T-04-10): main indexes the first
// value unguarded and panics on an empty enum; the len(Values)==0 guard is
// kept as the ONE deliberate crash-fix deviation. It is unreachable through
// the engine (protoc rejects empty enums, so the lint path sees only files
// that compile) and is unit-pinned in enum_first_value_zero_test.go.
type EnumFirstValueZeroLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *EnumFirstValueZeroLinked) LintRuleName() string {
	return "ENUM_FIRST_VALUE_ZERO"
}

// Message implements lint.Rule.
func (c *EnumFirstValueZeroLinked) Message() string {
	return "enum first value must be zero"
}

// Validate implements lint.Rule.
func (c *EnumFirstValueZeroLinked) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range info.File.Enums {
		// Documented deviation (research A7): main panics on an empty enum
		// (unguarded Values[0]); the guard keeps the port total. protoc rejects
		// empty enums, so the engine path never sees one.
		if len(enum.Values) == 0 {
			continue
		}
		val := enum.Values[0]
		if val.Number != "0" {
			// val.Pos is the value's first token — the same position main
			// reported as Meta.Pos (pinned by TestDeclPositionProbe).
			res = core.AppendIssue(res, c, val.Pos, val.Number)
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			if len(enum.Values) == 0 {
				continue
			}
			val := enum.Values[0]
			if val.Number != "0" {
				res = core.AppendIssue(res, c, val.Pos, val.Number)
			}
		}
	}

	return res, nil
}
