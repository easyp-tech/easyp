# Feature Research

**Domain:** Compiler-backed protobuf lint engines (buf lint as the reference; protolint as the parser-only counter-example)
**Researched:** 2026-11-17
**Confidence:** HIGH for buf behaviors (docs + source cross-verified), MEDIUM for protolint (single primary source)
**Project context:** Milestone `lint-easypc-infra` — migrate `easyp lint` from go-protoparser to the vendored `easypc` compiler + `protoast` model, keeping output format, exit codes and config surface unchanged. REQ ids below refer to `.spec/features/lint-easypc-infra/requirements.md`.

## Feature Landscape

### Table Stakes (Users Expect These)

What every compiler-backed protobuf lint engine must have. Users of buf-compatible tools assume these; EasyP already has most, and the milestone's job is to carry them across the engine swap without visible change.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Config-driven rule selection: `use`/`except` with **categories + rule IDs** | buf's core model: strictness hierarchy MINIMAL ⊂ BASIC ⊂ STANDARD (v1 name: DEFAULT) plus independent COMMENTS and UNARY_RPC. EasyP already implements all five groups (`internal/rules/builder.go`) | LOW | Exists; the only change is that new rule IDs join existing groups (REQ-4.1/4.2) and must appear in schema/MCP enums (REQ-4.4). Category placement matters: buf puts PACKAGE_NO_IMPORT_CYCLE in MINIMAL and SYNTAX_SPECIFIED + FIELD_NOT_REQUIRED in BASIC — exactly the placement REQ-4.1/4.2 specify |
| Path-based ignoring: `ignore` + `ignore_only` (per rule or category) | Every serious linter; buf docs treat `ignore_only` as the primary tool for systematic exceptions, comment ignores as the exception for one-off cases | LOW | Exists (`c.ignore`, `c.shouldIgnore` in `internal/core/lint.go`). Must survive the engine rewrite byte-identically (REQ-2.5). Note buf semantic: paths are prefix/dir or exact-file matches — EasyP's current `strings.HasPrefix` behavior is the contract, not buf's |
| Comment suppressions (`buf:lint:ignore RULE` above the element; EasyP additionally `easyp:disable`/`nolint` with block scopes) | The escape hatch users reach for when fixing a violation is worse than the violation; buf gates it behind `allow_comment_ignores` (v1: default off; v2: default on). EasyP defaults it on | MEDIUM | Exists (`commentSuppressions`, `CheckIsIgnored`). Highest parity risk in the migration: positions must thread through `protoast` exactly as go-protoparser produced them, and REQ-3.2 says where main evolved after the lab snapshot (comment suppressions), main wins over the lab. buf nuance worth mirroring: comment ignores only apply to annotations that have a source location — file-level rules (e.g. FILE_LOWER_SNAKE_CASE) cannot be comment-suppressed |
| **Compile diagnostics inside lint output (`COMPILE`)** | The defining capability of a compiler-backed engine: buf's `--error-format` is documented as "the format for build errors **or** check violations" — one machinery. Each compiler error becomes an annotation with `type: COMPILE`, the compiler's position and message (`import "x.proto": not found` special case), rules do not run on a corpus that fails to build, and the command exits with the findings exit code | MEDIUM | New in this milestone (REQ-2.2). Verified in buf source `bufprotocompile.go`: `typeString := "COMPILE"`, message from `ErrorWithPos`, zeroed position when unknown. Lab implements it in `internal/core/link.go`. EasyP's text format (`path:line:col:source message (RULE)`) stays; `COMPILE` appears as the rule name in that format (REQ-5.1) |
| Graceful degradation when linking is impossible (missing imports) | buf exits with an error; EasyP's legacy contract is stronger: fall back to per-file parsing and report read/syntax errors as before, without COMPILE issues (REQ-2.3/2.4). protocompile supports the parse-only pattern explicitly (parse + unlinked options) | MEDIUM | Differentiating robustness (see below) but also table stakes for EasyP specifically — removing it would break existing users' CI on incomplete checkouts |
| Stable output format + exit codes | CI scripts parse output; changing paths/lines/columns/codes breaks users silently. buf: text `path:line:col:message`, JSON with `type`; exit 100 for findings (source comment: "to distinguish user-parsable errors from system errors"). protolint: exit 0/1/2. EasyP: text `path:line:col:source message (rule)`, JSON, exit 1 for findings / 2 for import errors | LOW (contract is "don't change") | REQ-5.1 pins EasyP's own formats and codes. Note the deliberate divergence: EasyP keeps exit 1, not buf's 100 — "buf-compatible" for EasyP means diagnostic semantics and config, not exit-code values |
| **Cross-file rule correctness via linking** | IMPORT_USED, PACKAGE_SAME_DIRECTORY, PACKAGE_SAME_\<option\>, PACKAGE_NO_IMPORT_CYCLE, PACKAGE_DIRECTORY_MATCH need the linked corpus — a parser-only linter (protolint) cannot implement them, which is protolint's stated trade-off ("runs fast because this works without compiler") | HIGH | This IS the milestone: `protoc.Link` with `WithSyntaxTrees(true)`, `ModuleRule` (Collect/Finish) for aggregation rules, `ImportedFiles` on ProtoInfo. All 43 current rules keep identical findings (rule, file, line, column) on the acceptance corpora (REQ-3.1) |
| Rule parity (43 existing rules, identical findings) | Users judge an engine swap by zero diff in CI output; any position drift = regression | MEDIUM | Verified per-rule reconcile, not bulk copy: lab rules were rewritten from an older main base and main has since evolved (REQ-3.2). Colocated `<rule>_test.go` per AGENTS.md |
| Config validation errors for unknown rule/category names | Misconfigured `lint.use` must fail loudly, before any parsing | LOW | Exists; REQ-2.6 requires identical behavior. Enum source for the schema is `AllLintUseValues()` — new rules must flow into `schemas/` via `task schema:generate` (REQ-4.4) |
| Per-rule tuning options (`enum_zero_value_suffix`, `service_suffix`, `rpc_allow_*`) | buf tunes rule behavior through config keys rather than new rule variants | LOW | Exists in EasyP config v1; untouched by the milestone |
| Deterministic, ordered findings under parallelism | Parallel engines (buf compiles the whole image concurrently; lab uses 256-file windows via `internal/parallel`) must still emit findings in the same order as the sequential walk or diffs and CI caches break | MEDIUM | REQ-2.1 explicitly requires unchanged file traversal order. buf's own engine sorts annotations before printing; EasyP's order derives from WalkDir + rule order |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Speed on large corpora (parallel link + windows + import cache) | Lab measured 20 min → 0.08 s on a 4 MB `compute.proto`; buf's own engine parallelizes parsing/linking (protocompile resolvers are documented thread-safe for concurrent goroutines). EasyP's requirement is only "not slower than main" on googleapis-nocompute (7414 files, REQ-6.1) — the lab win is upside, not a gate | MEDIUM | Two-pass engine, 256-file parallel windows, import cache. Risk: parallelism must not reorder findings (see table stakes). The old main is sequential WalkDir — beating it is plausible; matching it is mandatory |
| In-repo vendored compiler, zero new external deps | Single static binary, no protoc, no new go.mod entries (REQ-1.2), no `replace`. Differentiates vs buf (large dependency tree) and vs protolint (no compiler at all). Paves the road for migrating breaking/generate/tidy/migrate later | LOW-MEDIUM | Decision already made: vendor `easypc` wholesale (incl. `cmd/`, `testdata/`, ~3.4 MB) + `protoast` (~1000 LOC). Cost is review size, not architecture |
| Richer suppression vocabulary than buf | EasyP accepts `easyp:disable`/`nolint`/`buf:lint:ignore` dialects with block scopes and re-enable directives — buf supports only single-line-above `buf:lint:ignore`, protolint has disable/enable blocks. Teams migrating from either tool keep their comments working | LOW | Exists; must be preserved through the rewrite (REQ-2.5). Parity check `check_lint_ignore_test.go` is the main-side evolution to honor (REQ-3.2) |
| Missing-import fallback (lint still runs per-file) | buf hard-fails when the graph cannot resolve; EasyP's legacy behavior reports the import error and still lints what it can parse. Better UX on partial checkouts and generated-file edge cases | MEDIUM | REQ-2.3: fallback must reproduce pre-change behavior exactly, including exit code 2 (`OpenImportFileError`) |
| Buf rule-category parity as product strategy | New rules land in the same categories buf uses (PACKAGE_NO_IMPORT_CYCLE → MINIMAL, SYNTAX_SPECIFIED/FIELD_NOT_REQUIRED → BASIC), so teams sharing policy configs between buf and easyp get equivalent defaults. Note buf's own quirk: PACKAGE_NO_IMPORT_CYCLE is categorized only in v2 configs; in v1/v1beta1 it is opt-in — EasyP's MINIMAL placement matches the v2 semantics | LOW | Schema enum + MCP description updates ride along (REQ-4.4) |
| Acceptance-grade parity harness (lint-diff on real corpora) | `compiler-lab/bench/lab.py` comparing full finding sets (rule, file, line, column) across old/new/lab CLIs on api-common (92) + googleapis-nocompute (7414) — stronger than unit tests, catches position drift | MEDIUM | Exists in lab; needs local reproduction (spike). Sibling design question: per-batch parity tests vs corpus-only finish line |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Auto-fix / rewriting proto files (protolint `-fix`) | "Fix the naming violations for me" | A lint engine that mutates sources breaks trust in CI; renames often change wire-visible names (packages, options) — exactly what a compatibility-first tool must not touch silently; hugely expands test surface | Keep lint read-only. Fixing belongs to a future explicit `easyp format`-style feature with its own review workflow |
| Auto-inserted suppression comments (protolint `-auto_disable`) | "Silence existing violations so the gate goes green" | Same source-mutation problem; also hides signal permanently | buf's approach: `--error-format=config-ignore-yaml` prints a ready-to-paste `ignore_only` block — the user edits config deliberately. Safe future differentiator for EasyP (additive output format), not for this milestone |
| Custom rule plugins (buf check plugins / protolint `-plugin`) | "Let teams add their own rules" | Freezes an internal rule API that the compiler migration would keep breaking; expands compatibility surface the milestone promises not to change | In-repo rules with colocated tests (current AGENTS.md convention); revisit only after all subsystems are on `easypc` |
| New output formats now (sarif, msvs, github-actions, junit, sonar) | CI vendor integration | REQ-5.1 freezes output; bundling format work into an engine-swap milestone makes the parity diff unverifiable | Ship the swap with zero visible change; add formats as additive follow-ups (buf and protolint both treat formats as a separate axis) |
| Migrating breaking/generate/mod/migrate in the same milestone | "While you're in there…" | `core.ProtoInfo` is shared with breaking (`breaking_check.go` reads the go-protoparser model); swapping both at once makes the diff unreviewable and doubles blast radius | Explicitly out of scope (REQ-5.2); go-protoparser and protocompile stay in go.mod; the series continues after lint stabilizes |
| Adopting buf's exit code 100 for findings | "Be more buf-compatible" | Breaks every existing EasyP CI script that checks exit 1; compatibility is EasyP's Core Value. buf's 100 exists to separate user findings from system errors — EasyP already separates via 1 vs 2 | Keep exit 1 (findings) / 2 (cannot import) / default (system). COMPILE findings flow through the existing findings path (REQ-5.1) |
| Persistent cross-run lint/compile cache | "Make re-runs instant" | Neither buf nor protolint persists lint caches; invalidation (config, rule set, dependency pins) is a correctness hazard next to a compatibility contract; the lab's in-run import cache already captures most of the win | In-run caching only; revisit after the whole toolkit shares one compiler |
| Speed beyond "not slower than main" as a requirement | Lab numbers are spectacular | Making the lab's 250x a hard gate couples the milestone to machine-specific benchmarks | REQ-6.1 wording is correct: not-slower is the gate, the lab figure is the north star |

## Feature Dependencies

```
[Vendored easypc + protoast (REQ-1.1/1.2)]
        └──requires──> [Two-pass lint engine: Link + parallel windows + import cache (REQ-2.1)]
                                ├──requires──> [COMPILE diagnostics (REQ-2.2)]
                                │                    └──conflicts──> [missing-import fallback (REQ-2.3)]
                                │                         (resolution: fallback path bypasses COMPILE — decided, not a real conflict)
                                ├──requires──> [43-rule parity on protoast (REQ-3.1)]
                                │       └──requires──> [suppressions/ignore parity (REQ-2.5, 3.2)]
                                └──requires──> [ModuleRule Collect/Finish]
                                        └──requires──> [PACKAGE_NO_IMPORT_CYCLE (REQ-4.1)]

[SYNTAX_SPECIFIED, FIELD_NOT_REQUIRED (REQ-4.2)]  ── single-file rules; depend only on protoast model

[New rules (REQ-4.1/4.2)] ──requires──> [Schema + MCP enum updates (REQ-4.4)]

[Everything above] ──gates──> [Compatibility acceptance (REQ-5.1/5.2) + performance (REQ-6.1)]
```

### Dependency Notes

- **COMPILE diagnostics require linking**: parse-only mode (REQ-2.3 fallback) cannot produce them — by design, the fallback reproduces the pre-compiler error behavior.
- **PACKAGE_NO_IMPORT_CYCLE requires ModuleRule**: it aggregates the corpus-wide import graph — the only new rule needing cross-file aggregation; SYNTAX_SPECIFIED and FIELD_NOT_REQUIRED are single-file and safe to port early.
- **Comment suppressions depend on position fidelity**: suppression matching is line-number-based (`suppressions.contains(rule, line)`), so any position drift between go-protoparser and protoast silently breaks suppressions before it visibly moves findings — parity tests must cover suppressed files, not just clean ones.
- **Schema updates enhance config validation**: new rule names must join `AllLintUseValues()` or valid configs start failing validation (REQ-2.6 regression).
- **Performance gate depends on all rules being ported**: measuring before the switch completes compares a hybrid path, not the engine.

## MVP Definition

### Launch With (this milestone — v1 of the compiler-backed lint)

- [x] Vendored `easypc` + `protoast`, no new external deps (REQ-1.1/1.2) — prerequisite for everything
- [x] Two-pass engine with linking, parallel windows, import cache, unchanged traversal order (REQ-2.1)
- [x] COMPILE diagnostics on non-compiling files, rules skipped, existing output format (REQ-2.2 + 5.1)
- [x] Missing-import fallback preserving legacy error behavior and exit codes (REQ-2.3/2.4)
- [x] Suppression/ignore/config-validation parity (REQ-2.5/2.6)
- [x] 43-rule findings parity on api-common + googleapis-nocompute (REQ-3.1/3.2)
- [x] Three new rules in buf-matching categories + schema/MCP regeneration (REQ-4.1–4.4)
- [x] breaking/generate/mod/migrate untouched (REQ-5.2); lint not slower than main (REQ-6.1)

(Checked = already committed as milestone scope; the roadmap phases execute them.)

### Add After Validation (v1.x / next features in the series)

- [ ] Breaking-change detection on the compiler — trigger: lint path stable; removes the dual-model `ProtoInfo` coupling flagged in the explore doc
- [ ] Generate / `mod tidy` / migrate on the compiler — trigger: breaking done; then delete go-protoparser from go.mod
- [ ] `config-ignore-yaml`-style output (generate `ignore_only` blocks) — trigger: user demand for suppression workflows; additive format, no contract break
- [ ] Additional output formats (json already exists; sarif/junit for CI vendors) — trigger: user request; additive
- [ ] More buf STANDARD-category rules (PROTOVALIDATE, STABLE_PACKAGE_NO_IMPORT_UNSTABLE) — trigger: users migrating fuller buf policies

### Future Consideration (v2+)

- [ ] Read-only "suggested fix" output (never auto-write) — only with a product decision on mutation policy
- [ ] Custom rule plugins — only after all subsystems share the compiler and the rule API is stable
- [ ] Persistent cross-run cache — only with a validated invalidation story across `protobuf.lock` changes

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| COMPILE diagnostics in lint | HIGH | MEDIUM | P1 (milestone scope, REQ-2.2) |
| 43-rule parity | HIGH | HIGH | P1 (the bulk of the work, REQ-3.1) |
| Suppressions/ignore/config parity | HIGH | MEDIUM | P1 (silent-breakage risk if missed) |
| Missing-import fallback | MEDIUM | MEDIUM | P1 (legacy contract, REQ-2.3) |
| New rules (3) + schema/MCP | MEDIUM | LOW | P1 (REQ-4.x; rides on finished engine) |
| Parallel windows + import cache | MEDIUM | LOW | P1 (already lab-proven) |
| Not-slower performance gate | MEDIUM | LOW | P1 (acceptance only, REQ-6.1) |
| breaking/generate migration | HIGH | HIGH | P2 (next features in series, out of scope now) |
| config-ignore-yaml output | MEDIUM | LOW | P2 (additive, post-stabilization) |
| Extra output formats | MEDIUM | LOW | P3 |
| Auto-fix / auto-disable | LOW | HIGH | P3 (anti-feature for now) |
| Rule plugins | LOW | HIGH | P3 (anti-feature for now) |

## Competitor Feature Analysis

| Feature | buf (reference) | protolint (parser-only) | EasyP (milestone target) |
|---------|-----------------|-------------------------|--------------------------|
| Engine base | protocompile: parse → link → options → sourceinfo; whole image built first | go-protoparser, no linking, by design | Vendored `easypc` (parse+link, syntax trees on) + `protoast` model |
| Compile diagnostics in lint | `COMPILE`-typed annotations, same formats as findings, exit 100, rules not run on failure | N/A (cannot detect compile errors) | `COMPILE` rule-name issues in EasyP's existing format, exit 1, rules not run (REQ-2.2) |
| Missing-import behavior | Hard error | N/A — never resolves imports | Fallback to per-file parse, legacy error + exit 2 (REQ-2.3) |
| Cross-file rules | Full set over linked image (IMPORT_USED, PACKAGE_NO_IMPORT_CYCLE, PACKAGE_SAME_*) | Impossible per its own README | ModuleRule Collect/Finish; PACKAGE_NO_IMPORT_CYCLE added (REQ-4.1) |
| Rule selection | use/except, categories MINIMAL/BASIC/STANDARD + COMMENTS/UNARY_RPC | .protolint.yaml lists, official defaults + opt-in unofficial | use/except-equivalent v1 config, groups MINIMAL/BASIC/DEFAULT/COMMENTS/UNARY_RPC — same taxonomy as buf v1 |
| Ignores | `ignore`, `ignore_only` (per rule/category) | config excludes, per-rule | `ignore`, per-rule `ignoreOnly` — preserved (REQ-2.5) |
| Comment suppressions | `// buf:lint:ignore RULE` above element; opt-in v1 / default v2 | `protolint:disable[:next|:this]` blocks with enable | `easyp:disable` / `nolint` / `buf:lint:ignore` dialects with block scopes — richest of the three |
| Auto-fix | No | Yes (`-fix`, `-auto_disable`) | No (deliberate anti-feature) |
| Custom rules | Check plugins (WASM/pluginrpc) | External plugin binaries | In-repo rules only (anti-feature to expand now) |
| Output formats | text, json, msvs, junit, github-actions, gitlab-code-quality, config-ignore-yaml | plain, junit, json, sarif, sonar, unix, tsc, CI flavors, Go templates | text, json — frozen this milestone |
| Exit codes | 100 findings (vs system errors) | 0 clean / 1 findings / 2 other | 1 findings / 2 cannot-import / default — frozen (REQ-5.1) |
| Config validation of rule names | Yes (fails on unknown) | Yes | Yes (REQ-2.6) |
| Performance posture | Parallel compile of whole image; no cross-run cache | Fast start, no compiler cost, but no semantics | Two-pass, 256-file parallel windows, in-run import cache; gate = not slower than main |

## Sources

- buf lint configuration and comment ignores — buf.build/docs/lint/usage (via Context7 `/websites/buf_build`), MEDIUM
- buf rule categories and rule lists incl. PACKAGE_NO_IMPORT_CYCLE v2-only categorization, FIELD_NOT_REQUIRED/SYNTAX_SPECIFIED in BASIC — buf.build/docs/lint/rules (via Context7), MEDIUM
- buf `--error-format` values and "build errors or check violations" semantics; text/JSON examples; `config-ignore-yaml` — buf.build/docs CLI reference + quickstart (via Context7), MEDIUM
- `COMPILE` annotation construction from compiler `ErrorWithPos` — bufbuild/buf source, `private/bufpkg/bufprotocompile/bufprotocompile.go` (raw.githubusercontent.com), cross-verified with docs: HIGH
- Lint command flow (build first, compile error aborts before rules, findings to stdout, `ErrFileAnnotation`) — bufbuild/buf source, `cmd/buf/internal/command/lint/lint.go`, HIGH
- Exit code 100 rationale — bufbuild/buf source, `private/buf/bufctl/bufctl.go` (`ExitCodeFileAnnotation = 100`; comment: distinguish user-parsable errors from system errors), HIGH
- bufcheck engine internals: RuleSpecBuilder, rule handlers over linked descriptors, comment-ignore matching only for located annotations, associated source paths — bufbuild/buf source, `bufcheckserverutil/rule_spec_builder.go`, `bufcheck/client.go`, `bufcheckserverhandle/lint.go` (via Context7 `/bufbuild/buf`), MEDIUM
- protolint full feature set (compiler-free trade-off, -fix, -auto_disable, disable/enable comments, plugins, reporters, exit codes 0/1/2) — github.com/yoheimuta/protolint README (raw fetch), MEDIUM
- protocompile stage pipeline, concurrent resolver semantics, error-accumulating Reporter, parse-without-link pattern — github.com/bufbuild/protocompile README + `resolver.go` + `reporter/reporter.go` (via Context7 `/bufbuild/protocompile`), MEDIUM
- EasyP current behavior (groups, output format `path:line:col:source message (rule)`, exit 1/2, suppression dialects, ignore/ignore_only) — repo source at `5ebab04`: `internal/rules/builder.go`, `internal/core/lint.go`, `internal/api/lint.go`, HIGH (first-party)

---
*Feature research for: compiler-backed protobuf lint engines (lint-easypc-infra milestone)*
*Researched: 2026-11-17*
