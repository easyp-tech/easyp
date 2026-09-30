# Descriptor-based breaking profiles

Nonempty <code>breaking.categories</code> selects real descriptor-based profiles: <code>FILE</code>, <code>PACKAGE</code>, <code>WIRE_JSON</code>, or <code>WIRE</code>. Omitted or empty categories retain the legacy checker; this is not an alias for WIRE. Multiple profiles apply their union without duplicate diagnostics.

| Change | FILE | PACKAGE | WIRE_JSON | WIRE |
|---|---|---|---|---|
| Move a declaration within its package, preserving options | Reject | Allow | Allow | Allow |
| Rename a field, even with its old JSON name retained | Reject | Reject | Reject | Allow |
| Change int32 to int64 | Reject | Reject | Reject | Allow |
| Change int32 to uint32 | Reject | Reject | Allow | Allow |
| Change string to bytes | Reject | Reject | Reject | Allow |
| Change bytes to string or int32 to sint32 | Reject | Reject | Reject | Reject |
| Delete a non-required field | Reject | Reject | Reserve number and name | Reserve number |
| Add/remove a required field or change a real oneof | Reject | Reject | Reject | Reject |
| Add explicit optional presence | Reject | Reject | Allow | Allow |
| Convert repeated entry messages to a compatible map | Reject | Reject | Reject | Allow |

A map is encoded as repeated entry messages; repeated scalar values are not compatible merely because their type equals the map value type. Existing reserved names and number ranges cannot be released in any profile. Enum aliases are checked by name/number pairs for source/JSON and by retained numbers for WIRE. RPC/service removal, request/response identity, streaming, idempotency, and actual default-value changes remain checked. Source profiles additionally check supported file/field code-generation options, enum openness, declaration and extension ranges. WIRE_JSON also checks JSON format capability.

Current and baseline sources are compiled separately using their own imports and pinned revisions. Diagnostics retain source coordinates and repository-relative paths; consumer managed options are not applied. Profiles work with section-scoped extends, ignores, local overlays and frozen verification. No source, manifest, lock or plugin output is written by breaking.

The precise implemented rule catalog is <code>internal/config/breaking_profiles.go</code>. This feature does not claim full rule-for-rule Buf equivalence. It checks standard descriptor properties and supported effective Edition features, not arbitrary application-defined option semantics or external generated-code behavior. Unsupported syntax/editions are rejected by the pinned compiler instead of being declared compatible.

## Selection and verification

~~~yaml
version: v1
breaking:
  baseline: git:main
  categories: [WIRE_JSON]
  ignore_unstable: true
~~~

Runtime and generated schema use the same profile catalog; MCP describes that selection. Existing direct FilesCheck callers and empty-category CLI policy retain the historical AST checker. Explicit profiles compare linked descriptors with normalized full type identities, real versus synthetic oneofs, and effective inherited features. Ranges use int64 interval coverage, including inclusive enum maximum endpoints; no iteration over reserved numbers.

The external matrix lives in tests/e2e/v1/breaking_profiles_parent_test.go in easyp-test. It includes real Git baselines, declaration moves/import aliases, unchanged files, max ranges, local historical dependencies, shared extends, frozen mode and JSONL diagnostics.
