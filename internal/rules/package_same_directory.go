package rules

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageSameDirectory)(nil)

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

// Message implements lint.Rule.
func (d *PackageSameDirectory) Message() string {
	return "different proto files in the same package should be in the same directory"
}

// Validate implements lint.Rule.
func (d *PackageSameDirectory) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	d.lazyInit()

	var res []core.Issue

	directory := filepath.Dir(protoInfo.Path)
	for _, packageInfo := range protoInfo.Info.ProtoBody.Packages {
		if d.cache[packageInfo.Name] == "" {
			d.cache[packageInfo.Name] = directory
			continue
		}

		if d.cache[packageInfo.Name] != directory {
			res = core.AppendIssue(res, d, packageInfo.Meta.Pos, packageInfo.Name, packageInfo.Comments)
		}
	}

	return res, nil
}
