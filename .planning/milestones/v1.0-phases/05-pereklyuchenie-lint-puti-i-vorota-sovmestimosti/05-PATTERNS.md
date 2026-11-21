# Phase 5: Переключение lint-пути и ворота совместимости - Pattern Map

**Mapped:** 2026-11-20
**Files analyzed:** 84 (12 production modify/trim, 13 test modify/trim, 44 production delete, 13 test delete, 2 create)
**Analogs found:** 84 / 84 — this phase is identity-based; every "change" is a rename/trim/delete whose analog is the pre-flip original or its linked twin already in the tree

**Tracked-source gate:** every analog path below was verified with `git ls-files --error-unmatch` — all TRACKED. No gitignored mirrors referenced.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/api/runtime.go` (seam, textually unchanged) | config/assembly | request-response | itself (`buildCore` :62-88) | exact |
| `internal/core/core.go` (trim) | model/DI container | — | itself (:18-128) | exact |
| `internal/core/lint_model.go` (rename) | model | transform | itself + old `dom.go` names | exact |
| `internal/core/lint_linked.go` (rename) | service (engine) | batch | old twin `internal/core/lint.go` (deleted same commit) | exact |
| `internal/core/lint_linked_read.go` (rename) | service (read path) | file-I/O | itself (2 names are permanent exceptions) | exact |
| `internal/core/lint_suppressions.go` (rename) | service (suppressions) | transform | old twin `internal/core/comment_suppressions.go` | exact |
| `internal/core/link.go` (WR-01/WR-02/IN-01 fixes) | service (classifier/resolver) | request-response | `internal/easypc/protoc/resolver.go` (ctx guard) + `internal/core/proto_info_read.go` (`openDiskSource` policy) | role-match |
| `internal/core/dom.go` (trim) | model | — | itself (keep-list verified) | exact |
| `internal/core/comment_suppressions.go` (trim) | utility | transform | itself (keep-list verified) | exact |
| `internal/rules/builder.go` (trim) | registry/config | transform | itself + `builder_linked.go` `NewLinked` (structural twin) | exact |
| `internal/rules/builder_linked.go` (rename) | registry | transform | `internal/rules/builder.go` `New` (:61-143) | exact |
| `internal/rules/lint_helpers.go` (trim) | utility | transform | itself (consumers verified by grep) | exact |
| `.golangci.yml` (remove exclusions) | config | — | itself | exact |
| 41 old rule files `internal/rules/<rule>.go` (delete) | component (rule) | transform | their identity twins `internal/rules/<rule>_linked.go` | exact |
| `internal/rules/casing.go` + `casing_test.go` (delete) | utility | transform | none needed — zero surviving consumers (grep-verified) | exact |
| `internal/core/lint.go` (delete; `close` moves out) | service (old engine) | batch | `lint_linked.go` (the renamed survivor) | exact |
| `internal/core/lint_proto_info.go` (delete) | service (old read path) | file-I/O | `lint_linked_read.go` `readSource` | exact |
| `internal/api/lint_link_matrix_test.go` (extend: class-1 COMPILE cells) | test | request-response | itself (existing cells are the template) | exact |
| `internal/core/lint_linked_matrix_test.go` (extend: WR-01/WR-02 fallback cells) | test | batch | itself (`compile_bypasses_suppression` cell :269) | exact |
| `internal/rules/init_test.go` (trim) | test helpers | file-I/O | itself (`parseLinkedFile`/`validate` stay) | exact |
| `internal/rules/builder_test.go`, `selection_test.go` | test | — | mechanical rename only | exact |
| `internal/config/v1/lint_selection_test.go`, `internal/migration/migration_test.go` | test | — | compile through `NewLinked → New` untouched | exact |
| `internal/core/lint_linked_test.go`, `lint_linked_readsource_test.go`, `lint_suppressions_parity_test.go` | test | batch | mechanical `LintLinked → Lint` rename | exact |
| `internal/rules/module_rules_guard_test.go` (keep) | test | — | untouched | exact |
| 10 parity test files `internal/rules/parity_*.go` (delete; full_registry re-hosted) | test | batch | die with the oracle; value transfers to goldens | exact |
| `internal/rules/check_lint_ignore_test.go`, `registry_parity_guard_test.go` (delete) | test | — | die with `core.CheckIsIgnored` / dual-registry | exact |
| `internal/core/comment_suppressions_test.go`, `comment_suppressions_eof_test.go` (delete) | test | batch | coverage carried by `lint_suppressions_parity_test.go` | exact |
| NEW `internal/rules/golden_full_registry_test.go` (planner names) | test (golden) | batch | `internal/rules/parity_full_registry_test.go` corpus + `parity_harness_test.go` fingerprint | exact (corpus), new assertions |
| NEW `internal/api/lint_stdout_golden_test.go` | test (subprocess golden) | request-response | `lint_link_matrix_test.go` `buildEasypBinary`/`runEasypLint` | role-match |

---

## Pattern Assignments

### `internal/api/runtime.go` — THE SEAM (config/assembly, request-response)

**Analog:** itself. Consolidation shape: source text is IDENTICAL post-flip; the bindings behind it change identity because the old constructors die in the same commit.

**Seam excerpt** (lines 62-88) — this is the only production constructor that populates lint rules:

```go
func buildCore(log logger.Logger, cfg config.Config, importRoots modules.SourceRoots) (*core.Core, error) {

	lintRules, ignoreOnly, err := rules.New(cfg.Lint) // post-flip: the renamed NewLinked body
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	return core.New(core.Options{
		Rules:                   lintRules, // post-flip: []core.Rule (linked contract, re-typed)
		Ignore:                  append(append([]string(nil), cfg.Lint.Ignore...), defaultVendorDir),
		IgnoreOnly:              ignoreOnly,
		AllowCommentIgnores:     cfg.Lint.AllowCommentIgnores,
		KnownLintRules:          rules.AllRuleNames(),
		...
```

**Binding identity table** (what each name binds to pre/post flip):

| Binding | Pre-flip | Post-flip |
|---------|----------|-----------|
| `rules.New(cfg)` (runtime.go:64) | old registry, builder.go:61 | renamed `NewLinked` body, builder_linked.go:71 |
| `Options.Rules` (runtime.go:70) | `[]core.Rule` old contract (dom.go:24-29) | re-typed `[]core.Rule` = etalon interface (lint_model.go:43-48) |
| `apps[key].Lint(...)` (lint_v1.go:205) | old sequential engine | renamed `LintLinked` (lint_linked.go:25) |

**Minimal-flip fallback** (if the planner rejects consolidation — two api edits instead of zero):
- runtime.go:64 → `rules.NewLinked(cfg.Lint)`; runtime.go:70 → `LinkedRules: lintRules` (field survives);
- lint_v1.go:205 → `apps[key].LintLinked(...)`.
Research recommends consolidation (rename-map headers in lint_model.go:9-33 and builder_linked.go:20-70 promise it; SC2 stays one-seam). Optionally add one comment at buildCore marking the seam ("the Phase-5 flip changed the bindings, not this file") — research Open Question 1.

**Engine call site** (`internal/api/lint_v1.go:205`, plus the user-visible SourceName rewrite the CLI golden gate must cover, :213-215):

```go
	for _, key := range batchOrder {
		fileIssues, err := apps[key].Lint(ctx.Context, newSelectedLintWalker(lintRoot, batchFiles[key]))
		if err != nil {
			return fmt.Errorf("Lint: %w", err)
		}
		for _, issue := range fileIssues {
			if v1IssueRuleExcluded(issue.RuleName, excludedByPath[filepath.ToSlash(issue.Path)]) {
				continue
			}
			if issue.RuleName == "PACKAGE_DIRECTORY_MATCH" && issue.SourceName != "" {
				issue.Path = filepath.ToSlash(issue.SourceName)
			}
			issues = append(issues, issue)
		}
	}
```

---

### `internal/core/lint_linked.go` (service, batch) — the engine that becomes `Lint`

**Analog:** itself + the old twin `lint.go` whose method name it inherits.

**Method rename target** (line 25): `func (c *Core) LintLinked(...)` → `Lint`. The doc comment above it is rewritten (drop "two-pass linked engine" self-reference only if desired — behavior text stays).

**Suppression-guard rename** (lines 227-241) — this becomes the only `shouldIgnore` once lint.go:79-93 dies:

```go
// shouldIgnoreLinked mirrors the old engine's shouldIgnore for the linked rule
// contract: the parameter type pins the original to the old rules, so the
// linked engine carries a byte-equivalent copy (see the lint_model.go rename
// map).
func (c *Core) shouldIgnoreLinked(rule LintRule, path string) bool {
	ruleName := LinkedRuleName(rule)
	ignoreFilesOrDirs := c.ignoreOnly[ruleName]

	for _, fileOrDir := range ignoreFilesOrDirs {
		switch {
		case fileOrDir == path:
			return true
		case strings.HasPrefix(path, fileOrDir):
			return true
		}
	}

	return false
}
```

Post-rename: `shouldIgnore(rule Rule, path string)`, body identical (compare old lint.go:79-93 — byte-equivalent shape, different param type). Callers to sweep in this file: lines 136, 187, 212. Field reads `c.linkedRules` (lines 130, 183, 210) → `c.rules`.

**D-01 total sort** (lines 146-160) — UNCHANGED; the golden corpus asserts its order:

```go
	slices.SortStableFunc(res, func(a, b IssueInfo) int {
		return cmp.Or(
			strings.Compare(a.Path, b.Path),
			strings.Compare(a.RuleName, b.RuleName),
			cmp.Compare(a.Position.Line, b.Position.Line),
			cmp.Compare(a.Position.Column, b.Position.Column),
			strings.Compare(a.SourceName, b.SourceName),
		)
	})
```

**Parity-pin literal that must stay byte-identical** (lines 112-117) — do NOT "fix" the stale callee name; it pins main's user-visible error chain:

```go
			// The wrapper text pins main's user-visible error chain (REQ-2.4
			// main-wins philosophy): the callee name c.lintProtoInfoRead in the
			// format string is a deliberate parity pin, not the actual callee.
			if infoErrs[i] != nil {
				return nil, fmt.Errorf("fs.WalkDir: c.lintProtoInfoRead: %w", infoErrs[i])
			}
```

---

### `internal/core/lint_model.go` (model, transform) — contract renames + header deletion

**Analog:** itself; post-flip names = the etalon's (`/tmp/gsd-research/compiler-lab/easyp-cli/internal/core/dom.go` @ 08cef10 confirms `Rule`/`ProtoInfo`/`AppendIssue` 4-arg/`GetRuleName`).

**Rename map to execute** (header comment lines 9-33 — DELETE the comment in the same commit once executed):

| Current (line) | Becomes | Enabler |
|---|---|---|
| `LintRule` interface (:43) | `Rule` | old `Rule` (dom.go:24-29) deleted same commit |
| `ModuleRule` (:67) | unchanged | — |
| `LinkedFileInfo` (:87) | `ProtoInfo` | old `ProtoInfo` (dom.go:57-64) deleted same commit |
| `AppendLinkedIssue` (:105) | `AppendIssue` (4-arg) | old 5-arg `AppendIssue` (dom.go:126-130) deleted same commit |
| `LinkedRuleName` (:119) | `GetRuleName` | old `GetRuleName` (dom.go:133-135) deleted same commit |

**The renamed name function** (lines 114-124) — keep the optional-override mechanism through the flip commit; it is what makes `GetRuleName(&SomeTypeLinked{})` return the canonical registry name before the optional type renames land:

```go
// LinkedRuleName returns the registry name of a linked-engine rule: the same
// UPPER_SNAKE spelling the old engine derives with GetRuleName. A rule may
// override the reflection-derived name with an optional LintRuleName method —
// suffixed port types (<Name>Linked) report their canonical registry name
// through it, so no _LINKED suffix leaks into findings, config, or schemas.
func LinkedRuleName(rule LintRule) string {
	if named, ok := rule.(interface{ LintRuleName() string }); ok {
		return named.LintRuleName()
	}
	return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())
}
```

---

### `internal/core/lint_linked_read.go` (service, file-I/O) — mostly mechanical

**Analog:** itself. `LinkedFileInfo` → `ProtoInfo` sweeps lines 30, 41, 52, 138, 155.

**PERMANENT EXCEPTIONS — do NOT rename** (breaking path owns the unsuffixed names):
- `linkedFilesFromImport` (:52) / `linkedFileFromImport` (:92): collide with `readFilesFromImport`/`readFileFromImport` surviving in `proto_info_read.go:71,89` (different signatures; renaming breaking-owned symbols is rejected — REQ-5.2 freeze).
- Their error wraps (`"readFileFromImport: %w"` :58, `"readProtoFile: %w"` :111) are parity pins — keep verbatim.

**`close` call sites** (:19, :107) — `close` must MOVE out of `lint.go` (:95-105) before that file is deleted; all surviving callers:

```
internal/core/lint_linked_read.go:19   defer c.close(ctx, f, path)
internal/core/lint_linked_read.go:107  defer c.close(ctx, f, importName)
internal/core/proto_info_read.go:25    defer c.close(ctx, f, path)        (breaking-owned — frozen file)
internal/core/proto_info_read.go:94    defer c.close(ctx, f, importName)  (breaking-owned — frozen file)
internal/core/link.go:139              defer r.core.close(ctx, f, path)
```

Suggested landing file for the moved `close`: `lint_linked_read.go` or `core.go` (package core; body verbatim from lint.go:95-105).

---

### `internal/core/lint_suppressions.go` (service, transform) — mirror-of-the-twin renames

**Analog:** old twin `internal/core/comment_suppressions.go:41-114` — the linked method is the same pairing machine reading protoast positions instead of go-protoparser Meta.

**Renames:** `linkedCommentSuppressions` (:48) → `commentSuppressions` (old name freed by the trim of comment_suppressions.go:41-114); `linkedParsedDirective` (:18) → `parsedDirective` (old type freed, comment_suppressions.go:33-37). `linkedDirectiveComment` (:12) **KEEPS its name** — `directiveComment` (*parser.Comment) survives via comment_suppressions.go:28-31 + `BreakingProtoInfo.directiveComments` (proto_info_read.go:49).

**Hidden cross-reference that must compile after the field rename** (lines 51-58) — this is why `Core.rules`/`GetRuleName` renames are load-bearing here:

```go
	known := make(map[string]bool)
	for _, name := range c.knownLintRules {
		known[name] = true
	}
	if len(known) == 0 {
		for _, rule := range c.rules {
			known[GetRuleName(rule)] = true
		}
	}
```

Post-flip `c.rules` is the renamed field and `GetRuleName` is the renamed function — no edit needed, but it breaks if either rename is skipped (research Pitfall 5).

---

### `internal/core/link.go` (service, request-response) — WR-01 / WR-02 / IN-01 fix target

**Analogs:** `internal/easypc/protoc/resolver.go` (the ctx guard every sibling resolver has) and `internal/core/proto_info_read.go` (`openDiskSource` file-selection policy being mirrored). **These fixes land PRE-FLIP**, on the dormant path, so the compat gate never sees the divergences.

**IN-01 pattern — ctx guard, copy from sibling resolvers** (`internal/easypc/protoc/resolver.go:27-30`, all four resolvers repeat it):

```go
func (m MemResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
```

**WR-02 pattern — the policy being mirrored** (`internal/core/proto_info_read.go:169-181`, frozen file, do not edit — replicate its check inside the resolver):

```go
func (c *Core) openDiskSource(disk FS, name string) (io.ReadCloser, error) {
	if rooted, ok := disk.(interface{ RootPath() string }); ok && c.sourceFileOpen != nil {
		full := filepath.Join(rooted.RootPath(), filepath.FromSlash(name))
		if c.importFileAllowed != nil && !c.importFileAllowed(full) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		...
```

**WR-01/WR-02 fix shape** — current `lintResolver` (link.go:118-145) gains `broken atomic.Bool`, the ctx guard, and the corpus-branch policy check; the fallback condition (:48) gains `resolver.broken.Load()`:

```go
// CURRENT (link.go:118-145) — the fix extends exactly this code:
type lintResolver struct {
	core    *Core
	disk    FS
	linted  map[string][]byte
	missing atomic.Bool // an import could not be found
}

func (r *lintResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if src, ok := r.linted[path]; ok {
		return src, nil
	}
	f, err := r.core.openImportFile(r.disk, path)
	if err != nil {
		var notFound *OpenImportFileError
		if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
			r.missing.Store(true)
			return nil, protoc.ErrFileNotFound
		}
		return nil, err
	}
	defer r.core.close(ctx, f, path)
	...
```

Fix sketch (from 02-REVIEW, validated against source):
1. Add `broken atomic.Bool` field.
2. First lines of `FindFileByPath`: `if ctx.Err() != nil { return nil, ctx.Err() }` (IN-01).
3. Corpus branch (`r.linted[path]` hit): consult `r.core.importFileAllowed` on the rooted full path, mirroring `openDiskSource`; denial → `r.missing.Store(true)` + return `protoc.ErrFileNotFound` (WR-02 — the per-file fallback then yields main's exact `os.ErrNotExist` carrier via `readSource`/`openDiskSource`).
4. Non-notFound open/read failure → `r.broken.Store(true)` + return the raw error (WR-01 — never becomes a COMPILE remark).
5. Fallback condition (link.go:48) becomes: `if resolver.missing.Load() || resolver.broken.Load() || !errors.Is(err, protoc.ErrCompilationFailed) || len(diags) == 0 {`.

**Do NOT add** `filepath.IsLocal` to the resolver — the non-local rejection already lives inside `openImportFile` (proto_info_read.go:104-106) and MUST continue to classify as COMPILE (etalon A4 cell stays green).

**IN-02 decision** (link.go:70-82): COMPILE bypasses ignore machinery — KEEP; record the decision in the `linkSources` doc comment at flip; pin with an ignored-broken-dependency cell.

---

### `internal/core/dom.go` (model) — trim boundaries (grep-verified keep-list)

**Analog:** itself. DELETE: old `Rule` interface (:24-29), old `ProtoInfo` struct (:57-64), 5-arg `AppendIssue` (:126-130), `buildError` (:157-164), old `GetRuleName` (:133-135). KEEP everything else — notably `Position` alias, `Issue`/`IssueInfo`, `ImportPath`/`PackageName`, `Collection`, `ProtoData`, wrapper types, `fromMetaPos` (:169-176, used by breaking_checker.go:370), `OpenImportFileError`, `GitRefNotFoundError`, `ConvertImportPath`, Plugin/Input types, and `toUpperSnakeCase` (:138-154 — the reflection fallback for the renamed `GetRuleName`).

---

### `internal/core/comment_suppressions.go` (utility, transform) — trim boundaries

**Analog:** itself. DELETE: `commentSuppressions` method (:41-114, the old reflection-fed walk) + `parsedDirective` type (:33-37). KEEP:
- `suppressionRange`/`suppressionRanges`/`contains` (:13-26 — shared with the linked engine);
- `directiveComment` (:28-31 — *parser.Comment; used by proto_comments.go + `BreakingProtoInfo.directiveComments`);
- `compactDirectiveNames` (:116-124), `parseDirectivePrefix` (:126-143), `collectDirectiveComments` (:147-231 — called by `readProtoFileWithDirectives`, proto_comments.go:21, breaking read path).

The `.golangci.yml` `internal/easypc` exclusion is unrelated; do not touch it.

---

### `internal/core/core.go` (model/DI) — collapse the dual rule surface

**Analog:** itself. DELETE `Options.LinkedRules` (:71-73) + dormancy comments; collapse `Core.rules []Rule` (:19) + `Core.linkedRules []LintRule` (:23) into ONE `rules []Rule` field wired from `Options.Rules` (:104-105). Post-flip `Options.Rules` is the linked contract (re-typed by lint_model.go renames). All other options/fields untouched.

---

### `internal/rules/builder_linked.go` (registry, transform) — `NewLinked` becomes `New`

**Analog:** `internal/rules/builder.go` `New` (:61-143) — the structural twin it mirrors (same validate call, same group order, same use/except/ignore_only plumbing).

**Constructor excerpt** (lines 71-153) — rename `NewLinked` → `New`, sweep `core.LintRule` → `core.Rule`, `core.LinkedRuleName` → `core.GetRuleName`; DELETE the header rename map (:22-70):

```go
func NewLinked(cfg config.LintConfig) ([]core.LintRule, map[string][]string, error) {
	if err := validateConfigRuleNames(cfg); err != nil {
		return nil, nil, fmt.Errorf("validateConfigRuleNames: %w", err)
	}
	allRules := []core.LintRule{
		//	minGroup
		&DirectorySamePackageLinked{},
		&PackageDefinedLinked{},
		&PackageDirectoryMatchLinked{
			Root:   ".",
			Prefix: cfg.PackageDirectoryPrefix,
		},
		...
	}
	rules := make(map[string]core.LintRule)
	for _, rule := range allRules {
		ruleName := core.LinkedRuleName(rule)
		rules[ruleName] = rule
	}
	use := unwrapLintGroups(cfg.Use)
	use = removeExcept(unwrapLintGroups(cfg.Except), use)
	...
	return res, unwrapIgnoreOnly(cfg.IgnoreOnly), nil
}
```

---

### `internal/rules/builder.go` (registry, transform) — trim, NOT delete

**Analog:** itself. DELETE only `New` (:60-143) — after which the `config` import (:8) becomes unused and must be dropped. KEEP:
- group constants (:12-18), `RuleGroup` (:21-25), `AllGroups` (:28-36), `AllRuleNames` (:39-45), `AllLintUseValues` (:49-58) — load-bearing for api/runtime.go:74, api/lint_v1.go:263, config/v1/schema.go:73-75, migration/convert.go:167;
- `unwrapIgnoreOnly` (:145-181), `unwrapLintGroups` (:183-204), `removeExcept` (:206-210), `defaultIfEmpty` (:213-218) — consumed by builder_linked.go:138-152;
- `addMinimal/addBasic/addDefault/addComments/addUnary` (:220-280) — **TRAP, see below**.

**MECHANICAL TRAP the research under-specifies:** the five `add*` helpers construct OLD rule types (`core.GetRuleName(&DirectorySamePackage{})` at :221-278). When the 41 old rule files die in the flip commit, these bodies reference dead types and `builder.go` stops compiling. The flip commit must ALSO re-point the five helpers at the port types:

```go
// flip commit: &DirectorySamePackage{} → &DirectorySamePackageLinked{}, ... (41 sites)
// name output is unchanged: GetRuleName (ex LinkedRuleName) prefers each port's
// LintRuleName() override, which reports the canonical registry name
```

After the OPTIONAL type-rename commit the bodies read `&DirectorySamePackage{}` again (now the renamed port type) — behavior-neutral either way, gated by `TestAllGroups`/`TestAllRuleNames` + `TestRegistryParity`'s proven name-set equality. Plan the flip-commit diff for builder.go accordingly (delete `New` + re-point 41 type literals + drop `config` import).

---

### `internal/rules/lint_helpers.go` (utility, transform) — trim + exclusion removal

**Analog:** itself; consumer check re-verified by grep this session — every D-01 helper has ZERO code consumers among surviving files (the only grep hits are comment mentions in `enum_first_value_zero_linked.go:13` and `directory_same_package_linked.go:17`).

**DISCREPANCY FLAG (planner decision):** research says "KEEP fileOption (:56)", but (a) the file's own header (:13-19) and fileOption's doc (:50-55) say "Phase 5 deletes it together with the old path", (b) fileOption has zero consumers, and (c) the phase removes the `.golangci.yml` paths exclusion (:24) while gating on `golangci-lint run` 0 issues — keeping an unused `fileOption` with the exclusion removed makes that gate fail. **Recommended disposition: DELETE `fileOption` too** (it is one of the nine names the exclusion comment lists). Keep-list is then exactly: `groupedValues` + `add` (:118-138) and `firstExemptMismatches` (:165-186). Delete-list: `fileOption` (:50-64), `optionString` (:66-75), `enumNumber` (:77-81), `leadingCommentDocuments` (:83-97), `isLintDirective` (:99-106), `packagePosition` (:108-116), `groupedValues.mismatches` (:140-163), `packageHasVersionSuffix` (:188-224), `versionNumber` (:226-229), plus the stale header comment (:1-39).

`.golangci.yml` in the same commit: remove the `internal/rules/lint_helpers.go` paths exclusion (lines 10-24). **Additional catch:** the QF1001 rule exclusion scoped to `internal/rules/casing.go` (lines 32-39) goes dead when `casing.go` is deleted — remove it in the same commit so the config carries no stale scopes.

---

### NEW `internal/rules/golden_full_registry_test.go` (test, batch) — Gate 1

**Analog:** `internal/rules/parity_full_registry_test.go` (corpus, lines 53-262) + `parity_harness_test.go` (drive shape, fingerprint). Re-host the Phase-4 corpus in-test (research Open Question 3) and replace parity-compare with a pinned full golden: (path, rule, line, col, source, **message**) + order — upgrading the Phase-4 fingerprint which deliberately excluded Message.

**Corpus to promote** (parity_full_registry_test.go:53-262, kept verbatim as the corpus map): 13 files / 9 packages, all 41 rules firing; rpc payload types deliberately declared so the corpus compiles under full type resolution (:166-168 comment — do not regress, research Pitfall 7). Coverage assertion (:271-283) and pinned count 73 (:285) carry over; capture the golden literals (including Message + order) from the byte-frozen old engine BEFORE the flip commit.

**Drive shape to copy** (parity_harness_test.go:59-98, minus the old-engine half and the fingerprint — order now comes from the engine itself):

```go
	rules, ignoreOnly, err := rules.NewLinked(cfg) // post-flip compiles as rules.New
	require.NoError(t, err)
	app := core.New(core.Options{
		Logger:              logger.NewNop(),
		LinkedRules:         rules,        // → Rules: rules post-flip
		IgnoreOnly:          ignoreOnly,
		AllowCommentIgnores: true,
		KnownLintRules:      rules.AllRuleNames(),
	})
	issues, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, ".")) // → app.Lint post-flip
	require.NoError(t, err)
	require.Len(t, issues, 73)          // pinned old-engine count
	require.Equal(t, goldenIssues, issues) // FULL compare incl. Message + order
```

The pre-flip capture version of this test drives the OLD engine (`rules.New` + `Options.Rules` + `app.Lint`) to freeze the golden; the flip commit re-points the drive to the linked constructor. If the golden test is captured pre-flip against the old engine and the flip re-points it, the file must not use the parity harness helpers (they die).

---

### `internal/api/lint_link_matrix_test.go` (test, request-response) — Gate 2 extension

**Analog:** itself — every new cell is a copy of an existing cell pattern with a new fixture.

**In-process cell pattern** (fixture :31-42, driver :47-58, cells :73-124). The NEW class-1 COMPILE cell lands IN the flip commit (physically unreachable pre-flip — the old engine runs):

```go
		{
			// class-1: the corpus fails compilation (not missing-import, not
			// per-file syntax; e.g. proto3 `required` field) — through actionV1
			// it prints COMPILE issues and returns ErrHasLintIssue.
			name: "compile_issue_exit1_class",
			files: map[string]string{
				"example_pkg/req.proto": "syntax = \"proto3\";\n\npackage example_pkg;\n\nmessage Req {\n  required string name = 1;\n}\n",
			},
			check: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrHasLintIssue)
			},
		},
```

**Subprocess cell pattern** (builder :140-157 `sync.OnceValues`, runner :163-180, sequential cells :185-237). NEW `compile_error_code_1` cell: exit 1 + stdout `path:line:col: message (COMPILE)` — reuse `buildEasypBinary`/`runEasypLint` as-is; the test stays sequential (go-testing process-state exception, documented at :182-184).

**Printer contract already pinned pre-flip** (:243-329, `compileIssue` helper + `TestPrintIssuesCompileGolden`): text byte-exact `"test.proto:4:3: ... (COMPILE)\n"` (:277), JSON zero Filename/Offset (:289-299), D-03 multi-issue order through the api path-sort (:305-316). The printer lives at `internal/api/lint.go:108-154`; exit routing (`ErrHasLintIssue` → 1, `OpenImportFileError` → errExit(2), default → main's log.Fatal → 1) at lint.go:80-98 + cmd/easyp/main.go:38.

**Existing cells re-run post-flip through the new engine — the primary SC1 exit-code gate:** missing_import → exit 2 + "Cannot import file" stderr (:194-206); syntax error → exit 1 (:207-217); clean → exit 0 empty stdout (:218-228); rule findings → exit 1 (:100-113).

---

### NEW `internal/api/lint_stdout_golden_test.go` (test, subprocess golden) — Gate 3

**Analog:** `lint_link_matrix_test.go` subprocess harness (`buildEasypBinary` :140-157 + `runEasypLint` :163-180) — reuse verbatim; the fixture map and pinned stdout lines are new. Corpus exercises findings + ignore_only + comment directives through the real binary; capture exact stdout (text, and optionally JSON) + exit code from the OLD binary pre-flip; assert post-flip. Covers path rendering, the `PACKAGE_DIRECTORY_MATCH` SourceName→Path rewrite (lint_v1.go:213-215), and format edge shapes. Committed fixture map follows `writeLinkMatrixFixture` (:31-42).

---

### `internal/core/lint_linked_matrix_test.go` (test, batch) — WR-01/WR-02/IN-02 cells

**Analog:** itself — the `compile_bypasses_suppression` cell (:269) is the IN-02 pin that must stay green. Add pre-flip: unreadable-import fallback cell (WR-01), blocked-corpus-file cell (WR-02 — restrictive `importFileAllowed` where the corpus file is denied and the run errors as before the flip), ignored-broken-dependency cell (IN-02 pin). Use the file's existing fixture/driver conventions.

---

### `internal/rules/init_test.go` (test helpers) — trim, do NOT delete

**Analog:** itself. DELETE `start` (:33-56) and `parseFile` (:58-99) — go-protoparser old-model helpers whose only consumer (`check_lint_ignore_test.go`) dies. KEEP `parseLinkedFile` (:101-137) and `validate` (:139-164) — they feed all 41 carried per-rule tests. Sweep their signatures' `core.LinkedFileInfo` → `core.ProtoInfo` and `core.LintRule` → `core.Rule`. The testdata constants (:16-31) stay — the repo-root `testdata/` corpora are referenced by surviving tests and MUST NOT be touched.

---

### Mechanical rename sweep (tests calling `LintLinked`)

**Direct callers verified by grep (5 files, not ~30):** `internal/core/lint_linked_matrix_test.go`, `lint_linked_readsource_test.go`, `lint_linked_test.go`, `lint_suppressions_parity_test.go`, `internal/rules/parity_harness_test.go` (the last dies). The wider surface is files calling the harness or `rules.NewLinked` — those die or compile through the rename. Method: `gofmt`/gopls rename or scripted sed sweep with compile+suite gates (Phase-1 precedent), covering `LintLinked`→`Lint`, `NewLinked`→`New`, `LinkedRules:`→`Rules:`, `LinkedFileInfo`→`ProtoInfo`, `LintRule`→`Rule`, `LinkedRuleName`→`GetRuleName`, `AppendLinkedIssue`→`AppendIssue` across both packages (bounded: no `Linked*` symbol is referenced outside internal/core and internal/rules — research-verified; `linkedFilesFromImport`/`linkedFileFromImport`/`linkedDirectiveComment` are the three documented exceptions).

---

## Shared Patterns

### Error wrapping (all edits)
**Source:** `.agents/skills/go-code-style/SKILL.md` + verified exemplars `internal/core/link.go:142` (`"ReadAll: %w"`), `proto_info_read.go:23` (`"Open: %w"`), `lint_linked_read.go:111` (parity-pin wraps).
**Apply to:** link.go fixes, the moved `close`, any new helper.
```go
return nil, fmt.Errorf("Open: %w", err)   // callee-only label, no receiver prefix
```
`errors.Is`/`errors.As` for identity (link.go:133-134 is the exemplar). Never bare-`defer` a Close — the `close` method exists precisely for that.

### Golden-master capture-then-flip (all gates)
**Source:** Phase-4 pinned-count pattern (`parity_full_registry_test.go:285` `assertBatchParity(..., 73)`) + `TestPrintIssuesCompileGolden` (lint_link_matrix_test.go:259-329).
**Apply to:** Gate 1 (corpus goldens incl. Message), Gate 3 (CLI stdout). Capture from the byte-frozen old engine BEFORE the flip commit; after deletion there is no oracle.

### Subprocess gate (exit codes / CLI goldens)
**Source:** `internal/api/lint_link_matrix_test.go:140-180`.
**Apply to:** Gate 2 subprocess cells + Gate 3. `sync.OnceValues` build; per-cell `cmd.Dir`/`cmd.Env`, parent state never mutated; sequential cells (go-testing exception).

### File-access policy funnel (security, ASVS V4)
**Source:** `internal/core/proto_info_read.go:103-125` (`openImportFile`: `filepath.IsLocal` boundary + `importFileAllowed`) and :169-181 (`openDiskSource`).
**Apply to:** the WR-02 fix — the linker's corpus branch becomes the third consumer of `importFileAllowed`; no file access outside these funnels. The frozen files themselves are zero-edit (gate: `git diff --exit-code -- internal/core/proto_info_read.go internal/core/proto_comments.go`).

### Same-run subsystem suite + carve-out (SC3)
**Source:** Phase 2-4 practice; `.planning/STATE.md` wazero blocker.
**Apply to:** phase closure: `go test -race -count=1 ./internal/core/... ./internal/rules/... ./internal/api/... ./internal/modules/... ./internal/migration/... ./internal/config/...`; EXCLUDE `./internal/adapters/plugin/` and `./internal/generation/` (pre-existing wazero race, ledger item 18) — document the carve-out in the plan and SUMMARY.

### Mechanical closure gates
**Source:** 05-RESEARCH Pattern 4 Gate 4, verified against go.mod (protocompile v0.14.1 line 8, go-protoparser v4.14.2 line 19).
```bash
go build ./... && go mod tidy && git diff --exit-code go.mod go.sum \
  && grep -q go-protoparser go.mod && grep -q protocompile go.mod
go run ./cmd/easyp schema-gen && git diff --exit-code schemas/   # rule names unchanged → drift must be empty
bin/golangci-lint run ./internal/...                              # validates exclusion removal + trims
go build -o /tmp/easyp-check ./cmd/easyp && /tmp/easyp-check lint --help   # exit 0
```

---

## No Analog Found

None — every file maps to a pre-flip original or identity twin. Two novel-content notes (analog harness exists, content is new):

| File | Role | Data Flow | Note |
|------|------|-----------|------|
| `internal/api/lint_stdout_golden_test.go` | test | request-response | Harness copied from `lint_link_matrix_test.go`; the committed fixture map + pinned stdout lines are new and must be captured pre-flip |
| class-1 COMPILE matrix cells | test | request-response | Existing cells are the template; the COMPILE class through `actionV1` is physically unreachable until the flip — cells land IN the flip commit (Phase-2 A1 execution) |

## Planner Decision Points (flagged during mapping)

1. **`fileOption` disposition** — research "KEEP" contradicts the file header, zero consumers, and the exclusion-removal + golangci-0-issues gate. Recommended: DELETE (see lint_helpers.go assignment).
2. **`builder.go` add\* helpers** — must be re-pointed to port types in the flip commit or the package does not compile after the 41 old rule files die (see builder.go assignment).
3. **`.golangci.yml` QF1001 block** (lines 32-39) — scoped to the deleted `casing.go`; remove in the same commit as the exclusion line 24 (research only names line 24).
4. **Consolidation vs minimal flip** (research A1 / Open Question 1) — PATTERNS.md documents consolidation as primary; both shapes' exact edits are in the seam assignment.
5. **Optional 41-type-rename commit** — after the flip, mechanical; name-neutrality proven by TestRegistryParity; `TestAllRuleNames`/`TestAllGroups` are the post-rename name gate. If landed, sweep includes `builder_linked.go` → folded/renamed file names and the 41 `<Name>Linked` → `<Name>` per the builder_linked.go:26-66 map.

## Metadata

**Analog search scope:** `internal/api`, `internal/core`, `internal/rules`, `internal/easypc/protoc`, `internal/config/v1`, `internal/migration`, repo root (`.golangci.yml`, `go.mod`, `cmd/easyp/main.go`)
**Files read in full this session:** runtime.go, lint.go, lint_linked.go, lint_linked_read.go, lint_model.go, lint_suppressions.go, link.go, core.go, dom.go-read-verify (keep-list via research + grep), comment_suppressions.go, proto_info_read.go, builder.go, builder_linked.go, lint_helpers.go, init_test.go, parity_harness_test.go, parity_full_registry_test.go, lint_link_matrix_test.go, lint.go (api), lint_v1.go (:180-259), resolver.go, .golangci.yml
**Grep gates run:** `LintLinked` test callers (5 files), `c.close(` callers (5 sites), D-01 helper consumers (zero code consumers; comment mentions only), go.mod module presence, parity file inventory (10 flat files)
**Pattern extraction date:** 2026-11-20
