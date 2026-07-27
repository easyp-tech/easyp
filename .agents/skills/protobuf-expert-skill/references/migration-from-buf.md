# Migration from buf.build to EasyP

EasyP is a drop-in replacement for buf.build. This guide maps buf concepts to EasyP equivalents.

## Command Mapping

| buf command | EasyP equivalent | Notes |
|-------------|-----------------|-------|
| `buf lint` | `easyp lint` | Same rule names and groups |
| `buf breaking` | `easyp breaking` | Same detection categories |
| `buf generate` | `easyp generate` | Supports local + remote plugins |
| `buf mod update` | `easyp mod update` | Git-native dependencies |
| `buf mod init` | `easyp init` | Interactive setup wizard |
| `buf build` | `easyp generate --descriptor_set_out` | FileDescriptorSet output |
| `buf format` | — | Not yet supported |
| `buf push` | — | Not needed (uses Git repos directly) |
| `buf registry` | — | Not needed (uses Git repos directly) |

## Config File Mapping

| buf file | EasyP file | Notes |
|----------|-----------|-------|
| `buf.yaml` | `easyp.yaml` | Single config for all features |
| `buf.gen.yaml` | `easyp.yaml` (`generate` section) | Merged into main config |
| `buf.lock` | — | Uses Git refs for pinning |

## Config Structure Migration

### buf.yaml → easyp.yaml: Lint

```yaml
# buf.yaml                          # easyp.yaml
version: v1                         # (no version field)
lint:                                lint:
  use:                                 use:
    - DEFAULT                            - DEFAULT
  except:                              except:
    - SERVICE_SUFFIX                     - SERVICE_SUFFIX
  ignore:                              ignore:
    - proto/vendor                       - proto/vendor
  allow_comment_ignores: true          allow_comment_ignores: true
  enum_zero_value_suffix: _UNSPECIFIED enum_zero_value_suffix: _UNSPECIFIED
  service_suffix: Service              service_suffix: Service
```

### buf.yaml → easyp.yaml: Breaking

```yaml
# buf.yaml                          # easyp.yaml
breaking:                            breaking:
  use:                                 # (no `use` — all checks enabled)
    - FILE                             ignore:
  ignore:                                - proto/internal
    - proto/internal                   against_git_ref: main
```

### buf.yaml → easyp.yaml: Dependencies

```yaml
# buf.yaml                          # easyp.yaml
deps:                                deps:
  - buf.build/googleapis/googleapis    - github.com/googleapis/googleapis
  - buf.build/grpc/grpc               - github.com/grpc/grpc@v1.60.0
```

Key difference: buf uses BSR module references, EasyP uses **Git repository URLs** with optional `@version` or `@commit` suffixes.

### buf.gen.yaml → easyp.yaml: Code Generation

```yaml
# buf.gen.yaml                      # easyp.yaml
version: v1                         generate:
plugins:                               inputs:
  - plugin: go                           - directory: proto
    out: gen/go                        plugins:
    opt: paths=source_relative           - name: go
  - plugin: buf.build/grpc/go              out: gen/go
    out: gen/go                            opts:
    opt: paths=source_relative               paths: source_relative
                                         - remote: buf.build/grpc/go
                                             out: gen/go
                                             opts:
                                               paths: source_relative
```

Key differences:
- `opt` (string) → `opts` (key-value map)
- `plugin` → `name`, `remote`, `path`, or `command`
- Inputs are declared explicitly in `generate.inputs`

## Lint Rule Compatibility

EasyP supports the **same rule names and groups** as buf:

| Group | Rules | Equivalent |
|-------|-------|-----------|
| `MINIMAL` | 4 rules | Identical |
| `BASIC` | 24 rules | Identical |
| `DEFAULT` | 32 rules | Identical |
| `COMMENTS` | 7 rules | Identical |
| `UNARY_RPC` | 2 rules | Identical |

Inline suppression uses `// easyp:off` / `// easyp:on` (instead of `// buf:lint:ignore`).

## Dependency Management Differences

| Concept | buf | EasyP |
|---------|-----|-------|
| Registry | BSR (buf.build) | Git repositories |
| Pinning | `buf.lock` | `@version` or `@commit` in deps URL |
| Download | `buf mod update` | `easyp mod download` |
| Vendoring | — | `easyp mod vendor` |

## Migration Steps

1. **Rename config**: Create `easyp.yaml` based on your `buf.yaml` + `buf.gen.yaml`
2. **Convert deps**: Replace BSR module refs with Git repository URLs
3. **Convert plugins**: Map `plugin` + `opt` to `name`/`remote` + `opts`
4. **Download deps**: `easyp mod download`
5. **Test lint**: `easyp lint` — should produce same results
6. **Test generate**: `easyp generate` — verify output matches
7. **Test breaking**: `easyp breaking --against main`
8. **Update CI**: Replace buf actions with [easyp-tech/actions](https://github.com/easyp-tech/actions)
9. **Remove buf files**: Delete `buf.yaml`, `buf.gen.yaml`, `buf.lock`
