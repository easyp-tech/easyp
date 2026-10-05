# Remaining review audit: validated configuration and reachable imports

## Correctness fixes

R-17 now rejects unknown managed file/field options, incorrect value types,
unsupported enum values and invalid selector structure. Generated schema uses
metadata derived from the actual managed-handler registry and protobuf option
descriptors. Parity tests compare the metadata with every registered handler.
No existing managed transformation or selector precedence is removed.

The policy and generation parsers validate their expanded YAML using the same
schema as validate-config. Environment expansion happens once per input path;
escaped placeholders remain literal. Policy conversion also validates settings
constructed directly in Go. A nonempty baseline must use git:<ref>. Unsupported
extends, categories and misspelled linter settings are not accepted by lint just
because it does not execute a breaking check.

B-11 (import validation and exit status): tidy follows the transitive import
closure reachable from local source files before writing manifest/lock. Missing
imports, illegal relative paths and invalid reachable source syntax fail with
the owning file. A google/protobuf prefix is not sufficient: bundled files must
exist. Unrelated invalid dependency files are not traversed. Cycles terminate
without duplicate parsing; this import scan is not a full protobuf type checker.

ls-files emits the complete JSON or text result, including accumulated errors,
and then returns exit status 1 when errors exist. include-imports=false remains
an intentional local-file listing rather than an import validation command.

R-14: an existing indirect requirement is inspected for direct imports instead
of skipped. Promotion removes only the indirect marker, keeping its version,
unrelated comments and unchanged lock. Repeating tidy remains idempotent.

C-16: legacy remote: host/plugin:version produces an actionable split-field
error even when the separate version is absent. Host ports remain valid.
C-14: mod update help explicitly describes refreshing within the current major
and rewriting the manifest and lock; exact commit behavior is unchanged.

R-21 / C-15: init cannot prompt through a hidden controlling terminal when its
input/output is redirected. Explicit or locally inferred identities still work.
An identity derived from origin reads the declared URL before insteadOf transport
rewriting, removes credentials, and only applies at the repository root. The
existing 0600 file mode is unchanged. Multiline identity injection is rejected.

C-17: lint executes once per effective policy/module batch, retains stateful
rule instances and full import access, applies named exclusions per file, and
sorts findings back to original lexical source order. Batching does not imply
that all imported files are themselves lint targets.

R-31: the existing AGENTS and top-level .spec documentation now describes native
v1 paths, commands, cache, error contracts and schema files. Old historical
review notes remain immutable; they record the state at their own commit.

## Remaining contracts and audit limits

X-20 is not implemented here. Local-replacement lock semantics, a frozen mode,
and unknown import-to-repository discovery need separate decisions. Buf registry
identities are not inferred as Git modules: B-11 registry metadata handling is
not closed by the import-integrity fix. Major-version identity policy remains
separate, as do shared extends, package selectors and extra breaking categories.

The documented 0600 init mode and non-offline tag revalidation remain deliberate
existing behavior, not silently modified to satisfy historical observations.
Existing service-oriented bundled skills and outdated development-task targets
are recorded separately from the repaired runtime issues. This audit does not
claim complete replay of every historical A/B/C/R script or security audit.

## Verification

~~~bash
go test -race -count=1 ./...
go vet ./...
task schema:check
task proto:check
~~~

The external suite has additional real CLI regressions under
 tests/e2e/v1/remaining_audit_test.go. EASYP_BIN and EASYP_SOURCE must identify
the same implementation. Red/green logs and exact commits live with the review
material, not in the source tree.
