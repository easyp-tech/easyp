package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumValueUpperSnakeCase)(nil)

// EnumValueUpperSnakeCase (research inventory #8, main-diverged): main's
// regexp wins over the lab's casing-helper equality (leading/trailing
// underscores and empty names fail main), and the position is the value's
// Meta.Pos ≡ Decl.Pos (the name start for keyword-less enum values), not a
// finer token. Walk shape: main's two-level enum iteration — top-level enums
// plus enums of TOP-LEVEL messages.
type EnumValueUpperSnakeCase struct{}

// Message implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Message() string {
	return "enum value must be in UPPER_SNAKE_CASE"
}

// Validate implements lint.Rule.
func (c *EnumValueUpperSnakeCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	upperSnakeCase := regexp.MustCompile("^[A-Z0-9]+(_[A-Z0-9]+)*$")

	for _, enum := range info.File.Enums {
		for _, field := range enum.Values {
			if !upperSnakeCase.MatchString(field.Name) {
				res = core.AppendIssue(res, c, field.Pos, field.Name)
			}
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			for _, field := range enum.Values {
				if !upperSnakeCase.MatchString(field.Name) {
					res = core.AppendIssue(res, c, field.Pos, field.Name)
				}
			}
		}
	}

	return res, nil
}
