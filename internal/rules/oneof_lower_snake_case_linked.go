package rules

import (
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*OneofLowerSnakeCaseLinked)(nil)

// OneofLowerSnakeCaseLinked is the linked-engine port of OneofLowerSnakeCase
// (research inventory #14, main-diverged): main's regexp wins over the lab's
// casing-helper equality (casing.go) — digits fail main ("oneof1") but pass
// the helper — and the position is the declaration keyword (oneof.Meta.Pos ≡
// Decl.Pos), not the lab's declared-name token. Walk shape: oneofs of
// TOP-LEVEL messages only (main's shallow loop); the recursion-capable
// Visitor walk is not used.
type OneofLowerSnakeCaseLinked struct{}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (c *OneofLowerSnakeCaseLinked) LintRuleName() string {
	return "ONEOF_LOWER_SNAKE_CASE"
}

// Message implements lint.LintRule.
func (c *OneofLowerSnakeCaseLinked) Message() string {
	return "oneof name should be lower_snake_case"
}

// Validate implements lint.LintRule.
func (c *OneofLowerSnakeCaseLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
	var res []core.Issue

	lowerSnakeCase := regexp.MustCompile("^[a-z]+(_[a-z]+)*$")

	for _, message := range info.File.Messages {
		for _, oneof := range message.Oneofs {
			if !lowerSnakeCase.MatchString(oneof.Name) {
				res = core.AppendLinkedIssue(res, c, oneof.Pos, oneof.Name)
			}
		}
	}

	return res, nil
}
