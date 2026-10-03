# Developer tooling and pinned-dependency diagnostics

## Developer task repairs

Tools live under the Taskfile root, not the caller's working directory. Paths
and explicit LOCAL_BIN overrides remain quoted through aggregate lint tasks.
init installs fixed tool versions and runs go mod download, never go get.
Go and root-Dockerfile lint checks have separate targets; lint attempts both and
returns failure if either fails. It does not hide existing source findings.
The old missing Docker/base and Docker/lint build helpers are removed. docker
and docker:build build the current root Dockerfile locally; neither publishes.
Release publishing stays in the existing release workflow.

Optional Mockery targets follow core.Rule, core.CurrentProjectGitWalker and
adapters/console.Console. The normal test suite uses handwritten doubles.
Generated mocks may introduce test dependencies: run go mod tidy after choosing
to add them. Verify generated code in a disposable checkout; do not commit
unused generated files merely to exercise a developer task.

scripts/check-dev-tools.sh checks command dispatch in an isolated directory with
spaces and quotes, using trace-only tools on an allowlisted PATH. It tests
failure propagation, missing executables, local-only Docker behavior, root-based
tool lookup and explicit overrides. It performs no downloads or real Docker/Go
operations. CI invokes the same dev-tools:check target.

## Remaining B-13 diagnostics

A cached content mismatch identifies the exact installed directory and instructs
the operator to inspect or quarantine that entry, retry mod download, and retain
the existing protobuf.lock. Nothing is automatically deleted, repaired or repinned.
A newly downloaded mismatch does not receive misleading local-cache repair advice.
Failure to query a repository is not reported as proof that a tag does not exist.
Successful history retrieval without the pinned commit advises restoring access
to that revision instead of substituting HEAD. Errors retain their Git causes;
cancellation during a Git subprocess remains detectable with errors.Is.

This does not add an offline fallback for tag validation, weaken integrity checks,
change version selection, or implement X-20/Buf registry mappings.

## Agent skills

The existing go-code-style, go-testing and legacy-named epctl-commands skills now
refer to the EasyP CLI module and urfave/cli v2. They use actual error/interface
owners and document process-state exceptions to parallel tests. Skill names and
paths are retained so existing references remain valid.

## Findings exposed by the repaired checks

The WASM runtime now propagates close failures rather than discarding them.
Removed an ineffectual map-value assignment in lint batching and simplified
string-builder formatting and YAML quote scanning without changing their scope.
Staticcheck ST1005 remains active for prose errors; its exclusion only matches
exported identifier-style wrapping labels required by the existing code style.
This is not a blanket disable for staticcheck or for arbitrary error messages.
Runtime-image APK packages have explicit overridable version build arguments,
verified against Alpine 3.22. Update those pins when updating runtime packages.
