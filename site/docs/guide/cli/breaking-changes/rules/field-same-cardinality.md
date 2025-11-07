# FIELD_SAME_CARDINALITY

Categories:

- **WIRE+**

This rule checks that message fields maintain the same cardinality (optionality). Changing a field's cardinality breaks both wire format compatibility and generated code, as the presence semantics and client code expectations differ between optional and required fields.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  int32 age = 3;
}

message CreateUserRequest {
  string name = 1;
  optional string email = 2;
  string phone = 3;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  optional string email = 2; // [!code --] Changed from required to optional
  int32 age = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2; // [!code --] Changed from optional to required
  string phone = 3;
}
```

**Error:**
```
user.proto:6:3: Field "2" with name "email" on message "User" became optional. (BREAKING_CHECK)
request.proto:7:3: Field "2" with name "email" on message "CreateUserRequest" became not optional. (BREAKING_CHECK)
```

### More Examples

**Proto2 to Proto3 migration issues:**

```proto
// Before (proto2)
syntax = "proto2";

message Order {
  required string id = 1;
  optional string notes = 2;
  repeated string tags = 3;
}

// After (proto3) - BREAKING CHANGES!
syntax = "proto3";

message Order {
  string id = 1;           // BREAKING: required -> implicit optional
  string notes = 2;        // BREAKING: explicit optional -> implicit optional  
  repeated string tags = 3; // OK: repeated unchanged
}
```

### Good

**Instead of changing cardinality, add new field:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2 [deprecated = true]; // [!code focus] // Keep original cardinality
  int32 age = 3;
  optional string email_optional = 4; // [!code focus] // New field with desired cardinality
}

message CreateUserRequest {
  string name = 1;
  optional string email = 2 [deprecated = true]; // [!code focus] // Keep original cardinality
  string phone = 3;
  string email_required = 4; // [!code focus] // New field with desired cardinality
}
```

**Or create a new message version:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus] // New package version

message User {
  string name = 1;
  optional string email = 2; // [!code focus] // Clean cardinality in v2
  int32 age = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2; // [!code focus] // Required in v2
  string phone = 3;
}
```

## Impact

- **Wire Format:** Field presence semantics change, breaking deserialization expectations
- **Generated Code:** Field accessor methods change (hasField(), clearField())
- **Validation:** Required field validation changes at runtime
- **Default Values:** Optional vs required fields handle defaults differently

## Real-World Example

**Client code breaks with cardinality changes:**
```go
// Before - required field (proto3 implicit)
user := &myapi.User{
    Name:  "John",
    Email: "john@example.com", // Must be set
    Age:   30,
}

// Check if email is set (implicit presence)
if user.Email != "" {
    // Email is provided
}

// After - optional field (explicit presence) 
user := &myapi.User{
    Name: "John",
    Age:  30,
    // Email can be nil/unset
}

// Check if email is set (explicit presence)
if user.Email != nil && *user.Email != "" {  // Different API!
    // Email is provided
}

// Setting email also changes
user.Email = &emailValue  // Pointer assignment vs direct
```

**Server validation breaks:**
```go
// Before - implicit required field
func validateUser(user *User) error {
    if user.Email == "" {  // Empty string check
        return errors.New("email is required")
    }
    return nil
}

// After - explicit optional field
func validateUser(user *User) error {
    if user.Email == nil {  // Nil pointer check  
        return errors.New("email is required")
    }
    if *user.Email == "" {  // Dereference to check empty
        return errors.New("email cannot be empty")
    }
    return nil
}
```

## Migration Strategy

1. **Add new field** with correct cardinality:
   ```proto
   string old_email = 2 [deprecated = true];
   optional string new_email = 5;  // New field with desired cardinality
   ```

2. **Dual-write period** - populate both fields during transition:
   ```go
   user := &User{
       Name:     "John",
       OldEmail: email,      // Legacy field
       NewEmail: &email,     // New optional field
   }
   ```

3. **Update clients** to use new field gradually:
   ```go
   // Check both fields during migration
   email := user.NewEmail
   if email == nil && user.OldEmail != "" {
       emailValue := user.OldEmail
       email = &emailValue  // Fallback to old field
   }
   ```

4. **Remove old field** in next major version:
   ```proto
   reserved 2, "old_email";
   optional string new_email = 5;
   ```

## Common Scenarios

### Proto2 to Proto3 Migration
```proto
// Instead of direct syntax migration
syntax = "proto3";
message Order {
  string id = 1;  // BREAKING: was required in proto2
}

// Maintain semantics during migration  
syntax = "proto3";
message Order {
  string id = 1;  // Keep as implicit required
  // Add explicit optional fields only when needed
  optional string notes = 2;
}
```

### Making Optional Fields Required
```proto
// Instead of changing existing field
message CreateUserRequest {
  string name = 1;
  string email = 2;  // BREAKING: was optional
}

// Add new required field
message CreateUserRequest {
  string name = 1;
  optional string email = 2 [deprecated = true];
  string required_email = 3;  // New required field
}
```

### Making Required Fields Optional
```proto
// Instead of changing existing field  
message User {
  optional string phone = 3;  // BREAKING: was required
}

// Add new optional field
message User {
  string phone = 3 [deprecated = true];  // Keep required
  optional string phone_optional = 4;   // New optional field
}
```

## Cardinality Types in Protobuf

### Proto3 Cardinality
- **Implicit optional**: `string name = 1;` (default in proto3)
- **Explicit optional**: `optional string email = 2;` 
- **Repeated**: `repeated string tags = 3;`
- **Map**: `map<string, string> metadata = 4;`

### Proto2 Cardinality  
- **Required**: `required string id = 1;`
- **Optional**: `optional string email = 2;`
- **Repeated**: `repeated string tags = 3;`

### Breaking Changes Matrix

| From | To | Result |
|------|----|---------| 
| Required | Optional | ❌ BREAKING |
| Optional | Required | ❌ BREAKING |
| Implicit | Explicit | ❌ BREAKING |  
| Explicit | Implicit | ❌ BREAKING |
| Single | Repeated | ❌ BREAKING |
| Repeated | Single | ❌ BREAKING |
| Field | Map | ❌ BREAKING |
| Map | Field | ❌ BREAKING |