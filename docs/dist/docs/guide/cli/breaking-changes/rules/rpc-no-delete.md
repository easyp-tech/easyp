# RPC_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no RPC methods are deleted from services. Deleting an RPC method breaks both generated code and client applications that call the method.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse); // [!code --]
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  // DeleteUser RPC was deleted - BREAKING CHANGE!
}
```

**Error:**
```
services.proto:7:1: Previously present RPC "DeleteUser" on service "UserService" was deleted. (BREAKING_CHECK)
```

### Good

**Instead of deleting, deprecate the RPC:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option deprecated = true; // [!code focus]
  }
}
```

## Impact

- **Generated Code:** RPC method stubs are removed, breaking client compilation
- **Client Applications:** Existing calls to the RPC fail to compile
- **Runtime:** gRPC clients lose method definition

## Migration Strategy

1. **Deprecate first:**
   ```proto
   rpc OldMethod(OldRequest) returns (OldResponse) {
     option deprecated = true;
   }
   ```

2. **Return error from implementation** indicating method is no longer supported

3. **Remove in next major version** after clients have migrated