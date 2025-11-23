# ENUM_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no enums are deleted from proto files. Deleting an enum breaks both wire format compatibility and generated code, as existing data may contain values from the deleted enum and client code depends on the generated enum types.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum UserRole { // [!code --]
  USER_ROLE_UNSPECIFIED = 0; // [!code --]
  USER_ROLE_ADMIN = 1; // [!code --]
  USER_ROLE_USER = 2; // [!code --]
} // [!code --]

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

// UserRole enum was deleted - BREAKING CHANGE!

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3; // ERROR: UserRole no longer exists!
}
```

**Error:**
```
user.proto:8:1: Previously present enum "UserRole" was deleted from file. (BREAKING_CHECK)
```

### More Examples

**Nested enum deletion:**

```proto
// Before
message User {
  string name = 1;
  Status status = 2;
  
  enum Status { // [!code --]
    STATUS_UNSPECIFIED = 0; // [!code --]
    STATUS_ACTIVE = 1; // [!code --]
    STATUS_INACTIVE = 2; // [!code --]
  } // [!code --]
}

// After - BREAKING CHANGE!
message User {
  string name = 1;
  Status status = 2;  // ERROR: Status enum deleted
  
  // Status nested enum was deleted
}
```

### Good

**Instead of deleting, deprecate the enum:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum UserRole {
  option deprecated = true; // [!code focus]
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
  USER_ROLE_USER = 2;
}

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3 [deprecated = true]; // [!code focus]
}
```

**Or replace with new enum:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum UserRole {
  option deprecated = true; // [!code focus] // Old enum
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
  USER_ROLE_USER = 2;
}

enum UserPermission { // [!code focus] // New enum with better design
  USER_PERMISSION_UNSPECIFIED = 0; // [!code focus]
  USER_PERMISSION_READ = 1; // [!code focus]
  USER_PERMISSION_WRITE = 2; // [!code focus]
  USER_PERMISSION_ADMIN = 3; // [!code focus]
} // [!code focus]

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3 [deprecated = true]; // [!code focus] // Old field
  UserPermission created_by_permission = 4; // [!code focus] // New field
}
```

## Impact

- **Wire Format:** Existing data with the deleted enum cannot be deserialized
- **Generated Code:** Enum types/constants are removed, breaking client compilation
- **Field References:** Fields using the deleted enum type become invalid
- **Switch Statements:** Client code with enum switch cases breaks

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
order := &myapi.Order{
    Id:            "order123",
    Status:        myapi.OrderStatus_ORDER_STATUS_PENDING,
    CreatedByRole: myapi.UserRole_USER_ROLE_ADMIN,  // ERROR after deletion
}

// Check role
switch order.CreatedByRole {
case myapi.UserRole_USER_ROLE_ADMIN:  // ERROR: undefined type
    // Admin logic
case myapi.UserRole_USER_ROLE_USER:   // ERROR: undefined type
    // User logic
}

// Generated code compilation fails:
// undefined: myapi.UserRole
// undefined: myapi.UserRole_USER_ROLE_ADMIN
```

**Existing data becomes unreadable:**
```json
// Serialized data before deletion
{
  "id": "order123",
  "status": "ORDER_STATUS_PENDING",
  "created_by_role": "USER_ROLE_ADMIN"
}

// After UserRole deletion - deserialization fails
// Parser doesn't know how to handle "created_by_role" field
```

**Server implementation breaks:**
```go
// Before - server handles enum
func validateOrder(order *Order) error {
    switch order.CreatedByRole {
    case UserRole_USER_ROLE_ADMIN:  // ERROR: type not found
        return nil  // Admins can create any order
    case UserRole_USER_ROLE_USER:   // ERROR: type not found
        // Additional validation for regular users
        return validateUserOrder(order)
    default:
        return errors.New("invalid user role")
    }
}
```

## Migration Strategy

1. **Deprecate the enum first:**
   ```proto
   enum OldEnum {
     option deprecated = true;
     // ... values
   }
   ```

2. **Stop using in new fields** and deprecate existing fields:
   ```proto
   OldEnum old_field = 5 [deprecated = true];
   ```

3. **Create replacement** if needed:
   ```proto
   enum NewEnum {
     // Better designed enum values
   }
   NewEnum new_field = 6;
   ```

4. **Update server code** to handle both during transition:
   ```go
   func handleRole(oldRole OldEnum, newRole NewEnum) {
       // Handle both enum types during migration
       if newRole != NEW_ENUM_UNSPECIFIED {
           // Use new enum
           return handleNewRole(newRole)
       }
       // Fall back to old enum
       return handleOldRole(oldRole)
   }
   ```

5. **Remove after migration period** in a new major version

## Common Scenarios

### Enum Redesign
```proto
// Instead of deleting old enum
enum Priority {
  option deprecated = true;  // Deprecate old design
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_HIGH = 2;
}

enum TaskPriority {  // Create new enum with better naming
  TASK_PRIORITY_UNSPECIFIED = 0;
  TASK_PRIORITY_LOW = 1;
  TASK_PRIORITY_MEDIUM = 2;
  TASK_PRIORITY_HIGH = 3;
  TASK_PRIORITY_URGENT = 4;
}
```

### Consolidating Enums
```proto
// Instead of deleting multiple enums
enum Status {
  option deprecated = true;  // Keep old enum
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}

enum State {
  option deprecated = true;  // Keep old enum
  STATE_UNSPECIFIED = 0;
  STATE_ENABLED = 1;
}

enum EntityStatus {  // New consolidated enum
  ENTITY_STATUS_UNSPECIFIED = 0;
  ENTITY_STATUS_ACTIVE = 1;
  ENTITY_STATUS_INACTIVE = 2;
  ENTITY_STATUS_ENABLED = 3;
  ENTITY_STATUS_DISABLED = 4;
}
```

### Moving Enum to Different Package
```proto
// Instead of deleting from current package
enum UserRole {
  option deprecated = true;  // Mark as deprecated
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
}

// Import from new location
import "common/roles.proto";

message Order {
  UserRole old_role = 3 [deprecated = true];
  common.UserRole new_role = 4;  // Reference from new package
}
```

### Version Migration
```proto
// Clean approach - new package version
package myapi.v2;

// Redesigned enums without deprecated ones
enum UserPermission {
  USER_PERMISSION_UNSPECIFIED = 0;
  USER_PERMISSION_READ = 1;
  USER_PERMISSION_WRITE = 2;
  USER_PERMISSION_ADMIN = 3;
}

message Order {
  string id = 1;
  UserPermission created_by_permission = 2;  // Clean design in v2
}
```
