package rules

import (
	"strings"
	"unicode"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumValuePrefix)(nil)

// EnumValuePrefix
// (research inventory #25, main-diverged; threat T-04-12): the prefix is
// main's private helper carried byte-verbatim below — an underscore before
// EVERY uppercase rune, so "HTTPCode" becomes "H_T_T_P_CODE" — and the
// check is a bare strings.HasPrefix WITHOUT a trailing underscore, so a
// value equal to the enum name passes. The lab's acronym-aware upper-snake
// helper (which passes HTTPCODE_OK on enum HTTPCode and fails a name-only
// value because it demands a trailing underscore) is not portable —
// casing.go helpers are forbidden in this port.
//
// The walk is main's two-level shape: top-level enums plus enums of
// TOP-LEVEL messages (direct field iteration, not the recursive Visitor —
// an enum inside a nested message is outside main's predicate). Positions
// are the value's first token (main's Meta.Pos ≡ Decl.Pos), never the lab's
// name token.
//
// pascalToUpperSnake is main's private helper carried byte-verbatim
// (body identical; the name carries the Linked suffix because the frozen old
// file enum_value_prefix.go still declares the original in this package —
// Phase 5 drops the suffix when it deletes the old types; rename map:
// pascalToUpperSnake -> pascalToUpperSnake).
type EnumValuePrefix struct{}

// Message implements lint.Rule.
func (e *EnumValuePrefix) Message() string {
	return "enum value prefix is not valid"
}

// Validate implements lint.Rule.
func (e *EnumValuePrefix) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range info.File.Enums {
		prefix := pascalToUpperSnake(enum.Name)

		for _, value := range enum.Values {
			if !strings.HasPrefix(value.Name, prefix) {
				res = core.AppendIssue(res, e, value.Pos, value.Name)
			}
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			prefix := pascalToUpperSnake(enum.Name)

			for _, value := range enum.Values {
				if !strings.HasPrefix(value.Name, prefix) {
					res = core.AppendIssue(res, e, value.Pos, value.Name)
				}
			}
		}
	}

	return res, nil
}

// pascalToUpperSnake converts a pascal-case name to the UPPER_SNAKE
// prefix main's EnumValuePrefix requires: an underscore precedes every
// uppercase rune. Carried byte-verbatim from main's enum_value_prefix.go.
func pascalToUpperSnake(s string) string {
	var result string

	for _, char := range s {
		if unicode.IsUpper(char) {
			if len(result) > 0 {
				result += "_"
			}
			result += string(char)
		} else {
			result += string(unicode.ToUpper(char))
		}
	}

	return result
}
