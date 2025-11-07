# RPC_SAME_RESPONSE_TYPE

Categories:

- **WIRE+**

This rule checks that RPC methods maintain the same response message type. Changing an RPC's response type breaks both wire format compatibility and generated code, as clients expect specific message structures when receiving responses from the method.

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

message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
}

message UpdateUserResponse {
  bool success = 1;
  string message = 2;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponseV2); // [!code --] Changed response type
  rpc UpdateUser(UpdateUserRequest) returns (UserUpdateResponse); // [!code --] Changed response type
}

message GetUserResponseV2 { // [!code --] Different message type
  string id = 1;
  UserProfile profile = 2;
  repeated string permissions = 3;
}

message UserUpdateResponse { // [!code --] Different message type  
  UpdateResult result = 1;
  UserProfile updated_profile = 2;
}
```

**Error:**
```
user_service.proto:6:3: RPC "GetUser" on service "UserService" changed response type from "GetUserResponse" to "GetUserResponseV2". (BREAKING_CHECK)
user_service.proto:7:3: RPC "UpdateUser" on service "UserService" changed response type from "UpdateUserResponse" to "UserUpdateResponse". (BREAKING_CHECK)
```

### More Examples

**Cross-package type changes:**

```proto
// Before
import "common/responses.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// After - BREAKING CHANGE!
import "v2/responses.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (v2.CreateUserResponse); // Different package
}
```

### Good

**Instead of changing response type, version the RPC:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc GetUserV2(GetUserRequest) returns (GetUserResponseV2); // [!code focus] // New RPC method
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc UpdateUserV2(UpdateUserRequest) returns (UserUpdateResponse); // [!code focus] // New RPC method
}

// Keep old messages for backward compatibility
message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
}

// Add new messages for enhanced functionality  
message GetUserResponseV2 { // [!code focus]
  string id = 1; // [!code focus]
  UserProfile profile = 2; // [!code focus]
  repeated string permissions = 3; // [!code focus]
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

message GetUserResponse { // [!code focus]
  string id = 1; // [!code focus] // Improved field names
  UserProfile profile = 2; // [!code focus]
  repeated string permissions = 3; // [!code focus]
} // [!code focus]
```

## Impact

- **Wire Format:** Server responses cannot be deserialized by existing clients
- **Generated Code:** Client stubs expect different response message types, breaking compilation
- **gRPC Calls:** Method signatures change in generated client code
- **Runtime Errors:** Type mismatches cause RPC response handling failures

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{
    UserId: "user123",
})
if err != nil {
    return err
}

// Access response fields
userID := resp.UserId  // This field exists
name := resp.Name      // This field exists

// After response type change - compilation fails  
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{
    UserId: "user123",
})
if err != nil {
    return err
}

// Access response fields - COMPILATION ERRORS
userID := resp.UserId  // ERROR: field not found
name := resp.Name      // ERROR: field not found

// Must use new fields:
userID := resp.Id                    // Different field name
profile := resp.Profile             // New structure
permissions := resp.Permissions     // New field
```

**Server implementation breaks:**
```go
// Before - server returns GetUserResponse
func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    return &GetUserResponse{
        UserId: "user123",
        Name:   "John Doe", 
        Email:  "john@example.com",
    }, nil
}

// After - server must return GetUserResponseV2
func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponseV2, error) {
    return &GetUserResponseV2{
        Id: "user123",                    // Different field name
        Profile: &UserProfile{           // New structure required
            Name:  "John Doe",
            Email: "john@example.com",
        },
        Permissions: []string{"read"},   // New field to populate
    }, nil
}
```

## Migration Strategy

1. **Add new RPC method** with improved response type:
   ```proto
   rpc GetUser(GetUserRequest) returns (GetUserResponse) {
     option deprecated = true;
   }
   rpc GetUserV2(GetUserRequest) returns (GetUserResponseV2);
   ```

2. **Implement both methods** on server during transition:
   ```go
   func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
       // Legacy implementation
       user := s.getUserInternal(req.UserId)
       return &GetUserResponse{
           UserId: user.ID,
           Name:   user.Name,
           Email:  user.Email,
       }, nil
   }
   
   func (s *server) GetUserV2(ctx context.Context, req *GetUserRequest) (*GetUserResponseV2, error) {
       // New implementation with enhanced response
       user := s.getUserInternal(req.UserId)
       return &GetUserResponseV2{
           Id: user.ID,
           Profile: &UserProfile{
               Name:  user.Name,
               Email: user.Email,
           },
           Permissions: user.Permissions,
       }, nil
   }
   ```

3. **Migrate clients** to use new RPC method:
   ```go
   // Update client calls to use V2
   resp, err := client.GetUserV2(ctx, &myapi.GetUserRequest{
       UserId: userID,
   })
   if err != nil {
       return err
   }
   
   // Handle new response structure
   profile := resp.Profile
   permissions := resp.Permissions
   ```

4. **Remove old RPC** in next major version:
   ```proto
   // Only keep new version
   rpc GetUserV2(GetUserRequest) returns (GetUserResponseV2);
   ```

## Common Scenarios

### Adding Fields to Response
```proto
// Instead of modifying existing response
message GetUserResponse {
  string user_id = 1;
  string name = 2;
  UserProfile profile = 3; // BREAKING: new required field
}

// Add new RPC method  
message GetUserResponseV2 {
  string user_id = 1;
  string name = 2;
  UserProfile profile = 3;  // Safe in new method
}
```

### Restructuring Response Data
```proto
// Instead of changing field structure
message UpdateUserResponse {
  UserDetails details = 1; // BREAKING: was individual fields
}

// Version the RPC
rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
  option deprecated = true;
}
rpc UpdateUserV2(UpdateUserRequest) returns (UpdateUserResponseV2);
```

### Error Handling Changes
```proto
// Instead of changing error structure
message CreateUserResponse {
  ErrorDetails error = 1; // BREAKING: different error format
}

// Keep old RPC, add new one
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
  option deprecated = true;  
}
rpc CreateUserV2(CreateUserRequest) returns (CreateUserResponseV2);
```

### Response Format Migration
```proto
// Instead of changing from simple to complex response
message GetUserResponse {
  // BREAKING: changed from simple fields to nested structure
  UserData data = 1;
}

// Provide both formats during transition
rpc GetUser(GetUserRequest) returns (GetUserResponse) {
  option deprecated = true;
}
rpc GetUserDetailed(GetUserRequest) returns (GetUserDetailedResponse);
```
