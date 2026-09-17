<!-- scope: **/*flag*, **/*toggle*, **/*feature* -->
# Feature Flags Documentation Template

## What This Generates

- `.spec/FEATURE_FLAGS.md` — documents the feature flag system, flag inventory, lifecycle, and rollout strategy

## Instructions

Analyze the project to identify any feature flag or feature toggle system. Look for:
- SDK imports (LaunchDarkly, Unleash, Flagsmith, Split.io)
- Custom flag implementations (config-based, env var-based, database-based)
- Flag evaluation calls in application code
- A/B testing or experimentation frameworks

If no feature flag system is found, **skip this template entirely** — do not generate an empty document.

### File: .spec/FEATURE_FLAGS.md

#### Structure:

##### 1. Overview
- **System**: which flag provider or custom implementation is used
- **Architecture**: how flags are loaded (SDK init, config file, env vars, database table)
- **Evaluation flow**: where in the request lifecycle flags are checked (middleware, service layer, per-handler)
- **Default behavior**: what happens when the flag service is unavailable (fail open / fail closed)

##### 2. Flag Inventory

Table of all active feature flags:

| Flag Name | Type | Description | Owner | Created | Expiry | Default |
|-----------|------|-------------|-------|---------|--------|---------|
| `enable_new_checkout` | boolean | New checkout flow | @payments | 2025-01-15 | 2025-04-15 | false |
| `search_algorithm` | multivariate | A/B test for search | @search | 2025-02-01 | — | `v1` |

Types: `boolean`, `multivariate` (string/number variants), `percentage` (0–100).

Mark flags as: 🟢 Active, 🟡 Rolling out, 🔴 Scheduled for removal.

##### 3. Lifecycle
Document the flag lifecycle in this project:
1. **Creation** — who can create flags, where they are defined, naming conventions
2. **Rollout** — gradual rollout strategy (percentage, user segments, regions)
3. **Stabilization** — when a flag is considered stable (metrics period, error rate threshold)
4. **Permanent or removal** — decision criteria: does the flag become permanent config or get removed?
5. **Removal process** — code cleanup steps, how to find all evaluation points

##### 4. Implementation Pattern
Show the code pattern used to evaluate flags in this project:
- **Evaluation call**: actual code example from the project (e.g., `flags.IsEnabled("feature_x", user)`)
- **Context passing**: how user/request context reaches the flag evaluation
- **Type safety**: is there a typed wrapper or raw string keys?
- **Fallback**: what value is returned if evaluation fails

##### 5. Testing with Flags
- **Unit tests**: how to override flag values in tests (mock, test config, env var)
- **Integration tests**: how to test both sides of a flag (enabled and disabled)
- **Test matrix**: which flag combinations are tested together
- **CI behavior**: are flags forced to a specific value in CI?

##### 6. Rollout Strategy
- **Percentage rollout**: how to do gradual rollout (1% → 10% → 50% → 100%)
- **User targeting**: can flags be enabled per user, per org, per region?
- **Kill switch**: how to instantly disable a flag in production
- **Monitoring**: what metrics to watch during rollout (error rate, latency, conversion)
- **Rollback**: how to roll back a flag without a code deploy

##### 7. Cleanup Policy
- **Stale flag detection**: how to find flags that are 100% rolled out but not removed from code
- **Linting / static analysis**: any tooling to detect dead flag references
- **Cleanup cadence**: how often the team reviews and removes old flags
- **Technical debt tracking**: how stale flags are tracked (issues, labels, dashboards)

##### 8. Configuration
How flags are configured per environment:

| Environment | Source | Sync | Notes |
|-------------|--------|------|-------|
| Development | `.env` / config file | Manual | All flags enabled by default |
| Staging | Flag service | Real-time | Mirrors production targeting |
| Production | Flag service | Real-time | Gradual rollout rules apply |

- **Local override**: how developers enable/disable flags locally

## General Rules

- Language: English
- Use actual flag names and code from the project — do not invent examples
- If no feature flag system exists, do not generate this file
- Mark flags scheduled for removal — this document should help cleanup efforts
