package rules

import (
	"reflect"

	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageDefined)(nil)

// PackageDefined this rule checks that all files have a package declaration.
type PackageDefined struct{}

// Name implements lint.Rule.
func (p *PackageDefined) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(p).Elem().Name())
}

// Validate implements lint.Rule.
func (p *PackageDefined) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	if len(protoInfo.Info.ProtoBody.Packages) == 0 {
		res = append(res, BuildError(protoInfo.Path, meta.Position{}, protoInfo.Path, lint.ErrPackageIsNotDefined))
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
