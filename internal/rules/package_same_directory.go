package rules

import (
	"errors"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageSameDirectory)(nil)

// PackageSameDirectory is a rule for checking consistency of directory and package names.
type PackageSameDirectory struct{}

// Validate implements core.Rule.
func (d *PackageSameDirectory) Validate(info core.ProtoInfo) []error {
	return []error{errors.New("implements me")}
}
