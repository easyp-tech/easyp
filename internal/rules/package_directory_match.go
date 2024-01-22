package rules

import (
	"errors"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageDirectoryMatch)(nil)

// PackageDirectoryMatch is a rule for checking consistency of directory and package names.
type PackageDirectoryMatch struct {
}

// Validate implements core.Rule.
func (d *PackageDirectoryMatch) Validate(info core.ProtoInfo) []error {
	return []error{errors.New("implements me")}
}
