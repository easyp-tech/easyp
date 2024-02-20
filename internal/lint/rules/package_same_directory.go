package rules

import (
	"errors"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameDirectory)(nil)

// PackageSameDirectory is a rule for checking consistency of directory and package names.
type PackageSameDirectory struct{}

// Validate implements core.Rule.
func (d *PackageSameDirectory) Validate(info lint.ProtoInfo) []error {
	return []error{errors.New("implements me")}
}
