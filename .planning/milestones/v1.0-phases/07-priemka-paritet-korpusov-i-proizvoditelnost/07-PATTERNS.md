# Phase 7: Приёмка — паритет корпусов и производительность - Pattern Map

**Mapped:** 2026-11-21
**Files analyzed:** 8 (3 modified, 5 created)
**Analogs found:** 8 / 8 (5 exact in-repo, 3 role-match via external instrument + 06-03 evidence precedent)

**Tracked-source gate (#3645):** every in-repo analog named below was verified with `git ls-files` (all print 1). The lab bench (`/tmp/gsd-research/compiler-lab/bench/lab.py`) is an **external instrument** — a clone of a *separate* repo (compiler-lab @ `08cef10`), not a path in this repo and not a copy-from source for committed code. Its shapes are quoted only as the invocation target for the throwaway `/tmp/p7` scripts whose committed artifact is the evidence record, never the script (06-03 precedent: `parity-evidence-06-03.json` is tracked, the 06-03 runner script was not).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/adapters/plugin/builtin_run.go` (modify) | service (WASM plugin runtime scope) | event-driven/concurrent | itself + `internal/adapters/plugin/builtin_impl.go` (fix is in-place serialization) | exact |
| `internal/adapters/plugin/builtin_impl.go` (modify — likely no edit needed, see note) | service (WASM plugin execution) | event-driven/concurrent | `internal/adapters/plugin/builtin_run.go` | exact |
| `internal/adapters/plugin/builtin_run_test.go` (modify — race regression pin) | test (concurrency) | event-driven/concurrent | existing tests in the same file (`synctest` + goroutine + `runtimeFactory` injection) | exact |
| `/tmp/p7/compare.py` (create, throwaway, NOT committed) | utility (filter/projection/Counter diff) | batch | RESEARCH Pattern 2 (live-validated this session); evidence contract = `parity-evidence-06-03.json` | role-match |
| `/tmp/p7/perf.py` (create, throwaway, NOT committed) | utility (measurement loop) | batch | external `lab.py` `timed()` + `bench()` repetition/median shape | role-match |
| `/tmp/p7/run-parity.sh` (create, throwaway, NOT committed) | utility (orchestration runner) | batch | `internal/api/lint_link_matrix_test.go` (subprocess binary build/run pattern); `06-USER-SETUP.md` verification commands | role-match |
| `.planning/phases/07-.../parity-evidence-07.json` (create, committed) | config/evidence artifact | batch | `.planning/phases/06-.../parity-evidence-06-03.json` | exact |
| `.planning/phases/07-.../perf-evidence-07.md` (create, committed) | evidence doc | batch | external `lab.py bench` markdown report shape (`results/vs-buf-*.md`) + saved `results/lint/*/​*.meta.json` fields | role-match |

**Note on `builtin_impl.go`:** both runtime-creation call sites (`builtin_run.go:144` and `builtin_impl.go:143`) call the *same* injected field `run.runtimeFactory`, which is assigned only once, at `builtin_run.go:59`. A single wrap of the default factory there covers both sites; `builtin_impl.go` should need no edit. The planner should still read both sites (below) to confirm no third creation path exists.

**Carried, no modification:** `internal/rules/golden_full_registry_test.go` (automated-core Gate 1, exists post-5be9f7e), `internal/api/lint_stdout_golden_test.go`, Taskfile-equivalent SC3 commands (Pattern: SC3 battery below).

## Pattern Assignments

### `internal/adapters/plugin/builtin_run.go` (service, event-driven/concurrent) — MODIFY

**Analog:** the file itself; fix inserts a package-level serialized default at the existing injection point.

**Imports pattern** (lines 3-15) — note `sync` is already imported; no new imports needed for a mutex fix:
```go
import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)
```

**The injection seam the fix must preserve** (lines 21-35 and 53-61): the `runtimeFactory` field is the test double seam — tests override it directly (`builtin_run_test.go:289`, `:335`, `:364`). The fix must wrap **only the default**, never change the field mechanism:
```go
type BuiltinRun struct {
	ctx            context.Context
	cancel         context.CancelFunc
	cache          wazero.CompilationCache
	config         wazero.RuntimeConfig
	runtimeFactory func(context.Context, wazero.RuntimeConfig) wazero.Runtime
	// ... mu sync.Mutex; closed bool; compiler wazero.Runtime; modules map[...]; work sync.WaitGroup; close sync.Once; closeErr error
}

// inside EnsureBuiltinRun (lines 53-61):
	run := &BuiltinRun{
		ctx: commandCtx, cancel: cancel, cache: cache,
		config: wazero.NewRuntimeConfig().
			WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads).
			WithCloseOnContextDone(true).
			WithCompilationCache(cache),
		runtimeFactory: wazero.NewRuntimeWithConfig,   // ← LINE 59: the fix wraps THIS default
		modules:        make(map[[sha256.Size]byte]*builtinCompilation),
	}
```

**Creation site 1** (lines 140-148, inside `compiledModule`, guarded by the *per-instance* `r.mu` — not global, hence the cross-instance race):
```go
	compilation, ok := r.modules[key]
	if !ok {
		compilation = &builtinCompilation{ready: make(chan struct{})}
		r.modules[key] = compilation
		if r.compiler == nil {
			r.compiler = r.runtimeFactory(r.ctx, r.config)   // ← LINE 144
		}
		r.work.Add(1)
		go r.compile(binary, compilation)
	}
```

**Suggested fix shape** (matches house style — English godoc, no new exports, keeps the seam):
```go
// runtimeCreationMu serializes wazero runtime construction: wazero v1.9.0's
// internal/version.GetWazeroVersion() writes a package global on first use
// and is not safe under concurrent NewRuntimeWithConfig calls (race reported
// from builtin_run.go:144 and builtin_impl.go:143).
var runtimeCreationMu sync.Mutex

func newRuntimeSerialized(ctx context.Context, config wazero.RuntimeConfig) wazero.Runtime {
	runtimeCreationMu.Lock()
	defer runtimeCreationMu.Unlock()
	return wazero.NewRuntimeWithConfig(ctx, config)
}
// then: runtimeFactory: newRuntimeSerialized,   // line 59
```

**Error-handling pattern to not break** (lines 90-104): cleanup joins errors with `errors.Join` and wraps with callee-name only (`fmt.Errorf("Close: %w", err)`):
```go
		err := module.Close(cleanupCtx)
		if err != nil {
			r.closeErr = errors.Join(r.closeErr, fmt.Errorf("Close: %w", err))
		}
```

---

### `internal/adapters/plugin/builtin_impl.go` (service, event-driven/concurrent) — MODIFY-POSSIBLE (no edit expected)

**Analog:** `internal/adapters/plugin/builtin_run.go`

**Creation site 2** (line 143 — routes through the same `run.runtimeFactory` field, so the one-point fix covers it):
```go
	ctx = experimental.WithMemoryAllocator(ctx, allocator.NewNonMoving())
	rt := run.runtimeFactory(ctx, run.config)   // ← LINE 143
```

**Cleanup semantics the fix must not disturb** (lines 106-121 and 145-152): owner close joins errors and zeroes output on failure; never a bare `defer Close()`:
```go
	ctx, owner := EnsureBuiltinRun(ctx)
	if owner != nil {
		defer func() {
			err := owner.Close(context.WithoutCancel(ctx))
			if err != nil {
				output = nil
				runErr = errors.Join(runErr, fmt.Errorf("Close: %w", err))
			}
		}()
	}
```

---

### `internal/adapters/plugin/builtin_run_test.go` (test, concurrency) — MODIFY (race regression pin)

**Analog:** the file itself — `TestBuiltinRunCompilesOnceAndIsolatesInvocations` is the live-verified race reproducer under a *package-level* `-race` run (the race only surfaces with two tests' runtimes in flight; single-test isolation is clean — planner must pin at package scope, see Shared Pattern "Regression pin").

**Concurrency test style to copy** (lines 23-66, abridged): `t.Parallel` + `synctest.Test` + goroutines + `atomic` counters + function-listener fixture:
```go
func TestBuiltinRunCompilesOnceAndIsolatesInvocations(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		const count = 4
		var compilations atomic.Int64
		// ... experimental.WithFunctionListenerFactory fixture ...
		ctx, run := EnsureBuiltinRun(ctx)
		// ...
		for i := range count {
			go func() {
				output, err := runBuiltinModule(ctx, module, []string{"echo"}, strings.NewReader(input))
				results <- result{input: input, output: output, err: err}
			}()
		}
		synctest.Wait()
		require.Len(t, started, count, "all invocations must overlap")
```

**The seam-override cells the fix must keep green** (lines 289-294; same at :335-340, :364-367):
```go
	run.runtimeFactory = func(ctx context.Context, config wazero.RuntimeConfig) wazero.Runtime {
		return &builtinRuntimeProbe{
			Runtime: wazero.NewRuntimeWithConfig(ctx, config), compilations: &compilations,
			closes: &runtimeCloses, closeContexts: closeContexts, closeErr: runtimeCloseErr,
		}
	}
```

**Test-double pattern** (lines 420-437): embedded interface + probe counters, `require`/`assert` split:
```go
type builtinRuntimeProbe struct {
	wazero.Runtime
	compilations  *atomic.Int64
	closes        *atomic.Int64
	closeContexts chan<- error
	closeErr      error
}

func (r *builtinRuntimeProbe) CompileModule(ctx context.Context, binary []byte) (wazero.CompiledModule, error) {
	r.compilations.Add(1)
	return r.Runtime.CompileModule(ctx, binary)
}
```

**Pin shape:** one new test creating **two independent `EnsureBuiltinRun` scopes concurrently** (the race needs two `NewRuntimeWithConfig` from different instances/paths) + the SC3 gate `go test -race -count=1 ./internal/adapters/plugin/`. Follow go-testing skill: `t.Parallel()`, per-case deps, no shared mutable state.

---

### `/tmp/p7/compare.py` (utility, batch; throwaway — output is the committed artifact)

**Analog:** RESEARCH Pattern 2 (live-validated 2026-11-21 on the mini fixture); evidence contract = `.planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/parity-evidence-06-03.json`.

Copy verbatim from RESEARCH Pattern 2 (already proven against real binaries this session):
- `NEW3 = {"PACKAGE_NO_IMPORT_CYCLE", "SYNTAX_SPECIFIED", "FIELD_NOT_REQUIRED"}` filter (MINIMAL auto-seed makes the new binary emit findings main cannot — unfiltered compare false-fails).
- Gate B assert inside `load()`: `assert not any(r["RuleName"] == "COMPILE" for r in rows)` — a COMPILE issue means the linked engine ran zero rules (spike: 19-finding false diff).
- Projection `key = (r["RuleName"], r["Path"], r["Position"]["Line"], r["Position"]["Column"])`; `collections.Counter` multiset compare both directions; Gate D = `Counter(key for r in rows if r["RuleName"] in NEW3)` reported per corpus (expected 0, nonzero is attributed, not fatal).
- JSON input rows are the lab snapshot shape: keys `Message, Path, Position{Column, Filename, Line, Offset}, RuleName, SourceName`.

Raw-line Counter diff is the reported diagnostic (not the gate): any nonzero raw-line delta with Gate C passing needs an explicit per-line disposition (Message/Offset drift).

---

### `/tmp/p7/perf.py` (utility, batch; throwaway)

**Analog:** external `/tmp/gsd-research/compiler-lab/bench/lab.py` — copy the shape, do not modify the stand.

**Timing helper to copy** (`lab.py` lines 224-232):
```python
def timed(argv, cwd, accepted):
    """Runs argv under /usr/bin/time -l; returns (stdout, seconds, peak_rss_mb)."""
    start = time.perf_counter()
    proc = subprocess.run(["/usr/bin/time", "-l", *argv], cwd=cwd, capture_output=True, text=True)
    elapsed = time.perf_counter() - start
    if proc.returncode not in accepted:
        raise RuntimeError(f"{argv[0]} in {cwd} exited {proc.returncode}:\n{proc.stderr[-2000:]}")
    rss = int(re.search(r"(\d+)\s+maximum resident set size", proc.stderr).group(1)) / 2**20
    return proc.stdout, elapsed, rss
```

**Repetition/median structure to copy** (`lab.py` lines 88-96, from `bench`):
```python
            samples = []
            for i in range(args.warmup + args.runs):
                # ... one sample ...
                if i >= args.warmup:
                    samples.append(sample)
            times = [s[0] for s in samples]
            row[tool] = {"median_s": statistics.median(times), "min_s": min(times), "max_s": max(times),
                         "rss_mb": statistics.median(s[1] for s in samples) / 2**20, ...}
```

**Adaptation the phase loop must add (lab.py does NOT do it for lint):** `lab.py bench`'s TOOLS are `["buf", "easyp", "easyp-protoc"]` measuring descriptor-set builds (`argv_for`, line 50) — it cannot measure `easyp` vs `easyp-baseline` lint. The phase loop uses `timed()` with the lint argv `["<binary>", "--format", "json", "lint"]` (cwd = corpus root; accepted = `{0, 1}`) and **interleaves** (m,n,m,n,...) — `bench` runs all runs per tool back-to-back. Also cross-check `re.search(r'msg="lint completed" issues=(\d+)', proc.stderr)` against parsed issue count (lab.py lint lines 178-181) and require equal issue counts across all runs of a binary.

---

### `/tmp/p7/run-parity.sh` (utility, batch; throwaway runner)

**Analog (in-repo, tracked):** `internal/api/lint_link_matrix_test.go` — the house pattern for building the CLI binary and running it as a subprocess with observable exit codes.

**Build-once-with-provenance pattern** (lines 155-176):
```go
// buildEasypBinary compiles ./cmd/easyp once per test run into a temp dir and
// returns the binary path. The binary lives outside the repository and is never
// committed; ...
var buildEasypBinary = sync.OnceValues(func() (string, error) {
	// ...
	build := exec.Command("go", "build", "-o", binPath, "./cmd/easyp")
	build.Dir = repoRoot
	if out, buildErr := build.CombinedOutput(); buildErr != nil {
		return "", fmt.Errorf("go build: %w: %s", buildErr, out)
	}
	return binPath, nil
})
```

**Subprocess run with exit-code capture** (lines 178-199) — the mini-fixture Gate A–D rehearsal uses the same shape (child gets its own `Dir`/`Env`; parent never mutated):
```go
func runEasypLint(t *testing.T, binPath, dir string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binPath, "lint", "--path", ".")
	cmd.Dir = dir
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
	return stdout.String(), stderr.String(), code
}
```

**Runner order constraints (external lab.py, live-verified):** baseline via `git worktree add <tmp> 1e9a596 && go build` then `git worktree remove` immediately (no registration leak); **setup.sh first, phase binaries second** — `lab.py main()` (lines 366-368) hard-exits unless all five `bin/{buf,easyp,easyp-baseline,easyp-protoc,fdsdump}` exist, and re-running setup.sh overwrites the two phase binaries:
```python
    for name in ("buf", "easyp", "easyp-baseline", "easyp-protoc", "fdsdump"):
        if not os.path.exists(os.path.join(BIN, name)):
            sys.exit(f"missing {os.path.join(BIN, name)}; run bench/setup.sh")
```
`lint` accepts only `--bin easyp|easyp-baseline` (argparse choices, line 351). Gate order per corpus: A (main-r1 vs main-r2 → 0/0) → B (zero COMPILE) → C (4-tuple multiset after NEW3 filter) → D (delta accounting); ×2 runs per binary; plain (no `-race`) builds for parity AND perf.

**Mini-corpus fixture (in-repo, tracked):** `internal/rules/golden_full_registry_test.go` lines 463-473 show the corpus-from-map pattern (`t.TempDir()` + file map) used to validate the runner without corpora; the runner fixture is the same `testdata/{import_cycle,no_syntax,required_field}` trees + `easyp.yaml` STANDARD the spike used.

---

### `.planning/phases/07-.../parity-evidence-07.json` (evidence artifact, batch; committed)

**Analog (exact):** `.planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/parity-evidence-06-03.json` (47 lines, tracked). Copy the whole block structure:

```json
{
  "lab":     { "checkout_sha": "08cef1...", "binary_path": "/tmp/p6-parity/bin/easyp-lab", "exit_code": 1, "finding_count_total": 9 },
  "new":     { "commit_sha": "d0525b...", "binary_path": "/tmp/p6-parity/bin/easyp-new", "exit_code": 1, "finding_count_total": 14 },
  "corpus": ["easyp.yaml", "import_cycle/a.proto", "..."],
  "config": "version v1 / linters default STANDARD",
  "projection": "(RuleName, Path, Position.Line, Position.Column)",
  "per_rule_counts": { "PACKAGE_NO_IMPORT_CYCLE": { "lab": 2, "new": 2 }, "...": {} },
  "multiset_equal": true,
  "teeth_anti_vacuity": ">=1 per rule both sides",
  "corpora_leg": {
    "status": "deferred",
    "reason": "... not accessible ...",
    "clone_attempts": ["gh repo clone ... -> GraphQL: Could not resolve to a Repository", "..."],
    "options_presented": ["PROVIDE: ...", "PROVIDE: ...", "DEFER: ..."],
    "choice": "deferred (...)",
    "carries_requirement": "..."
  },
  "timestamp": "2026-11-21T00:20:38Z"
}
```

**Phase-7 deltas on this shape:** `new` block gains a `baseline` block (commit `1e9a596`, build command, `go version`, no-race); per-corpus gate states `gate_a/b/c/d` with main-vs-main first; rule counts stated from registries as **41 (main) / 44 (new)** — never "43"; the `corpora_leg` block carries the explicit anti-ratification state: `corpora: UNREACHABLE (gh+SSH, 2026-11-21 spike) — leg pending user unblock` unless actually fired.

---

### `.planning/phases/07-.../perf-evidence-07.md` (evidence doc, batch; committed)

**Analog (role-match):** `lab.py bench` report writer (lines 103-118) — markdown table with median (min-max), ratio vs baseline, RSS; plus the meta fields `{"bin", "wall_s", "rss_mb", "issues", "rc"}` written by `lab.py lint` (lines 186-189).

```python
    lines = [f"# buf vs easyp {stamp}", "",
             f"runs={args.runs} warmup={args.warmup}. Wall time median (min-max), CPU user/sys median, "
             "peak RSS median. ...", "",
             "| corpus | tool | wall | vs buf | user | sys | RSS |", "|---|---|---|---|---|---|---|"]
    for r in rows:
        base = r["buf"]["median_s"]
        for tool in TOOLS:
            t = r[tool]
            lines.append(f"| {r['corpus']} ({r['files']}) | {tool} | {t['median_s']:.2f}s "
                         f"({t['min_s']:.2f}-{t['max_s']:.2f}) | {base / t['median_s']:.2f}x | ...")
```

Phase-7 report must add: binary provenance (SHAs `1e9a596`/HEAD, build commands, no `-race`), machine context (CPU model, `nproc`, load), corpus identity (92/7414 file counts), per-run table (all 5 measured + warmup noted), gate `median(new) <= median(main)` on googleapis-nocompute (api-common recorded, not gated), and the explicit corpora-gate state if unfired.

---

## Shared Patterns

### Runtime-creation serialization (the only production-code change)
**Source:** `internal/adapters/plugin/builtin_run.go:59` (default assignment), `:144` (site 1), `internal/adapters/plugin/builtin_impl.go:143` (site 2)
**Apply to:** the race fix only; both sites route through one field — wrap the default, keep the seam.
**House sync style to mirror:** `builtin_run.go:28,33` (`mu sync.Mutex`, `close sync.Once`) and `internal/api/lint_link_matrix_test.go:159` (`sync.OnceValues`).

### Error wrapping and deferred Close
**Source:** `.agents/skills/go-code-style/SKILL.md` + `.spec/agent-rules.md` (mandatory); exemplars `builtin_run.go:92-104`, `builtin_impl.go:108-115,146-152`
**Apply to:** all phase-added Go code.
```go
err := module.Close(cleanupCtx)
if err != nil {
    r.closeErr = errors.Join(r.closeErr, fmt.Errorf("Close: %w", err))
}
```
Callee-name-only `%w` wraps; never bare `defer x.Close()`; `errors.Join` for multi-cleanup.

### Test conventions for the regression pin
**Source:** `.agents/skills/go-testing/SKILL.md` + `.spec/agent-rules.md`; exemplars `builtin_run_test.go:23-98` (synctest + goroutines + atomics), `:420-448` (embedded-interface probe doubles)
**Apply to:** any new/modified test file.
`t.Parallel()` top-level and isolated subtests; `require` fatal / `assert` independent; named table cases with lowercase underscore names; per-case mutable deps; process-state behavior (real exit codes, concurrent binaries) stays subprocess-based (`exec.Cmd.Dir/Env`) or package-scope `-race`, never parent mutation.

### Zero-COMPILE purity gate (anti-vacuity) before any compare
**Source (in-repo proof of Gate B):** `internal/rules/golden_full_registry_test.go:492-516`
**Apply to:** compare.py load path and the golden carried gate.
```go
	for _, iss := range got {
		require.NotEqual(t, core.CompileRuleName, iss.RuleName,
			"the corpus must compile under the linked engine's full type resolution: a COMPILE finding vacates the rule-finding golden (Pitfall 7)")
	}
	require.Len(t, got, goldenFindingCount, "pinned old-engine finding count, asserted BEFORE the golden compare")
```
(plus the coverage invariant: every registry name fires at least once; and `requireTotalOrder`, lines 479-490, halting on any five-term tie instead of papering over it.)

### Multiset compare with sort-neutralized order noise
**Source:** external `lab.py` `lint` (line 180: `issues = sorted(line for line in proc.stdout.splitlines() if line.startswith("{"))`) and `lint_diff` (lines 200-206: `only_base, only_new = sorted((base - new).elements()), sorted((new - base).elements())`); Counter projection in RESEARCH Pattern 2
**Apply to:** compare.py, runner; never compare unsorted streams (main's emission order is nondeterministic).

### Evidence provenance + honest checkpoint disposal
**Source:** `.planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/parity-evidence-06-03.json` (`corpora_leg` block: status/reason/clone_attempts/options_presented/choice/carries_requirement) + `.planning/phases/06-novye-pravila-i-regeneratsiya-shem-mcp/06-USER-SETUP.md` (env-var table, one-command verification path with expected results)
**Apply to:** parity-evidence-07.json, perf-evidence-07.md, and the checkpoint item. A corpora leg "passed" without corpora does not exist — evidence states the gate explicitly. The unblock nuance to propagate: `GOOGLEAPIS` must be a checkout **compatible with the args lists** (a fresh public googleapis clone; the stale yadro 2024-02 checkout has 3889 protos vs required 7416).

### SC3 battery (direct equivalents — `task` absent)
**Source:** `Taskfile.yml` (tracked; `test:` at line 66 with gotestsum precondition, `lint:go:` at line 83, `schema:check:` at line 57); RESEARCH Pattern 4 verbatim commands:
```bash
GOBIN="$PWD/bin" go install gotest.tools/gotestsum@v1.13.0   # or plain: go test -race -count=1 ./...
bin/gotestsum --format pkgname -- -coverprofile=coverage.out -race -count=1 ./...
bin/golangci-lint run ./...
tmpd=$(mktemp -d) && go run ./cmd/easyp schema-gen --out-dir "$tmpd" && diff -r schemas "$tmpd"; rm -rf "$tmpd"
```
**Apply to:** phase gate. `coverage.out` is gitignored — never commit; hadolint leg honestly recorded as environment-unavailable.

## No Analog Found

No file is left without guidance, but three throwaway scripts have **no in-repo tracked code analog** (by design — 06-03 precedent: the committed artifact is the evidence record, not the script):

| File | Role | Data Flow | Reason + substitute |
|------|------|-----------|---------------------|
| `/tmp/p7/compare.py` | utility | batch | No committed diff-script exists in repo; the design was live-validated 2026-11-21 (RESEARCH Pattern 1 evidence) and the evidence contract is `parity-evidence-06-03.json` |
| `/tmp/p7/perf.py` | utility | batch | No measurement loop in repo; copy `lab.py timed()`/`bench()` shape (external instrument) + add interleaving |
| `/tmp/p7/run-parity.sh` | utility | batch | No orchestration script committed; in-repo subprocess/build pattern lives in `internal/api/lint_link_matrix_test.go`; command sequence in RESEARCH "Corpus legs" |

## Metadata

**Analog search scope:** `internal/adapters/plugin/`, `internal/rules/`, `internal/api/`, `internal/core/`, `.planning/phases/{02,05,06}-*/`, `Taskfile.yml`, `.agents/skills/{go-code-style,go-testing}/SKILL.md`, `.spec/agent-rules.md`; external instrument `/tmp/gsd-research/compiler-lab/bench/` (existence + line cites re-verified 2026-11-21)
**Files scanned:** ~20 (12 read in full or in targeted sections; all named analogs `git ls-files`-verified tracked)
**Pattern extraction date:** 2026-11-21

**Corrections/additions vs RESEARCH.md the planner should know:**
1. `lab.py` has more verbs than RESEARCH described: `bench` (`--runs`/`--warmup`, median/min/max/RSS per tool), `parity` (buf-vs-easyp descriptor dumps), `lint-buf`, `breaking-buf`. This does NOT change the perf design — `bench`'s TOOLS are `["buf","easyp","easyp-protoc"]` measuring descriptor-set builds, not lint — but the repetition/median machinery is copyable instead of hand-rolled, and only **interleaving** remains phase-added.
2. `lab.py lint` already records `{"bin", "wall_s", "rss_mb", "issues", "rc"}` meta per run and cross-checks stderr `msg="lint completed" issues=N` — the perf loop and evidence can reuse these fields verbatim.
3. The wazero fix is a **one-point change**: both racing call sites (`builtin_run.go:144`, `builtin_impl.go:143`) call the same `run.runtimeFactory` field assigned once at `builtin_run.go:59`; wrap only that default so the test-injection seam (`builtin_run_test.go:289/335/364`) stays intact.
4. `lab.py main()` (lines 366-368) requires all five bins (`buf`, `easyp`, `easyp-baseline`, `easyp-protoc`, `fdsdump`) even for a lint-only session — runner must invoke setup.sh first, then overwrite the two phase binaries.
