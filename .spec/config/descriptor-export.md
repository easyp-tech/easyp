# v1 descriptor exports

## CLI contract

The two export flags belong to the generate command, not to easyp.gen.yaml.
They are mutually exclusive. Paths are resolved from the command working
directory, independently of the selected generation project.

~~~bash
# One compatible graph containing all selected modules.
easyp generate --project backend --descriptor_set_out build/backend.pb --include_imports

# One independent graph for each selected project/module pair.
easyp generate --descriptor_set_out_dir build/descriptors --include_imports
~~~

Without either flag, ordinary code generation is unchanged. With an export
flag, configured plugins still run, but only after all requested graphs and
output paths pass preflight. Pluginless projects can export descriptors too.

The include_imports flag controls the saved contents only. Compilation and
compatibility validation always use the full transitive graph, even when
imports are omitted from the exported files.

## Single-file export

The descriptor_set_out flag combines compatible graphs. Identical descriptors
are deduplicated by proto filename and content. Different definitions under
one import path, or conflicting symbols across different files, cause an
error. Version numbers alone do not determine compatibility.

Diagnostics identify both generation projects, selected modules, owning source
modules, and the first differing descriptor field. Resolved dependency versions
and commits are included when available from the lock actually used by the
consumer. Local replacements are labelled as replacements, not as resolved
remote versions.

Consumer managed options are preserved. The exporter does not drop go_package
or other options just to make conflicting graphs appear compatible.

## Directory export and naming

The descriptor_set_out_dir flag validates each project/module graph separately;
it does not combine graphs with each other. It therefore supports consumers
with different managed options and modules with conflicting dependency versions.

Within the output directory, generation-project paths mirror their paths from
the working directory. A project at the root writes directly into the output
directory. Each filename consists of the module identity's final path component,
sanitized to a lowercase filesystem-safe name, followed by the first 12 hex
digits of the SHA-256 of the full module identity, and the suffix .pb.

~~~text
descriptors/
  backend/
    user-<identity-hash>.pb
    order-<identity-hash>.pb
  frontend/
    user-<identity-hash>.pb
~~~

The identity suffix is always present: adding another module with the same
basename cannot rename or overwrite existing targets. Filenames do not depend
on selection order or the module version. Without a manifest, the identity is
formed from the module directory relative to the working directory. Explicit
projects outside that directory are placed in an _external namespace with a
safe, identity-suffixed project name; parent traversal is never used in outputs.

Selecting the same resolved directory more than once in one project exports it
and runs its plugins only once. Different source directories claiming the same
output identity are rejected before plugins execute. Existing ancestor symlinks
are resolved for collision checks; dangling output ancestors and symlink output
files are rejected before plugins execute. Existing unrelated files
are not deleted. Use a dedicated empty directory when a directory's entire
contents must represent exactly one invocation.

This mode changes descriptor destinations only, not configured plugin output
paths. Independent generated SDKs still need suitable plugin output layouts.

## Preparation and effects

The generation layer prepares core.GenerationPlan values in memory: resolve
sources, compile, and apply managed options. The full graphs are then validated
and export bytes are prepared before the first plugin is invoked. Plugin
requests and exported descriptors are independent copies of that prepared graph;
there is no second compilation after validation.

A descriptor conflict, invalid later module, invalid export path, or conflicting
export flags must not invoke any plugin or change existing generated/descriptor
outputs. Resolving sources can still populate the configured dependency cache.

After successful preflight, plugins run. Descriptor files are written only after
all selected plugin executions succeed, using an atomic replacement per file
and preserving an existing regular file's permissions. Missing output parent
directories are created. This is not a transaction across all plugin side
effects or all exported files: plugin errors and filesystem failures after
preflight do not promise a global rollback.

An explicit Go-options-only ancestor with child generation projects and no
protobuf.mod is an inheritance container, not an implicit module. Incidental
legacy proto files do not turn that container into an export target. A declared
root module still participates, as do standalone pluginless projects.

## Regression coverage

Unit and CLI tests live in internal/core, internal/generation and internal/api.
The independent easyp-test repository contains file-first v1 fixtures and their
checks in tests/e2e/v1/descriptor_exports_test.go. All v1 checks require an
explicit absolute EASYP_BIN and never silently use a released binary.
