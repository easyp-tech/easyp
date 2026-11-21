---
schema_version: 1
open_count: 0
waived_count: 0
fixed_count: 17
total_count: 17
last_updated: 2026-11-21T09:25:52.405Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 01 | deviation | internal/adapters/plugin/builtin_run_test.go |  | Pre-existing wazero v1.9.0 data race (GetWazeroVersion) fails internal/adapters/plugin under -race; reproduced on pristine HEAD, unrelated to 01-01 vendor drop; full-suite race leg red on this machine | fixed |  | 2026-11-18T10:33:54.860Z | 2026-11-21T09:25:47.178Z |
| 2 | 01 | deviation | .golangci.yml | 20 | Upstream QF1001 in byte-verbatim casing.go silenced by finding-scoped exclusion rule (plan 01-04 Rule 3; prohibition letter vs gate-green conflict) - ratify interpretation, re-evaluate at next re-sync | fixed |  | 2026-11-18T21:54:53.254Z | 2026-11-21T09:25:47.313Z |
| 3 | 04 | deviation | internal/rules/enum_value_prefix_linked.go | 38 | pascalToUpperSnake carried byte-verbatim under rename-mapped name pascalToUpperSnakeLinked (frozen old file owns the original name); Phase 5 drops the suffix | fixed |  | 2026-11-19T20:52:18.040Z | 2026-11-21T09:25:47.654Z |
| 4 | 04 | deviation | internal/rules/parity_default_test.go | 139 | v1alpha.proto cell verdict follows the live old-engine oracle (issues at 0:0), not the plan's 'clean' annotation - the main-vs-lab tooth (lab strips the extension) | fixed |  | 2026-11-19T20:52:24.164Z | 2026-11-21T09:25:47.798Z |
| 5 | 04 | deviation | internal/rules/parity_comments_test.go |  | presence_basics fixture corrected against the live old-engine oracle: DOC_ENUM_OTHER was left undocumented (COMMENT_ENUM_VALUE co-fired, old-side 9 vs hand-derived 8); the documenting comment was added so the documented enum's values are all documented | fixed |  | 2026-11-19T21:06:40.363Z | 2026-11-21T09:25:48.148Z |
| 6 | 04 | deviation | internal/rules/parity_comments_test.go |  | shallow_shapes fixture corrected: NestedEnum was a direct child of the TOP-LEVEL message so main's +1 level correctly reported it (old-side 4 vs planned 2); the enum moved inside the nested message so the cell isolates the two-level zero-findings boundary | fixed |  | 2026-11-19T21:06:40.510Z | 2026-11-21T09:25:48.461Z |
| 7 | 04 | deviation | internal/rules/parity_full_registry_test.go |  | Task-3 corpus file created although the task's <files> list is empty (Rule 3): the full-registry closure acceptance criterion requires the broad 41-rule corpus to live in a test file; parity_full_registry_test.go (TestLinkedParityFullRegistry) is its home | fixed |  | 2026-11-20T15:23:46.354Z | 2026-11-21T09:25:48.863Z |
| 8 | 04 | deviation | internal/rules/parity_full_registry_test.go |  | Fixture corrections before GREEN (live old-engine oracle): the dir_only_msg suppression directive was missing from the first draft and the public/weak imports used unqualified references (parsed against the SOURCE package), so all three imports read unused; directive added and references qualified (dep.v1.Outer, wdep.v1.WDepMsg) | fixed |  | 2026-11-20T15:23:46.565Z | 2026-11-21T09:25:49.120Z |
| 9 | 04 | deviation | internal/rules/parity_full_registry_test.go |  | Fixture correction before GREEN: the rpc payload messages (GetReq/GetResp/DoStreamRequest/DoStreamResponse/DupRequest/DupResponse/DupAgainResponse) were undeclared, so the linked engine's full type resolution routed the corpus to COMPILE (8 findings) while the syntax-only old engine reported rules; documented PascalCase declarations restored compiling parity | fixed |  | 2026-11-20T15:23:46.927Z | 2026-11-21T09:25:49.605Z |
| 10 | 05 | deviation | internal/rules/golden_full_registry_test.go |  | 05-01: plan's Task-1 verify chain (git status \| grep -vc \| grep -q under pipefail) cannot exit 0 when the guard PASSES (grep -c exits 1 on zero selected lines); gate run with corrected plumbing preserving the exact intended semantics (count == 0), test green; re-run before any narration | fixed |  | 2026-11-20T19:21:06.311Z | 2026-11-21T09:25:50.073Z |
| 11 | 05 | deviation | internal/core/lint_linked_matrix_test.go |  | 05-02: WR-01 cell re-derived against live source (Rule 3): the plan's directory fixture cannot drive the resolver error path — FSWalker.Open refuses non-regular files with ErrUnsupported and a directory named dep.proto is walked as a corpus file so pass-1 readErrs skip linking before the resolver runs; permission bits are not root-safe in CI; the cell injects an *os.PathError (os.ErrPermission) through Options.OpenSourceFile exercising the identical openDiskSource funnel, and the post-fix expectation is re-derived from the plan's 'returns no error' to a command error carrying the OS cause — the plan's own must-have truth ('behaves like the old per-file engine') and threat model T-05-02 ('the OS cause surfaces as a command error') both require the command-error outcome; verified pre-fix RED shows the misleading COMPILE remark | fixed |  | 2026-11-20T19:51:54.957Z | 2026-11-21T09:25:50.556Z |
| 12 | 05 | deviation | internal/core/lint_linked_matrix_test.go |  | 05-02: WR-02 cell re-derived against live source (Rule 3): an engine-level RED is impossible because pass-1 readSource consults the same openDiskSource policy funnel, so the plan's 'run succeeds and lints b.proto from memory' pre-fix state does not exist (probe: the run errors with the exact os.ErrNotExist carrier pre-fix); the cell drives the lintResolver corpus branch directly (pre-fix it serves the denied content from memory — the actual WR-02 bypass) and pins openDiskSource's exact carrier plus the engine-level end-to-end denial | fixed |  | 2026-11-20T19:51:55.295Z | 2026-11-21T09:25:50.933Z |
| 13 | 05 | deviation | internal/api/lint_stdout_golden_test.go |  | 05-03: Gate-3 twin fixture corrected (Rule 3): the 05-01 corpus declared the same message (CleanFile) in two same-package files — legal for the old per-file engine but a redeclaration under the linked engine's full corpus linking, routing the whole run to COMPILE and vacating the rule-finding golden (research Pitfall 7 applied to Gate 3); Excluded_File.proto's message renamed to ExcludedMsg so the corpus compiles; the pinned stdout literals were re-captured from the byte-frozen pre-flip binary (built from flip-HEAD via git archive) over the corrected fixture and are byte-identical to the 05-01 literals, md5-stable across 3 runs; TestLintStdoutGolden green post-flip | fixed |  | 2026-11-20T20:29:31.645Z | 2026-11-21T09:25:51.068Z |
| 14 | 05 | deviation | internal/core/proto_info_read_test.go |  | 05-03: plan surface addition (Rule 3, blocking compile): internal/core/proto_info_read_test.go and link_test.go call the deleted lintProtoInfoRead directly but are absent from the plan's files_modified list; TestProtoInfoReadTwinsMirroring + TestLintProtoInfoReadErrorLabels died with their oracle, the two link_test.go fallback cells re-pointed to protoInfoFromSource (same openImportFile funnel, same typed OpenImportFileError carrier) — assertions preserved verbatim in intent | fixed |  | 2026-11-20T20:29:31.985Z | 2026-11-21T09:25:51.410Z |
| 15 | 05 | deviation | internal/core/lint_linked.go |  | 05-03: acceptance-criterion baselines reconciled against the plan's own prohibitions (Rule 3): (a) grep lintProtoInfoRead lands at 8 lines, not the criterion's EXPECT==1 — the byte-identical parity-pin comment+literal (2), the frozen proto_info_read.go comment 'twin of lintProtoInfoRead' (1), and the five exact-string proof cells asserting the pin text (5) are all mandated by the plan's canonical-six C1 prohibition and Phase-3 main-wins proof; the achieved invariant is zero LIVE references (callable call sites) to the deleted function; (b) the api runtime.go removed-lines gate's raw grep -c '^- ' counts the diff header, so the corrected plumbing is grep -c '^-[^-]' = 0 (comment-only addition preserved); (c) the zero-ref grep HEAD baseline was 399 at execution time vs the plan's recorded 384 (05-02's cells postdate the plan); EXPECT==0 achieved | fixed |  | 2026-11-20T20:29:32.513Z | 2026-11-21T09:25:51.750Z |
| 16 | 05 | deviation | internal/core/lint_linked_read.go | 139 | 05-04 Rule 3: protoInfoFromLinked renamed to protoInfoFromLink (plan action listed only the 41 port types; the zero-reference criterion 'LintRuleName or word-Linked' == 0 forced the rename, 4 occurrences: def + 2 comments + engine call) plus comment-only factual refreshes forced by the same criterion (41 port headers de-self-referentialized after the sweep, builder_linked.go registry doc, package_directory_match registry-name prose, GetRuleName doc comment) | fixed |  | 2026-11-20T20:59:41.829Z | 2026-11-21T09:25:52.067Z |
| 17 | 07 | deviation | internal/adapters/plugin/builtin_run_test.go |  | 07-04 Rule 1 deviation: REVIEW WR-01 window-probe snippet replaced by held-mutex ordering probe (false-positive under -race with correct mutex measured; mutex-block not synctest-durable) | fixed |  | 2026-11-21T05:32:04.987Z | 2026-11-21T09:25:52.405Z |

````json
[
  {
    "id": 1,
    "kind": "deviation",
    "phase": "01",
    "file": "internal/adapters/plugin/builtin_run_test.go",
    "line": null,
    "description": "Pre-existing wazero v1.9.0 data race (GetWazeroVersion) fails internal/adapters/plugin under -race; reproduced on pristine HEAD, unrelated to 01-01 vendor drop; full-suite race leg red on this machine",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-18T10:33:54.860Z",
    "resolved_at": "2026-11-21T09:25:47.178Z",
    "milestone": null
  },
  {
    "id": 2,
    "kind": "deviation",
    "phase": "01",
    "file": ".golangci.yml",
    "line": 20,
    "description": "Upstream QF1001 in byte-verbatim casing.go silenced by finding-scoped exclusion rule (plan 01-04 Rule 3; prohibition letter vs gate-green conflict) - ratify interpretation, re-evaluate at next re-sync",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-18T21:54:53.254Z",
    "resolved_at": "2026-11-21T09:25:47.313Z",
    "milestone": null
  },
  {
    "id": 3,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/enum_value_prefix_linked.go",
    "line": 38,
    "description": "pascalToUpperSnake carried byte-verbatim under rename-mapped name pascalToUpperSnakeLinked (frozen old file owns the original name); Phase 5 drops the suffix",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-19T20:52:18.040Z",
    "resolved_at": "2026-11-21T09:25:47.654Z",
    "milestone": null
  },
  {
    "id": 4,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_default_test.go",
    "line": 139,
    "description": "v1alpha.proto cell verdict follows the live old-engine oracle (issues at 0:0), not the plan's 'clean' annotation - the main-vs-lab tooth (lab strips the extension)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-19T20:52:24.164Z",
    "resolved_at": "2026-11-21T09:25:47.798Z",
    "milestone": null
  },
  {
    "id": 5,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_comments_test.go",
    "line": null,
    "description": "presence_basics fixture corrected against the live old-engine oracle: DOC_ENUM_OTHER was left undocumented (COMMENT_ENUM_VALUE co-fired, old-side 9 vs hand-derived 8); the documenting comment was added so the documented enum's values are all documented",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-19T21:06:40.363Z",
    "resolved_at": "2026-11-21T09:25:48.148Z",
    "milestone": null
  },
  {
    "id": 6,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_comments_test.go",
    "line": null,
    "description": "shallow_shapes fixture corrected: NestedEnum was a direct child of the TOP-LEVEL message so main's +1 level correctly reported it (old-side 4 vs planned 2); the enum moved inside the nested message so the cell isolates the two-level zero-findings boundary",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-19T21:06:40.510Z",
    "resolved_at": "2026-11-21T09:25:48.461Z",
    "milestone": null
  },
  {
    "id": 7,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_full_registry_test.go",
    "line": null,
    "description": "Task-3 corpus file created although the task's <files> list is empty (Rule 3): the full-registry closure acceptance criterion requires the broad 41-rule corpus to live in a test file; parity_full_registry_test.go (TestLinkedParityFullRegistry) is its home",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T15:23:46.354Z",
    "resolved_at": "2026-11-21T09:25:48.863Z",
    "milestone": null
  },
  {
    "id": 8,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_full_registry_test.go",
    "line": null,
    "description": "Fixture corrections before GREEN (live old-engine oracle): the dir_only_msg suppression directive was missing from the first draft and the public/weak imports used unqualified references (parsed against the SOURCE package), so all three imports read unused; directive added and references qualified (dep.v1.Outer, wdep.v1.WDepMsg)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T15:23:46.565Z",
    "resolved_at": "2026-11-21T09:25:49.120Z",
    "milestone": null
  },
  {
    "id": 9,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_full_registry_test.go",
    "line": null,
    "description": "Fixture correction before GREEN: the rpc payload messages (GetReq/GetResp/DoStreamRequest/DoStreamResponse/DupRequest/DupResponse/DupAgainResponse) were undeclared, so the linked engine's full type resolution routed the corpus to COMPILE (8 findings) while the syntax-only old engine reported rules; documented PascalCase declarations restored compiling parity",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T15:23:46.927Z",
    "resolved_at": "2026-11-21T09:25:49.605Z",
    "milestone": null
  },
  {
    "id": 10,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/rules/golden_full_registry_test.go",
    "line": null,
    "description": "05-01: plan's Task-1 verify chain (git status | grep -vc | grep -q under pipefail) cannot exit 0 when the guard PASSES (grep -c exits 1 on zero selected lines); gate run with corrected plumbing preserving the exact intended semantics (count == 0), test green; re-run before any narration",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T19:21:06.311Z",
    "resolved_at": "2026-11-21T09:25:50.073Z",
    "milestone": null
  },
  {
    "id": 11,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/core/lint_linked_matrix_test.go",
    "line": null,
    "description": "05-02: WR-01 cell re-derived against live source (Rule 3): the plan's directory fixture cannot drive the resolver error path — FSWalker.Open refuses non-regular files with ErrUnsupported and a directory named dep.proto is walked as a corpus file so pass-1 readErrs skip linking before the resolver runs; permission bits are not root-safe in CI; the cell injects an *os.PathError (os.ErrPermission) through Options.OpenSourceFile exercising the identical openDiskSource funnel, and the post-fix expectation is re-derived from the plan's 'returns no error' to a command error carrying the OS cause — the plan's own must-have truth ('behaves like the old per-file engine') and threat model T-05-02 ('the OS cause surfaces as a command error') both require the command-error outcome; verified pre-fix RED shows the misleading COMPILE remark",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T19:51:54.957Z",
    "resolved_at": "2026-11-21T09:25:50.556Z",
    "milestone": null
  },
  {
    "id": 12,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/core/lint_linked_matrix_test.go",
    "line": null,
    "description": "05-02: WR-02 cell re-derived against live source (Rule 3): an engine-level RED is impossible because pass-1 readSource consults the same openDiskSource policy funnel, so the plan's 'run succeeds and lints b.proto from memory' pre-fix state does not exist (probe: the run errors with the exact os.ErrNotExist carrier pre-fix); the cell drives the lintResolver corpus branch directly (pre-fix it serves the denied content from memory — the actual WR-02 bypass) and pins openDiskSource's exact carrier plus the engine-level end-to-end denial",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T19:51:55.295Z",
    "resolved_at": "2026-11-21T09:25:50.933Z",
    "milestone": null
  },
  {
    "id": 13,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/api/lint_stdout_golden_test.go",
    "line": null,
    "description": "05-03: Gate-3 twin fixture corrected (Rule 3): the 05-01 corpus declared the same message (CleanFile) in two same-package files — legal for the old per-file engine but a redeclaration under the linked engine's full corpus linking, routing the whole run to COMPILE and vacating the rule-finding golden (research Pitfall 7 applied to Gate 3); Excluded_File.proto's message renamed to ExcludedMsg so the corpus compiles; the pinned stdout literals were re-captured from the byte-frozen pre-flip binary (built from flip-HEAD via git archive) over the corrected fixture and are byte-identical to the 05-01 literals, md5-stable across 3 runs; TestLintStdoutGolden green post-flip",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T20:29:31.645Z",
    "resolved_at": "2026-11-21T09:25:51.068Z",
    "milestone": null
  },
  {
    "id": 14,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/core/proto_info_read_test.go",
    "line": null,
    "description": "05-03: plan surface addition (Rule 3, blocking compile): internal/core/proto_info_read_test.go and link_test.go call the deleted lintProtoInfoRead directly but are absent from the plan's files_modified list; TestProtoInfoReadTwinsMirroring + TestLintProtoInfoReadErrorLabels died with their oracle, the two link_test.go fallback cells re-pointed to protoInfoFromSource (same openImportFile funnel, same typed OpenImportFileError carrier) — assertions preserved verbatim in intent",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T20:29:31.985Z",
    "resolved_at": "2026-11-21T09:25:51.410Z",
    "milestone": null
  },
  {
    "id": 15,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/core/lint_linked.go",
    "line": null,
    "description": "05-03: acceptance-criterion baselines reconciled against the plan's own prohibitions (Rule 3): (a) grep lintProtoInfoRead lands at 8 lines, not the criterion's EXPECT==1 — the byte-identical parity-pin comment+literal (2), the frozen proto_info_read.go comment 'twin of lintProtoInfoRead' (1), and the five exact-string proof cells asserting the pin text (5) are all mandated by the plan's canonical-six C1 prohibition and Phase-3 main-wins proof; the achieved invariant is zero LIVE references (callable call sites) to the deleted function; (b) the api runtime.go removed-lines gate's raw grep -c '^- ' counts the diff header, so the corrected plumbing is grep -c '^-[^-]' = 0 (comment-only addition preserved); (c) the zero-ref grep HEAD baseline was 399 at execution time vs the plan's recorded 384 (05-02's cells postdate the plan); EXPECT==0 achieved",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T20:29:32.513Z",
    "resolved_at": "2026-11-21T09:25:51.750Z",
    "milestone": null
  },
  {
    "id": 16,
    "kind": "deviation",
    "phase": "05",
    "file": "internal/core/lint_linked_read.go",
    "line": 139,
    "description": "05-04 Rule 3: protoInfoFromLinked renamed to protoInfoFromLink (plan action listed only the 41 port types; the zero-reference criterion 'LintRuleName or word-Linked' == 0 forced the rename, 4 occurrences: def + 2 comments + engine call) plus comment-only factual refreshes forced by the same criterion (41 port headers de-self-referentialized after the sweep, builder_linked.go registry doc, package_directory_match registry-name prose, GetRuleName doc comment)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-20T20:59:41.829Z",
    "resolved_at": "2026-11-21T09:25:52.067Z",
    "milestone": null
  },
  {
    "id": 17,
    "kind": "deviation",
    "phase": "07",
    "file": "internal/adapters/plugin/builtin_run_test.go",
    "line": null,
    "description": "07-04 Rule 1 deviation: REVIEW WR-01 window-probe snippet replaced by held-mutex ordering probe (false-positive under -race with correct mutex measured; mutex-block not synctest-durable)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-11-21T05:32:04.987Z",
    "resolved_at": "2026-11-21T09:25:52.405Z",
    "milestone": null
  }
]
````
