package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageNoImportCycle)(nil)

// PackageNoImportCycle this is an extra uncategorized rule that detects package import cycles.
// The Protobuf compiler outlaws circular file imports, but it's still possible to introduce package cycles, such as these:
type PackageNoImportCycle struct {
	// cache is a map of package name to a slice of package names that it imports
	cache map[string][]string
}

// Name implements lint.Rule.
func (p *PackageNoImportCycle) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(p).Elem().Name())
}

func (p *PackageNoImportCycle) lazyInit() {
	if p.cache == nil {
		p.cache = make(map[string][]string)
	}
}

// Validate implements lint.Rule.
func (p *PackageNoImportCycle) Validate(protoInfo lint.ProtoInfo) []error {
	p.lazyInit()
	panic("implement me")
}
