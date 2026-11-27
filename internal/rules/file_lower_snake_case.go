package rules

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
	"path/filepath"
)

var _ core.Rule = (*FileLowerSnakeCase)(nil)

// FileLowerSnakeCase this rule checks that file names are lower_snake_case.
type FileLowerSnakeCase struct{}

// Message implements lint.Rule.
func (f *FileLowerSnakeCase) Message() string {
	return "file name should be lower_snake_case.proto"
}

// Validate implements lint.Rule.
func (f *FileLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	base := filepath.Base(protoInfo.Path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if name != toLowerSnakeCase(name) {
		res = core.AppendIssue(res, f, protoast.FilePosition(protoInfo.Path), protoInfo.Path)
	}

	return res, nil
}
