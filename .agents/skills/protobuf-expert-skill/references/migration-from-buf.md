# Migration from buf to EasyP v1

The EasyP v1 layout is close to buf's:

| buf | EasyP v1 |
|---|---|
| module identity in `buf.yaml` | `protobuf.mod` (`module`, `roots`, `require`) |
| `buf.lock` | `protobuf.lock` |
| `buf.yaml` lint/breaking | `easyp.yaml` (`linters`, `issues`, `breaking`) |
| `buf.gen.yaml` | `easyp.gen.yaml` |

The main differences:
- Dependencies are **Git repositories**, not BSR modules.
- `easyp init` refuses to run next to `buf.yaml`, so the conversion is manual.

Coming from EasyP v0 instead? Use [migration-v0-to-v1.md](./migration-v0-to-v1.md).

## Command mapping

| buf | EasyP v1 | Notes |
|---|---|---|
| `buf lint` | `easyp lint` | Same rule names |
| `buf breaking --against …` | `easyp breaking --against <git-ref>` | Profiles FILE / PACKAGE / WIRE_JSON / WIRE |
| `buf generate` | `easyp generate` | `name` / `path` / `command` / `remote` plugins |
| `buf dep update` / `buf mod update` | `easyp mod update` | Git tags/commits |
| `buf dep prune` / tidy | `easyp mod tidy` | |
| add a dep | `easyp get github.com/org/repo[@v1.2.3]` | |
| `buf config init` / `buf mod init` | `easyp init --module …` | Remove `buf.yaml` first |
| `buf build -o x.binpb` | `easyp generate --descriptor_set_out x.binpb [--include_imports]` | |
| `buf format`, `buf push`, `buf registry` | — | No formatter; no registry (Git is the registry) |

## Step by step

1. **Pick the module identity.** Usually the Git path (`github.com/acme/contracts`, plus a subdirectory for nested modules). Confirm it with the user.
2. **Move buf files aside:** `git mv buf.yaml buf.yaml.old`, the same for `buf.gen.yaml` and `buf.lock`. Then run `easyp init --module github.com/acme/contracts`.
3. **`protobuf.mod`.** Copy the module roots: buf v2 `modules[].path`, or the v1 workspace `directories`. Then convert dependencies (table below), list them all in `require ( … )`, and run `easyp mod tidy`.
4. **`easyp.yaml`.** Convert lint and breaking (below).
5. **`easyp.gen.yaml`.** Convert plugins (below).
6. **Verify.** Run `easyp validate-config`, `easyp lint`, then `easyp generate`, and diff the generated code against the buf output. Run `easyp breaking --against main`.
7. **Update CI.** See [ci-cd-integration.md](./ci-cd-integration.md). Then delete the `*.old` buf files.

## Dependencies: BSR → Git

| buf.yaml `deps` | `protobuf.mod` |
|---|---|
| `buf.build/googleapis/googleapis` | `require github.com/googleapis/googleapis` |
| `buf.build/grpc-ecosystem/grpc-gateway` | `require github.com/grpc-ecosystem/grpc-gateway` (v2 tags without a native manifest get `+incompatible` — let `easyp get` add it) |
| `buf.build/bufbuild/protovalidate` | `require github.com/bufbuild/protovalidate` |
| `buf.build/envoyproxy/protoc-gen-validate` | `require github.com/envoyproxy/protoc-gen-validate` |
| `buf.build/mercari/grpc-federation` | `require github.com/mercari/grpc-federation` |
| Other BSR modules | Find the source Git repo, then `easyp get <repo>`; use `--import-root` when the protos sit in a subdirectory |

A Git dependency may itself carry a `buf.yaml` with BSR deps. EasyP maps the known modules above to fixed Git compatibility snapshots and logs a warning. An unknown BSR module fails explicitly.

## Lint: buf.yaml → easyp.yaml

```yaml
# buf.yaml (v1/v2)                   # easyp.yaml
lint:                                 version: v1
  use: [STANDARD]   # or DEFAULT      linters:
  except:                               default: STANDARD
    - SERVICE_SUFFIX                    disable: [SERVICE_SUFFIX]
  ignore: [proto/vendor]                allow_comment_ignores: true
  ignore_only:                        issues:
    FIELD_LOWER_SNAKE_CASE:             exclude-rules:
      - proto/legacy                      - path: proto/vendor
  allow_comment_ignores: true             - path: proto/legacy
  enum_zero_value_suffix: _UNSPECIFIED      linters: [FIELD_LOWER_SNAKE_CASE]
  service_suffix: Service             linters-settings:
                                        ENUM_ZERO_VALUE_SUFFIX: {suffix: _UNSPECIFIED}
                                        SERVICE_SUFFIX: {suffix: Service}
```

| buf `use` | `linters.default` |
|---|---|
| `MINIMAL` | `MINIMAL` |
| `BASIC` | `BASIC` |
| `DEFAULT` (v1) / `STANDARD` (v2) | `STANDARD` |
| `+ COMMENTS` | `default: COMMENTS`, or `enable: [COMMENTS]` |
| `+ UNARY_RPC` | `enable: [UNARY_RPC]` |

`// buf:lint:ignore RULE` comments keep working. The native form is `// easyp:disable RULE` / `// easyp:enable RULE`.

## Breaking: buf.yaml → easyp.yaml

```yaml
# buf.yaml                            # easyp.yaml
breaking:                             breaking:
  use: [FILE]                           categories: [FILE]     # FILE | PACKAGE | WIRE_JSON | WIRE
  ignore: [proto/internal]              ignore: [proto/internal]
  ignore_unstable_packages: true        ignore_unstable: true
                                        baseline: git:main     # buf has no equivalent; --against overrides
```

## Generation: buf.gen.yaml → easyp.gen.yaml

```yaml
# buf.gen.yaml (v2)                    # easyp.gen.yaml
version: v2                            version: v1
managed:                               generate:
  enabled: true                          managed:
  override:                                enabled: true
    - file_option: go_package_prefix       override:
      value: github.com/acme/gen/go          - file_option: go_package_prefix
plugins:                                       value: github.com/acme/gen/go
  - local: protoc-gen-go               plugins:
    out: gen/go                          - name: go
    opt: paths=source_relative             out: gen/go
  - remote: buf.build/grpc/go:v1.6.2       opts:
    out: gen/go                              paths: source_relative
    opt:                                 - remote: plugins.beta.easyp.tech/grpc/go
      - paths=source_relative              version: v1.6.2
      - require_unimplemented_servers=false out: gen/go
inputs:                                    opts:
  - directory: proto                         paths: source_relative
                                             require_unimplemented_servers: false
```

- `local: protoc-gen-X` → `name: X`. `local: ./bin/x` → `path: ./bin/x`.
- `remote: buf.build/…:v` → an EasyP remote plus a separate `version`; or a local plugin.
- `opt` strings → the `opts` map: `a=b` becomes `a: b`; a repeated key becomes a list.
- `inputs` → targets come from the local module (`protobuf.mod` roots). Use `generate.paths` / `generate.packages` / `generate.modules` to narrow or add targets.

## Dependency management differences

| Concept | buf | EasyP v1 |
|---|---|---|
| Registry | BSR | Any Git host (system `git` credentials) |
| Manifest | `buf.yaml` `deps` | `protobuf.mod` `require` |
| Pinning | `buf.lock` | `protobuf.lock` (commit + `h1:` hash) |
| CI reproducibility | `buf.lock` | `easyp --frozen …` |
| Vendoring | — | `easyp mod vendor` → `easyp_vendor/` |
| Local overrides | workspaces | `replace module => ../path` |
