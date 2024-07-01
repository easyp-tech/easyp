package rules

import (
	"reflect"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*ImportNoPublic)(nil)

// ImportNoPublic this rule outlaws declaring imports as public.
// If you didn't know that was possible, forget what you just learned in this sentence.
type ImportNoPublic struct{}

// Name implements lint.Rule.
func (i *ImportNoPublic) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(i).Elem().Name())
}

// Validate implements lint.Rule.
func (i *ImportNoPublic) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, imp := range protoInfo.Info.ProtoBody.Imports {
		if imp.Modifier == parser.ImportModifierPublic {
			res = append(res, BuildError(protoInfo.Path, imp.Meta.Pos, imp.Location, lint.ErrImportIsPublic))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
