<!-- generated: 2026-09-30, template: cli.md -->

# EasyP CLI

## Overview

EasyP is a Protocol Buffers CLI for linting, compatibility checks, code generation, and Git-based dependency management. The entry point is [cmd/easyp/main.go](../cmd/easyp/main.go); handlers use <code>urfave/cli/v2</code> and implement [api.Handler](../internal/api/interface.go).

Build a local binary and run the usual checks from a configured project:

~~~bash
go build -o easyp ./cmd/easyp
easyp lint --path .
easyp generate
~~~

The native v1 files have separate roles: <code>protobuf.mod</code> declares module identity, roots, and dependencies; <code>protobuf.lock</code> records resolved Git dependencies; <code>easyp.yaml</code> defines producer lint/breaking policy; <code>easyp.gen.yaml</code> defines consumer generation. Legacy configuration requires explicit migration.

## Command Tree

These handlers and aliases are registered by the application:

~~~text
easyp
├── lint (l)                         Lint Protocol Buffer files
├── generate (g)                     Generate selected consumer projects
├── breaking                         Check API compatibility against Git
├── get                              Add or promote a direct Git module requirement
├── mod (m)                          Manage protobuf.mod / protobuf.lock
│   ├── download                     Install the exact locked dependencies
│   ├── update                       Update compatible versions / versionless HEAD
│   ├── tidy                         Resolve requirements and write the lock
│   └── vendor                       Copy locked sources into easyp_vendor
├── init (i)                         Initialize the three v1 project files
├── migrate                          Preview or apply v0-to-v1 migration
├── ls-files (ls)                    List module sources and reachable imports
├── validate-config (validate)       Validate one file or a directory tree
├── schema-gen                       Generate six v1 JSON Schema artifacts
└── completion
    ├── bash                         Print Bash completion script
    └── zsh                          Print Zsh completion script
~~~

The framework also provides help and <code>--version</code>; there is no registered standalone <code>version</code> command.

## Global Flags

| Flag | Type | Default | Environment variable | Meaning |
|---|---|---|---|---|
| <code>--cfg</code>, <code>--config</code> | string | <code>easyp.yaml</code> | <code>EASYP_CFG</code> | Explicit policy path for lint/breaking, or validation target. Discovery applies when unset. |
| <code>--debug</code>, <code>-d</code> | bool | <code>false</code> | <code>EASYP_DEBUG</code> | Enable debug logs. |
| <code>--format</code>, <code>-f</code> | enum | <code>text</code> in flag definition | <code>EASYP_FORMAT</code> | Select <code>text</code> or <code>json</code> for commands supporting structured output. |
| <code>--frozen</code> | bool | <code>false</code> | — | Require the selected native manifest and locked graph without dependency re-resolution or manifest/lock writes. |
| <code>--version</code> | bool | <code>false</code> | — | Framework version output from <code>internal/version</code>. |
| <code>--help</code> | bool | <code>false</code> | — | Framework help. |

The application registers config, debug, format, and frozen; their definitions are in [internal/flags/flags.go](../internal/flags/flags.go) and [frozen.go](../internal/flags/frozen.go). Config is optional. Put global flags before the command; <code>ls-files</code> also registers format locally, and <code>validate-config</code> registers both config and format locally. <code>flags.GetFormat</code> in [internal/flags/format.go](../internal/flags/format.go) selects the nearest explicitly set format, including an explicit global or environment value. Without an explicit selection, lint/breaking use text and ls-files/validate-config use JSON.

## Frozen dependency mode

<code>--frozen</code> is explicit and defaults to false; <code>CI=true</code> never enables it. Put it before the command, or directly on <code>generate</code>, <code>lint</code>, <code>breaking</code>, <code>ls-files</code>, <code>get</code>, <code>init</code>, or <code>migrate</code>. Module commands accept it on <code>mod</code> or its <code>download</code>, <code>tidy</code>, <code>vendor</code>, and <code>update</code> subcommands. A true value anywhere in the command lineage enables frozen mode; a child <code>--frozen=false</code> cannot disable a parent's true value.

~~~bash
easyp --frozen generate
easyp generate --frozen --project services/backend
easyp mod --frozen download
easyp mod download --frozen
easyp lint --frozen --path proto
easyp breaking --frozen --against main
easyp ls-files --frozen
~~~

Every selected native module whose dependency graph is used must have both <code>protobuf.mod</code> and <code>protobuf.lock</code>, including an empty lock for a module with no dependencies. Plain source directories without a manifest are rejected. All root <code>replace</code> directives, including unused entries, are rejected before accessing their targets or fetching dependencies. The shared validation checks direct and transitive requirements against the lock and rejects missing, invalid, incomplete, or stale graphs.

Frozen operations preserve manifest and lock bytes and do not resolve dependency versions, query HEAD/tags, or update requirements. They may fill the cache using exact locked commits and verify content hashes on both cold and warm caches. Frozen mode is not offline; cache fills and configured remote plugins can require network access.

<code>mod download</code> validates and installs the locked graph. <code>mod vendor</code> validates it before copying sources into <code>easyp_vendor</code>. <code>mod tidy</code>, <code>get</code>, <code>mod update</code>, and <code>init</code> refuse frozen execution; <code>migrate</code> also refuses it, including preview and interactive mode. Prepare or refresh the manifest/lock explicitly outside frozen mode.

Generation validates every selected graph before any plugin executes, including descriptor-only runs. An explicit workspace-relative module path uses that module's manifest and lock; a directory containing only <code>easyp.gen.yaml</code> does not need a separate pair. For example, <code>backend/easyp.gen.yaml</code> selecting <code>proto/user</code> validates <code>proto/user/protobuf.mod</code> and <code>proto/user/protobuf.lock</code>. A dependency selected by identity still requires its consumer's manifest and lock, and does not need its own lock inside the downloaded directory. Unselected projects are ignored; frozen does not broaden <code>--project</code> or <code>--all</code> selection or change their plugin authorization. Lint keeps effective policy scoping. Breaking validates current and historical graphs with their own manifests and locks; the baseline never borrows the current lock. <code>ls-files --include-imports=false</code> still enforces frozen graph validation even though it lists only local files.

## Commands Reference

### <code>easyp lint [flags]</code>

Lints sources with the effective v1 policy. Flags come from [internal/api/lint.go](../internal/api/lint.go).

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--path</code>, <code>-p</code> | string | <code>.</code> | Source path relative to the operation root. The definition is marked required but its default is already marked set, so an explicit argument is optional. |
| <code>--root</code>, <code>-r</code> | string | selected policy's directory | Override the operation root; a relative value is resolved from the working directory. |

~~~bash
easyp lint --path proto
easyp --format json lint --path api
~~~

Policy discovery starts at the working directory or the explicit root; see Configuration below. Descendant policies apply to their sources. Module roots and native module dependencies provide imports. Findings go to stdout: text uses <code>path:line:column:source message (rule)</code>; JSON emits one issue object per line, rather than one array.

### <code>easyp generate [flags]</code>

Generates according to the selected <code>easyp.gen.yaml</code> files. [internal/api/generate.go](../internal/api/generate.go) registers these generation flags, plus the shared <code>--frozen</code> flag:

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--project</code> | repeatable string | nearest ancestor generator | Select a consumer project directory containing <code>easyp.gen.yaml</code>; relative paths start at the working directory. |
| <code>--all</code> | bool | <code>false</code> | Recursively select generation projects below the working directory; exclusive with <code>--project</code>. |
| <code>--workspace</code> | string | discovered workspace boundary | Set the root for repository-relative module selectors and bounded ancestor lookup; it must contain the working directory. |
| <code>--descriptor_set_out</code> | string | empty | Write one combined binary <code>FileDescriptorSet</code>. |
| <code>--descriptor_set_out_dir</code> | string | empty | Write one binary descriptor set per selected project/module; exclusive with <code>--descriptor_set_out</code>. |
| <code>--include_imports</code> | bool | <code>false</code> | Include transitive dependencies in exported descriptor sets. |

~~~bash
easyp generate
easyp generate --project services/backend --project services/frontend
easyp generate --all --descriptor_set_out_dir descriptors --include_imports
easyp generate --project services/backend --descriptor_set_out descriptors.pb
~~~

Automatic selection finds the nearest ancestor generator within the workspace boundary. It does not recursively execute all generators; use <code>--all</code> for that. Recursive discovery skips hidden directories, <code>easyp_vendor</code>, <code>node_modules</code>, and nested Git repositories. The global config flag does not select a generator. This command has no <code>--path</code>, <code>--root</code>, or <code>EASYP_ROOT_GENERATE_PATH</code> input.

Generation targets come from <code>generate.modules</code>, or from the local module enclosing the selected generator when the list is empty. Native <code>protobuf.mod</code> supplies roots and dependencies; <code>generate.inputs</code> is removed. Plugin output directories are relative to the generator file. Relative descriptor destinations are relative to the working directory. Export still runs configured plugins; an empty plugin list permits descriptor-only export.

All selected descriptor graphs are prepared before plugin execution. A single combined export rejects conflicting definitions with the same import path; separate per-project/module exports retain independent graphs. See [descriptor exports](config/descriptor-export.md) for stable filenames and output guarantees, and the implementations in [internal/generation/discovery.go](../internal/generation/discovery.go), [selection.go](../internal/generation/selection.go), and [descriptor_set.go](../internal/generation/descriptor_set.go).

### <code>easyp breaking [flags]</code>

Checks compatibility against a Git revision. Flags and exit handling are in [internal/api/breaking_check.go](../internal/api/breaking_check.go).

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--path</code>, <code>-p</code> | string | <code>.</code> | Source path relative to the operation root; shares lint's already-set default. |
| <code>--against</code> | string | <code>master</code> fallback | Explicit Git ref overrides <code>breaking.baseline</code>; an explicitly empty value is an error. |
| <code>--root</code>, <code>-r</code> | string | selected policy's directory | Override the operation root. |

~~~bash
easyp breaking --path proto --against main
easyp --format json breaking --against HEAD~1
~~~

Without an explicit <code>--against</code>, the effective policy's <code>breaking.baseline</code> wins, with its required <code>git:</code> prefix removed; only an empty baseline falls back to <code>master</code>. Nonempty policy baselines must use <code>git:&lt;ref&gt;</code>, while the CLI flag accepts the raw ref. The v0 <code>breaking_check.against_git_ref</code> YAML key is obsolete. Findings use lint's text/JSON formats.

Current and baseline sources use their own module manifests and locks. Descendant policies and <code>breaking.ignore</code> are evaluated with policy ownership; ignore paths are relative to the policy file. Explicit FILE/PACKAGE/WIRE_JSON/WIRE profiles compare linked descriptors; omitted or empty categories default to FILE for v1 policies. See [breaking-profiles.md](config/breaking-profiles.md). <code>breaking.ignore_unstable</code> is supported. Section-scoped <code>extends</code> uses bounded local policies or verified consumer dependencies. See [internal/api/breaking_v1.go](../internal/api/breaking_v1.go), [breaking_scope.go](../internal/api/breaking_scope.go), and [breaking_ignore.go](../internal/api/breaking_ignore.go).

### <code>easyp get &lt;module&gt;[@version|@tag|@commit]</code>

Adds a new direct requirement or promotes an existing indirect requirement, then resolves its transitive dependencies and writes <code>protobuf.mod</code> and <code>protobuf.lock</code>. Exactly one positional module argument is required; the shared <code>--frozen</code> flag is registered in [internal/api/get_v1.go](../internal/api/get_v1.go).

An explicit query accepts a semantic version, named Git tag or full Git commit. Named tags are resolved into full commit pins in the manifest and lock; a branch name is not a tag. Module identities use Go major suffixes: v0/v1 are unsuffixed, v2+ uses matching <code>/vN</code>. For unsuffixed legacy repositories, get automatically marks verified pre-native v2+ releases with <code>+incompatible</code>; it cannot bypass a native manifest. See [major versions and Git mapping](config/dependency.md#major-versions-and-git-mapping). A new requirement without a suffix resolves Git HEAD. Repeating an existing requirement without a suffix preserves its specified version. Module lookup starts at the working directory and searches upward within the workspace.

~~~bash
easyp get github.com/googleapis/googleapis
easyp get github.com/googleapis/googleapis@common-protos-1_3_1
easyp get github.com/acme/contracts@v1.2.3
~~~

### <code>easyp mod &lt;subcommand&gt;</code>

The <code>mod</code> parent and all four subcommands accept <code>--frozen</code>. They use the nearest ancestor <code>protobuf.mod</code> found within the workspace, not a YAML dependency list or an unconditional current-directory root. Registration is in [internal/api/mod.go](../internal/api/mod.go); lookup is in [module_root.go](../internal/api/module_root.go).

| Subcommand | Behavior |
|---|---|
| <code>download</code> | Read the existing lock, validate remote requirements against it, and install/verify its exact pinned dependencies in the cache. A missing lock reports that tidy is required. |
| <code>update</code> | Advance tagged requirements within their current major version, refresh versionless requirements to HEAD, retain explicit commit pins, and rewrite manifest/lock resolution. Stable versions skip prerelease upgrades. |
| <code>tidy</code> | Resolve manifest requirements and transitive requirements, check import collisions/unresolved imports, update manifest requirements, and write exact commits/content hashes to the lock. Existing versionless pins are preserved. |
| <code>vendor</code> | Copy verified locked dependency proto files into <code>easyp_vendor</code>, replacing that directory through staging/backup handling. |

~~~bash
easyp mod tidy
easyp mod download
easyp mod update
easyp mod vendor
~~~

Outside frozen mode, with local replacements, <code>tidy</code> validates the ephemeral effective graph without writing manifest or lock. <code>get</code>/<code>update</code> may edit explicit requirements, but all operations preserve an existing lock byte-for-byte and do not create a local-graph lock. <code>vendor</code> copies the effective graph without changing the lock. Unknown imports are reported; automatic import-to-Git discovery is intentionally excluded. See [Dependency Management](config/dependency.md).

### <code>easyp init [flags]</code>

Initializes <code>protobuf.mod</code>, <code>easyp.yaml</code>, and <code>easyp.gen.yaml</code>. The policy starts with <code>linters.default: STANDARD</code> and <code>breaking.baseline: git:main</code>; generation starts with <code>plugins: []</code>. It does not create a lock or prompt for lint rule groups.

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--dir</code>, <code>-d</code> | string | <code>.</code> | Directory to initialize; also set by <code>EASYP_INIT_DIR</code>. |
| <code>--module</code> | string | inferred or prompted | Canonical module identity. |

Identity precedence is explicit <code>--module</code>, an existing native manifest, workspace Git identity, then an interactive prompt. Git inference applies when the target is the repository root and derives the identity from the literal <code>remote.origin.url</code>, before any transport URL rewrite. The result is validated as one module identity before preparing the manifest. Different existing file contents require individual overwrite confirmation; all choices are collected before writing. Identical files remain unchanged. Existing <code>buf.yaml</code> or <code>buf.yml</code> is rejected because Buf-to-Git dependency mapping is not implemented.

Prompts require both input and output to be terminals. Noninteractive initialization works with an explicit or inferred valid identity and no required overwrite confirmations; missing identity or a conflicting existing file returns an error instead of opening a controlling terminal.

~~~bash
easyp init --dir . --module github.com/acme/contracts
~~~

See [internal/api/init.go](../internal/api/init.go), [init_v1.go](../internal/api/init_v1.go), [init_prompt.go](../internal/api/init_prompt.go), and [gitmodules/identity.go](../internal/adapters/gitmodules/identity.go).

### <code>easyp migrate [flags]</code>

Prepares actual v1 candidates from legacy configuration without executing plugins. Positional arguments are rejected. The command and wizard are in [internal/api/migrate.go](../internal/api/migrate.go), [migrate_interactive.go](../internal/api/migrate_interactive.go), and [migrate_prompt.go](../internal/api/migrate_prompt.go).

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--dir</code> | string | <code>.</code> | Directory containing legacy <code>easyp.yaml</code>. |
| <code>--module</code> | string | unset | Canonical local module identity; required in flag-only mode. |
| <code>--interactive</code> | bool | automatic on a terminal without explicit module | Request the terminal wizard; <code>--interactive=false</code> disables automatic prompting. |
| <code>--resolve-lock</code> | bool | <code>false</code> | In flag-only mode, authorize dependency resolution/integrity verification and cache writes, including during preview. |
| <code>--write</code> | bool | <code>false</code> | In flag-only mode, apply validated candidates with historical backups. |

The wizard requires both input and output to be terminals. It collects missing directory/identity values, displays an initial local preview, separately asks for dependency/cache access if required, displays the verified preview, and asks before applying files. Enter means no for both confirmations; prefilled flags never bypass them. EOF is cancellation, not consent. Explicit <code>--module</code> preserves flag-only preview unless interactive mode is explicitly requested.

~~~bash
easyp migrate
easyp migrate --dir . --module github.com/acme/contracts
easyp migrate --dir . --module github.com/acme/contracts --resolve-lock --write
~~~

Flag-only invocation previews without file writes unless <code>--write</code> is supplied. If dependency integrity checks are required, application remains blocked until <code>--resolve-lock</code> authorizes them. Existing native outputs are not overwritten with conflicting candidates; legacy replaced files receive <code>.v0.bak</code> backups and <code>easyp.lock</code> is retained unchanged. Legacy <code>version</code> metadata such as <code>v1alpha</code> is accepted and omitted with a warning; <code>deps: null</code> means an empty dependency list. Keys that the v0 parser ignored are also omitted with source-path warnings instead of being assigned invented v1 semantics, while the byte-identical v0 backup preserves them. Known fields with invalid or ambiguous values still block migration. The plan rechecks observed files before applying and rolls back ordinary write failures, but multiple replacements are not process-crash atomic. See [internal/migration/migration.go](../internal/migration/migration.go), [apply.go](../internal/migration/apply.go), and [migration review context](config/review-migration-and-polish.md).

Local directory selection may become literal <code>generate.paths</code> selectors
without moving sources. The plan compares whole roots, then the original
module-directory-relative paths, then complete protobuf packages for cases
that cannot be represented by literal paths.
Each candidate must preserve the exact import-name-to-physical-source map.
Omitted or empty legacy roots keep <code>.</code>; native manifests with no
<code>roots</code> use that same default. Same-package files outside selected
paths do not become generation targets, including ignored Gradle build copies.
Changed import names, hidden/vendor/nested-module boundary changes and inferred
local filters mixed with whole-module Git inputs stay blocked. The plan rechecks
sources, paths and packages before applying. See [source selection](config/package-selection.md).

Historical <code>easyp.lock</code> entries may use a full SemVer tag followed by
Git's peeled-ref suffix <code>^{}</code>. Migration verifies the corresponding
tag and legacy content hash before writing the native lock; it never rewrites
the historical lock or switches to HEAD. Peeled branches, abbreviated refs,
repeated suffixes and pseudo-version-shaped peeled tags are rejected.
Released v0 lock hashes cover the installed <code>git archive '*.proto'</code>
contents after legacy root rewrites, while the new lock covers the materialized v1
snapshot. Migration verifies either the historical archive hash or the existing
whole-tree hash at the pinned revision. It rejects archive attributes that omit
or alter proto sources rather than silently changing their contracts.
Internal file, directory, import-root and metadata symlinks are supported.
Logical paths keep their protobuf import names. Git targets resolve only from
the pinned tree; installed snapshots contain regular resolved bytes and work
with <code>core.symlinks=false</code>. Selected inputs cannot escape their owner,
cross undeclared nested repositories/submodules or form cycles. Invalid unused
auxiliary links are omitted. Migration verifies historical archive contents
separately; it never substitutes the new snapshot hash for a v0 input hash.

### <code>easyp ls-files [flags]</code>

Lists sources from the working directory's native manifest, or a default local <code>.</code> root if that directory has no manifest. Unlike <code>mod</code>/<code>get</code>, this handler does not search upward for a module. The global config flag does not choose its root.

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--include-imports</code>, <code>-I</code> | bool | <code>true</code> | Include reachable imports from dependency roots and embedded well-known protos. |
| <code>--format</code>, <code>-f</code> | enum | <code>json</code> when unset | Command-local registration of the shared format flag. |

~~~bash
easyp ls-files
easyp ls-files --include-imports=false --format text
~~~

JSON is one indented object with <code>files</code>, <code>roots</code>, and optional <code>errors</code>. Each file contains <code>abs_path</code>, <code>import_path</code>, <code>source</code>, and <code>root</code>. Text prints tab-separated <code>import_path</code>, <code>source</code>, and <code>abs_path</code> to stdout, plus collected errors to stderr; it does not print the roots table. Sources are <code>workspace</code>, <code>dependency</code>, or <code>wellknown</code>. Embedded well-known paths use the synthetic <code>/wellknownimports</code> prefix, not a host directory.

After printing the report, collected parse/import errors return <code>cli.Exit(..., 1)</code> with an error count. Both JSON and text reports therefore produce status 1 when their error collection is nonempty. Setup, cache, indexing, and output-write failures also return errors. Source: [internal/api/ls_files_v1.go](../internal/api/ls_files_v1.go).

### <code>easyp validate-config [flags]</code>

Recursively validates <code>easyp.yaml</code>, <code>easyp.gen.yaml</code>, <code>protobuf.mod</code>, and <code>protobuf.lock</code> below the current directory. An explicit config argument or environment value chooses one file or a directory instead. Directory scanning reports an error if no recognized configuration files are found.

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--cfg</code>, <code>--config</code> | string | current directory when unset | Command-local registration of the global config selector. |
| <code>--format</code>, <code>-f</code> | enum | <code>json</code> when unset | Command-local registration of the shared format selector. |

~~~bash
easyp validate-config --config easyp.yaml
easyp validate-config --format text
easyp --format text validate-config --config ./services
~~~

The default JSON object contains <code>valid</code> and optional <code>errors</code>/<code>warnings</code>. Text begins with <code>VALID: true</code> or <code>VALID: false</code>, followed by issue tables when present. Invalid reports return <code>ErrHasValidateIssue</code> after printing, giving process status 1. Validation checks configuration structure and supported semantics; it does not install dependencies or execute generators. Read [internal/api/validate.go](../internal/api/validate.go), [internal/config/v1/validate.go](../internal/config/v1/validate.go), and [validate_path.go](../internal/config/v1/validate_path.go) for the current checks.

### <code>easyp schema-gen [flags]</code>

Generates the versioned schemas and latest aliases for policy YAML, generator YAML, and lock YAML.

| Flag | Type | Default | Meaning |
|---|---|---|---|
| <code>--out-dir</code> | string | <code>schemas</code> | Output directory for all six artifacts. |

~~~bash
easyp schema-gen
easyp schema-gen --out-dir /tmp/easyp-schemas
~~~

[internal/schemagen/schemagen.go](../internal/schemagen/schemagen.go) writes exactly:

- [schemas/easyp-v1.schema.json](../schemas/easyp-v1.schema.json) and [schemas/easyp.schema.json](../schemas/easyp.schema.json).
- [schemas/easyp.gen-v1.schema.json](../schemas/easyp.gen-v1.schema.json) and [schemas/easyp.gen.schema.json](../schemas/easyp.gen.schema.json).
- [schemas/protobuf.lock-v1.schema.json](../schemas/protobuf.lock-v1.schema.json) and [schemas/protobuf.lock.schema.json](../schemas/protobuf.lock.schema.json).

<code>protobuf.mod</code> uses its native parser, not a generated JSON Schema. Edit types/schema construction under [internal/config/v1](../internal/config/v1), then regenerate artifacts; do not hand-edit the generated files.

### <code>easyp completion &lt;shell&gt;</code>

Only Bash and Zsh subcommands are registered. They take no command-specific flags and print a script to stdout. Dynamic suggestions use the framework's <code>--generate-bash-completion</code> support.

~~~bash
easyp completion bash > "$(brew --prefix)/etc/bash_completion.d/easyp"
easyp completion zsh > "${fpath[1]}/_easyp"
~~~

These shell-specific examples assume an existing completion installation and writable destination. Source: [internal/api/completion.go](../internal/api/completion.go).

## Configuration and Command Context

Explicit config selection follows flag, then <code>EASYP_CFG</code>; relative values resolve from the working directory. The flag's nominal <code>easyp.yaml</code> default does not mean every command reads that file from the working directory:

| Command | Context resolution |
|---|---|
| lint / breaking | Explicit config wins. Otherwise search for an ancestor policy from cwd or <code>--root</code>, then accept one unambiguous outermost descendant policy. Multiple independent policies require selection. The policy directory becomes project root; <code>--root</code> overrides the operation root. |
| generate | Discover the nearest generator or explicitly select projects; <code>--all</code> opts into recursive discovery. <code>--workspace</code> sets its boundary. |
| mod / get | Find the nearest ancestor native module within the workspace. |
| ls-files | Read the working directory's manifest or use its default local root. |
| validate-config | Scan cwd unless an explicit config selector chooses a file/directory. |
| init / migrate | Use <code>--dir</code>. |

[internal/workspace/discovery.go](../internal/workspace/discovery.go) bounds searches at the nearest Git repository. Without Git metadata, it uses the outermost EasyP ancestor below the home directory. Policy root handling is in [internal/api/roots.go](../internal/api/roots.go).

| Environment variable | Meaning | Default |
|---|---|---|
| <code>EASYP_CFG</code> | Explicit config selection for the relevant commands | unset; command discovery applies |
| <code>EASYP_DEBUG</code> | Debug logging | <code>false</code> |
| <code>EASYP_FORMAT</code> | Shared explicit output format | command-specific default |
| <code>EASYP_INIT_DIR</code> | Init directory | <code>.</code> |
| <code>EASYPPATH</code> | Module cache root | <code>$HOME/.easyp</code> |

The cache path is resolved to an absolute path by [internal/api/runtime.go](../internal/api/runtime.go). Import consumers may install missing locked sources in that cache. Native dependency declarations belong in <code>protobuf.mod</code>, not the removed generation input configuration.

Current feature boundaries remain explicit: <code>generate.packages</code> selects exact protobuf names; FILE/PACKAGE/WIRE_JSON/WIRE are selectable breaking profiles. Unknown-import-to-Git-module discovery is intentionally excluded. Local replacement overlays apply in normal mode; explicit <code>--frozen</code> rejects them. Current parsing, execution, and schemas must all be consulted before documenting further support; see [policy.go](../internal/config/v1/policy.go), [generate.go](../internal/config/v1/generate.go), and [schema.go](../internal/config/v1/schema.go).

## Exit Codes and Error Flow

| Code | Current executable behavior |
|---:|---|
| 0 | Handler success, including successful reports and a wizard confirmation declined without an input error. |
| 1 | Lint/breaking findings use explicit <code>os.Exit(1)</code>; ls-files reports collected parse/import errors before returning <code>cli.Exit(..., 1)</code>. Any ordinary error returned from <code>app.Run</code> reaches standard-library <code>log.Fatal</code>, which prints it and exits 1; this includes invalid configuration, module/cache errors, usage errors, and failed initialization/migration. |
| 2 | Lint/breaking wrappers recognize <code>*core.OpenImportFileError</code>. Breaking also recognizes <code>*core.GitRefNotFoundError</code> and <code>core.ErrRepositoryDoesNotExist</code>. |

An import-related message alone does not imply status 2. [internal/core/proto_info_read.go](../internal/core/proto_info_read.go) returns the typed missing-import error after disk, dependency roots, and embedded well-known lookup fail. Invalid paths, permission/read failures, and imported-file parse errors remain ordinary returned errors. Generation and module handlers do not use lint/breaking's status-2 switch. Keep these contracts tied to the checked-in source when changing import handling.

## I/O Contracts

- **stdin/terminal:** init may ask for module identity and overwrite choices through its interactive prompter, only when input and output are terminals. The migration wizard reads line-based input from the application reader (normally stdin) and also requires terminal input/output; flag-only migration accepts no stdin document.
- **stdout:** lint/breaking findings; ls-files and validation reports; migration previews/prompts; shell completion; framework help/version. Schema generation primarily writes files.
- **stderr:** structured text logger output (including debug records when enabled), fatal returned errors, ls-files text-mode collected errors, and its CLI exit error count for either report format.
- **files:** init writes the three native project files; migration writes validated candidates/backups when authorized; generation writes configured plugin outputs and optional descriptors; get/tidy/update write manifest/lock results; download populates the cache; vendor writes <code>easyp_vendor</code>; schema-gen writes six schemas. Commands resolving imports may also populate the cache.

~~~bash
easyp --format json lint --path proto > lint-issues.json
easyp ls-files --format json | jq '.files[]'
easyp validate-config --format json > validation.json
~~~

## Troubleshooting

| Message / condition | Interpretation and next step |
|---|---|
| <code>no selected easyp.gen.yaml</code> | Select a consumer directory with <code>--project</code>, or intentionally opt into recursive generation with <code>--all</code>. |
| <code>multiple independent policies found</code> | Select the intended policy with global <code>--cfg</code>. |
| Missing <code>protobuf.lock</code> during download | Resolve the manifest explicitly with <code>mod tidy</code> before downloading its pinned graph. |
| Frozen rejects missing/stale locks or local replacements | Remove replacements and prepare the manifest/lock with <code>mod tidy</code> outside frozen mode; commit both files before retrying. |
| Legacy configuration rejection | Use <code>migrate</code> to inspect a concrete v1 conversion; generation does not silently convert v0 inputs. |
| <code>--interactive requires terminal input and output</code> | Run the wizard in a terminal or use explicit module/write/resolve flags for scripted migration. |
| Conflicting descriptor in a combined export | Select a compatible project set or use <code>--descriptor_set_out_dir</code> to retain separate graphs. |

## Development

~~~bash
go run ./cmd/easyp lint --path .
go build -o easyp ./cmd/easyp
~~~

To add a command, implement <code>api.Handler.Command() *cli.Command</code>, define only its actual command flags, and register the handler in <code>buildCommand</code> in [cmd/easyp/main.go](../cmd/easyp/main.go). Shared flags and output precedence live in [internal/flags](../internal/flags). Keep command help, discovery, error behavior, and this reference aligned with the implementation.

For additional reviewed behavior, see [context and generation](config/review-context-and-generation.md), [generation and baselines](config/review-generation-and-baselines.md), and [migration and reserved contracts](config/review-migration-and-polish.md). Those are retained review records; current source remains authoritative.

### Shared policy references

<code>linters.extends</code> and <code>breaking.extends</code> resolve independently after nearest-section cascading. See [policy-extends](config/policy-extends.md) for the file/module grammar, precedence, consumer-relative baseline/ignore behavior and non-networked validation.

## MCP configuration reference

<code>easyp_config_describe</code> covers all four v1 formats. <code>protobuf.mod</code> returns text grammar and examples; YAML files, including <code>protobuf.lock</code>, return their actual JSON Schemas. The reference never reads project files or resolves dependencies. See [MCP reference contract](config/mcp-module-reference.md).
