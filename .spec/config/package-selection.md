# Exact protobuf package selection

A nonempty generate.packages selects exact protobuf package names within the
modules selected by a generation project. Empty or omitted selects all source
files, preserving the existing behavior. Names follow protobuf identifier
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

All files declaring a selected package participate. Matching is across the
project's selected modules, not separately required in every module. Modules
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

The migration wizard can infer these selectors for legacy local directory
inputs. It preserves the existing roots (including the default `.`) and source
locations, requiring exact equality of current import names and physical files.
It rejects an empty inferred list, a partial package, boundary changes or a
local filter combined with whole-module Git generation inputs. Both the source
map and inferred selectors are rechecked before apply. The preview warns that
future files in a selected package participate regardless of their directory.

Regressions verify exact/prefix distinction, multiple files, multiple modules,
no-plugin failures, invalid unselected/required sources, custom options, per-
plugin imports, both descriptor modes and compiled Go output with a local
major-version dependency.
