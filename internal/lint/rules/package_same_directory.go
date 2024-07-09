package rules

import (
	"path/filepath"
	"reflect"

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

// Name implements lint.Rule.
func (d *PackageSameDirectory) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(d).Elem().Name())
}

// Message implements lint.Rule.
func (d *PackageSameDirectory) Message() string {
	return "different proto files in the same package should be in the same directory"
}

// Validate implements lint.Rule.
func (d *PackageSameDirectory) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	d.lazyInit()

	var res []lint.Issue

	directory := filepath.Dir(protoInfo.Path)
	for _, packageInfo := range protoInfo.Info.ProtoBody.Packages {
		if d.cache[packageInfo.Name] == "" {
			d.cache[packageInfo.Name] = directory
			continue
		}

		if d.cache[packageInfo.Name] != directory {
			res = append(res, lint.BuildError(packageInfo.Meta.Pos, packageInfo.Name, d.Message()))
		}
	}

	return res, nil
}
