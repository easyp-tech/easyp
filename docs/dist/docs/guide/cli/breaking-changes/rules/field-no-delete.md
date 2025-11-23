# FIELD_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no message fields are deleted. Deleting a field breaks both wire format compatibility and generated code, as existing data may contain the deleted field and client code may reference it.

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
  string phone = 4; // [!code --]
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  int32 age = 3;
  // phone field was deleted - BREAKING CHANGE!
}
```

**Error:**
```
user.proto:6:3: Previously present field "4" with name "phone" on message "User" was deleted. (BREAKING_CHECK)
```

### Good

**Instead of deleting, deprecate the field:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  int32 age = 3;
  string phone = 4 [deprecated = true]; // [!code focus]
}
```

**Or reserve the field number after removal:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  reserved 4; // [!code focus]
  reserved "phone"; // [!code focus]
  
  string name = 1;
  string email = 2;
  int32 age = 3;
}
```

## Impact

- **Wire Format:** Old messages with the deleted field cannot be properly deserialized
- **Generated Code:** Field accessors are removed, breaking client compilation
- **Data Loss:** Existing serialized data loses information
- **JSON Compatibility:** JSON parsers expect the field to exist

## Migration Strategy

1. **Deprecate first:**
   ```proto
   string old_field = 5 [deprecated = true];
   ```

2. **Stop writing to the field** in your application code

3. **Reserve the field** in next version to prevent accidental reuse:
   ```proto
   reserved 5;
   reserved "old_field";
   ```

4. **Never reuse field numbers** - they must remain reserved permanently