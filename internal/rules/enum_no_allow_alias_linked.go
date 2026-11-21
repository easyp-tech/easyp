package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*EnumNoAllowAlias)(nil)

// EnumNoAllowAlias
// (research inventory #6, main-diverged): main issues for ANY enum-level
// option named allow_alias, regardless of its value (allow_alias = false
// included), at the ENUM keyword position (main's Meta.Pos ≡ Decl.Pos) with
// the enum name as the source, re-reporting per matching option occurrence —
// exactly main's loop shape. The lab's Value=="true"-at-option-position
// reading is not ported.
type EnumNoAllowAlias struct{}

// Message implements lint.Rule.
func (e *EnumNoAllowAlias) Message() string {
	return "enum must not allow alias"
}

// Validate implements lint.Rule.
func (e *EnumNoAllowAlias) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	for _, enum := range info.File.Enums {
		for _, opt := range enum.Options {
			if opt.Name == "allow_alias" {
				res = core.AppendIssue(res, e, enum.Pos, enum.Name)
			}
		}
	}

	for _, msg := range info.File.Messages {
		for _, enum := range msg.Enums {
			for _, opt := range enum.Options {
				if opt.Name == "allow_alias" {
					res = core.AppendIssue(res, e, enum.Pos, enum.Name)
				}
			}
		}
	}

	return res, nil
}
