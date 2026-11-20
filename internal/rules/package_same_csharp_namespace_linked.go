package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*PackageSameCsharpNamespace)(nil)

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
//   - the first option occurrence per package (walk order) is exempt; every
//     later occurrence of a differing value — within the same file included —
//     issues at its own option position
//     (groupedValues.firstExemptMismatches — the lab's
//     report-everyone Finish is not reachable from ports);
//   - a file without a package never joins a group (main returns early).
type PackageSameCsharpNamespace struct {
	files groupedValues
}

// Message implements lint.Rule.
func (p *PackageSameCsharpNamespace) Message() string {
	return "different proto files in the same package should have the same csharp_namespace"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageSameCsharpNamespace) Validate(core.ProtoInfo) ([]core.Issue, error) {
	return nil, nil
}

// Collect implements core.ModuleRule.
func (p *PackageSameCsharpNamespace) Collect(info core.ProtoInfo) error {
	if len(info.File.Packages) == 0 {
		// main returns early: packageless files never join groups.
		return nil
	}
	pkg := info.File.Packages[0]
	// main iterates EVERY occurrence of the compared option (04-REVIEW
	// WR-02): the first seeds the per-package value, every later occurrence —
	// same file included — compares at its own position. A file without the
	// matching option adds nothing — the old skip-optionless semantics
	// survive as the empty loop.
	for _, opt := range info.File.Options {
		if opt.Name != "csharp_namespace" {
			continue
		}
		// opt.Value is the raw written constant; opt.Pos is Decl.Pos — the
		// `option` keyword, the same position main reported as option.Meta.Pos.
		p.files.add(pkg.Name, groupedValue{
			path:   info.Path,
			value:  opt.Value,
			source: opt.Value,
			pos:    opt.Pos,
		})
	}
	return nil
}

// Finish implements core.ModuleRule.
func (p *PackageSameCsharpNamespace) Finish() []core.IssueInfo {
	return p.files.firstExemptMismatches(p)
}
