package rules

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*DirectorySamePackage)(nil)

// DirectorySamePackage this rule checks that all files in a given directory are in the same package.
type DirectorySamePackage struct {
	// dir => package
	cache map[string]string
}

func (d *DirectorySamePackage) lazyInit() {
	if d.cache == nil {
		d.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (d *DirectorySamePackage) Validate(protoInfo lint.ProtoInfo) []error {
	d.lazyInit()
	var res []error

	directory := filepath.Dir(protoInfo.Path)
	for _, pack := range protoInfo.Info.ProtoBody.Packages {
		if d.cache[directory] == "" {
			d.cache[directory] = pack.Name
			continue
		}

		if d.cache[directory] != pack.Name {
			res = AppendError(res, DIRECTORY_SAME_PACKAGE, pack.Meta.Pos, pack.Name, pack.Comments)
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
