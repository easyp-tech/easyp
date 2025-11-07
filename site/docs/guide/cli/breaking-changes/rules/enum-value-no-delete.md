# ENUM_VALUE_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no enum values are deleted from enums. Deleting an enum value breaks both wire format compatibility and generated code, as existing data may contain the deleted enum value and client code may reference the generated constants.

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
  ORDER_STATUS_CANCELLED = 5;
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  PRIORITY_URGENT = 4;
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
  // ORDER_STATUS_CANCELLED = 5; // [!code --] Deleted enum value - BREAKING!
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  // PRIORITY_URGENT = 4; // [!code --] Deleted enum value - BREAKING!
}
```

**Error:**
```
order.proto:9:3: Previously present enum value "5" on enum "OrderStatus" was deleted. (BREAKING_CHECK)
priority.proto:8:3: Previously present enum value "4" on enum "Priority" was deleted. (BREAKING_CHECK)
```

### More Examples

**Multiple value deletions:**

```proto
// Before
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_GUEST = 1;
  USER_ROLE_USER = 2;
  USER_ROLE_MODERATOR = 3;
  USER_ROLE_ADMIN = 4;
  USER_ROLE_SUPERADMIN = 5;
}

// After - BREAKING CHANGES!
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_USER = 2;
  USER_ROLE_ADMIN = 4;
  // USER_ROLE_GUEST = 1;      // BREAKING: deleted
  // USER_ROLE_MODERATOR = 3;  // BREAKING: deleted  
  // USER_ROLE_SUPERADMIN = 5; // BREAKING: deleted
}
```

### Good

**Instead of deleting, deprecate the enum values:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
  ORDER_STATUS_CANCELLED = 5 [deprecated = true]; // [!code focus]
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  PRIORITY_URGENT = 4 [deprecated = true]; // [!code focus]
}
```

**Or reserve the enum values after removal:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  reserved 5; // [!code focus]
  reserved "ORDER_STATUS_CANCELLED"; // [!code focus]
  
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum Priority {
  reserved 4; // [!code focus]
  reserved "PRIORITY_URGENT"; // [!code focus]
  
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
}
```

## Impact

- **Wire Format:** Existing data with deleted enum values cannot be properly deserialized
- **Generated Code:** Enum constants are removed, breaking client compilation
- **Client Applications:** Code referencing deleted enum values fails to compile
- **Switch Statements:** Case statements for deleted values cause compilation errors
- **Default Handling:** Unknown enum values may be handled differently

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
order := &myapi.Order{
    Status: myapi.OrderStatus_ORDER_STATUS_CANCELLED,  // ERROR after deletion
}

// Switch statement breaks
switch order.Status {
case myapi.OrderStatus_ORDER_STATUS_PENDING:
    // Handle pending
case myapi.OrderStatus_ORDER_STATUS_CANCELLED:  // ERROR: undefined constant
    // Handle cancellation
default:
    // Handle unknown
}

// Generated code compilation fails:
// undefined: myapi.OrderStatus_ORDER_STATUS_CANCELLED
```

**Existing data becomes problematic:**
```json
// Serialized data before deletion
{
  "id": "order123",
  "status": "ORDER_STATUS_CANCELLED"
}

// After enum value deletion
// Data deserializes but status becomes UNSPECIFIED (0)
// or causes parsing errors depending on implementation
```

**Server validation breaks:**
```go
// Before - server handles all enum values
func validateOrderStatus(status OrderStatus) error {
    switch status {
    case OrderStatus_ORDER_STATUS_PENDING,
         OrderStatus_ORDER_STATUS_CONFIRMED,
         OrderStatus_ORDER_STATUS_SHIPPED,
         OrderStatus_ORDER_STATUS_DELIVERED,
         OrderStatus_ORDER_STATUS_CANCELLED:  // ERROR: undefined
        return nil
    default:
        return errors.New("invalid order status")
    }
}

// Server logic that depends on cancelled status
func canModifyOrder(order *Order) bool {
    return order.Status != OrderStatus_ORDER_STATUS_CANCELLED  // ERROR: undefined
}
```

## Migration Strategy

1. **Deprecate enum values first:**
   ```proto
   ORDER_STATUS_CANCELLED = 5 [deprecated = true];
   ```

2. **Stop using deprecated values** in new code:
   ```go
   // Don't set deprecated values in new code
   order.Status = OrderStatus_ORDER_STATUS_DELIVERED  // Use alternative
   ```

3. **Update client code** to handle deprecated values gracefully:
   ```go
   switch order.Status {
   case OrderStatus_ORDER_STATUS_CANCELLED:
       // Handle legacy cancelled orders
       log.Warn("Processing deprecated cancelled status")
       // Convert to new status or handle specially
   }
   ```

4. **Reserve the values** after sufficient migration period:
   ```proto
   reserved 5, "ORDER_STATUS_CANCELLED";
   ```

5. **Never reuse enum numbers** - they must remain reserved permanently

## Common Scenarios

### Business Logic Changes
```proto
// Instead of deleting discontinued status
enum ProductStatus {
  PRODUCT_STATUS_UNSPECIFIED = 0;
  PRODUCT_STATUS_ACTIVE = 1;
  PRODUCT_STATUS_INACTIVE = 2;
  // PRODUCT_STATUS_DISCONTINUED = 3;  // Don't delete!
  
  // Better approach:
  PRODUCT_STATUS_DISCONTINUED = 3 [deprecated = true];
}
```

### Workflow Simplification
```proto
// Instead of removing intermediate states
enum TaskStatus {
  TASK_STATUS_UNSPECIFIED = 0;
  TASK_STATUS_TODO = 1;
  TASK_STATUS_IN_PROGRESS = 2;
  // TASK_STATUS_IN_REVIEW = 3;    // Don't delete intermediate states!
  // TASK_STATUS_APPROVED = 4;     // Don't delete!
  TASK_STATUS_DONE = 5;
  
  // Better approach:
  TASK_STATUS_IN_REVIEW = 3 [deprecated = true];
  TASK_STATUS_APPROVED = 4 [deprecated = true];
}
```

### Enum Consolidation
```proto
// When merging similar enum values
enum NotificationLevel {
  NOTIFICATION_LEVEL_UNSPECIFIED = 0;
  NOTIFICATION_LEVEL_INFO = 1;
  NOTIFICATION_LEVEL_WARNING = 2;
  NOTIFICATION_LEVEL_ERROR = 3;
  // Don't delete these even if consolidating:
  NOTIFICATION_LEVEL_DEBUG = 4 [deprecated = true];  // Keep deprecated
  NOTIFICATION_LEVEL_TRACE = 5 [deprecated = true];  // Keep deprecated
}
```

## Wire Format Considerations

### Enum Value Reuse Prevention
```proto
enum Status {
  // Never reuse numbers, even after deletion
  reserved 2;  // Was STATUS_DELETED, now reserved forever
  reserved "STATUS_DELETED";
  
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  // STATUS_DELETED = 2;  // This number can never be reused
  STATUS_INACTIVE = 3;
}
```

### JSON Compatibility
```proto
// JSON uses enum names, so both number and name should be reserved
enum Color {
  reserved 2;      // Reserve the number
  reserved "RED";  // Reserve the name for JSON compatibility
  
  COLOR_UNSPECIFIED = 0;
  COLOR_BLUE = 1;
  // COLOR_RED = 2;  // Was deleted, now reserved
  COLOR_GREEN = 3;
}
```

### Proto2 vs Proto3 Behavior
```proto
// Proto2: Unknown enum values are preserved in unknown fields
// Proto3: Unknown enum values are preserved as-is
// Both cases: Don't delete values that might exist in stored data

// Safe approach for both syntax versions:
enum MyEnum {
  MY_ENUM_UNSPECIFIED = 0;
  MY_ENUM_VALUE_OLD = 1 [deprecated = true];  // Don't delete
  MY_ENUM_VALUE_NEW = 2;
}
```
