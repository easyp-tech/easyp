package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.LintRule = (*PackageDirectoryMatchLinked)(nil)

// PackageDirectoryMatchLinked is the linked-engine port of
// PackageDirectoryMatch (research inventory #3, verbatim-from-lab): a rule for
// checking consistency of directory and package names. The struct tags stay
// byte-verbatim from main (config plumbing unchanged); the registry name is
// pinned through LintRuleName because the suffixed port type would otherwise
// reflect to PACKAGE_DIRECTORY_MATCH_LINKED.
type PackageDirectoryMatchLinked struct {
	Root   string `json:"root" yaml:"root" env:"PACKAGE_DIRECTORY_MATCH_ROOT"`
	Prefix string `json:"-" yaml:"-"`
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (d *PackageDirectoryMatchLinked) LintRuleName() string {
	return "PACKAGE_DIRECTORY_MATCH"
}

// Message implements lint.LintRule.
func (d *PackageDirectoryMatchLinked) Message() string {
	return "package does not match directory path"
}

// Validate implements lint.LintRule.
func (d *PackageDirectoryMatchLinked) Validate(info core.LinkedFileInfo) ([]core.Issue, error) {
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
		res = core.AppendLinkedIssue(res, d, pkgInfo.Pos, sourceName)
		if qualifiedPackage != expectedPackage {
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q or module-qualified %q", pkgInfo.Name, preparePath, expectedPackage, qualifiedPackage)
		} else {
			res[len(res)-1].Message = fmt.Sprintf("package %q does not match module-relative directory %q; expected %q", pkgInfo.Name, preparePath, expectedPackage)
		}
	}

	return res, nil
}
