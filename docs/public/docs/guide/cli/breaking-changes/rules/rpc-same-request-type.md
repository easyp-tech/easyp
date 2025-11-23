# RPC_SAME_REQUEST_TYPE

Categories:

- **WIRE+**

This rule checks that RPC methods maintain the same request message type. Changing an RPC's request type breaks both wire format compatibility and generated code, as clients expect specific message structures when calling the method.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message UpdateUserRequest {
  string user_id = 1;
  string name = 2;
  string email = 3;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequestV2) returns (GetUserResponse); // [!code --] Changed request type
  rpc UpdateUser(UserUpdateRequest) returns (UpdateUserResponse); // [!code --] Changed request type
}

message GetUserRequestV2 { // [!code --] Different message type
  string id = 1;
  bool include_profile = 2;
}

message UserUpdateRequest { // [!code --] Different message type  
  string id = 1;
  UserProfile profile = 2;
}
```

**Error:**
```
user_service.proto:6:3: RPC "GetUser" on service "UserService" changed request type from "GetUserRequest" to "GetUserRequestV2". (BREAKING_CHECK)
user_service.proto:7:3: RPC "UpdateUser" on service "UserService" changed request type from "UpdateUserRequest" to "UserUpdateRequest". (BREAKING_CHECK)
```

### More Examples

**Cross-package type changes:**

```proto
// Before
import "common/user.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// After - BREAKING CHANGE!
import "v2/user.proto";

service UserService {
  rpc CreateUser(v2.CreateUserRequest) returns (CreateUserResponse); // Different package
}
```

### Good

**Instead of changing request type, version the RPC:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc GetUserV2(GetUserRequestV2) returns (GetUserResponse); // [!code focus] // New RPC method
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc UpdateUserV2(UserUpdateRequest) returns (UpdateUserResponse); // [!code focus] // New RPC method
}

// Keep old messages for backward compatibility
message GetUserRequest {
  string user_id = 1;
}

// Add new messages for enhanced functionality  
message GetUserRequestV2 { // [!code focus]
  string id = 1; // [!code focus]
  bool include_profile = 2; // [!code focus]
} // [!code focus]
```

**Or create a new service version:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus] // New package version

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse); // [!code focus] // Clean interface in v2
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse); // [!code focus]
}

message GetUserRequest { // [!code focus]
  string id = 1; // [!code focus] // Improved field names
  bool include_profile = 2; // [!code focus]
} // [!code focus]
```

## Impact

- **Wire Format:** Existing client requests cannot be deserialized by new server
- **Generated Code:** Client stubs expect different request message types, breaking compilation
- **gRPC Calls:** Method signatures change in generated client code
- **Runtime Errors:** Type mismatches cause RPC call failures

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{
    UserId: "user123",
})

// After request type change - compilation fails  
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{  // ERROR: type not found
    UserId: "user123",  // ERROR: field not found
})

// Must use new type:
resp, err := client.GetUser(ctx, &myapi.GetUserRequestV2{
    Id: "user123",               // Different field name
    IncludeProfile: true,        // New required field
})
```

**Server implementation breaks:**
```go
// Before - server expects GetUserRequest
func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    userID := req.UserId  // This field exists
    // ... implementation
}

// After - server must handle GetUserRequestV2
func (s *server) GetUser(ctx context.Context, req *GetUserRequestV2) (*GetUserResponse, error) {
    userID := req.Id              // Different field name
    includeProfile := req.IncludeProfile  // New field to handle
    // ... implementation must change
}
```

## Migration Strategy

1. **Add new RPC method** with improved request type:
   ```proto
   rpc GetUser(GetUserRequest) returns (GetUserResponse) {
     option deprecated = true;
   }
   rpc GetUserV2(GetUserRequestV2) returns (GetUserResponse);
   ```

2. **Implement both methods** on server during transition:
   ```go
   func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
       // Legacy implementation
       return s.getUserInternal(req.UserId, false)
   }
   
   func (s *server) GetUserV2(ctx context.Context, req *GetUserRequestV2) (*GetUserResponse, error) {
       // New implementation  
       return s.getUserInternal(req.Id, req.IncludeProfile)
   }
   ```

3. **Migrate clients** to use new RPC method:
   ```go
   // Update client calls to use V2
   resp, err := client.GetUserV2(ctx, &myapi.GetUserRequestV2{
       Id: userID,
       IncludeProfile: true,
   })
   ```

4. **Remove old RPC** in next major version:
   ```proto
   // Only keep new version
   rpc GetUserV2(GetUserRequestV2) returns (GetUserResponse);
   ```

## Common Scenarios

### Adding Required Fields
```proto
// Instead of modifying existing request
message GetUserRequest {
  string user_id = 1;
  bool include_deleted = 2; // BREAKING: new required field
}

// Add new RPC method  
message GetUserRequestV2 {
  string user_id = 1;
  bool include_deleted = 2;  // Safe in new method
}
```

### Restructuring Request Data
```proto
// Instead of changing field structure
message UpdateUserRequest {
  UserProfile profile = 1; // BREAKING: was individual fields
}

// Version the RPC
rpc UpdateUser(UpdateUserRequest) returns (...) {
  option deprecated = true;
}
rpc UpdateUserV2(UpdateUserRequestV2) returns (...);
```

### Package Migrations
```proto
// Instead of changing import packages
import "v2/messages.proto";  // BREAKING: different package

// Keep old RPC, add new one
rpc OldMethod(v1.Request) returns (...) {
  option deprecated = true;  
}
rpc NewMethod(v2.Request) returns (...);
```
