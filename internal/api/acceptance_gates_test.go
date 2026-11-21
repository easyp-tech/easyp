package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// acceptanceBaselineCommit is the RATIFIED baseline oracle commit (07-02
// flagged assumption A2): 1e9a596 — verified zero Go-source delta to the first
// milestone commit 8f969d8 (two docs-only commits intervene — a0a0ed9, 26924b1;
// the 1e9a596..8f969d8 diff is .planning/* + go.mod metadata only, go-billy
// v5.6.2 indirect->direct at the same version). "Current main" for parity is
// therefore exactly this commit's binary; the oracle's presence is the gate —
// the test fails loudly when the commit is absent and never skips.
const acceptanceBaselineCommit = "1e9a596"

// acceptanceCorpusConfig is the teeth corpus's whole config: exactly two keys,
// mirroring 06-03's corpus and bench/setup.sh's materialized-project
// convention. STANDARD expands to MINIMAL+BASIC+DEFAULT (policy.go:109) and
// enables 35 of the 44-name HEAD registry, 32 of the 41-name baseline registry;
// the COMMENTS (7) and UNARY_RPC (2) registry rules are never enabled by this
// config. All three NEW3 rules are inside MINIMAL/BASIC, so Gate D is
// unaffected. Rule counts are stated from the registries everywhere
// (lint_model.go:28-29 flags the stale "43" REQUIREMENTS figure — never
// propagate it).
const acceptanceCorpusConfig = "version: v1\nlinters:\n  default: STANDARD\n"

// acceptanceNewRules is the NEW3 filter: the three Phase-6 rule names the
// baseline 41-rule registry lacks (HEAD's registry is 44). Every compare
// filters these from the HEAD side before the 4-tuple multiset compare, and
// Gate D accounts for them separately.
var acceptanceNewRules = map[string]bool{
	"PACKAGE_NO_IMPORT_CYCLE": true,
	"SYNTAX_SPECIFIED":        true,
	"FIELD_NOT_REQUIRED":      true,
}

// acceptanceCorpusFiles is the teeth corpus: the five tracked fixtures copied
// verbatim preserving subdirectory layout. sp1/sp2 are same-package negatives
// that 06-03 excluded from the corpus — kept excluded here.
var acceptanceCorpusFiles = []string{
	"import_cycle/a.proto",
	"import_cycle/a2.proto",
	"import_cycle/b/b.proto",
	"no_syntax/no_syntax.proto",
	"required_field/required_field.proto",
}

// acceptanceReportedIssues matches the stderr cross-check line the lint run
// emits — the same line lab.py greps (msg="lint completed" issues=N).
var acceptanceReportedIssues = regexp.MustCompile(`msg="lint completed" issues=(\d+)`)

// acceptanceRepoRoot resolves the repository root the same way
// buildEasypBinary does: runtime.Caller on this file, three dirs up from
// internal/api.
func acceptanceRepoRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(thisFile))), nil
}

// lintIssue is the JSON issue shape the CLI prints with --format json: one
// object per line, the embedded Issue fields promoted beside Path — the same
// shape lab.py snapshots and the RESEARCH spike verified field-identical
// across the two binaries (Message, Path, Position{Filename, Offset, Line,
// Column}, RuleName, SourceName).
type lintIssue struct {
	Position struct {
		Filename string `json:"Filename"`
		Offset   int    `json:"Offset"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Position"`
	SourceName string `json:"SourceName"`
	Message    string `json:"Message"`
	RuleName   string `json:"RuleName"`
	Path       string `json:"Path"`
}

// lintKey is the Gate C identity projection: (RuleName, Path, Position.Line,
// Position.Column) — the LINT-PARITY-01 letter. Message text and Offset sit
// outside the gate; drift there belongs to the reported raw-line diagnostic,
// never to a silent pass and never to this gate.
type lintKey struct {
	Rule string
	Path string
	Line int
	Col  int
}

func (i lintIssue) key() lintKey {
	return lintKey{Rule: i.RuleName, Path: i.Path, Line: i.Position.Line, Col: i.Position.Column}
}

// removeWorktreeRegistration tears down a baseline worktree registration and
// prunes stale leftovers. Used on the build-failure path (a cached
// sync.OnceValues error would otherwise keep the leak forever) and from the
// test's t.Cleanup.
func removeWorktreeRegistration(repoRoot, worktreeDir string) {
	remove := exec.Command("git", "worktree", "remove", "--force", worktreeDir)
	remove.Dir = repoRoot
	_, _ = remove.CombinedOutput()
	prune := exec.Command("git", "worktree", "prune")
	prune.Dir = repoRoot
	_, _ = prune.CombinedOutput()
}

// baselineBinary is the baseline oracle build result: the binary path plus the
// worktree dir (non-empty exactly when a git worktree registration may still
// exist for the caller's t.Cleanup to remove).
type baselineBinary struct {
	binPath  string
	worktree string
}

// buildBaselineBinary compiles the baseline oracle once per test run: a plain
// go build of ./cmd/easyp (no -race — detector overhead invalidates timing,
// and parity runs use the same binaries as perf) from a temporary git worktree
// at the ratified 1e9a596. The worktree registration is removed via the
// caller's t.Cleanup so failures never leak one; the binary lives outside the
// repository and is never committed.
var buildBaselineBinary = sync.OnceValues(func() (baselineBinary, error) {
	repoRoot, err := acceptanceRepoRoot()
	if err != nil {
		return baselineBinary{}, err
	}

	// The oracle's presence is the gate: fail loudly when the ratified commit
	// is absent — never skip.
	check := exec.Command("git", "cat-file", "-e", acceptanceBaselineCommit+"^{commit}")
	check.Dir = repoRoot
	if out, checkErr := check.CombinedOutput(); checkErr != nil {
		return baselineBinary{}, fmt.Errorf("baseline oracle commit %s is absent (never skip): %w: %s", acceptanceBaselineCommit, checkErr, out)
	}

	binDir, err := os.MkdirTemp("", "easyp-acceptance-baseline")
	if err != nil {
		return baselineBinary{}, err
	}
	// The worktree dir does not pre-exist (git creates it); its basename is
	// unique per run because binDir is, so the .git/worktrees/<id> admin name
	// derived from the basename never collides with a stale registration.
	worktree := filepath.Join(binDir, filepath.Base(binDir)+"-wt")
	add := exec.Command("git", "worktree", "add", "--detach", worktree, acceptanceBaselineCommit)
	add.Dir = repoRoot
	if out, addErr := add.CombinedOutput(); addErr != nil {
		return baselineBinary{}, fmt.Errorf("git worktree add %s: %w: %s", worktree, addErr, out)
	}

	binPath := filepath.Join(binDir, "easyp-baseline")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/easyp")
	build.Dir = worktree
	if out, buildErr := build.CombinedOutput(); buildErr != nil {
		removeWorktreeRegistration(repoRoot, worktree)
		return baselineBinary{}, fmt.Errorf("go build (baseline %s): %w: %s", acceptanceBaselineCommit, buildErr, out)
	}
	return baselineBinary{binPath: binPath, worktree: worktree}, nil
})

// writeAcceptanceCorpus materializes the teeth corpus into a fresh temp root:
// the five tracked fixtures copied verbatim preserving subdirectory layout
// (corpus-root-relative imports resolve) plus the two-key STANDARD easyp.yaml.
func writeAcceptanceCorpus(t *testing.T) string {
	t.Helper()

	repoRoot, err := acceptanceRepoRoot()
	require.NoError(t, err)

	root := t.TempDir()
	for _, name := range acceptanceCorpusFiles {
		content, err := os.ReadFile(filepath.Join(repoRoot, "testdata", filepath.FromSlash(name)))
		require.NoError(t, err)
		dst := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o700))
		require.NoError(t, os.WriteFile(dst, content, 0o600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte(acceptanceCorpusConfig), 0o600))
	return root
}

// runAcceptanceLint runs one binary over the corpus with the lab.py lint
// shape: argv = binary, --format, json, lint with cwd = corpus root; the child
// gets its own Dir and an inherited Env, the parent process is never mutated.
// Exit 0/1 accepted; exit 2 = config/compile error = fatal
// (parity-evidence-06-03.json exit-code contract — the corpus must stay
// linker-clean). Every stdout line must be a JSON object (a non-JSON line
// hard-fails — output-contract drift), and the stderr msg="lint completed"
// issues=N cross-check must agree with the parsed issue count. Returns the
// parsed issues and the raw JSON lines (index-aligned).
func runAcceptanceLint(t *testing.T, binPath, corpusRoot string) ([]lintIssue, []string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binPath, "--format", "json", "lint")
	cmd.Dir = corpusRoot
	cmd.Env = os.Environ()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		code = exitErr.ExitCode()
	}
	require.LessOrEqual(t, code, 1,
		"lint exited %d: exit 2 = config/compile error — the corpus must stay linker-clean (parity-evidence-06-03.json exit-code contract); stderr: %s", code, stderr.String())

	var issues []lintIssue
	var lines []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(line), &raw),
			"every stdout line must be a JSON object — a non-JSON line is output-contract drift")
		require.NotNil(t, raw, "a JSON null line is not an issue object")
		var issue lintIssue
		require.NoError(t, json.Unmarshal([]byte(line), &issue))
		issues = append(issues, issue)
		lines = append(lines, line)
	}

	reported := acceptanceReportedIssues.FindStringSubmatch(stderr.String())
	require.NotNil(t, reported,
		"stderr must carry the msg=%q issues=N cross-check line; stderr: %s", "lint completed", stderr.String())
	require.Equal(t, strconv.Itoa(len(issues)), reported[1],
		"stderr msg=%q issues=N must agree with the parsed issue count", "lint completed")

	return issues, lines
}

// TestAcceptanceGates is the automated-core parity datapoint: the baseline
// oracle (current main, built from the ratified 1e9a596) vs HEAD on the teeth
// corpus, Gates A-D in order, each binary run twice. It is permanent machinery
// in the normal suite — independent of acceptance-corpus availability (the
// api-common/googleapis corpora legs are the gated 07-03 scripts behind the
// documented unblock path). Rule counts everywhere come from the registries:
// 41 (baseline main) / 44 (HEAD); the stale "43" figure is never asserted.
//
// Provenance (07-02 flagged assumption A2, RATIFIED): baseline = 1e9a596 —
// verified zero Go-source delta to the first milestone commit 8f969d8 (two
// docs-only commits intervene — a0a0ed9, 26924b1; the 1e9a596..8f969d8 diff is
// .planning/* + go.mod metadata only). Plain builds only: -race detector
// overhead would invalidate timing, and parity runs use the same binaries as
// perf.
//
// Gate order (the ratified compare design, RESEARCH Pattern 1):
//   - Gate A — main-vs-main determinism, captured first per ROADMAP: the same
//     binary twice, sorted raw JSON-line multisets equal on both sides.
//     Emission ORDER is never asserted (main's is nondeterministic by map
//     iteration; the engine's own determinism was proven under -race in
//     Phases 3/5) — content is what must repeat.
//   - Gate B — zero RuleName==COMPILE in all four runs: a COMPILE issue means
//     the linked engine ran no rules — either the corpus violates the
//     all-files-compile precondition or the engine broke. This mirrors the
//     golden purity gate (internal/rules/golden_full_registry_test.go
//     requireGoldenPurity): "the corpus must compile under the linked engine's
//     full type resolution: a COMPILE finding vacates the rule-finding golden".
//   - Gate C — 4-tuple (RuleName, Path, Position.Line, Position.Column)
//     multiset identity after filtering NEW3 from the HEAD side, both
//     directions, shared multiset non-empty (anti-vacuity: an all-empty
//     compare is a failure).
//   - Gate D — HEAD-side NEW3 counts exactly PACKAGE_NO_IMPORT_CYCLE=2,
//     SYNTAX_SPECIFIED=1, FIELD_NOT_REQUIRED=1, matching
//     parity-evidence-06-03.json and the RESEARCH spike.
func TestAcceptanceGates(t *testing.T) {
	t.Parallel()

	headBin, err := buildEasypBinary()
	require.NoError(t, err)

	baseline, err := buildBaselineBinary()
	require.NoError(t, err)
	t.Cleanup(func() {
		if baseline.worktree == "" {
			return
		}
		repoRoot, rootErr := acceptanceRepoRoot()
		if rootErr != nil {
			t.Logf("resolve repo root for worktree cleanup: %v", rootErr)
			return
		}
		removeWorktreeRegistration(repoRoot, baseline.worktree)
	})

	corpusRoot := writeAcceptanceCorpus(t)

	type gateRun struct {
		name   string
		issues []lintIssue
		lines  []string
	}
	// main-vs-main baseline captured first per ROADMAP: the same baseline
	// binary twice, then the HEAD binary twice — each with the lab.py lint
	// shape.
	capture := func(label, bin string) gateRun {
		issues, lines := runAcceptanceLint(t, bin, corpusRoot)
		return gateRun{name: label, issues: issues, lines: lines}
	}
	baseR1 := capture("main-r1", baseline.binPath)
	baseR2 := capture("main-r2", baseline.binPath)
	headR1 := capture("new-r1", headBin)
	headR2 := capture("new-r2", headBin)

	sortedLines := func(r gateRun) []string {
		out := slices.Clone(r.lines)
		slices.Sort(out)
		return out
	}

	// Gate A — main-vs-main determinism, both sides, x2: sorted raw JSON-line
	// multisets equal. Never compare unsorted streams (Pitfall 5).
	require.Equal(t, sortedLines(baseR1), sortedLines(baseR2),
		"Gate A: baseline main-vs-main determinism (sorted raw JSON-line multisets, r1 vs r2)")
	require.Equal(t, sortedLines(headR1), sortedLines(headR2),
		"Gate A: head determinism (sorted raw JSON-line multisets, r1 vs r2)")

	// Gate B — zero-COMPILE precondition in all four runs, before any compare.
	for _, r := range []gateRun{baseR1, baseR2, headR1, headR2} {
		for _, iss := range r.issues {
			require.NotEqual(t, "COMPILE", iss.RuleName,
				"Gate B (%s): the corpus must compile under the linked engine's full type resolution: a COMPILE finding vacates the rule-finding compare (golden_full_registry_test.go requireGoldenPurity) — either the corpus violates the all-files-compile precondition or the engine broke", r.name)
		}
	}

	projection := func(issues []lintIssue, filterNew3 bool) map[lintKey]int {
		out := make(map[lintKey]int)
		for _, iss := range issues {
			if filterNew3 && acceptanceNewRules[iss.RuleName] {
				continue
			}
			out[iss.key()]++
		}
		return out
	}

	for _, pair := range []struct {
		label      string
		base, head gateRun
	}{
		{"r1", baseR1, headR1},
		{"r2", baseR2, headR2},
	} {
		// Gate C — filtered 4-tuple multiset identity, both directions.
		baseCount := projection(pair.base.issues, false)
		headCount := projection(pair.head.issues, true)
		require.True(t, maps.Equal(baseCount, headCount),
			"Gate C (%s): filtered 4-tuple (RuleName, Path, Position.Line, Position.Column) multiset identity must hold in both directions — baseline registry 41 rules vs HEAD 44 with the NEW3 names filtered", pair.label)

		shared := 0
		for key, n := range baseCount {
			if m, ok := headCount[key]; ok {
				shared += min(n, m)
			}
		}
		require.Greater(t, shared, 0,
			"Gate C (%s) anti-vacuity: the shared multiset must be non-empty — an all-empty compare proves nothing", pair.label)

		// Gate D — exact HEAD-side new-rule delta (reported rules, not the
		// stale "43" figure; HEAD registry is 44).
		headDelta := make(map[string]int)
		for _, iss := range pair.head.issues {
			if acceptanceNewRules[iss.RuleName] {
				headDelta[iss.RuleName]++
			}
		}
		require.Equal(t, map[string]int{
			"PACKAGE_NO_IMPORT_CYCLE": 2,
			"SYNTAX_SPECIFIED":        1,
			"FIELD_NOT_REQUIRED":      1,
		}, headDelta,
			"Gate D (%s): HEAD-side new-rule counts must match parity-evidence-06-03.json and the RESEARCH spike exactly", pair.label)
	}

	// Reported, never asserted and never a gate: the raw JSON-line Counter
	// diff between baseline-r1 and head-r1 after the NEW3 filter. Message
	// text and Offset sit outside the 4-tuple gate; every future delta line
	// gets an explicit Message/Offset disposition at evidence time (07-03).
	rawBase := make(map[string]int)
	for _, line := range baseR1.lines {
		rawBase[line]++
	}
	rawHead := make(map[string]int)
	for idx, iss := range headR1.issues {
		if acceptanceNewRules[iss.RuleName] {
			continue
		}
		rawHead[headR1.lines[idx]]++
	}
	rawDiff := func(base, head map[string]int, sign string) []string {
		out := make([]string, 0, len(base))
		for line, n := range base {
			if d := n - head[line]; d > 0 {
				out = append(out, fmt.Sprintf("%s%d %s", sign, d, line))
			}
		}
		slices.Sort(out)
		return out
	}
	onlyBase := rawDiff(rawBase, rawHead, "-")
	onlyHead := rawDiff(rawHead, rawBase, "+")
	t.Logf("raw JSON-line diagnostic after NEW3 filter (reported, not a gate): only-main %d, only-head %d", len(onlyBase), len(onlyHead))
	for _, line := range onlyBase {
		t.Logf("  %s", line)
	}
	for _, line := range onlyHead {
		t.Logf("  %s", line)
	}
}
