package rules

import (
	"go.redsock.ru/protopack/internal/core"
)

var _ core.Rule = (*PackageNoImportCycle)(nil)

// PackageNoImportCycle this is an extra uncategorized rule that detects package import cycles.
// The Protobuf compiler outlaws circular file imports, but it's still possible to introduce package cycles, such as these:
type PackageNoImportCycle struct {
	// cache is a map of package name to a slice of package names that it imports
	cache map[string][]string
}

func (p *PackageNoImportCycle) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string][]string)
	}
}

// Message implements lint.Rule.
func (p *PackageNoImportCycle) Message() string {
	return "package should not have import cycles"
}

// Validate implements lint.Rule.
func (p *PackageNoImportCycle) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	p.lazyInit()
	panic("implement me")
}
