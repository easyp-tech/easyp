package rules

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*DirectorySamePackage)(nil)

// DirectorySamePackage is a rule for checking consistency of directory and package names.
type DirectorySamePackage struct {
	// dir => package
	cache map[string]string
}

func (d *DirectorySamePackage) lazyInit() {
	if d.cache == nil {
		d.cache = make(map[string]string)
	}
}

// Validate implements core.Rule.
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
			res = append(res, buildError(pack.Meta.Pos, pack.Name, lint.ErrDirectorySamePackage))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
