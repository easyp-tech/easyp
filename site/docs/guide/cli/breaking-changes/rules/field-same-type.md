# FIELD_SAME_TYPE

Categories:

- **WIRE+**

This rule checks that message fields maintain the same type. Changing a field's type breaks wire format compatibility and generated code, as the binary representation and client code expectations differ between types.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

message Product {
  string name = 1;
  int32 price = 2;
  bool available = 3;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message Product {
  string name = 1;
  string price = 2; // [!code --] Changed from int32 to string
  bool available = 3;
}
```

**Error:**
```
product.proto:5:3: Field "2" with name "price" on message "Product" changed type from "int32" to "string". (BREAKING_CHECK)
```

### More Examples

**Incompatible type changes:**

```proto
// Before
message Order {
  string id = 1;
  int64 timestamp = 2;
  repeated string tags = 3;
  OrderStatus status = 4;
}

// After - ALL BREAKING CHANGES!
message Order {
  int32 id = 1;           // string -> int32: BREAKING
  string timestamp = 2;   // int64 -> string: BREAKING  
  string tags = 3;        // repeated -> singular: BREAKING
  int32 status = 4;       // enum -> int32: BREAKING
}
```

### Good

**Instead of changing type, add a new field:**
```proto
syntax = "proto3";

package myapi.v1;

message Product {
  string name = 1;
  int32 price = 2 [deprecated = true]; // [!code focus]
  bool available = 3;
  string price_formatted = 4; // [!code focus] // New field for string price
}
```

**Or create a new message version:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus] // New package version

message Product {
  string name = 1;
  string price = 2; // [!code focus] // Now string in v2
  bool available = 3;
}
```

## Impact

- **Wire Format:** Binary data cannot be deserialized correctly between versions
- **Generated Code:** Field types change, breaking client compilation
- **Runtime Errors:** Type mismatches cause parsing failures
- **Data Corruption:** Incorrect interpretation of binary data

## Common Type Change Issues

### Numeric Types
```proto
// BREAKING: Different wire representations
int32 -> int64   // Different encoding
int32 -> uint32  // Different sign interpretation
int32 -> string  // Completely different format
```

### Collection Types
```proto
// BREAKING: Structure changes
string -> repeated string     // Singular to collection
repeated int32 -> map<string, int32>  // Array to map
```

### Message Types
```proto
// BREAKING: Different message structure
UserInfo -> UserProfile  // Different message types
string -> UserInfo       // Scalar to message
```

## Migration Strategy

1. **Add new field** with correct type:
   ```proto
   int32 old_price = 2 [deprecated = true];
   string new_price = 5;  // New field
   ```

2. **Dual-write period** - populate both fields during transition

3. **Update clients** to use new field

4. **Remove old field** in next major version:
   ```proto
   reserved 2, "old_price";
   string new_price = 5;
   ```

## Safe Type Changes (Wire Compatible)

Note: EasyP currently treats ALL type changes as breaking. However, some changes are technically wire-compatible:

- `int32` ↔ `uint32` (same encoding)
- `int64` ↔ `uint64` (same encoding)  
- `string` → `bytes` (if UTF-8 valid)

These may be supported in future EasyP versions with different strictness levels.