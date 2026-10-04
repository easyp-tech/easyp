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
	Root string `json:"root" yaml:"root" env:"PACKAGE_DIRECTORY_MATCH_ROOT"`
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
	preparePath := filepath.Dir(name)
	preparePath = filepath.ToSlash(preparePath)
	expectedPackage := strings.ReplaceAll(preparePath, "/", ".")

	for _, pkgInfo := range protoInfo.Info.ProtoBody.Packages {
		if pkgInfo.Name != expectedPackage {
			sourceName := protoInfo.Path
			if protoInfo.ImportPath != "" {
				sourceName = protoInfo.ImportPath
			}
			res = core.AppendIssue(res, d, pkgInfo.Meta.Pos, sourceName, pkgInfo.Comments)
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q", pkgInfo.Name, preparePath, expectedPackage)
		}
	}

	return res, nil
}
