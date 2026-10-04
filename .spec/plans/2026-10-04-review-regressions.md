# v1 review regressions and fixes

## Goal

Address every finding from the review of `d70401e`, including documentation.
Add CLI regressions in `easyp-test` before changing EasyP. Successful Go
generation must produce an SDK that compiles, not merely exit successfully.

## Contracts to preserve

- `options.go.package_prefix` works with managed mode omitted, enabled, or disabled.
- Managed disable selectors and explicit overrides keep their precedence.
- Path markers retain their existing file path semantics.
- Ordinary option prefixes derive multi-segment Go paths from protobuf packages.
- Import-only local replacements do not contribute consumer lint/breaking policies.
- Explicitly selecting a replacement still validates its own policy.
- Git modules are resolved from the requested revision, independently of current
  checkout metadata. Unrelated parent worktrees must not be treated as modules.
- The v1 default breaking profile is FILE; low-level legacy callers stay supported.

## Work

1. **Regressions first**
   - Add a managed/prefix matrix: omitted/enabled/disabled, ordinary prefixes,
     single-segment packages, path markers, disabled Go options, and overrides.
     Check descriptor `go_package`, generated locations, and compile each Go SDK.
   - Add overlapping proto basenames (`user.proto`, `user_types.proto`) with
     distinct protobuf packages and compile the generated SDK.
   - Add consumer lint and breaking cases with replacement-local legacy/invalid
     policies; verify ignored policies and explicit selection separately.
   - Add local nested Git modules resolved by historical tag and commit after
     removal of their live manifest, including cold-cache frozen download.
   - Run the new suites against the unmodified review binary and record failures.
2. **Go output placement (Must Fix + Should Fix)**
   - Use the effective Go import path after managed rules to place Go outputs.
   - Match the generating proto unambiguously; preserve plugin-specific filenames
     and outputs whose location is already controlled by the plugin.
   - Relocate only explicit `paths=source_relative`; cover omitted/default import
     paths when the proto source directory coincides with the Go import path.
     Repeated `paths` options use the plugin's last-value priority.
   - Add focused table-driven unit tests for relocation and rule interactions.
3. **Consumer policies (Should Fix)**
   - Filter import-only replacement modules before reading their policies.
   - Preserve explicit target selection and error propagation.
4. **Historical local Git modules (Should Fix)**
   - Discover the repository without requiring live module metadata.
   - Validate nested module identity at the fetched revision; retain isolation
     from unrelated parent repositories.
5. **Documentation (Nice to Have)**
   - Correct stale v1 breaking-default statements in AGENTS, agent rules, and
     architecture documentation. Check other current references for consistency.
6. **Verification and final review**
   - Run the new regressions with the fixed binary, then the full v1 E2E suite.
   - Run Go formatting, `go test -race -count=1 ./...`, `go vet ./...`, project
     Go lint, schema drift checks, module verification, and the ARM build gate.
   - Review both final diffs, preserve unrelated local files, and report exact
     commands/results and any unavailable checks.

## Progress

- [x] Regression suites added and baseline failures confirmed.
- [x] Effective Go output placement and ambiguous basenames fixed.
- [x] Replacement policy ordering fixed.
- [x] Historical local Git module discovery fixed.
- [x] Documentation corrected.
- [x] Full verification and final diff review complete.

## Implementation

| Concern | EasyP files | Regression suites in `easyp-test/tests/e2e/v1` |
| --- | --- | --- |
| Effective Go paths and source identity | `internal/generation/config.go`, `internal/core/core.go`, `generate.go`, `go_package_output.go`, `go_package_output_test.go` | `managed_go_layout_test.go`: 42 managed/layout cases, 21 plugin-layout cases, 3 overlapping-name cases; all 66 SDKs compile. |
| Import-only replacement policies | `internal/api/lint_v1.go`, `breaking_v1.go` | `review_consumer_policies_test.go`: legacy/invalid policies, lint/breaking, explicit-selection controls. |
| Historical local Git discovery | `internal/adapters/gitmodules/git_source.go`, `git_source_test.go` | `review_historical_local_module_test.go`: 4 historical revisions/layouts, 2 unrelated-worktree rejection cases. |
| Current breaking documentation | `AGENTS.md`, `.spec/agent-rules.md`, `ARCHITECTURE.md`, `ERRORS.md`, `DOMAIN.md` | `review_breaking_docs_test.go`: all 5 current references. |

The local candidate unit test now checks the nearest worktree boundary; module
acceptance is checked against the requested revision by the CLI regressions.
No live manifest is required to propose a repository candidate.

Independent final review found and verified two additional branches of the Go
layout defect: omitted `paths` uses import layout, and repeated `paths` use the
last value. Both first failed a new regression against the intermediate binary,
then passed after correction. No confirmed findings remained in final review.

## Verification — 2026-10-04

Source baseline: EasyP `d70401e`; easyp-test `f71ac96`. Checks below cover the
working-tree fixes. Binaries: `/tmp/easyp-review-fixes` and
`/tmp/easyp-review-fixes-mcp`. Go cache: `/tmp/easyp-codex-go-cache`.

| Command | Result |
| --- | --- |
| Targeted `go test -race -tags=v1 ./tests/e2e/v1 -run '^TestReview(ManagedGoLayout\|GoPluginImportLayout\|GoLayoutOverlappingProtoNames\|ReplacementPolicyIsImportOnly\|HistoricalLocalModule\|UnrelatedLocalWorktreeIsNotModule\|BreakingDefaultDocumentation)$' -count=1 -parallel=4 -v` | PASS: 81 cases, including 66 Go/Go gRPC SDK compilations. |
| EasyP `go test -race -count=1 ./...` | PASS. |
| easyp-test `go test -race -tags=v1 ./tests/e2e/v1 -count=1 -parallel=4` | PASS: 77.482s with the final CLI/MCP binaries and `EASYP_SOURCE` set to the tested checkout. |
| easyp-test `go test -race -count=1 ./internal/... ./cmd/...` | PASS. |
| EasyP `go vet ./...`; easyp-test `go vet -tags=v1 ./tests/e2e/v1` | PASS. |
| `task lint:go LOCAL_BIN=/tmp/easyp-quality-tools/golangci-lint-2.14.0-darwin-arm64` | PASS: 0 issues, pinned version 2.14.0. |
| `task schema:check` | PASS: no schema drift. |
| `task proto:check` | PASS: validation, lint, lock and deterministic generation. |
| `go mod verify` | PASS: all modules verified. |
| `GOOS=linux GOARCH=arm GOARM=7 go build ./cmd/easyp ./cmd/easyp-mcp` | PASS. |
| Go formatting and `git diff --check` in both repositories | PASS. |

Baseline failure logs are `/tmp/easyp-review-fixes-red.log`,
`/tmp/easyp-review-fixes-unit-red.log`, and
`/tmp/easyp-review-fixes-docs-red.log`. Final logs use
`/tmp/easyp-review-fixes-final-{regressions,unit,v1,lint,schema,proto}.log`.
Docker lint and opt-in live upstream/Python/plugin tests were not run; the
standard v1 suite and all new local regressions were run. After verification,
the user requested committing and pushing these fixes directly for review:
EasyP to `v1.0`, easyp-test regressions to `master`.
