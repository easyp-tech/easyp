package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*PackageDirectoryMatch)(nil)

// PackageDirectoryMatch (research inventory #3, verbatim-from-lab): a rule for
// checking consistency of directory and package names. The struct tags stay
// byte-verbatim from main (config plumbing unchanged); the registry name
// reflects directly from the canonical (unsuffixed) type name.
type PackageDirectoryMatch struct {
	Root   string `json:"root" yaml:"root" env:"PACKAGE_DIRECTORY_MATCH_ROOT"`
	Prefix string `json:"-" yaml:"-"`
}

// Message implements lint.Rule.
func (d *PackageDirectoryMatch) Message() string {
	return "package does not match directory path"
}

// Validate implements lint.Rule.
func (d *PackageDirectoryMatch) Validate(info core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue

	name := strings.TrimPrefix(info.Path, d.Root)
	if info.ImportPath != "" {
		name = info.ImportPath
	}
	preparePath := filepath.ToSlash(filepath.Dir(name))
	expectedPackage := strings.ReplaceAll(preparePath, "/", ".")
	qualifiedPackage := expectedPackage
	if d.Prefix != "" && expectedPackage != "." && expectedPackage != d.Prefix && !strings.HasPrefix(expectedPackage, d.Prefix+".") {
		qualifiedPackage = d.Prefix + "." + expectedPackage
	}

	for _, pkgInfo := range info.File.Packages {
		if pkgInfo.Name == expectedPackage || pkgInfo.Name == qualifiedPackage {
			continue
		}
		sourceName := info.Path
		if info.ImportPath != "" {
			sourceName = info.ImportPath
		}
		// pkgInfo.Pos is the declaration's first token — the same position
		// main reported as Meta.Pos (pinned by TestDeclPositionProbe).
		res = core.AppendIssue(res, d, pkgInfo.Pos, sourceName)
		if qualifiedPackage != expectedPackage {
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q or module-qualified %q", pkgInfo.Name, preparePath, expectedPackage, qualifiedPackage)
		} else {
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q", pkgInfo.Name, preparePath, expectedPackage)
		}
	}

	return res, nil
}
