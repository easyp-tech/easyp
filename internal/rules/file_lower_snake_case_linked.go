package rules

import (
	"path/filepath"
	"regexp"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*FileLowerSnakeCase)(nil)

// FileLowerSnakeCase
// (research inventory #27, main-diverged position + regexp; threat T-04-12):
// main checks the FULL filename including ".proto" against the
// package-style lower-snake regexp — no extension stripping (the lab's
// TrimSuffix+helper-equality would pass a file main fails, e.g. a digit
// directly after the leading lowercase run of the stem). The finding is
// reported at the zero Position literal (Filename only; Line 0, Column 0 —
// main's meta.Position zero value), never the lab's 1:1 file-level position
// and never a name-token position.
type FileLowerSnakeCase struct{}

// Message implements lint.Rule.
func (f *FileLowerSnakeCase) Message() string {
	return "file name should be lower_snake_case.proto"
}

// Validate implements lint.Rule.
func (f *FileLowerSnakeCase) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	// main's matchLowerSnakeCase regexp, checked against the full filename
	// including the .proto extension. Declared locally: the frozen old file
	// owns the package-level declaration of the same name.
	lowerSnakeCase := regexp.MustCompile("^[a-z]+([_|[.][a-z0-9]+)*$")

	fileName := filepath.Base(info.Path)
	if !lowerSnakeCase.MatchString(fileName) {
		res = core.AppendIssue(res, f, core.Position{
			Filename: info.Path,
		}, info.Path)
	}

	return res, nil
}
