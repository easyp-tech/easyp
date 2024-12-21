package rules

import (
	"path/filepath"
	"regexp"

	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*FileLowerSnakeCase)(nil)

// FileLowerSnakeCase this rule says that all .proto files must be named as lower_snake_case.proto.
// This is the widely accepted standard.
type FileLowerSnakeCase struct {
}

// Message implements lint.Rule.
func (f *FileLowerSnakeCase) Message() string {
	return "file name should be lower_snake_case.proto"
}

// Validate implements lint.Rule.
func (f *FileLowerSnakeCase) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	fileName := filepath.Base(protoInfo.Path)
	if !isLowerSnakeCase(fileName) {
		res = core.AppendIssue(res, f, meta.Position{
			Filename: protoInfo.Path,
			Offset:   0,
			Line:     0,
			Column:   0,
		}, protoInfo.Path, nil)
	}

	return res, nil
}

var matchLowerSnakeCase = regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")

func isLowerSnakeCase(s string) bool {
	return matchLowerSnakeCase.MatchString(s)
}
