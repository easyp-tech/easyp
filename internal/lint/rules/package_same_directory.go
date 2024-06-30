package rules

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageSameDirectory)(nil)

// PackageSameDirectory this rule checks that all files with a given package are in the same directory.
type PackageSameDirectory struct {
	// dir => package
	cache map[string]string
}

func (d *PackageSameDirectory) lazyInit() {
	if d.cache == nil {
		d.cache = make(map[string]string)
	}
}

// Validate implements lint.Rule.
func (d *PackageSameDirectory) Validate(protoInfo lint.ProtoInfo) []error {
	d.lazyInit()

	var res []error

	directory := filepath.Dir(protoInfo.Path)
	for _, packageInfo := range protoInfo.Info.ProtoBody.Packages {
		if d.cache[packageInfo.Name] == "" {
			d.cache[packageInfo.Name] = directory
			continue
		}

		if d.cache[packageInfo.Name] != directory {
			res = AppendError(res, PACKAGE_SAME_DIRECTORY, packageInfo.Meta.Pos, packageInfo.Name, packageInfo.Comments)
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
