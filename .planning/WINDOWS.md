---
schema_version: 1
open_count: 1
waived_count: 0
fixed_count: 0
total_count: 1
last_updated: 2026-11-18T10:33:54.860Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 01 | deviation | internal/adapters/plugin/builtin_run_test.go |  | Pre-existing wazero v1.9.0 data race (GetWazeroVersion) fails internal/adapters/plugin under -race; reproduced on pristine HEAD, unrelated to 01-01 vendor drop; full-suite race leg red on this machine | open |  | 2026-11-18T10:33:54.860Z |  |

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
  }
]
````
