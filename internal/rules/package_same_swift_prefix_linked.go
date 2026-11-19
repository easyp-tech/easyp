package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*PackageSameSwiftPrefixLinked)(nil)

// PackageSameSwiftPrefixLinked is the linked-engine port of
// PackageSameSwiftPrefix (research inventory #22, main-diverged + adapt +
// BUG PIN, threat T-04-09): all files of one package must declare the same
// swift_prefix option — except that MAIN COMPARES THE ruby_package OPTION.
// Main's option-name literal is a copy-paste bug (the rule named swift_prefix
// reads ruby_package); per the orchestrator-ratified bug-for-bug resolution
// (research Open Question 1, Pitfall 5) this port reproduces it EXACTLY:
// Collect reads the "ruby_package" option, so differing swift_prefix values
// with matching ruby_package values pass and differing ruby_package values
// issue at the ruby_package option position. FIXING IT HERE IS FORBIDDEN —
// it would be a behavior change that breaks Phase-7 corpus parity. The
// corrected behavior is a DEFERRED POST-PARITY DECISION (ledgered in the
// 04-07 deferred-decisions summary), never a silent improvement. The
// swift_prefix_bug_pin parity cell and the carried regression cells pin the
// bug from both sides.
//
// The rest of the semantics are main's (threat T-04-11):
//
//   - a file WITHOUT the compared option is skipped entirely — never added as
//     an empty value (the lab's add-always is forbidden);
//   - the comparison basis is the option constant AS WRITTEN (raw
//     opt.Value ≡ main's option.Constant; string-unquoting helpers are
//     forbidden — Pitfall 10);
//   - the first option value per package (walk order) is exempt, later
//     differing files issue at their own option positions
//     (groupedValues.firstExemptMismatches — the lab's report-everyone
//     Finish is not reachable from ports);
//   - a file without a package never joins a group (main returns early).
type PackageSameSwiftPrefixLinked struct {
	files groupedValues
}

// LintRuleName returns the canonical registry name (no _LINKED suffix).
func (p *PackageSameSwiftPrefixLinked) LintRuleName() string {
	return "PACKAGE_SAME_SWIFT_PREFIX"
}

// Message implements lint.LintRule.
func (p *PackageSameSwiftPrefixLinked) Message() string {
	return "all files in the same package must have the same swift_prefix option"
}

// Validate implements lint.LintRule. Lint calls Collect and Finish instead.
func (p *PackageSameSwiftPrefixLinked) Validate(core.LinkedFileInfo) ([]core.Issue, error) {
	return nil, nil
}

// Collect implements core.ModuleRule.
func (p *PackageSameSwiftPrefixLinked) Collect(info core.LinkedFileInfo) error {
	if len(info.File.Packages) == 0 {
		// main returns early: packageless files never join groups.
		return nil
	}
	pkg := info.File.Packages[0]

	// BUG PIN (T-04-09): the option NAME below is main's literal — the rule
	// registered as swift_prefix compares ruby_package. Do not "fix" it here;
	// see the type header for the deferred post-parity decision.
	opt := fileOption(info.File, "ruby_package")
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
func (p *PackageSameSwiftPrefixLinked) Finish() []core.IssueInfo {
	return p.files.firstExemptMismatches(p)
}
