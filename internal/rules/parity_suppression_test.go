package rules_test

// Suppression-parity cells (04-08 gap closure; 04-REVIEW CR-01 + WR-01):
// ignore_only must change GROUP MEMBERSHIP exactly like main — main skips
// Validate for a file listed in ignore_only for a rule (core/lint.go:48), so
// an ignored file never joins a group rule's cache and can never seed its
// exempt-first value. The linked engine reached the same semantics only with
// the CR-01 collectFile fix; before it, the cells below failed the parity
// compare with old=0/new=1 at opts/v1/beta.proto:5:1.
//
// Probe provenance (04-REVIEW CR-01, re-proven by the verifier via
// `go test -overlay`, suite TestScratchIgnoreOnlyGroupParityProbe): corpus
// opts/v1/alpha.proto (go_package "github.com/alpha/v1") + opts/v1/beta.proto
// (go_package "github.com/beta/v1") under PACKAGE_SAME_GO_PACKAGE with
// ignore_only {PACKAGE_SAME_GO_PACKAGE: [opts/v1/alpha.proto]} — old engine
// 0 findings (alpha never seeds the cache; beta becomes the exempt first
// value), linked engine 1 finding at opts/v1/beta.proto:5:1 (alpha still
// seeded the group; beta's path is not ignored, so it survives the drain
// filter). The divergence affects all nine group ModuleRules; this suite pins
// the engine-side membership fix on one representative group rule and rides
// the harness threading fix (04-REVIEW WR-01): assertBatchParity feeds
// NewLinked's unwrapped ignoreOnly into the linked app's core.Options.
// IgnoreOnly, so both engines run identical suppression state by construction.
//
// The two non-CR-01 cells are regression pins — they failed only while the
// harness ran nil-vs-set suppression state (WR-01) and stay green after the
// threading + engine fix:
//   - mismatching_file_ignored: ignoring the NON-exempt file drops exactly
//     that file's finding on both sides — main via the Validate skip, the
//     linked engine via the collectFile membership skip plus the drain's
//     path filter;
//   - per_file_rule_ignored: a per-file rule's ignored file is skipped at
//     Validate (main) / checkFile (linked), the reported file's finding
//     survives.

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/config"
)

// TestLinkedParitySuppressions pins ignore_only parity: group membership
// (04-REVIEW CR-01) and the per-path findings filters (regression pins).
func TestLinkedParitySuppressions(t *testing.T) {
	t.Parallel()

	// optsCorpus is the 04-REVIEW CR-01 corpus, reproduced in-suite: two
	// files in one package with differing go_package values. The files start
	// at `syntax` with no leading blank line, so the option statement sits at
	// 5:1 — the probe's reported position.
	optsCorpus := map[string]string{
		"opts/v1/alpha.proto": `syntax = "proto3";

package opts.v1;

option go_package = "github.com/alpha/v1";

message Alpha {}
`,
		"opts/v1/beta.proto": `syntax = "proto3";

package opts.v1;

option go_package = "github.com/beta/v1";

message Beta {}
`,
	}

	t.Run("first_exempt_file_ignored", func(t *testing.T) {
		t.Parallel()

		// CR-01 probe cell: the ignored file is the group's exempt-first
		// contributor. Main never seeds its cache with alpha, so beta becomes
		// the exempt first value and NEITHER engine reports (old len 0).
		// Pre-fix the linked engine collected alpha anyway and reported beta
		// at opts/v1/beta.proto:5:1 — the RED failure (old len 0 vs new 1).
		// Hand-derived old-side count: 0.
		cfg := config.LintConfig{
			Use: []string{"PACKAGE_SAME_GO_PACKAGE"},
			IgnoreOnly: map[string][]string{
				"PACKAGE_SAME_GO_PACKAGE": {"opts/v1/alpha.proto"},
			},
		}
		assertBatchParity(t, cfg, optsCorpus, 0)
	})

	t.Run("mismatching_file_ignored", func(t *testing.T) {
		t.Parallel()

		// Regression pin: the ignored file is the MISMATCHING one. Main drops
		// its finding via the Validate skip (old 0); the fixed linked engine
		// agrees — the collectFile membership skip plus the drain's path
		// filter both gate beta. Hand-derived old-side count: 0.
		cfg := config.LintConfig{
			Use: []string{"PACKAGE_SAME_GO_PACKAGE"},
			IgnoreOnly: map[string][]string{
				"PACKAGE_SAME_GO_PACKAGE": {"opts/v1/beta.proto"},
			},
		}
		assertBatchParity(t, cfg, optsCorpus, 0)
	})

	t.Run("per_file_rule_ignored", func(t *testing.T) {
		t.Parallel()

		// Regression pin for per-file rules: the ignored file's finding is
		// skipped at Validate (main) / checkFile (linked — the per-path skip
		// in the concurrent window), the reported file's finding survives.
		// Hand-derived old-side count: 1 at mp/reported.proto.
		files := map[string]string{
			"mp/ignored.proto": `syntax = "proto3";

package mp;

message lower_ignored {}
`,
			"mp/reported.proto": `syntax = "proto3";

package mp;

message lower_reported {}
`,
		}
		cfg := config.LintConfig{
			Use: []string{"MESSAGE_PASCAL_CASE"},
			IgnoreOnly: map[string][]string{
				"MESSAGE_PASCAL_CASE": {"mp/ignored.proto"},
			},
		}
		assertBatchParity(t, cfg, files, 1)
	})
}
