package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageDefined)(nil)

// PackageDefined is a rule for checking package is defined.
type PackageDefined struct{}

// Validate implements core.Rule.
func (p *PackageDefined) Validate(protoInfo core.ProtoInfo) []error {
	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		res = append(res, buildError(meta.Position{}, protoInfo.Path, core.ErrPackageIsNotDefined))
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
