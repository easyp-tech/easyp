package rules

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageNoImportCycle)(nil)

// PackageNoImportCycle is reserved and is not registered as an executable rule.
// Package-cycle validation requires a complete import graph and is not implemented.
type PackageNoImportCycle struct{}

// Message implements lint.Rule.
func (p *PackageNoImportCycle) Message() string {
	return "package should not have import cycles"
}

// Validate implements lint.Rule.
func (p *PackageNoImportCycle) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	return nil, fmt.Errorf("%w: PACKAGE_NO_IMPORT_CYCLE is not implemented", core.ErrInvalidRule)
}
