---
schema_version: 1
open_count: 2
waived_count: 0
fixed_count: 0
total_count: 2
last_updated: 2026-11-18T21:54:53.254Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 01 | deviation | internal/adapters/plugin/builtin_run_test.go |  | Pre-existing wazero v1.9.0 data race (GetWazeroVersion) fails internal/adapters/plugin under -race; reproduced on pristine HEAD, unrelated to 01-01 vendor drop; full-suite race leg red on this machine | open |  | 2026-11-18T10:33:54.860Z |  |
| 2 | 01 | deviation | .golangci.yml | 20 | Upstream QF1001 in byte-verbatim casing.go silenced by finding-scoped exclusion rule (plan 01-04 Rule 3; prohibition letter vs gate-green conflict) - ratify interpretation, re-evaluate at next re-sync | open |  | 2026-11-18T21:54:53.254Z |  |

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
  }
]
````
