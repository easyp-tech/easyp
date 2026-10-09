# EasyP v1 Configuration Reference

In v1 a project has four files, each with one job:

| File | Role | Format | Written by |
|---|---|---|---|
| `protobuf.mod` | Module identity, source roots, dependencies, local replacements | Text (Go-mod-like) | `easyp init`, `easyp get`, `easyp mod tidy/update`, by hand |
| `protobuf.lock` | Exact commit + content hash of every Git dependency | YAML | `easyp get`, `easyp mod tidy/update`, `easyp migrate` — **never by hand** |
| `easyp.yaml` | Producer policy: lint and breaking checks | YAML | `easyp init`, by hand |
| `easyp.gen.yaml` | Consumer generation: targets, plugins, managed mode | YAML | `easyp init`, by hand |

Commit all four. Validate with `easyp validate-config`, which checks the whole tree recursively. JSON Schemas for the YAML files come from `easyp schema-gen --out-dir schemas`.

If you see `lint:`, `deps:`, `generate.inputs` or `breaking.against_git_ref` in `easyp.yaml`, it is a **v0** config. Go to [migration-v0-to-v1.md](./migration-v0-to-v1.md).

Environment variables (`$VAR`, `${VAR}`) are expanded in `easyp.yaml` and `easyp.gen.yaml` only, never in `protobuf.mod`/`protobuf.lock`.

---

## protobuf.mod

```text
module github.com/acme/contracts/orders

roots (
	proto
)

require (
	github.com/acme/contracts/common
	github.com/googleapis/googleapis 03a91044136a014466d4293eb1fe91f2b02075d2
	github.com/grpc-ecosystem/grpc-gateway v2.31.0+incompatible // indirect
	github.com/acme/types v1.2.3
)

replace github.com/acme/types => ../types
```

| Directive | Rules |
|---|---|
| `module <identity>` | Exactly one. Includes the subdirectory for a module nested in a repository. v2+ modules end with `/vN`. |
| `roots <dir>` / `roots ( … )` | Import roots relative to this file; default `.`. Cannot escape the module directory. |
| `require <module> [version]` | One per dependency, as a line or inside a `( … )` block. The version is a **separate token**: omitted (Git HEAD, pinned in the lock on first resolve), semver tag `v1.2.3`, or a full 40/64-hex commit. `module@v1.2.3` is rejected. |
| `replace <module> => <path>` | Local development overlay; relative to this file. Never written to the lock. `--frozen` rejects any `replace`. |
| `// indirect` | Added by `get`/`tidy` for transitive requirements. |

Major versions follow Go rules:
- v0/v1 are unsuffixed; v2+ requires `/vN` in the identity.
- `+incompatible` marks a v2+ tag of a repository that has no native `protobuf.mod`. `easyp get` adds it automatically.

For a nested module, semver tags are directory-prefixed (`common/v1.2.3`).

Prefer editing through commands:

```bash
easyp get github.com/googleapis/googleapis                # add or promote a requirement, resolve, write mod + lock
easyp get github.com/acme/contracts@v1.2.3
easyp get github.com/googleapis/googleapis@common-protos-1_3_1   # named tag → full commit pin
easyp get --import-root api/svc/v1 gitlab.com/acme/svc   # hint where the dependency's import root is
easyp mod tidy                                            # after hand edits: resolve, add indirects, write lock
```

## protobuf.lock

```yaml
# protobuf.lock - GENERATED FILE, DO NOT EDIT MANUALLY
version: 1
modules:
  - source: github.com/googleapis/googleapis
    version: 03a91044136a014466d4293eb1fe91f2b02075d2
    commit: 03a91044136a014466d4293eb1fe91f2b02075d2
    hash: h1:…
```

| Field | Meaning |
|---|---|
| `version: 1` | Lock format version (integer). |
| `modules[].source` | Module identity. |
| `modules[].version` | Requested semver or the commit. |
| `modules[].commit` | Full commit hash. |
| `modules[].hash` | `h1:` directory hash of the installed snapshot, verified on every cache hit. |
| `modules[].roots` | Import roots inferred for a dependency without metadata (or given via `get --import-root`). |
| `modules[].bsr` | Record of a BSR dependency mapped to a fixed Git compatibility snapshot. |

A module without dependencies still needs a lock (`modules: []`) for `--frozen` runs.

## easyp.yaml — lint and breaking policy

```yaml
version: v1

linters:
  default: STANDARD            # MINIMAL | BASIC | STANDARD | COMMENTS (cumulative); default STANDARD
  enable:                      # extra rules or groups: COMMENTS, UNARY_RPC, any rule name
    - COMMENTS
    - UNARY_RPC
  disable:                     # rules or groups to switch off
    - SERVICE_SUFFIX
  allow_comment_ignores: true  # honour // easyp:off / // easyp:on (default: true)
  extends: ./.policies/lint.yaml   # optional base policy for this section

linters-settings:
  ENUM_ZERO_VALUE_SUFFIX:
    suffix: _UNSPECIFIED       # NONE = no suffix required
  SERVICE_SUFFIX:
    suffix: Service

issues:
  exclude-rules:
    - path: proto/legacy               # skip every rule under this path
    - path: proto/old/**
      linters: [FIELD_LOWER_SNAKE_CASE] # skip only these rules here
    - linters: [PACKAGE_VERSION_SUFFIX]  # no path = disable these rules everywhere

breaking:
  baseline: git:main           # default ref for `easyp breaking`; must start with git:
  categories: [FILE]           # FILE (default) | PACKAGE | WIRE_JSON | WIRE
  ignore_unstable: false       # skip packages whose last component is unstable (v1alpha1, v1beta2, v1test)
  ignore:
    - proto/internal
  extends: example.com/policies#breaking/base.yaml
```

Presets (`linters.default`):

| Preset | Rules |
|---|---|
| `MINIMAL` | 4 package-layout rules |
| `BASIC` | MINIMAL + 20 naming/import rules |
| `STANDARD` | BASIC + 8 API-shape rules: enum prefix and zero suffix, file names, RPC request/response names, package version suffix, service suffix |
| `COMMENTS` | STANDARD + 7 comment rules |

`UNARY_RPC` (no client or server streaming) is never in a preset; add it via `enable`. Rule details are in [lint-rules.md](./lint-rules.md).

`linters-settings` accepts only `ENUM_ZERO_VALUE_SUFFIX.suffix` and `SERVICE_SUFFIX.suffix`.

`issues.exclude-rules[].path` is relative to this file:
- it supports `*`, `?`, `[...]` and whole-segment `**`;
- a literal directory covers its descendants.

`extends` (in `linters` or `breaking`) loads a base for that section only:
- `./path.yaml` or a directory, meaning the `easyp.yaml` inside it;
- `module#path/in/module.yaml` for a policy shipped in a declared, locked dependency.

**Policy discovery.** `lint`/`breaking` use the nearest ancestor `easyp.yaml`. Descendant `easyp.yaml` files apply to their own subtrees. Pass `--cfg` when several independent policies exist.

## easyp.gen.yaml — generation

```yaml
version: v1

generate:
  modules:                       # what to generate; empty = the local module enclosing this file
    - github.com/acme/contracts  # module identity (local or a required dependency)
    - module: github.com/googleapis/googleapis
      paths: [google/api]        # per-module file/dir filter
      packages: [google.api]     # per-module package filter
  packages: [acme.orders.v1]     # global exact protobuf package filter
  paths: [proto/orders]          # global module-relative file/dir filter
  managed:
    enabled: true
    disable:
      - module: github.com/googleapis/googleapis
    override:
      - file_option: go_package_prefix
        value: github.com/acme/contracts/gen/go

plugins:
  - name: go                     # protoc-gen-go from PATH; builtins: python, java, kotlin, cpp, csharp, objc, php, ruby, grpc_*
    out: gen/go                  # relative to this file
    opts:
      paths: source_relative
  - name: go-grpc
    out: gen/go
    opts:
      paths: source_relative
      require_unimplemented_servers: false
  - remote: plugins.beta.easyp.tech/grpc-ecosystem/gateway
    version: v2.29.0             # required for remote, forbidden for name/path/command
    out: gen/go
    opts:
      paths: source_relative
  - path: ./bin/protoc-gen-custom
    out: gen/custom
  - command: ["go", "run", "./cmd/protoc-gen-mcp"]
    out: .
    with_imports: false          # also generate for imported files

options:
  go:
    package_prefix: github.com/acme/contracts/gen/go   # derive go_package from the protobuf package
```

Plugin rules:
- Exactly one of `name`, `path`, `command`, `remote`.
- `out` is required.
- `version` must be a pinned semver and is allowed **only** with `remote`. `remote: host/p:v1` is rejected: split it into `remote` + `version`.
- `opts` values may be scalars or lists; a list repeats the option.

Generation rules:
- Dependencies are available for imports, but are generated only when listed in `generate.modules`.
- `managed` has the same shape as in v0: `enabled`, `disable[]` selectors (`module`, `package`, `path`, `file_option`, `field_option`, `field`), and `override[]` (`file_option` or `field_option` + `value`, optionally scoped).
- `easyp generate` picks the nearest ancestor `easyp.gen.yaml`. `--project DIR` (repeatable) selects others; `--all` runs every one below the working directory.

## Layouts

Single module, generation in the same repo:

```text
repo/
├── protobuf.mod        # module github.com/acme/contracts, roots proto
├── protobuf.lock
├── easyp.yaml
├── easyp.gen.yaml
└── proto/acme/orders/v1/orders.proto
```

API repo consumed by a service repo:

```text
contracts/ (producer)            service/ (consumer)
├── protobuf.mod                 ├── protobuf.mod   # module github.com/acme/service
├── protobuf.lock                │                  # require github.com/acme/contracts v1.4.0
├── easyp.yaml                   ├── protobuf.lock
└── proto/...                    └── easyp.gen.yaml # generate.modules: [github.com/acme/contracts]
```

Monorepo: one `protobuf.mod` per module directory. Generators sit next to the consumers (`services/billing/easyp.gen.yaml`); run them with `easyp generate --all` or `--project services/billing`.
