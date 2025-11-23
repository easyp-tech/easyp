# MESSAGE_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no messages are deleted from proto files. Deleting a message breaks both wire format compatibility and generated code, as existing data may reference the deleted message and client code depends on the generated types.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3;
}

message Address { // [!code --]
  string street = 1; // [!code --]
  string city = 2; // [!code --]
  string country = 3; // [!code --]
} // [!code --]

message Order {
  string id = 1;
  User user = 2;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3;  // ERROR: Address message no longer exists!
}

// Address message was deleted - BREAKING CHANGE!

message Order {
  string id = 1;
  User user = 2;
}
```

**Error:**
```
user.proto:8:1: Previously present message "Address" was deleted from file. (BREAKING_CHECK)
```

### More Examples

**Nested message deletion:**

```proto
// Before
message User {
  string name = 1;
  Profile profile = 2;
  
  message Profile { // [!code --]
    string bio = 1; // [!code --]
    string avatar_url = 2; // [!code --]
  } // [!code --]
}

// After - BREAKING CHANGE!
message User {
  string name = 1;
  Profile profile = 2;  // ERROR: Profile message deleted
  
  // Profile nested message was deleted
}
```

### Good

**Instead of deleting, deprecate the message:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3 [deprecated = true]; // [!code focus]
}

message Address {
  option deprecated = true; // [!code focus]
  string street = 1;
  string city = 2;
  string country = 3;
}

message Order {
  string id = 1;
  User user = 2;
}
```

**Or replace with new message version:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3 [deprecated = true]; // [!code focus] // Old field
  AddressV2 address_v2 = 4; // [!code focus] // New field with better structure
}

message Address {
  option deprecated = true; // [!code focus]
  string street = 1;
  string city = 2;
  string country = 3;
}

message AddressV2 { // [!code focus]
  string street_address = 1; // [!code focus]
  string city = 2; // [!code focus]
  string state = 3; // [!code focus]
  string postal_code = 4; // [!code focus]
  string country = 5; // [!code focus]
} // [!code focus]

message Order {
  string id = 1;
  User user = 2;
}
```

## Impact

- **Wire Format:** Existing data with the deleted message cannot be deserialized
- **Generated Code:** Message classes/structs are removed, breaking client compilation
- **Field References:** Fields using the deleted message type become invalid
- **Nested Dependencies:** All messages referencing the deleted message break

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
user := &myapi.User{
    Name:  "John Doe",
    Email: "john@example.com",
    Address: &myapi.Address{  // ERROR after deletion
        Street:  "123 Main St",
        City:    "New York",
        Country: "USA",
    },
}

// Generated code compilation fails:
// undefined: myapi.Address
```

**Existing data becomes unreadable:**
```json
// Serialized data before deletion
{
  "name": "John Doe",
  "email": "john@example.com", 
  "address": {
    "street": "123 Main St",
    "city": "New York",
    "country": "USA"
  }
}

// After Address deletion - deserialization fails
// Parser doesn't know how to handle "address" field
```

## Migration Strategy

1. **Deprecate the message first:**
   ```proto
   message OldMessage {
     option deprecated = true;
     // ... fields
   }
   ```

2. **Stop using in new fields** and deprecate existing fields that use it:
   ```proto
   OldMessage old_data = 5 [deprecated = true];
   ```

3. **Create replacement** with new structure if needed:
   ```proto
   NewMessage new_data = 6;  // Better design
   ```

4. **Remove after migration period** in a new major version

## Common Scenarios

### Refactoring Message Structure
```proto
// Instead of deleting UserInfo and creating UserProfile
message UserInfo {
  option deprecated = true;  // Deprecate old
  string name = 1;
  string email = 2;
}

message UserProfile {  // Create new alongside
  string full_name = 1;
  string email_address = 2;
  string phone = 3;
  // Better field organization
}
```

### Removing Unused Messages
```proto
// Even if message seems unused, it might be in serialized data
message OldConfig {
  option deprecated = true;  // Always deprecate first
  // Don't delete immediately - data might exist
}
```

### Version Migration
```proto
// Clean approach - new package version
package myapi.v2;

message User {
  // Redesigned message structure
  // No deprecated Address references
}
```
