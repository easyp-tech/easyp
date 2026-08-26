# Protobuf Best Practices (EasyP-aligned)

Best practices for writing `.proto` files that are idiomatic, maintainable, and pass EasyP lint rules.

## File Organization

### File naming
- Use `lower_snake_case.proto` for all file names → enforced by `FILE_LOWER_SNAKE_CASE`
- One service per file (or group tightly related messages)
- Name the file after the primary entity: `user_service.proto`, `order.proto`

### Package structure
- Use dot-separated, lower_snake_case packages → enforced by `PACKAGE_LOWER_SNAKE_CASE`
- End with a version suffix → enforced by `PACKAGE_VERSION_SUFFIX`
- Match directory layout → enforced by `PACKAGE_DIRECTORY_MATCH`

```
proto/
  myapp/
    user/
      v1/
        user_service.proto    → package myapp.user.v1;
        user.proto            → package myapp.user.v1;
    order/
      v1/
        order_service.proto   → package myapp.order.v1;
```

### File options
- Set `go_package`, `java_package`, etc. consistently within each package
- Enforced by `PACKAGE_SAME_GO_PACKAGE`, `PACKAGE_SAME_JAVA_PACKAGE`, etc.
- Consider using EasyP's managed mode to auto-set these

---

## Naming Conventions

| Entity | Convention | Example | EasyP Rule |
|--------|-----------|---------|------------|
| Message | PascalCase | `UserProfile` | `MESSAGE_PASCAL_CASE` |
| Field | lower_snake_case | `first_name` | `FIELD_LOWER_SNAKE_CASE` |
| Service | PascalCase + suffix | `UserService` | `SERVICE_PASCAL_CASE`, `SERVICE_SUFFIX` |
| RPC | PascalCase | `GetUser` | `RPC_PASCAL_CASE` |
| Enum type | PascalCase | `UserStatus` | `ENUM_PASCAL_CASE` |
| Enum value | UPPER_SNAKE_CASE | `USER_STATUS_ACTIVE` | `ENUM_VALUE_UPPER_SNAKE_CASE` |
| Oneof | lower_snake_case | `auth_method` | `ONEOF_LOWER_SNAKE_CASE` |
| Package | lower_snake_case | `myapp.user.v1` | `PACKAGE_LOWER_SNAKE_CASE` |
| File | lower_snake_case | `user_service.proto` | `FILE_LOWER_SNAKE_CASE` |

---

## Enums

### Zero value
Every enum must have a zero value with a suffix like `_UNSPECIFIED`:

```protobuf
enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0;  // Default/unknown
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
}
```

- Zero value rule → `ENUM_ZERO_VALUE_SUFFIX` (configurable suffix)
- First value must be 0 → `ENUM_FIRST_VALUE_ZERO`

### Value prefixing
Prefix all values with the enum type in UPPER_SNAKE_CASE → `ENUM_VALUE_PREFIX`:

```protobuf
// Good
enum Color {
  COLOR_UNSPECIFIED = 0;
  COLOR_RED = 1;
}

// Bad — values lack prefix
enum Color {
  UNSPECIFIED = 0;
  RED = 1;
}
```

### No aliases
Avoid `allow_alias = true` → `ENUM_NO_ALLOW_ALIAS`. Use separate values instead.

---

## Services and RPCs

### Request/Response pattern
Each RPC should have unique, method-named request and response types:

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}
```

- Unique types → `RPC_REQUEST_RESPONSE_UNIQUE`
- Naming convention → `RPC_REQUEST_STANDARD_NAME`, `RPC_RESPONSE_STANDARD_NAME`

### Service suffix
Use a consistent suffix (default: `Service`) → `SERVICE_SUFFIX`

### Streaming considerations
If using `UNARY_RPC` rules, streaming is disallowed. Design APIs as unary where possible. Use streaming only when truly needed (large data transfers, real-time updates).

---

## Comments and Documentation

Document all public entities with leading comments:

```protobuf
// UserService handles user account operations.
service UserService {
  // GetUser retrieves a user by their unique identifier.
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

// UserStatus represents the current state of a user account.
enum UserStatus {
  // USER_STATUS_UNSPECIFIED is the default value.
  USER_STATUS_UNSPECIFIED = 0;
  // USER_STATUS_ACTIVE means the user can log in.
  USER_STATUS_ACTIVE = 1;
}

// GetUserRequest contains the parameters for retrieving a user.
message GetUserRequest {
  // user_id is the unique identifier of the user.
  string user_id = 1;
}
```

Enforced by `COMMENT_SERVICE`, `COMMENT_RPC`, `COMMENT_ENUM`, `COMMENT_ENUM_VALUE`, `COMMENT_MESSAGE`, `COMMENT_FIELD`, `COMMENT_ONEOF`.

---

## Imports

- Remove unused imports → `IMPORT_USED`
- Never use `import public` → `IMPORT_NO_PUBLIC`
- Never use `import weak` → `IMPORT_NO_WEAK`
- Avoid circular package imports → `PACKAGE_NO_IMPORT_CYCLE`

---

## Backward Compatibility

### Never do

- Remove or rename fields (reserve instead)
- Change field types or numbers
- Remove enum values (reserve instead)
- Remove or rename RPCs
- Change RPC request/response types

### Safe changes

- Add new fields (with new field numbers)
- Add new enum values
- Add new RPCs to existing services
- Add new services
- Add new messages
- Deprecate fields with `[deprecated = true]`

### Reservations

When removing a field or enum value, reserve both the number and the name:

```protobuf
message User {
  reserved 3, 7;
  reserved "old_field", "legacy_field";

  string name = 1;
  string email = 2;
  // field 3 was old_field (removed in v1.2)
}
```

---

## API Design Patterns

### Use wrapper messages
Always wrap request and response in dedicated messages (never use primitives directly):

```protobuf
// Good
rpc GetUser(GetUserRequest) returns (GetUserResponse);

// Bad — cannot evolve without breaking
rpc GetUser(google.protobuf.StringValue) returns (User);
```

### Pagination
For list endpoints, use cursor-based or offset pagination:

```protobuf
message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListUsersResponse {
  repeated User users = 1;
  string next_page_token = 2;
}
```

### Field masks
Use `google.protobuf.FieldMask` for partial updates:

```protobuf
message UpdateUserRequest {
  User user = 1;
  google.protobuf.FieldMask update_mask = 2;
}
```

### Standard method names
Follow consistent verb patterns: `Get`, `List`, `Create`, `Update`, `Delete`, `BatchGet`, `BatchCreate`.

---

## Versioning

- Use package version suffixes: `myapp.user.v1`, `myapp.user.v2`
- Never make breaking changes within a version — create `v2` instead
- Keep old versions running until all clients migrate
- Enforced by `PACKAGE_VERSION_SUFFIX`
