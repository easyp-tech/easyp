# EasyP Lint Rules Reference

EasyP v1 has 41 lint rules. Configure them in `easyp.yaml` under `linters` (see [config-reference.md](./config-reference.md#easypyaml--lint-and-breaking-policy)).

## Presets and groups

`linters.default` picks a **cumulative preset**:

| `linters.default` | Rules | Contains |
|---|---|---|
| `MINIMAL` | 4 | MINIMAL group |
| `BASIC` | 24 | MINIMAL + BASIC groups |
| `STANDARD` (default) | 32 | MINIMAL + BASIC + DEFAULT groups |
| `COMMENTS` | 39 | STANDARD + COMMENTS group |

`linters.enable` / `linters.disable` / `issues.exclude-rules[].linters` accept rule names **and group names**. Groups are disjoint sets:

| Group | Rules | Topic |
|---|---|---|
| `MINIMAL` | 4 | Package/directory consistency |
| `BASIC` | 20 | Naming conventions, import hygiene, package option consistency |
| `DEFAULT` | 8 | Enum/RPC/service/file API standards |
| `COMMENTS` | 7 | Documentation on every entity |
| `UNARY_RPC` | 2 | No streaming RPCs; never part of a preset |

```yaml
version: v1
linters:
  default: STANDARD
  enable: [COMMENTS, UNARY_RPC]
  disable: [COMMENT_FIELD]
```

> **v0 note:** v0 `lint.use` took the *disjoint* groups, so `use: [DEFAULT]` enabled only the 8 DEFAULT rules, not 32. `easyp migrate` keeps the exact v0 rule set, so the result usually reads `default: MINIMAL` plus long `enable`/`disable` lists. Switching to `default: STANDARD` turns on more rules: run lint before proposing it.

---

## MINIMAL Group (4 rules)

| Rule | What it checks |
|------|----------------|
| `DIRECTORY_SAME_PACKAGE` | All `.proto` files in the same directory must declare the same package |
| `PACKAGE_DEFINED` | Every file must have a `package` declaration |
| `PACKAGE_DIRECTORY_MATCH` | Package name must match the directory path |
| `PACKAGE_SAME_DIRECTORY` | All files declaring the same package must be in the same directory |

---

## BASIC Group (20 rules)

### Naming

| Rule | What it checks |
|------|----------------|
| `ENUM_PASCAL_CASE` | Enum type names must be PascalCase |
| `ENUM_VALUE_UPPER_SNAKE_CASE` | Enum values must be UPPER_SNAKE_CASE |
| `FIELD_LOWER_SNAKE_CASE` | Message field names must be lower_snake_case |
| `MESSAGE_PASCAL_CASE` | Message type names must be PascalCase |
| `ONEOF_LOWER_SNAKE_CASE` | Oneof field names must be lower_snake_case |
| `PACKAGE_LOWER_SNAKE_CASE` | Package names must be lower_snake_case |
| `RPC_PASCAL_CASE` | RPC method names must be PascalCase |
| `SERVICE_PASCAL_CASE` | Service names must be PascalCase |

### Enums

| Rule | What it checks |
|------|----------------|
| `ENUM_FIRST_VALUE_ZERO` | First enum value must have number 0 |
| `ENUM_NO_ALLOW_ALIAS` | Enums must not use `allow_alias = true` |

### Imports

| Rule | What it checks |
|------|----------------|
| `IMPORT_NO_PUBLIC` | No `import public` statements |
| `IMPORT_NO_WEAK` | No `import weak` statements |
| `IMPORT_USED` | All imports must be referenced |

### Cross-file Package Consistency

| Rule | What it checks |
|------|----------------|
| `PACKAGE_SAME_CSHARP_NAMESPACE` | Files in same package must have same `csharp_namespace` |
| `PACKAGE_SAME_GO_PACKAGE` | Files in same package must have same `go_package` |
| `PACKAGE_SAME_JAVA_MULTIPLE_FILES` | Files in same package must have same `java_multiple_files` |
| `PACKAGE_SAME_JAVA_PACKAGE` | Files in same package must have same `java_package` |
| `PACKAGE_SAME_PHP_NAMESPACE` | Files in same package must have same `php_namespace` |
| `PACKAGE_SAME_RUBY_PACKAGE` | Files in same package must have same `ruby_package` |
| `PACKAGE_SAME_SWIFT_PREFIX` | Files in same package must have same `swift_prefix` |

---

## DEFAULT Group (8 rules)

| Rule | What it checks | Config |
|------|----------------|--------|
| `ENUM_VALUE_PREFIX` | Enum values must be prefixed with the enum type name in UPPER_SNAKE_CASE | — |
| `ENUM_ZERO_VALUE_SUFFIX` | Zero-value enum entry must end with a specific suffix | `lint.enum_zero_value_suffix` (default: `_UNSPECIFIED`) |
| `FILE_LOWER_SNAKE_CASE` | `.proto` file names must be lower_snake_case | — |
| `RPC_REQUEST_RESPONSE_UNIQUE` | Each RPC request/response type must be used by only one RPC | — |
| `RPC_REQUEST_STANDARD_NAME` | RPC request type must be named `<MethodName>Request` | — |
| `RPC_RESPONSE_STANDARD_NAME` | RPC response type must be named `<MethodName>Response` | — |
| `PACKAGE_VERSION_SUFFIX` | Package must end with a version (e.g., `.v1`, `.v2beta1`) | — |
| `SERVICE_SUFFIX` | Service names must end with a configurable suffix | `lint.service_suffix` (default: `Service`) |

---

## COMMENTS Group (7 rules)

| Rule | What it checks |
|------|----------------|
| `COMMENT_ENUM` | Enum types must have a non-empty leading comment |
| `COMMENT_ENUM_VALUE` | Enum values must have a non-empty leading comment |
| `COMMENT_FIELD` | Message fields must have a non-empty leading comment |
| `COMMENT_MESSAGE` | Message types must have a non-empty leading comment |
| `COMMENT_ONEOF` | Oneof fields must have a non-empty leading comment |
| `COMMENT_RPC` | RPC methods must have a non-empty leading comment |
| `COMMENT_SERVICE` | Services must have a non-empty leading comment |

---

## UNARY_RPC Group (2 rules)

| Rule | What it checks |
|------|----------------|
| `RPC_NO_CLIENT_STREAMING` | RPCs must not use client streaming |
| `RPC_NO_SERVER_STREAMING` | RPCs must not use server streaming |

---

## Removed in v1

| Rule | Status |
|---|---|
| `PACKAGE_NO_IMPORT_CYCLE` | Removed. Listing it anywhere is an error: `invalid rule: PACKAGE_NO_IMPORT_CYCLE is not implemented`. |

---

## Suppression

### Excluding paths and rules

```yaml
issues:
  exclude-rules:
    - path: proto/vendor                  # all rules, whole subtree
    - path: proto/legacy/**
      linters: [FIELD_LOWER_SNAKE_CASE]   # one rule, one subtree
    - linters: [PACKAGE_VERSION_SUFFIX]   # no path: disable everywhere
```

Paths are relative to the `easyp.yaml` and support `*`, `?`, `[...]` and whole-segment `**`. A plain directory covers everything below it.

### Disabling rules

```yaml
linters:
  default: STANDARD
  disable:
    - SERVICE_SUFFIX
    - PACKAGE_VERSION_SUFFIX
```

### Inline comment directives

Allowed when `linters.allow_comment_ignores` is true (the v1 default):

```protobuf
// easyp:disable MESSAGE_PASCAL_CASE, FIELD_LOWER_SNAKE_CASE
message legacy_message {          // directive annotates this declaration only
  string BadField = 1;
}

// easyp:disable FIELD_LOWER_SNAKE_CASE
message A { string BadOne = 1; }
message B { string BadTwo = 1; }
// easyp:enable FIELD_LOWER_SNAKE_CASE   // paired: covers every line in between
message C { string good_one = 1; }
```

Directives are read from comments attached to declarations. If the closing `// easyp:enable` is the last thing in the file, with no declaration after it, it is silently ignored (as of `v1.0.0-nightly.20261008.1`). The `disable` then covers only the declaration right below it. Close a range before a following declaration, or annotate each declaration separately.

Also accepted for compatibility, scoped to the next declaration: `// buf:lint:ignore RULE` and `// nolint:RULE`.

Rule names must exist. Any other `easyp:` directive is a **hard error** that stops the whole lint run, for example `// easyp:off` / `// easyp:on`: `malformed lint directive "easyp:off"`. Replace those with `easyp:disable` / `easyp:enable` pairs that name the rules.
