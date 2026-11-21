# Phase 1: Merge main и примирение путей - Pattern Map

**Mapped:** 2026-11-21
**Files analyzed:** 10 merge-affected paths (3 conflicted, 4 auto-merged/adopted, 2 module files, 1 gate surface) + 4 gate/reference styles
**Analogs found:** 10/10 (no greenfield source files; the phase's "analogs" are the two pre-merge parents and the repo's existing verification idioms)

## Read this first

This is a **merge-only phase**. No new production code is written; every source file either comes from `origin/main` verbatim, stays on the branch verbatim, or is a two-line import union. The planner must treat the paths below as **merge-resolution targets**, not as files to re-implement:

- **main wins** (D-03): `internal/core/breaking_profiles.go`, all of `internal/sourceview/*`, `internal/api/breaking_v1.go`, `go.mod`/`go.sum` pins.
- **branch wins** (D-03): `internal/core/dom.go` and the whole lint contour (`internal/rules`, `internal/easypc`, `internal/protoast`, `internal/core/lint_*`, `link.go`, `proto_info_read.go`) — these must not appear in the merge diff.
- **Union**: `internal/core/core.go` (both sides add disjoint state; only the import block conflicts).

State observed at mapping time (re-verify before execution):

- `HEAD` = `2acac25cbe87b707dbb449e0b803547e42f8d9e6` on `feat/lint-easypc-infra`, working tree clean (`git status --porcelain --untracked-files=no` empty). The pre-flight docs commit from RESEARCH step 1a already exists as `9408042 docs: archive v1.0 phase artifacts; sync v1.1 state`; the dirty-index blocker is therefore gone, but the clean-tree gate must still be asserted.
- `origin/main` = `5d36266d71e85f8f0fc3fa41f787a03ee5c68a79` (#253).
- `git merge-tree --write-tree HEAD origin/main` still yields exactly three conflicts: `breaking_profiles.go` (ours `8a3a10d`, theirs `6e45a05`), `core.go` (ours `9b63472`, theirs `a2419c8`), `dom.go` (ours `52bfbd7`, theirs `3a35dc5`).
- The three rehearsed resolution blobs still exist in this repository's object database (verified with `git cat-file -t`): `6e45a055c1e12efb25d14d210fdfa5ee0c52ba90` (main's `breaking_profiles.go`), `52bfbd7d645a4c2d587ea84d08c9bbf27e7e0f67` (branch's `dom.go`), `34987d1824df1c4c14337df4b04c2625761e7b65` (union `core.go`). The first two are exactly the parents' blobs, so the producer gates are `git rev-parse origin/main:<path>` / `git rev-parse HEAD:<path>`; the third is an optional restore/verification shortcut, valid only while both parents are unchanged.

Tracked-source verification: every analog path below is git-tracked (`git ls-files` on the branch) or tracked in `origin/main`'s tree (`git ls-tree origin/main`). No gitignored mirrors are referenced.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/core/breaking_profiles.go` (CONFLICT → take `origin/main`) | service (breaking engine) | file-I/O + transform (proto sources → descriptors → issues) | `origin/main:internal/core/breaking_profiles.go` (blob `6e45a05`, 1265 lines) | exact (side-select; branch delta subsumed) |
| `internal/core/dom.go` (CONFLICT → keep branch) | model | transform | `HEAD:internal/core/dom.go` (blob `52bfbd7`, 181 lines) | exact (side-select; main delta rejected) |
| `internal/core/core.go` (CONFLICT → union) | provider/config | request-response (constructor wiring) | union blob `34987d1`; parents `HEAD:core.go` (`9b63472`) and `origin/main:core.go` (`a2419c8`) | exact (two-line import union; rest auto-merges) |
| `go.mod` (MODIFY, MVS) | config | transform | `origin/main:go.mod` (delta: +easyprotoc pin, protobuf rc.1) | exact (take main's pins) |
| `go.sum` (MODIFY, MVS) | config | transform | `origin/main:go.sum` | exact (take main's sums) |
| `internal/api/breaking_v1.go` (MODIFY, auto-merge) | controller / command wiring | request-response + file-I/O | `origin/main:internal/api/breaking_v1.go` (session lifecycle at 74-79, 155, 306-315, 479) | exact (verify byte-equal to main) |
| `internal/sourceview/local.go` (MODIFY, auto-merge) | utility / library | file-I/O | `origin/main:internal/sourceview/local.go` (`OpenLocalSelected` 27-37, `openSelected` 39-60) | exact (branch never touched it) |
| `internal/sourceview/root_tree.go` (NEW from main) | utility | file-I/O | `origin/main:internal/sourceview/root_tree.go`; conceptual predecessor: branch `local.go` using raw `*os.Root` | exact upstream (no branch-side analog) |
| `internal/sourceview/session.go` (NEW from main) | store/provider (per-command root cache) | file-I/O | `origin/main:internal/sourceview/session.go`; same package's `OpenLocalSelected` package function is the nil-session fallback | exact upstream (no branch-side analog) |
| `internal/sourceview/session_test.go` (NEW from main) | test | file-I/O | `origin/main:internal/sourceview/session_test.go`; repo Go-test style in `internal/protoast/position_test.go` | role-match upstream; contains a pre-existing errcheck finding |
| *(no new file)* merge/ancestry/golden gates | test/integration | request-response (subprocess) + shell | `scripts/validate-release-tag.sh` (shell ancestry), `internal/api/acceptance_gates_test.go`, `internal/rules/golden_full_registry_test.go` | exact gate style |

## Pattern Assignments

### `internal/core/breaking_profiles.go` (service, file-I/O + transform)

**Analog:** `origin/main:internal/core/breaking_profiles.go` — take wholesale. Do **not** hand-merge: the branch's entire delta to this file is two hunks and both are subsumed by the branch's `dom.go` alias (RESEARCH Pattern 1, verified):

```diff
-	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
 	"google.golang.org/protobuf/reflect/protoreflect"
 	...
+	"github.com/easyp-tech/easyp/internal/protoast"
-			Issue: Issue{RuleName: rule, Message: message, Position: meta.Position{
+			Issue: Issue{RuleName: rule, Message: message, Position: protoast.Position{
```

Main spells the same position as `Position{...}` (line 373), and `type Position = protoast.Position` in the kept `dom.go` makes that identical to the branch's spelling.

**Imports pattern** (main lines 3-21) — copy exactly:

```go
import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/easyp-tech/easyprotoc/protoc"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)
```

**Core compile pattern** (main lines 122-139, `compiled, err := protoc.Compile(...)`):

```go
compiled, err := protoc.Compile(ctx, targets,
	protoc.WithResolver(breakingImportResolver{
		importPaths: imports,
		open:        c.breakingSourceAccessor(ctx, root),
	}),
	protoc.WithStandardImports(true),
	protoc.WithIncludeImports(true),
	protoc.WithIncludeSourceInfo(true),
	protoc.WithErrorHandler(func(diagnostic protoc.Diagnostic) {
		if diagnostic.IsWarn {
			return
		}
		diagnosticsMu.Lock()
		diagnostics = append(diagnostics, diagnostic.String())
		diagnosticsMu.Unlock()
	}),
)
```

**Resolver pattern** (main lines 161-203) — implements `protoc.Resolver`, preserves import-path order, maps `os.ErrNotExist` to `protoc.ErrFileNotFound`:

```go
type breakingImportResolver struct {
	importPaths []string
	open        func(string) (io.ReadCloser, error)
}

func (r breakingImportResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(r.importPaths) == 0 {
		return r.read(path)
	}
	for _, importPath := range r.importPaths {
		data, err := r.read(filepath.Join(importPath, path))
		if errors.Is(err, protoc.ErrFileNotFound) {
			continue
		}
		return data, err
	}
	return nil, protoc.ErrFileNotFound
}
```

**Source accessor / session coupling** (main lines 205-224):

```go
func (c *Core) breakingSourceAccessor(ctx context.Context, root string) func(string) (io.ReadCloser, error) {
	return func(filename string) (io.ReadCloser, error) {
		if c.importFileAllowed != nil && !c.importFileAllowed(filename) {
			return nil, &os.PathError{Op: "open", Path: filename, Err: os.ErrNotExist}
		}
		if c.sourceFileOpen != nil {
			reader, err := c.sourceFileOpen(filename)
			if err == nil || !errors.Is(err, os.ErrNotExist) {
				return reader, err
			}
		}
		relative, err := filepath.Rel(root, filename)
		if err == nil && filepath.IsLocal(relative) {
			return c.sourceSession.OpenLocal(ctx, root, relative)
		}
		return c.openSourceFile(filename)
	}
}
```

**Gate (resolution evidence, not tests):**

```bash
git checkout origin/main -- internal/core/breaking_profiles.go
test "$(git rev-parse :internal/core/breaking_profiles.go)" = "$(git rev-parse origin/main:internal/core/breaking_profiles.go)"
test -z "$(git diff origin/main -- internal/core/breaking_profiles.go)"
grep -q 'easyprotoc/protoc' internal/core/breaking_profiles.go
```

**Anti-pattern:** `git checkout --ours -- internal/core/breaking_profiles.go` compiles and can pass branch tests while silently dropping #253 (RESEARCH Pitfall 2). Gate by blob/diff, never by green build.

---

### `internal/core/dom.go` (model, transform)

**Analog:** `HEAD:internal/core/dom.go` — keep wholesale (blob `52bfbd7`, byte-identical to pre-merge). Main's only delta is the alias we reject (`type Position = meta.Position`, main lines 14-15).

**Branch pattern to preserve** (branch `dom.go:15-19`):

```go
// Position is a source position carried by issues. It is an alias of
// protoast.Position so lint findings already speak the type the future engine
// emits; go-protoparser values are converted at the AppendIssue boundary via
// fromMetaPos.
type Position = protoast.Position
```

**Conversion boundary** (branch `dom.go:125-135`) — this is what makes main's `Position{...}` spelling compatible:

```go
// fromMetaPos converts a go-protoparser source position to the issue position
// type. The field sets are identical in name, type, and order, so the
// conversion copies every field and changes no value.
func fromMetaPos(p meta.Position) Position {
	return Position{
		Filename: p.Filename,
		Offset:   p.Offset,
		Line:     p.Line,
		Column:   p.Column,
	}
}
```

**Gate:**

```bash
git checkout "$PRE_MERGE_BASE" -- internal/core/dom.go
test "$(git rev-parse :internal/core/dom.go)" = "$(git rev-parse "$PRE_MERGE_BASE":internal/core/dom.go)"
grep -n 'type Position = protoast.Position' internal/core/dom.go
```

**Anti-patterns** (RESEARCH Pitfall 3): taking `type Position = meta.Position` breaks every `core.AppendIssue(..., protoast.Position, ...)` call site; declaring both aliases is a duplicate-declaration compile error.

---

### `internal/core/core.go` (provider/config, request-response)

**Analog:** union blob `34987d1824df1c4c14337df4b04c2625761e7b65`; parents are `HEAD:internal/core/core.go` (adds `importCache`/`parsedImports`, branch lines 45-55) and `origin/main:internal/core/core.go` (adds `sourceSession`/`SourceSession`, main lines 13, 30, 69-71, 101). `git merge` auto-merges everything except the import block, so the hand edit is exactly two import lines.

**Union import block** (union blob lines 4-16):

```go
import (
	"errors"
	"io"
	"maps"
	"slices"
	"sync"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/internal/sourceview"
)
```

**Union state** (union blob lines 29-57) — both sides coexist:

```go
	sourceFileOpen        func(string) (io.ReadCloser, error)
	sourceSession         *sourceview.Session
	sourceFilesWalk       func(string, func(string) error) (bool, error)
	...
	// importCache holds parsed import outlines for lint; files are parsed in
	// parallel, so it is shared and locked.
	importCache *parsedImports
}

// parsedImports holds parsed import outlines for lint; files are parsed in
// parallel, so it is shared and locked.
type parsedImports struct {
	mu    sync.Mutex
	files map[string]*protoast.File
}
```

**Option + constructor wiring** (union blob lines 82-84, 114):

```go
	// SourceSession, if set, is used for local opens that bypass
	// OpenSourceFile, so they share the command's open roots.
	SourceSession *sourceview.Session
	...
		sourceSession:           options.SourceSession,
```

**Gate:** hand-merge the two imports, then assert the union blob:

```bash
# optional deterministic shortcut once parents are confirmed unchanged:
# git cat-file -p 34987d1824df1c4c14337df4b04c2625761e7b65 > internal/core/core.go
test "$(git rev-parse :internal/core/core.go)" = "34987d1824df1c4c14337df4b04c2625761e7b65"
```

---

### `go.mod` / `go.sum` (config, MVS transform)

**Analog:** `origin/main:go.mod` / `origin/main:go.sum`; no hand edits (D-05). Expected delta (main's commit):

```diff
 	github.com/charmbracelet/x/term v0.2.1
+	github.com/easyp-tech/easyprotoc v0.0.0-20261104120806-f378f3f3a05f
 	github.com/easyp-tech/service/api v1.0.0
...
-	google.golang.org/protobuf v1.36.11
+	google.golang.org/protobuf v1.37.0-rc.1
```

**Gate:** `go list -m github.com/easyp-tech/easyprotoc google.golang.org/protobuf` prints the two expected versions; `go mod tidy -diff` is empty; `grep -q 'replace' go.mod` is false.

---

### `internal/api/breaking_v1.go` (controller, request-response + file-I/O)

**Analog:** `origin/main:internal/api/breaking_v1.go` — auto-merges; the branch's only delta in this file (relative to merge-base) is a comment, so verify byte-equality instead of editing. Main adds the per-command session and passes it into `core.Options`.

**Session lifecycle pattern** (main lines 74-79; identical shape at 306-315):

```go
	session := sourceview.NewSession()
	defer func() {
		err := session.Close()
		if err != nil {
			log.Error(ctx.Context, "Close", slog.Any("error", err))
		}
	}()
```

**Wiring pattern** (main line 155, repeated at 479):

```go
app := core.New(core.Options{Logger: log, ImportRoots: currentImports.Paths(), ImportFileAllowed: currentImports.FileAllowed(), OpenSourceFile: currentImports.OpenSourceFile, SourceSession: session, BreakingCheckConfig: core.BreakingCheckConfig{
```

**Gate:** `test -z "$(git diff origin/main -- internal/api/breaking_v1.go)"`.

---

### `internal/sourceview/local.go` (utility, file-I/O)

**Analog:** `origin/main:internal/sourceview/local.go` — auto-merge; branch never touched this file (`git diff HEAD origin/main` is main's #253 delta only). Key structural change: `localView` returns `*rootTree` instead of `*os.Root`, and the open body is factored so both the package function and `Session` share it.

**Boundary pattern** (main lines 27-60):

```go
func OpenLocalSelected(ctx context.Context, rootPath, logical string, selected func(Resolution) bool, nestedAllowed ...func(Resolution) bool) (fs.File, error) {
	root, view, err := localView(rootPath)
	if err != nil {
		return nil, fmt.Errorf("localView: %w", err)
	}
	file, err := openSelected(ctx, root, view, logical, selected, nestedAllowed...)
	if err != nil {
		return nil, errors.Join(err, root.Close())
	}
	return &localFile{File: file, root: root}, nil
}

// openSelected resolves and opens logical below root. The caller owns root.
func openSelected(ctx context.Context, root *rootTree, view *View, logical string, selected func(Resolution) bool, nestedAllowed ...func(Resolution) bool) (fs.File, error) {
	resolved, err := view.Resolve(ctx, filepath.ToSlash(logical))
	if err != nil {
		return nil, fmt.Errorf("Resolve: %w", localResolutionError(logical, resolved, err))
	}
	if err := checkLocalResolution(root.Lstat, resolved); err != nil {
		permitted := errors.Is(err, ErrNestedRepository) && len(nestedAllowed) > 0 && nestedAllowed[0] != nil && nestedAllowed[0](resolved)
		if !permitted {
			return nil, err
		}
	}
	...
```

**Gate:** `test -z "$(git diff origin/main -- internal/sourceview/local.go)"`.

---

### `internal/sourceview/root_tree.go` (utility, file-I/O; NEW from main)

**Analog:** `origin/main:internal/sourceview/root_tree.go` (main-only; no branch-side file). Take verbatim; do not "improve" it inside the merge.

**Pattern** (main lines 18-52) — per-directory handles opened from the parent handle, double-checked insertion:

```go
type rootTree struct {
	root *os.Root

	mu   sync.Mutex
	dirs map[string]*os.Root
}

func newRootTree(root *os.Root) *rootTree {
	return &rootTree{root: root, dirs: map[string]*os.Root{".": root}}
}

// dir returns the handle for a directory, opening missing ancestors.
func (t *rootTree) dir(name string) (*os.Root, error) {
	t.mu.Lock()
	handle, ok := t.dirs[name]
	t.mu.Unlock()
	if ok {
		return handle, nil
	}
	parent, err := t.dir(path.Dir(name))
	if err != nil {
		return nil, err
	}
	handle, err = parent.OpenRoot(path.Base(name))
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, ok := t.dirs[name]; ok {
		return existing, handle.Close()
	}
	t.dirs[name] = handle
	return handle, nil
}
```

---

### `internal/sourceview/session.go` (store/provider, file-I/O; NEW from main)

**Analog:** `origin/main:internal/sourceview/session.go` (main-only). Take verbatim.

**Pattern** (main lines 19-58, 65-80) — one root entry per canonical root, `sync.Once` initialization, nil-receiver fallback to the package functions:

```go
type Session struct {
	mu    sync.Mutex
	roots map[string]*sessionRoot
}

type sessionRoot struct {
	once sync.Once
	tree *rootTree
	view *View
	err  error
}

func NewSession() *Session {
	return &Session{roots: make(map[string]*sessionRoot)}
}

func (s *Session) root(rootPath string) (*rootTree, *View, error) {
	s.mu.Lock()
	entry, ok := s.roots[rootPath]
	if !ok {
		entry = &sessionRoot{}
		s.roots[rootPath] = entry
	}
	s.mu.Unlock()
	entry.once.Do(func() { entry.tree, entry.view, entry.err = localView(rootPath) })
	return entry.tree, entry.view, entry.err
}

func (s *Session) OpenLocalSelected(ctx context.Context, rootPath, logical string, selected func(Resolution) bool, nestedAllowed ...func(Resolution) bool) (fs.File, error) {
	if s == nil {
		return OpenLocalSelected(ctx, rootPath, logical, selected, nestedAllowed...)
	}
	root, view, err := s.root(filepath.Clean(rootPath))
	if err != nil {
		return nil, fmt.Errorf("localView: %w", err)
	}
	return openSelected(ctx, root, view, logical, selected, nestedAllowed...)
}
```

**Close pattern** (main lines 65-80) — join all root closes and reset the map:

```go
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for _, entry := range s.roots {
		if entry.tree != nil {
			errs = append(errs, entry.tree.Close())
		}
	}
	s.roots = nil
	return errors.Join(errs...)
}
```

---

### `internal/sourceview/session_test.go` (test, file-I/O; NEW from main)

**Analog:** `origin/main:internal/sourceview/session_test.go`. Take verbatim. This is upstream test style (std `t.Fatalf`, `t.TempDir`, `t.Context()`), not the testify style used by most branch tests:

```go
func TestSessionSharesRootsAndKeepsChecks(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	...
	session := NewSession()
	defer session.Close()
	for _, name := range []string{"one.proto", "two.proto"} {
		file, err := session.OpenLocal(t.Context(), root, "a/b/"+name)
		...
	}
	if len(session.roots) != 1 {
		t.Fatalf("roots = %d, want one shared root", len(session.roots))
	}
	if _, err := session.OpenLocal(t.Context(), root, "a/b/escape.proto"); !errors.Is(err, ErrOutsideRoot) {
		t.Fatalf("escape through the session = %v, want ErrOutsideRoot", err)
	}
}
```

**Important:** line 29 is main's pre-existing golangci-lint errcheck finding (`defer session.Close()`), which violates `.spec/agent-rules.md:10`. It is red on `origin/main` already and is **not** a Phase 1 gate (D-09). Do not fix it inside the merge commit (RESEARCH Pitfall 4/5); a separate follow-up commit is the only sanctioned place.

**Gate:** `test -z "$(git diff origin/main -- internal/sourceview)"`.

## Shared Patterns

### 1. Shell resolution/ancestry gates — analog `scripts/validate-release-tag.sh`

**Source:** `scripts/validate-release-tag.sh:1-27` (tracked)
**Apply to:** every shell `<verify>` block for D-09 ancestry, blob, and containment checks.

Repo shell-gate conventions: strict mode, explicit `git rev-parse --verify --quiet`, `git merge-base --is-ancestor`, human-readable failure to stderr, non-zero exit. Reuse this shape instead of ad-hoc one-liners:

```bash
#!/usr/bin/env bash
set -euo pipefail

if ! git rev-parse --verify --quiet "refs/tags/$tag^{commit}" > /dev/null; then
  echo "Release tag $tag does not resolve to a commit." >&2
  exit 1
fi

for branch in origin/main origin/v1.0; do
  if git rev-parse --verify --quiet "refs/remotes/$branch^{commit}" > /dev/null &&
     git merge-base --is-ancestor "refs/tags/$tag^{commit}" "refs/remotes/$branch"; then
    printf '%s\n' nightly
    exit 0
  fi
done

echo "Nightly tag $tag is not in origin/v1.0 or origin/main history." >&2
exit 1
```

Apply the same style to the phase gates:

```bash
test -z "$(git status --porcelain --untracked-files=no)"
git merge-base --is-ancestor origin/main HEAD
test "$(git rev-list --count HEAD..origin/main)" -eq 0
test -z "$(git grep -n -E '^(<<<<<<< |>>>>>>> )' -- '*.go')"
```

For byte-exact file/set comparison, the repo's other shell gate (`scripts/check-proto-example.sh:15-21`) uses `cmp` / `diff -r` with a temp dir + `trap`; use the same for "lint contour byte-identical" and "golden unchanged" evidence.

### 2. Pinned-oracle Go test gate — analog `internal/api/acceptance_gates_test.go`

**Source:** `internal/api/acceptance_gates_test.go:22-29`, `144-150`, `284-342` (tracked)
**Apply to:** `TestAcceptanceGates` (Gates A–D ×2).

The repo's gate idiom: a ratified/pinned constant with a comment explaining why, presence is the gate (never skip), then explicit named gates with diagnostic messages.

```go
// acceptanceBaselineCommit is the RATIFIED baseline oracle commit ... the oracle's
// presence is the gate — the test fails loudly when the commit is absent and never skips.
const acceptanceBaselineCommit = "1e9a596"
```

```go
check := exec.Command("git", "cat-file", "-e", acceptanceBaselineCommit+"^{commit}")
check.Dir = repoRoot
if out, checkErr := check.CombinedOutput(); checkErr != nil {
	return baselineBinary{}, fmt.Errorf("baseline oracle commit %s is absent (never skip): %w: %s", acceptanceBaselineCommit, checkErr, out)
}
```

```go
// Gate A — main-vs-main determinism, both sides, x2: sorted raw JSON-line
// multisets equal. Never compare unsorted streams (Pitfall 5).
require.Equal(t, sortedLines(baseR1), sortedLines(baseR2),
	"Gate A: baseline main-vs-main determinism (sorted raw JSON-line multisets, r1 vs r2)")
...
// Gate B — zero-COMPILE precondition in all four runs, before any compare.
for _, r := range []gateRun{baseR1, baseR2, headR1, headR2} { ... }
```

Planner: cite the exact test name plus `-run` regex in the verify block, and state the expected sentinel (exit 0 / PASS), not "tests pass".

### 3. Pinned-count golden pattern — analog `internal/rules/golden_full_registry_test.go`

**Source:** `internal/rules/golden_full_registry_test.go:57-60`, `529` (tracked)
**Apply to:** the 94-findings byte-exact gate (expected expectations must not be edited).

```go
// goldenFindingCount pins the captured finding total, asserted before the
// golden compare on both drives (T-04-02 anti-vacuity).
const goldenFindingCount = 94
```

Gate: `go test -race -count=1 ./internal/rules -run '^TestLintGoldenFullRegistry$' -v`, and independently `git diff "$PRE_MERGE_BASE" HEAD -- internal/rules` must be empty (expectations untouched).

### 4. Position gates — analogs `internal/protoast/position_test.go`, `internal/core/position_parity_test.go`

**Source:** `internal/protoast/position_test.go:12-22`; `internal/core/position_parity_test.go:14-21` (tracked)
**Apply to:** D-07a/D-07b gates; these are the regression tripwire for choosing the wrong `Position` alias.

Both use `t.Parallel()`, a hand-computed table, and comments tying the test to the decision. Gate commands:

```bash
go test -race -count=1 ./internal/protoast -run '^TestPositionRuneColumns$' -v
go test -race -count=1 ./internal/core -run '^TestProtoastPositionsMatchLegacyContract$' -v
```

### 5. Merge commit message — repo precedent + CONTEXT override

**Sources:** git history (tracked objects)
**Apply to:** the single merge commit.

Observed precedents:

- Local merge with explanatory body: `8bd9626 Merge main into feature/v1.0-pilot` — body states what was adapted and ends with `Validate with go test -race -count=1 ./..., go vet, both CLI builds, ...`. This is the closest structural analog.
- Minimal local merge: `b17b948 Merge branch 'main' into fix/missing-dependency` (subject only).
- GitHub-created merges: `Merge pull request #239 from easyp-tech/fix/nightly-v1-regressions`, body = the PR title line (`fa16a22`).
- There is **no prior `merge:`-prefixed commit** in this repo (`git log --all --grep='^merge:'` empty). CONTEXT D-01/D-02 explicitly asks for `merge: origin/main (#253 Use compiler for breaking check)`, so follow CONTEXT and pair it with the `8bd9626` body shape: state the semantic rule, then list the three resolved files one per bullet, then the validation commands.

### 6. Docs commit message — repo precedent

**Apply to:** only if any planning-artifact commit is still needed (the archive commit already exists as `9408042`).

Observed forms: `docs: archive v1.0 phase artifacts; sync v1.1 state`, `docs(state): sync session pointer after phase 01 context`, `docs(01): research phase domain`, `docs(phase-01): add validation strategy`, `docs(07): ship milestone ... [ci skip]`. Convention: `docs:` or `docs(scope):`, imperative, semicolon-separated clauses for multiple effects; `[ci skip]` is used for docs-only ship commits. Do not add a new docs commit if the tree is already clean.

### 7. Error-handling / style conventions for any hand-touched code

**Source:** `.spec/agent-rules.md:8-13`; adopted code already follows it (main's `breaking_profiles.go`, `sourceview/*`).
**Apply to:** the only hand edit (`core.go` import block) and any review of adopted code.

- Wrap as `fmt.Errorf("<callee>: %w", err)`; join multi-error closes with `errors.Join`.
- Never bare `defer x.Close()` — main's `session_test.go` is the known exception and stays untouched in Phase 1.
- Imports ordered stdlib → third-party → project; godoc in English.

## No Analog Found

| File/Behavior | Role | Data Flow | Reason |
|---------------|------|-----------|--------|
| `internal/sourceview/root_tree.go` | utility | file-I/O | New in #253; no branch-side equivalent (closest conceptual predecessor is branch `local.go` using raw `*os.Root`) — take `origin/main` verbatim |
| `internal/sourceview/session.go` | store/provider | file-I/O | New in #253; nil-session fallback is the bridge to the package-function API |
| `internal/sourceview/session_test.go` | test | file-I/O | New in #253; std-test style, and it carries upstream's pre-existing errcheck finding — no compliant in-repo analog to copy |
| Go-test-based ancestry/blob gates | test | integration | Repo has no Go test asserting git ancestry; the closest analog is shell (`scripts/validate-release-tag.sh`), so these gates belong in plan `<verify>` blocks, not in `*_test.go` |
| `merge:`-prefixed commit subject | — | — | No repo precedent; sanctioned by CONTEXT D-01/D-02, modeled on `8bd9626` body style |

## Metadata

**Analog search scope:** `internal/core`, `internal/api`, `internal/sourceview`, `internal/rules`, `internal/protoast`, `scripts/`, `Taskfile.yml`, `.github/workflows/`, `.spec/`, git object database (`HEAD`, `origin/main`, rehearsal blobs).
**Files/objects scanned:** 30 (10 merge-affected paths + 4 gate files + 2 shell scripts + Taskfile/CI + commit messages + rehearsed blobs).
**Pattern extraction date:** 2026-11-21
**Re-validate before execution:** `origin/main` SHA (`5d36266`), clean index, and the three conflict blobs. If `origin/main` moves, abort and re-research — do not reuse these blob gates.
