---
schema_version: 1
open_count: 9
waived_count: 0
fixed_count: 0
total_count: 9
last_updated: 2026-11-20T15:23:46.927Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 01 | deviation | internal/adapters/plugin/builtin_run_test.go |  | Pre-existing wazero v1.9.0 data race (GetWazeroVersion) fails internal/adapters/plugin under -race; reproduced on pristine HEAD, unrelated to 01-01 vendor drop; full-suite race leg red on this machine | open |  | 2026-11-18T10:33:54.860Z |  |
| 2 | 01 | deviation | .golangci.yml | 20 | Upstream QF1001 in byte-verbatim casing.go silenced by finding-scoped exclusion rule (plan 01-04 Rule 3; prohibition letter vs gate-green conflict) - ratify interpretation, re-evaluate at next re-sync | open |  | 2026-11-18T21:54:53.254Z |  |
| 3 | 04 | deviation | internal/rules/enum_value_prefix_linked.go | 38 | pascalToUpperSnake carried byte-verbatim under rename-mapped name pascalToUpperSnakeLinked (frozen old file owns the original name); Phase 5 drops the suffix | open |  | 2026-11-19T20:52:18.040Z |  |
| 4 | 04 | deviation | internal/rules/parity_default_test.go | 139 | v1alpha.proto cell verdict follows the live old-engine oracle (issues at 0:0), not the plan's 'clean' annotation - the main-vs-lab tooth (lab strips the extension) | open |  | 2026-11-19T20:52:24.164Z |  |
| 5 | 04 | deviation | internal/rules/parity_comments_test.go |  | presence_basics fixture corrected against the live old-engine oracle: DOC_ENUM_OTHER was left undocumented (COMMENT_ENUM_VALUE co-fired, old-side 9 vs hand-derived 8); the documenting comment was added so the documented enum's values are all documented | open |  | 2026-11-19T21:06:40.363Z |  |
| 6 | 04 | deviation | internal/rules/parity_comments_test.go |  | shallow_shapes fixture corrected: NestedEnum was a direct child of the TOP-LEVEL message so main's +1 level correctly reported it (old-side 4 vs planned 2); the enum moved inside the nested message so the cell isolates the two-level zero-findings boundary | open |  | 2026-11-19T21:06:40.510Z |  |
| 7 | 04 | deviation | internal/rules/parity_full_registry_test.go |  | Task-3 corpus file created although the task's <files> list is empty (Rule 3): the full-registry closure acceptance criterion requires the broad 41-rule corpus to live in a test file; parity_full_registry_test.go (TestLinkedParityFullRegistry) is its home | open |  | 2026-11-20T15:23:46.354Z |  |
| 8 | 04 | deviation | internal/rules/parity_full_registry_test.go |  | Fixture corrections before GREEN (live old-engine oracle): the dir_only_msg suppression directive was missing from the first draft and the public/weak imports used unqualified references (parsed against the SOURCE package), so all three imports read unused; directive added and references qualified (dep.v1.Outer, wdep.v1.WDepMsg) | open |  | 2026-11-20T15:23:46.565Z |  |
| 9 | 04 | deviation | internal/rules/parity_full_registry_test.go |  | Fixture correction before GREEN: the rpc payload messages (GetReq/GetResp/DoStreamRequest/DoStreamResponse/DupRequest/DupResponse/DupAgainResponse) were undeclared, so the linked engine's full type resolution routed the corpus to COMPILE (8 findings) while the syntax-only old engine reported rules; documented PascalCase declarations restored compiling parity | open |  | 2026-11-20T15:23:46.927Z |  |

````json
[
  {
    "id": 1,
    "kind": "deviation",
    "phase": "01",
    "file": "internal/adapters/plugin/builtin_run_test.go",
    "line": null,
    "description": "Pre-existing wazero v1.9.0 data race (GetWazeroVersion) fails internal/adapters/plugin under -race; reproduced on pristine HEAD, unrelated to 01-01 vendor drop; full-suite race leg red on this machine",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-18T10:33:54.860Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 2,
    "kind": "deviation",
    "phase": "01",
    "file": ".golangci.yml",
    "line": 20,
    "description": "Upstream QF1001 in byte-verbatim casing.go silenced by finding-scoped exclusion rule (plan 01-04 Rule 3; prohibition letter vs gate-green conflict) - ratify interpretation, re-evaluate at next re-sync",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-18T21:54:53.254Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 3,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/enum_value_prefix_linked.go",
    "line": 38,
    "description": "pascalToUpperSnake carried byte-verbatim under rename-mapped name pascalToUpperSnakeLinked (frozen old file owns the original name); Phase 5 drops the suffix",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-19T20:52:18.040Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 4,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_default_test.go",
    "line": 139,
    "description": "v1alpha.proto cell verdict follows the live old-engine oracle (issues at 0:0), not the plan's 'clean' annotation - the main-vs-lab tooth (lab strips the extension)",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-19T20:52:24.164Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 5,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_comments_test.go",
    "line": null,
    "description": "presence_basics fixture corrected against the live old-engine oracle: DOC_ENUM_OTHER was left undocumented (COMMENT_ENUM_VALUE co-fired, old-side 9 vs hand-derived 8); the documenting comment was added so the documented enum's values are all documented",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-19T21:06:40.363Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 6,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_comments_test.go",
    "line": null,
    "description": "shallow_shapes fixture corrected: NestedEnum was a direct child of the TOP-LEVEL message so main's +1 level correctly reported it (old-side 4 vs planned 2); the enum moved inside the nested message so the cell isolates the two-level zero-findings boundary",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-19T21:06:40.510Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 7,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_full_registry_test.go",
    "line": null,
    "description": "Task-3 corpus file created although the task's <files> list is empty (Rule 3): the full-registry closure acceptance criterion requires the broad 41-rule corpus to live in a test file; parity_full_registry_test.go (TestLinkedParityFullRegistry) is its home",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-20T15:23:46.354Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 8,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_full_registry_test.go",
    "line": null,
    "description": "Fixture corrections before GREEN (live old-engine oracle): the dir_only_msg suppression directive was missing from the first draft and the public/weak imports used unqualified references (parsed against the SOURCE package), so all three imports read unused; directive added and references qualified (dep.v1.Outer, wdep.v1.WDepMsg)",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-20T15:23:46.565Z",
    "resolved_at": null,
    "milestone": null
  },
  {
    "id": 9,
    "kind": "deviation",
    "phase": "04",
    "file": "internal/rules/parity_full_registry_test.go",
    "line": null,
    "description": "Fixture correction before GREEN: the rpc payload messages (GetReq/GetResp/DoStreamRequest/DoStreamResponse/DupRequest/DupResponse/DupAgainResponse) were undeclared, so the linked engine's full type resolution routed the corpus to COMPILE (8 findings) while the syntax-only old engine reported rules; documented PascalCase declarations restored compiling parity",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-11-20T15:23:46.927Z",
    "resolved_at": null,
    "milestone": null
  }
]
````
