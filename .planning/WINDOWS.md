---
schema_version: 1
open_count: 6
waived_count: 0
fixed_count: 0
total_count: 6
last_updated: 2026-11-19T21:06:40.510Z
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
  }
]
````
