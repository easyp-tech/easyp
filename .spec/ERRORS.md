<!-- generated: 2026-09-30, template: errors.md -->
# Error Handling

EasyP is a command-line application. It does not expose an HTTP or gRPC
server API, so it has no HTTP status mapping, gRPC status mapping, or
server-side JSON error envelope. The API layer in this document means the
command handlers in <code>internal/api</code>.

## Error Architecture

~~~text
CLI command handlers (internal/api)
  -> configuration parsing/validation (internal/config/v1)
  -> module operations and source resolution (internal/modules)
  -> generation orchestration (internal/generation)
  -> lint/breaking/compiler/plugin engines (internal/core)
  -> migration planning/application (internal/migration)
  -> adapters: Git/cache, module metadata, console and plugins
  -> filesystem, Git subprocesses, parsers and remote executors

Errors return upward with operation context; command-specific handlers
print findings, classify selected errors, or return the error to main.
~~~

These are cooperating packages, not one mandatory call chain. For example,
<code>mod tidy</code> calls <code>modules.Tidy</code> with a
<code>gitmodules.Cache</code>; it does not call a dependency method on
<code>core.Core</code>. Generation resolves module sources before constructing
the core compiler/plugin engine. Migration uses its own plan and filesystem
transaction.

Go <code>error</code> values normally propagate upward. Use
<code>errors.Is</code> for sentinel identity and <code>errors.As</code> for
structured error types; wrapping with <code>%w</code> preserves both. Ordinary
errors returned by <code>app.Run</code> reach <code>log.Fatal</code> in
<code>cmd/easyp/main.go</code> and exit 1. An action returning
<code>cli.Exit</code> instead supplies an <code>ExitCoder</code>: the default
urfave/cli handler prints its message and exits with the specified status
before <code>main</code>'s fallback. The explicit
<code>errExit</code> helper in <code>internal/api/runtime.go</code> logs with the
application logger and calls <code>os.Exit</code>. It is used by selected lint
and breaking error branches, not as a global mapper.

## Sentinel Catalog

These sentinels exist in the current source. Their Go identifiers do not
constitute a serialized error-code protocol.

| Name | Defined in | Meaning and handling |
|---|---|---|
| <code>core.ErrInvalidRule</code> | <code>internal/core/core.go</code> | Unknown or unavailable lint rule. Wrapped by rule selection/building and returned; no dedicated CLI exit mapping. |
| <code>core.ErrRepositoryDoesNotExist</code> | <code>internal/core/core.go</code> | Git repository is unavailable for breaking checks. The breaking handler logs an actionable message and exits 2. |
| <code>core.ErrEmptyInputFiles</code> | <code>internal/core/core.go</code> | <code>Core.PrepareGeneration</code> found no input protos. Current generation preparation returns the error; there is no warning-and-success special case in the v1 CLI. |
| <code>core.ErrRootOutsideProject</code> | <code>internal/core/breaking_check.go</code> | Breaking scan root is outside the Git repository. No dedicated CLI mapping; returned failures exit 1. |
| <code>v1.ErrLegacyConfiguration</code> | <code>internal/config/v1/legacy_detection.go</code> | Recognized legacy configuration requires explicit migration. The message points to <code>easyp migrate --module &lt;identity&gt;</code>; native parsers do not silently convert files. |
| <code>modules.ErrLockedVersionChanged</code> | <code>internal/modules/immutable_versions.go</code> | A previously locked semantic version now resolves to a different commit or content hash. The wrapped error includes old/new identities and asks for restoration or a new version; it returns through the command and exits 1. |
| <code>api.ErrHasLintIssue</code> | <code>internal/api/lint.go</code> | Lint findings were printed; the lint handler exits 1. |
| <code>api.ErrHasValidateIssue</code> | <code>internal/api/lint.go</code> | <code>validate-config</code> printed an invalid report; the returned error reaches <code>main</code> and exits 1. |
| <code>api.ErrBreakingCheckIssue</code> | <code>internal/api/breaking_check.go</code> | Breaking findings were printed; the breaking handler exits 1. |

Dependency resolution no longer relies on the removed model sentinels for
version lookup, installed-module metadata, lock entries, or hash mismatch.
Those cases are now ordinary contextual errors in the v1 module/cache flow,
except for the explicit <code>ErrLockedVersionChanged</code> identity above.
The <code>mod</code> handlers return those errors; they have no special
version-not-found <code>os.Exit</code> branch.

## Typed Errors and Diagnostic Records

| Type | Defined in | Handling |
|---|---|---|
| <code>*core.OpenImportFileError</code> | <code>internal/core/dom.go</code> | Carries <code>FileName</code>. Lint and breaking use <code>errors.As</code>, log “Cannot import file” with a file-name attribute, and exit 2. This mapping applies to this type, not every error mentioning an import. |
| <code>*core.GitRefNotFoundError</code> | <code>internal/core/dom.go</code> | Carries <code>GitRef</code>. Breaking logs “Cannot find git ref” with the ref and exits 2. |
| <code>console.RunError</code> | <code>internal/adapters/console/error.go</code> | Holds command, parameters, working directory, underlying error and stderr. Its error string includes command/error/stderr; it has no <code>Unwrap</code> method, so the underlying <code>Err</code> is not automatically discoverable with <code>errors.Is</code>. |
| <code>config.ValidationIssue</code> | <code>internal/config/validation_issue.go</code> | Diagnostic record with file/code/message/location/severity; not an <code>error</code> implementation. |
| <code>v1ListError</code> | <code>internal/api/ls_files_v1.go</code> | File-listing diagnostic with code/message; not an <code>error</code> implementation. |

No project-wide error interface or central domain-to-status conversion exists.
Choose a sentinel for identity-only branching and a typed error when a caller
needs structured context. Do not invent a sentinel or a status mapping merely
to mirror a diagnostic message.

## Module and Cache Error Flow

Native <code>protobuf.mod</code> supplies requirements, roots and local
replacements; <code>protobuf.lock</code> supplies immutable remote selections.
The source files below own their respective failure paths.

| Operation | Source | Failure behavior |
|---|---|---|
| Read native manifest | <code>internal/modules/project_files.go</code> | Read/parse errors retain context. Only <code>ReadModuleOrDefault</code> treats a missing manifest as a local root with no dependencies; <code>ReadManifest</code> callers require the manifest. |
| Read/validate lock | <code>internal/modules/locked_sources.go</code> | Missing or unsatisfied remote requirements instruct the user to run <code>easyp mod tidy</code>. A missing native lock alongside legacy <code>easyp.lock</code> wraps <code>ErrLegacyConfiguration</code>; renaming the old file is not conversion. |
| Resolve versions | <code>internal/modules/resolve.go</code> | Invalid versions, conflicting commit constraints, incompatible tag/commit selections and unstable versionless constraints return contextual errors. Deferred errors from provisional HEAD dependency graphs are reported only if that graph remains selected. |
| Guard recorded versions | <code>internal/modules/immutable_versions.go</code> | Reusing a semantic version whose commit/hash changed wraps <code>ErrLockedVersionChanged</code>. |
| Read local replacements | <code>internal/modules/local_sources.go</code> | Metadata/root errors and local dependency cycles return errors. Relative paths resolve from the owning manifest; absolute replacement paths remain absolute. |
| Check imports | <code>internal/modules/collisions.go</code>, <code>internal/modules/imports.go</code> | Duplicate paths, nonportable import paths and unreadable/invalid reachable sources fail the operation. Import validation traverses root sources and reachable dependency/built-in imports, resolving against local/dependency roots and embedded well-known files. Unrelated dependency files are not traversed. |
| Install locked content | <code>internal/adapters/gitmodules/download.go</code> | Validate the lock, install missing snapshots and verify hashes of both new and existing cached content. Hash mismatches or non-directory cache entries are errors, not successful cache misses. |
| Obtain pinned Git objects | <code>internal/adapters/gitmodules/object_cache.go</code> | Verify the requested commit after fetching. Unavailable pins do not silently become HEAD. Cancellation and Git failures propagate. |
| Persist manifest/lock | <code>internal/modules/project_files.go</code> | If lock writing fails after a manifest edit, restore the original manifest; join a restoration failure with the original error. |

### Local Overlay and Remaining X20 Boundaries

Local overlays are implemented without changing the published lock:

- <code>modules.EnsureEffectiveGraph</code> applies only main-module replacements,
  validates used local identities and resolves needed remote requirements.
  Tidy/download never write project files in local mode; get/update may edit
  explicit requirements but preserve the lock. Vendor snapshots the effective graph.
  Historical versionless remote edges require a historical pin rather than HEAD.
- Explicit <code>--frozen</code> validates native manifests and their existing
  locks, including every reachable transitive requirement and exact cached hash.
  It rejects main-module replacements and mutating commands. Missing snapshots
  may be downloaded only at locked commits; dependency versions are not resolved.
  Missing/stale locks fail without project writes. See [CLI.md](CLI.md).
- Unknown import paths are checked against known roots. There is no implemented
  automatic mapping from an arbitrary missing import to its owning Git module.
  <code>resolveV1Lock</code> returns <code>cannot resolve imports</code> rather
  than discovering a dependency identity from that path.

CLI exit classification for missing-import diagnostics is a separate concern
from X20 dependency discovery. Follow the actual handler's typed-error or
result-record branch, not the words in its message.

## Exit Code Behavior

| Exit code | Current command behavior |
|---|---|
| <code>0</code> | A command action returns nil. Migration preview and a declined wizard confirmation can succeed without writing project files. Generation can return success when there are no executable targets, but an attempted compilation with no input files returns an error. |
| <code>1</code> | Printed lint/breaking findings; an invalid validation report; <code>ls-files</code> collection diagnostics via <code>cli.Exit</code>; or an otherwise unhandled returned error, including module/cache, generation and migration failures. |
| <code>2</code> | The explicit lint/breaking <code>OpenImportFileError</code> branches and breaking's missing Git ref/repository branches. It is not the universal code for all invalid configuration or import errors. |

There is no central exit-code mapper. Inspect <code>Action</code> as well as its
internal worker functions: tests that call only an internal function cannot
establish the process exit code. In <code>internal/api/ls_files_v1.go</code>,
<code>Action</code> first writes the complete JSON/text result, then returns
<code>cli.Exit</code> with status 1 when the collected <code>errors</code> are
nonempty. Output failures return their own wrapped errors. This exit behavior
does not implement automatic import-to-module discovery.

## Configuration Validation and Reserved Fields

<code>internal/config/v1/validate_path.go</code> discovers the four native files:
<code>easyp.yaml</code>, <code>easyp.gen.yaml</code>, <code>protobuf.mod</code> and
<code>protobuf.lock</code>. <code>ValidateFile</code> in
<code>internal/config/v1/validate.go</code> selects the parser/semantic checks;
YAML schema diagnostics come from
<code>internal/config/v1/validate_yaml.go</code>. Filesystem failures are returned
as errors rather than embedded in a validation report.

- <code>yaml_validation</code> identifies schema diagnostics, with source
  coordinates where available; legacy-policy detection also uses this code.
- <code>v1_validation</code> identifies semantic parser/conversion failures and
  environment-expansion or schema-loading failures. The old
  <code>ValidateRaw</code>/<code>envsubst_error</code> description is not the v1
  path.
- <code>config.HasErrors</code> and the CLI report use
  <code>SeverityError</code> to decide invalidity. Warnings alone leave
  <code>valid</code> true.

Field presence in a Go struct is not a promise of implemented runtime behavior:

| Field | Verified boundary |
|---|---|
| <code>linters.extends</code> | Invalid reference syntax fails parsing; missing bases, cycles and boundary violations fail context-aware resolution. Direct engine conversion rejects an unresolved base. CLI validation never downloads dependencies. |
| <code>breaking.extends</code> | Resolves independently in the checked module context. Invalid or unresolved bases fail explicitly; baseline and ignore paths belong to the consumer. |
| <code>generate.packages</code> | Reserved; nonempty values fail <code>ParseGenerate</code> in <code>internal/config/v1/generate.go</code>. Select complete modules with <code>generate.modules</code>. |
| <code>breaking.categories</code> | <code>FILE</code> is implemented and adds declaration-move checks. Other categories are unsupported/reserved and rejected by shared policy semantics; do not claim arbitrary category selection is implemented. |

<code>ParsePolicy</code> and <code>ParseGenerate</code> also validate already-expanded
YAML through <code>validateExpandedV1YAML</code>; environment substitution is not
repeated during that schema pass. <code>yamlValidationError</code> in
<code>internal/config/v1/semantic_validation.go</code> joins schema errors with
file/location context for runtime callers. Policy parsing and both
<code>LintConfig</code>/<code>BreakingConfig</code> invoke the shared semantics check,
including unsupported extensions/categories, settings and baseline syntax.
A nonempty baseline must be <code>git:&lt;ref&gt;</code>; an empty value permits fallback.
Recursive validation still returns structured diagnostics through its own entry
point; field presence alone is insufficient to assert feature support.

## Wrapping Convention

Add the called operation and preserve the cause with <code>%w</code>. Use a
function/method prefix rather than a redundant layer/package label. A current
example from <code>internal/modules/repository.go</code> is:

~~~go
roots, err := LocalSources(moduleDir, module)
if err != nil {
    return nil, fmt.Errorf("LocalSources: %w", err)
}
locked, err := EnsureLockedSources(ctx, moduleDir, module, cache)
if err != nil {
    return nil, fmt.Errorf("EnsureLockedSources: %w", err)
}
return append(roots, locked...), nil
~~~

Expected absence is checked before wrapping, as in
<code>ReadModuleOrDefault</code> in <code>internal/modules/project_files.go</code>:

~~~go
_, module, err := ReadManifest(root)
if errors.Is(err, os.ErrNotExist) {
    return v1.Module{Roots: []string{"."}}, nil
}
if err != nil {
    return v1.Module{}, fmt.Errorf("ReadManifest: %w", err)
}
return module, nil
~~~

Retain the primary failure when cleanup or rollback also fails.
<code>errors.Join</code> is used by manifest restoration, vendor restoration and
migration transactions. Migration also joins close errors into named return
errors. Do not replace a failing operation's error with a later cleanup error,
or discard identity by using only <code>%v</code> when callers inspect it.

## Migration Error Flow

<code>internal/api/migrate.go</code> wraps <code>migration.Build</code> and
<code>Plan.Apply</code> failures. The flag-only path previews without writes;
applying requires <code>--write</code>, and dependency integrity work requires
<code>--resolve-lock</code>. A plan needing unapproved lock resolution can be
shown with warnings, but its <code>Apply</code> fails. The terminal wizard in
<code>internal/api/migrate_interactive.go</code> asks separately for dependency
access/cache writes and applying the displayed plan; flags do not bypass its
confirmations.

Invalid explicit directory/module values fail directly; the wizard can prompt
again for invalid values entered interactively. A declined confirmation prints
cancellation and returns nil. EOF, incomplete input, context cancellation, or
forced interactive mode without terminal input/output returns an error
(<code>internal/api/migrate_prompt.go</code>).

<code>internal/migration/migration.go</code> rejects unrepresentable legacy
semantics, conflicting native outputs and changed inputs rather than silently
converting them. <code>internal/migration/apply.go</code> stages outputs/backups,
rechecks observed state and restores prior files after ordinary installation
errors. A rollback failure is joined with the original error and reports the
retained recovery directory. Multiple filesystem replacements are not a
process-crash-atomic transaction; legacy backups and the retained legacy lock
are recovery inputs.

## User-Facing Output Formats

EasyP does not serialize a general error response object:

- Lint/breaking print each <code>core.IssueInfo</code> to stdout. Text uses
  <code>path:line:column:source message (rule)</code>; JSON encodes one issue per
  line in <code>internal/api/lint.go</code>.
- <code>validate-config</code> prints a report with <code>valid</code> and optional
  <code>errors</code>/<code>warnings</code> arrays. Issues contain
  <code>code</code>/<code>message</code>, with file/location/severity as provided
  by the record. Text output starts with <code>VALID: true|false</code> and adds
  error/warning tables. See <code>internal/api/validate.go</code>.
- <code>ls-files</code> JSON has <code>files</code>, <code>roots</code> and optional
  <code>errors</code>. Collection codes are <code>parse_error</code>,
  <code>invalid_import</code>, <code>import_not_found</code> and
  <code>open_error</code>. Text mode writes file rows to the application writer
  and diagnostics to its error writer (stdout/stderr by default). Collection
  errors preserve the result output and cause exit 1; setup/indexing failures
  return ordinary errors.
- Migration prints warnings and candidate contents through its application
  writer. These warnings are plan information, not validation-record arrays.
- Explicit <code>errExit</code> cases log to stderr through the configured text
  <code>slog</code> handler. The default CLI handler prints <code>cli.Exit</code>
  summaries to stderr; ordinary returned failures use <code>main</code>'s
  standard <code>log.Fatal</code> path.

## Retry and Error Logging

There is no shared retry/backoff policy or retryable-error classification.
The Git object cache does have a compatibility fallback: if a shallow fetch of
the exact commit fails, it fetches advertised history and verifies that same
commit again (<code>internal/adapters/gitmodules/object_cache.go</code>). This
is not permission to substitute HEAD or a different version. Its object-lock
wait loop honors context cancellation.

<code>cmd/easyp/main.go</code> installs a text <code>slog</code> logger on stderr,
with debug logging enabled by the debug flag. Preserve useful operation
context, including source/module/ref/path details when known. Log a failure at
the boundary that terminates or meaningfully handles it; lower layers should
normally wrap and return instead of logging the same failure repeatedly.
