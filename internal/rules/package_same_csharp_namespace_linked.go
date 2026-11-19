package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*PackageSameCsharpNamespaceLinked)(nil)

// PackageSameCsharpNamespaceLinked is the linked-engine port of
// PackageSameCsharpNamespace (research inventory #16, main-diverged + adapt):
// all files of one package must declare the same csharp_namespace option.
// The Collect/Finish mechanics come from the etalon's ModuleRule shape; the
// semantics are main's (threat T-04-11):
//
//   - a file WITHOUT the option is skipped entirely — never added as an empty
//     value (the lab's add-always is forbidden);
//   - the comparison basis is the option constant AS WRITTEN (raw
//     opt.Value ≡ main's option.Constant; string-unquoting helpers are
//     forbidden — Pitfall 10);
//   - the first option value per package (walk order) is exempt, later
//     differing files issue at their own option positions
//     (groupedValues.firstExemptMismatches — the lab's
//     report-everyone Finish is not reachable from ports);
//   - a file without a package never joins a group (main returns early).
type PackageSameCsharpNamespaceLinked struct {
	files groupedValues
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (p *PackageSameCsharpNamespaceLinked) LintRuleName() string {
	return "PACKAGE_SAME_CSHARP_NAMESPACE"
}

// Message implements lint.LintRule.
func (p *PackageSameCsharpNamespaceLinked) Message() string {
	return "different proto files in the same package should have the same csharp_namespace"
}

// Validate implements lint.LintRule. Lint calls Collect and Finish instead.
func (p *PackageSameCsharpNamespaceLinked) Validate(core.LinkedFileInfo) ([]core.Issue, error) {
	return nil, nil
}

// Collect implements core.ModuleRule.
func (p *PackageSameCsharpNamespaceLinked) Collect(info core.LinkedFileInfo) error {
	if len(info.File.Packages) == 0 {
		// main returns early: packageless files never join groups.
		return nil
	}
	pkg := info.File.Packages[0]
	opt := fileOption(info.File, "csharp_namespace")
	if opt == nil {
		// main skips optionless files entirely.
		return nil
	}
	// opt.Value is the raw written constant; opt.Pos is Decl.Pos — the
	// `option` keyword, the same position main reported as option.Meta.Pos.
	p.files.add(pkg.Name, groupedValue{
		path:   info.Path,
		value:  opt.Value,
		source: opt.Value,
		pos:    opt.Pos,
	})
	return nil
}

// Finish implements core.ModuleRule.
func (p *PackageSameCsharpNamespaceLinked) Finish() []core.IssueInfo {
	return p.files.firstExemptMismatches(p)
}
