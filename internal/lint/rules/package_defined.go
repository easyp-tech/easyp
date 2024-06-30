package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageDefined)(nil)

// PackageDefined this rule checks that all files have a package declaration.
type PackageDefined struct{}

// Validate implements lint.Rule.
func (p *PackageDefined) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		res = AppendError(res, PACKAGE_DEFINED, meta.Position{}, protoInfo.Path, nil)
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
