# Section-scoped producer policy inheritance

The nearest definition still selects each producer section as a whole. Explicit
linters.extends or breaking.extends then loads a base for that section only.
No generation plugin or base issues section is inherited or executed.

## References and boundaries

~~~yaml
version: v1
linters:
  extends: ./.policies/lint.yaml
  disable: [PACKAGE_VERSION_SUFFIX]
breaking:
  extends: example.com/policies#breaking/base.yaml
  ignore: []
~~~

Local ./ or ../ references resolve relative to the referring file, inside the
workspace. A directory means its easyp.yaml. The explicit module#path form and
the longest known module/path prefix form refer only to identities in the
consumer's declared, verified graph. Exact module identity selects easyp.yaml.
Module-relative paths cannot escape the actual manifest directory; nested /vN
layouts are resolved using that directory, never a proto import root. Relative
references inside a dependency remain inside that same module. Symlink escapes,
cycles and chains deeper than 16 base policies are errors with provenance.

Versions occur only in protobuf.mod/protobuf.lock, never in extends. Loading a
policy may install an exact existing pin, but never resolves a tag, selects HEAD,
adds a dependency or changes a lock. Main-module replacements are allowed only
in normal mode; remote edges introduced by a replacement still need existing
pins. Unreachable lock entries are not acquired. Frozen additionally rejects
replacements and unreachable pins and checks the complete reached graph.

## Precedence

Base policies resolve recursively. Explicit local default replaces the base
preset; inherited explicit enable/disable actions remain. Local enable can undo
a base disable, including an expanded group; local disable wins in the same
layer. Selections are deduplicated deterministically. Linter settings merge by
rule/key; an explicitly empty section clears all base settings, and an empty
rule map clears that rule's settings. YAML aliases/merge keys retain presence.
Explicit false, empty scalar baseline and empty category/ignore lists remain
local decisions rather than being replaced by inherited values. Environment
expansion is performed only for consumer-owned policy files and their local
extends. A policy reached through a declared dependency module treats ${...}
placeholders literally, including relative extends inside that dependency and
local-replacement modules physically stored inside the workspace. Values are
never expanded again after merging.

Breaking baseline and ignore paths are templates at the consuming policy, not
paths inside the dependency cache. Each checked module resolves its own graph,
so a common ancestor policy may select different pinned bases for two modules.
Deleted source/module scopes are compared using the chosen baseline snapshot.
A wholly removed module with a remote-only base may need explicit --against to
locate the historical graph; it never silently succeeds without comparison.

## Validation and ownership

ParsePolicy, ValidateFile and ValidatePath validate syntax without acquisition.
The CLI validate-config adds bounded resolution of references, checks cached
hashes and reports missing content with instructions to run mod download. It
never downloads policy dependencies. Structured errors include the consuming
file/field plus the chain of base sources. A root policy inherited by native
child modules is checked in those consumer contexts.

The reference MCP tool only describes syntax and examples; it never reads the
referenced files. No locks, policies or proto sources are written by resolution.

Tests live in internal/modules/policy_parent_test.go, internal/config/v1 policy
merge/reference tests and easyp-test/tests/e2e/v1/policy_extends*.go.
