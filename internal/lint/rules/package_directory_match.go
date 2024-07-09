package rules

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*PackageDirectoryMatch)(nil)

// PackageDirectoryMatch is a rule for checking consistency of directory and package names.
type PackageDirectoryMatch struct {
	Root string `json:"root" yaml:"root" env:"PACKAGE_DIRECTORY_MATCH_ROOT"`
}

// Name implements lint.Rule.
func (d *PackageDirectoryMatch) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(d).Elem().Name())
}

// Message implements lint.Rule.
func (d *PackageDirectoryMatch) Message() string {
	return "package is not matched with path"
}

// Validate implements lint.Rule.
func (d *PackageDirectoryMatch) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	preparePath := filepath.Dir(strings.TrimPrefix(protoInfo.Path, d.Root))
	expectedPackage := strings.Replace(preparePath, "/", ".", -1)

	for _, pkgInfo := range protoInfo.Info.ProtoBody.Packages {
		if pkgInfo.Name != expectedPackage {
			res = append(res, lint.BuildError(pkgInfo.Meta.Pos, protoInfo.Path, d.Message()))
		}
	}

	return res, nil
}
