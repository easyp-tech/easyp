# SERVICE_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no services are deleted from proto files. Deleting a service breaks both generated code and client applications that depend on the service.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

service OrderService { // [!code --]
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse); // [!code --]
} // [!code --]
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// OrderService was deleted - BREAKING CHANGE!
```

**Error:**
```
services.proto:8:1: Previously present service "OrderService" was deleted from file. (BREAKING_CHECK)
```

### Good

**Instead of deleting, deprecate the service:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

service OrderService {
  option deprecated = true; // [!code focus]
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse) {
    option deprecated = true; // [!code focus]
  }
}
```

## Impact

- **Generated Code:** Service classes/interfaces are removed, breaking client compilation
- **Client Applications:** Existing client code fails to compile
- **Runtime:** gRPC clients lose service definition

## Migration Strategy

1. **Deprecate first:**
   ```proto
   service OldService {
     option deprecated = true;
     // ... methods
   }
   ```

2. **Notify clients** about deprecation timeline

3. **Remove after migration period** in a new major version