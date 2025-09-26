package rules

import (
	"log/slog"
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
	return "package is not matched with path"
}

// Validate implements lint.Rule.
func (d *PackageDirectoryMatch) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	preparePath := filepath.Dir(strings.TrimPrefix(protoInfo.Path, d.Root))
	preparePath = filepath.ToSlash(preparePath)
	expectedPackage := strings.ReplaceAll(preparePath, "/", ".")

	for _, pkgInfo := range protoInfo.Info.ProtoBody.Packages {
		slog.Info("data", "preparePath", preparePath, "expectedPackage", expectedPackage, "pkgInfo.Name", pkgInfo.Name)
		if pkgInfo.Name != expectedPackage {
			res = core.AppendIssue(res, d, pkgInfo.Meta.Pos, protoInfo.Path, pkgInfo.Comments)
		}
	}

	return res, nil
}
