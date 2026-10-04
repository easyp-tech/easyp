package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageDirectoryMatch)(nil)

// PackageDirectoryMatch is a rule for checking consistency of directory and package names.
type PackageDirectoryMatch struct {
	Root   string `json:"root" yaml:"root" env:"PACKAGE_DIRECTORY_MATCH_ROOT"`
	Prefix string `json:"-" yaml:"-"`
}

// Message implements lint.Rule.
func (d *PackageDirectoryMatch) Message() string {
	return "package does not match directory path"
}

// Validate implements lint.Rule.
func (d *PackageDirectoryMatch) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	name := strings.TrimPrefix(protoInfo.Path, d.Root)
	if protoInfo.ImportPath != "" {
		name = protoInfo.ImportPath
	}
	preparePath := filepath.ToSlash(filepath.Dir(name))
	expectedPackage := strings.ReplaceAll(preparePath, "/", ".")
	qualifiedPackage := expectedPackage
	if d.Prefix != "" && expectedPackage != "." && expectedPackage != d.Prefix && !strings.HasPrefix(expectedPackage, d.Prefix+".") {
		qualifiedPackage = d.Prefix + "." + expectedPackage
	}

	for _, pkgInfo := range protoInfo.Info.ProtoBody.Packages {
		if pkgInfo.Name == expectedPackage || pkgInfo.Name == qualifiedPackage {
			continue
		}
		sourceName := protoInfo.Path
		if protoInfo.ImportPath != "" {
			sourceName = protoInfo.ImportPath
		}
		res = core.AppendIssue(res, d, pkgInfo.Meta.Pos, sourceName, pkgInfo.Comments)
		if qualifiedPackage != expectedPackage {
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q or module-qualified %q", pkgInfo.Name, preparePath, expectedPackage, qualifiedPackage)
		} else {
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q", pkgInfo.Name, preparePath, expectedPackage)
		}
	}

	return res, nil
}
