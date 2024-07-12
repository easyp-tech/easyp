package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageDefined)(nil)

// PackageDefined this rule checks that all files have a package declaration.
type PackageDefined struct{}

// Message implements lint.Rule.
func (p *PackageDefined) Message() string {
	return "package should be defined"
}

// Validate implements lint.Rule.
func (p *PackageDefined) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		res = append(res, lint.BuildError(meta.Position{
			Filename: protoInfo.Path,
		}, protoInfo.Path, p.Message()))
	}

	return res, nil
}
