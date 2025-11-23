# ENUM_VALUE_SAME_NAME

Categories:

- **WIRE+**

This rule checks that enum values maintain the same name for each number. Changing an enum value's name (while keeping the same number) breaks both JSON compatibility and generated code, as clients expect specific constant names.

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
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_APPROVED = 2; // [!code --] Changed from ORDER_STATUS_CONFIRMED
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_COMPLETED = 4; // [!code --] Changed from ORDER_STATUS_DELIVERED
}
```

**Error:**
```
order.proto:7:3: Enum value "2" on enum "OrderStatus" changed name from "ORDER_STATUS_CONFIRMED" to "ORDER_STATUS_APPROVED". (BREAKING_CHECK)
order.proto:9:3: Enum value "4" on enum "OrderStatus" changed name from "ORDER_STATUS_DELIVERED" to "ORDER_STATUS_COMPLETED". (BREAKING_CHECK)
```

### More Examples

**Multiple renames breaking clients:**

```proto
// Before
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
  USER_ROLE_MODERATOR = 2;
  USER_ROLE_USER = 3;
}

// After - ALL BREAKING CHANGES!
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMINISTRATOR = 1;  // BREAKING: ADMIN -> ADMINISTRATOR
  USER_ROLE_MOD = 2;           // BREAKING: MODERATOR -> MOD  
  USER_ROLE_MEMBER = 3;        // BREAKING: USER -> MEMBER
}
```

### Good

**Instead of renaming, add new values and deprecate old ones:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2 [deprecated = true]; // [!code focus]
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4 [deprecated = true]; // [!code focus]
  ORDER_STATUS_APPROVED = 5; // [!code focus] // New value instead of renaming
  ORDER_STATUS_COMPLETED = 6; // [!code focus] // New value instead of renaming
}
```

**Or create a new enum version:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus] // New package version

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_APPROVED = 2; // [!code focus] // Clean names in v2
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_COMPLETED = 4; // [!code focus]
}
```

## Impact

- **Generated Code:** Constant names change, breaking client compilation
- **JSON Compatibility:** JSON serialization uses enum names, breaking parsers
- **Client Applications:** Code referencing old constant names fails to compile
- **Documentation:** API docs become outdated with wrong enum names

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
status := myapi.OrderStatus_ORDER_STATUS_CONFIRMED
if order.Status == myapi.OrderStatus_ORDER_STATUS_DELIVERED {
    // handle delivery
}

// After rename - compilation fails
status := myapi.OrderStatus_ORDER_STATUS_CONFIRMED  // ERROR: undefined constant
if order.Status == myapi.OrderStatus_ORDER_STATUS_DELIVERED {  // ERROR: undefined constant
    // handle delivery
}
```

**JSON compatibility breaks:**
```json
// Before - JSON uses old names
{
  "status": "ORDER_STATUS_CONFIRMED"
}

// After - parser expects new names
{
  "status": "ORDER_STATUS_APPROVED"  // Old JSON fails to parse
}
```

## Migration Strategy

1. **Add new enum values** instead of renaming:
   ```proto
   ORDER_STATUS_CONFIRMED = 2 [deprecated = true];
   ORDER_STATUS_APPROVED = 5;  // New value
   ```

2. **Update server code** to handle both old and new values during transition

3. **Migrate clients** to use new enum values

4. **Remove deprecated values** in next major version:
   ```proto
   reserved 2, "ORDER_STATUS_CONFIRMED";
   ORDER_STATUS_APPROVED = 5;
   ```

## Allow Alias Exception

With `allow_alias = true`, you can temporarily have multiple names for the same value:

```proto
enum Status {
  option allow_alias = true;
  STATUS_UNSPECIFIED = 0;
  STATUS_OLD_NAME = 1 [deprecated = true];
  STATUS_NEW_NAME = 1;  // Same number, different name - allowed with alias
}
```

Note: EasyP currently detects this as breaking even with `allow_alias`. This may be refined in future versions.