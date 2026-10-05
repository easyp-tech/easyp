# v1 review correctness fixes

## Lint selection (X-9, R-16, R-17)

The runtime, v1 schema, and MCP share the implemented rule/group catalog.
Repeated rule names or group names are idempotent: the first occurrence is
retained, and disabling takes precedence over enabling. The same semantics
apply to presets and pathless issue exclusions.

Unknown names in enable, disable, and exclusion lists are errors, even when
the same name also appears in the opposite list. Configuration validation
reports all invalid entries with their YAML coordinates and rejected names.
The reserved PACKAGE_NO_IMPORT_CYCLE rule is not advertised or executable;
selecting it reports that it is not implemented rather than panicking.

## Local replacements (X-13)

An absolute replacement target is used directly. A relative target is resolved
from the directory of its owning protobuf.mod, not from the generation consumer
or command working directory. Source resolution and explicit dependency
selection use the same path helper.

## Breaking baseline precedence (X-17)

Precedence is: explicitly supplied --against, then breaking.baseline, then the
existing CLI default master. Explicit --against master still overrides a
configured baseline. An explicitly empty or whitespace-only flag is an error;
a missing explicit Git ref does not silently fall back to the configured ref.
These changes do not expand supported Git ref kinds.

## Legacy dependency metadata (X-2)

When adapting a published pre-v1 dependency, easyp.yaml deps and generation Git
inputs both contribute requirements. Exact source/version duplicates are
removed in stable order; different minimum versions remain visible to MVS.
Malformed deps entries fail explicitly. Roots retain their previous behavior.
A matching native protobuf.mod takes precedence over the legacy configuration.
Only dependency metadata is read; dependency plugins are never executed.

## Verification

Unit tests cover runtime selection, validation locations and schema parity,
absolute/relative paths, flag precedence, native manifest precedence, invalid
legacy metadata, and combined deps/Git-input requirements. End-to-end fixtures
are in easyp-test/specs/scenarios/v1 and run through TestFiveReviewFixes using
an explicit EASYP_BIN. The pre-fix binary fails all five defect cases while
the relative-replacement control passes.
