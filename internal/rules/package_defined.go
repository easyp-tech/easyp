package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*PackageDefined)(nil)

// PackageDefined this rule checks that all files have a package declaration.
type PackageDefined struct{}

// Message implements lint.Rule.
func (p *PackageDefined) Message() string {
	return "package should be defined"
}

// Validate implements lint.Rule.
func (p *PackageDefined) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		res = core.AppendIssue(res, p, meta.Position{
			Filename: protoInfo.Path,
		}, protoInfo.Path, nil)
	}

	return res, nil
}
