# Protobuf package and path selection

A nonempty generate.packages selects exact protobuf package names within the
modules selected by a generation project. Empty or omitted means all packages
within generate.paths; without either filter, all module sources participate. Names follow protobuf identifier
syntax; they are not prefixes, filesystem paths, globs or expressions.

~~~yaml
version: v1
generate:
  modules: [proto/orders, proto/users]
  packages: [orders.v1, users.v1]
plugins:
  - name: go
    out: gen
    with_imports: true
~~~

All files declaring a selected package within the selected paths participate.
Matching is across the project's selected modules, not separately required in
every module. Modules
with no matching package do not execute plugins or produce empty descriptor
sets. Unknown names fail before any plugin, including partially matched lists.
Duplicate names are idempotent. Explicit package validation applies even to an
options-only project without plugins.

Discovery lexes top-level package declarations, leaving unrelated malformed
message bodies uncompiled. Selected sources and their entire reachable import
closure are compiled by the existing protobuf compiler, so required invalid
imports still fail. Existing module-root and import collision checks remain.
Compilation preserves source information, custom options and managed rules.

Plugin FileToGenerate contains only selected sources unless that plugin opts
into with_imports. Other packages remain in ProtoFile when they are required
for linking. One plugin's with_imports does not affect another plugin or the
independent descriptor export include_imports flag.

Both descriptor export modes use the same prepared selection. Without imports
only selected source descriptors are written; with imports the entire selected
closure is written. Compatible merging, collision diagnostics and preflight
before execution remain unchanged. Selection never modifies dependency state,
proto packages, import names or a published lock. Frozen checks still verify
the selected module graph; local replacements remain forbidden in frozen mode.

## Literal paths

<code>generate.paths</code> selects exact proto import-relative files or directory
subtrees across the project's selected modules. Empty means all module sources;
<code>.</code> explicitly selects all. Matching is component-bounded: <code>mcp</code>
does not select <code>mcp-copy</code>. Canonical relative paths are required;
absolute paths, traversal, backslashes and globs are rejected. These selectors
are separate from plugin <code>opts.paths</code>, which controls output layout.

~~~yaml
version: v1
generate:
  paths: [mcp]
plugins:
  - name: go
    out: .
    opts: {paths: source_relative}
~~~

Paths and packages intersect. Each selector must match a resulting source in
at least one selected module. Unknown or partially unmatched lists fail before
plugins or descriptor writes, including no-plugin projects and parents selected
with <code>--all</code>. Source discovery respects existing module/Buf filters;
required imports outside the selected paths still compile. No automatic Git
ignore policy is introduced. Roots, source locations and import names stay fixed.

The migration wizard prefers exact directory paths after whole-root equality.
It falls back to complete package selectors only when paths cannot preserve a
mixed-root selection. The current import-name/physical-file maps must match,
and sources plus fixed paths/packages are rechecked before apply. Empty inferred
selection, alias/boundary changes and local filters combined with whole-module
Git generation inputs remain rejected. Future same-package build copies outside
a selected directory do not widen that directory's targets. Package fallback
previews still warn that future files declaring selected packages participate.

Regressions verify exact/prefix distinction, multiple files, multiple modules,
no-plugin failures, invalid unselected/required sources, custom options, per-
plugin imports, both descriptor modes and compiled Go output with a local
major-version dependency.
