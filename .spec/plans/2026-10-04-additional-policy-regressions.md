# Additional v1 policy regressions

## Confirmed defects

The audit continues from EasyP `ece51fe` and easyp-test `5796d16`.

1. Consumer lint and breaking select local replacement policies when the
   dependency has legacy EasyP metadata, Buf metadata, or a nested native module.
   The previous filter only recognizes the replacement's exact native module
   directory.
2. A plain replacement under `roots .` becomes both a checked repository file
   and an import. Lint reports dependency-local findings; breaking compilation
   fails because the same protobuf symbols are compiled twice.
3. Absolute replacements inside the repository are excluded in the current
   checkout but not its baseline snapshot, producing duplicate descriptors or
   false deletions. Both plain and native dependency layouts are affected.

## Implementation and preserved contracts

- Identify import-only sources by containment in a declared replacement path,
  before reading their module or policy metadata.
- Inspect consuming manifests first; only native manifests declare replacements.
- An explicitly selected replacement file/directory remains a policy target.
- Selecting a replacement does not implicitly select its own local replacements.
- Filesystem aliases and deleted selected paths retain their physical identity.
- Baseline scope discovery receives the original repository root explicitly,
  rebasing absolute in-repository replacements into the scanned snapshot.
- Independent workspace modules and similarly named sibling directories remain
  checked. Compatibility adapters still provide dependency roots and requirements.
- Breaking retains comparison of imported contracts under the consumer policy;
  the fix removes duplicate targets and unrelated policies, not import checking.
- Frozen mode continues to reject local replace directives.

Independent review also caught a false positive in the initial containment fix:
an unused replacement pointing to a proto file hid that consumer file. Target
classification now requires a directory. CLI regression tests first reproduced
the failure; sibling and independent-module cases serve as negative controls.

## Verification plan

- [x] Extend CLI regressions before changing production code: native,
  legacy-with/without-manifest, Buf v1/v1-workspace/v2-workspace, nested modules,
  and plain replacements. Include explicit-selection and file-preservation checks.
- [x] Run against the unchanged `ece51fe` CLI and confirm failures.
- [x] Add focused parallel table-driven unit cases for source selection, sibling
  isolation, nested replacement ownership, aliases, and missing selected paths.
- [x] Run the complete source tests with race detection.
- [x] Run the complete v1 regression suite, public dependency regressions, and
  Python gRPC runtime regressions using freshly built CLI/MCP binaries.
- [x] Run vet, pinned Go lint, schema/proto checks, module verification, and ARM
  compilation. Review the final changes independently before committing/pushing.

Baseline failures: `/tmp/easyp-additional-baseline-all-red.log` (18 failures,
12 passing controls against unchanged `ece51fe`). The legacy protobuf manifest
fixture is a passing compatibility control.
Focused fixed regressions: `/tmp/easyp-additional-policies-green.log`.
Missing-alias unit failure and correction:
`/tmp/easyp-additional-unit-{red,green}.log`.
Absolute-path and file-target failures:
`/tmp/easyp-additional-{absolute,scope}-red.log`.
The final targeted suite has 30 CLI cases, including 6 absolute/plain/native/
alias/inherited-policy cases and 6 scope-isolation controls.

## Final verification — 2026-10-04

Tested binaries: `/tmp/easyp-additional-fixes` and
`/tmp/easyp-additional-fixes-mcp`, built from the final working tree. Commands
used `GOCACHE=/tmp/easyp-codex-go-cache`; E2E also used `EASYP_SOURCE` pointing
to this checkout and `EASYP_TEST_PYTHON=/tmp/easyp-python-grpc-runtime/bin/python`.

| Command | Result |
| --- | --- |
| EasyP `go test -race -count=1 ./...` | PASS. |
| easyp-test `go test -race -tags=v1,live_dependencies,python_grpc ./tests/e2e/v1 -count=1 -parallel=4 -v` | PASS, 151.555s. Includes the complete standard v1 suite, 30 targeted policy cases, 66 Go SDK compilations, 11 real upstream dependency cases, and 8 Python gRPC runtime cases. |
| easyp-test `go test -race -tags=v1,python_grpc ./tests/e2e/v1 -count=1 -parallel=4` | PASS, 84.726s after the final error-guard readability change; source unit/race, lint, vet and ARM build were also rerun. |
| easyp-test `go test -race -count=1 ./internal/... ./cmd/...` | PASS. |
| EasyP `go vet ./...`; easyp-test `go vet -tags=v1,live_dependencies,python_grpc ./tests/e2e/v1` | PASS. |
| `task lint:go LOCAL_BIN=/tmp/easyp-quality-tools/golangci-lint-2.14.0-darwin-arm64` | PASS, 0 issues. |
| `task lint:docker` | PASS. |
| `task schema:check`; `task proto:check` | PASS. |
| `go mod verify` | PASS. |
| `GOOS=linux GOARCH=arm GOARM=7 go build ./cmd/easyp ./cmd/easyp-mcp` | PASS. |
| Go formatting and `git diff --check` in both repositories | PASS. |

Independent read-only review reproduced the absolute-baseline defect and the
initial file-target false positive. After regressions and corrections, it found
no remaining confirmed defects, including additional deleted/aliased selections.
Full verification logs are `/tmp/easyp-additional-final-*.log`.
The optional external Python gRPC plugin suite was not run; the builtin Python
and builtin gRPC Python runtime cases were run.
