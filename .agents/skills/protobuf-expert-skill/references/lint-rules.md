# EasyP Lint Rules Reference

EasyP provides 42 lint rules organized into 5 groups. Groups are cumulative — `BASIC` includes `MINIMAL`, `DEFAULT` includes `BASIC`.

## Rule Groups

| Group | Rules | Description |
|-------|-------|-------------|
| `MINIMAL` | 4 | Package consistency only |
| `BASIC` | 24 | MINIMAL + naming conventions and import hygiene |
| `DEFAULT` | 32 | BASIC + enum/RPC/service/file standards |
| `COMMENTS` | 7 | Documentation requirements for all entities |
| `UNARY_RPC` | 2 | Disallow streaming RPCs |

Use a group name directly in `lint.use` to enable all its rules.

---

## MINIMAL Group (4 rules)

| Rule | What it checks |
|------|----------------|
| `DIRECTORY_SAME_PACKAGE` | All `.proto` files in the same directory must declare the same package |
| `PACKAGE_DEFINED` | Every file must have a `package` declaration |
| `PACKAGE_DIRECTORY_MATCH` | Package name must match the directory path |
| `PACKAGE_SAME_DIRECTORY` | All files declaring the same package must be in the same directory |

---

## BASIC Group (adds 20 rules)

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

## DEFAULT Group (adds 8 rules)

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

## Uncategorized (1 rule)

| Rule | What it checks |
|------|----------------|
| `PACKAGE_NO_IMPORT_CYCLE` | Packages must not have circular import dependencies |

---

## Suppression

### Ignoring paths

```yaml
lint:
  ignore:
    - proto/vendor        # Ignore entire directories
  ignore_only:
    FIELD_LOWER_SNAKE_CASE:
      - proto/legacy      # Ignore specific rule for specific paths
```

### Inline comment ignores

Enable with `lint.allow_comment_ignores: true`, then use in `.proto` files:

```protobuf
// easyp:off
message legacy_message {  // This won't trigger MESSAGE_PASCAL_CASE
  string BadField = 1;    // This won't trigger FIELD_LOWER_SNAKE_CASE
}
// easyp:on
```

### Excluding rules

```yaml
lint:
  use:
    - DEFAULT
  except:
    - SERVICE_SUFFIX          # Disable individual rules from a group
    - PACKAGE_VERSION_SUFFIX
```
