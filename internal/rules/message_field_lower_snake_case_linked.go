package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*FieldLowerSnakeCase)(nil)

// FieldLowerSnakeCase
// (research inventory #9, main-diverged): main's regexp wins over the lab's
// casing-helper equality (casing.go), and the position is the field's
// Meta.Pos ≡ Decl.Pos — the TYPE token start (main's test pins column 3), not
// the lab's declared-name token. Walk shape: fields of TOP-LEVEL messages
// only, plain fields only —
// map, group and oneof-member fields are outside main's predicate, so the
// recursion-capable Visitor walk (which visits Maps, Groups and oneof
// members) is not used.
type FieldLowerSnakeCase struct{}

// Message implements lint.Rule.
func (c *FieldLowerSnakeCase) Message() string {
	return "message field should be lower_snake_case"
}

// Validate implements lint.Rule.
func (c *FieldLowerSnakeCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	lowerSnakeCase := regexp.MustCompile("^[a-z0-9]+(_[a-z0-9]+)*$")

	for _, message := range info.File.Messages {
		for _, field := range message.Fields {
			if !lowerSnakeCase.MatchString(field.Name) {
				res = core.AppendIssue(res, c, field.Pos, field.Name)
			}
		}
	}

	return res, nil
}
