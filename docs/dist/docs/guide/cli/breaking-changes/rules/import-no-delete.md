# IMPORT_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no import statements are deleted from proto files. Deleting an import breaks both wire format compatibility and generated code, as the imported types may be referenced in the current file and removing the import makes those types unavailable.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "common/user.proto";
import "common/address.proto";

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Duration processing_time = 3;
  common.User customer = 4;
  common.Address shipping_address = 5;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
// import "google/protobuf/duration.proto";  // [!code --] Deleted import - BREAKING!
// import "common/user.proto";               // [!code --] Deleted import - BREAKING!
// import "common/address.proto";            // [!code --] Deleted import - BREAKING!

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Duration processing_time = 3;  // ERROR: Duration not available
  common.User customer = 4;                      // ERROR: User not available
  common.Address shipping_address = 5;           // ERROR: Address not available
}
```

**Error:**
```
order.proto:5:1: Previously import "google/protobuf/duration.proto" was deleted. (BREAKING_CHECK)
order.proto:6:1: Previously import "common/user.proto" was deleted. (BREAKING_CHECK)
order.proto:7:1: Previously import "common/address.proto" was deleted. (BREAKING_CHECK)
```

### More Examples

**Service definition breaks with deleted imports:**

```proto
// Before
import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "common/auth.proto";

service OrderService {
  rpc GetOrder(GetOrderRequest) returns (Order) {
    option (google.api.http) = {
      get: "/v1/orders/{id}"
    };
  }
  
  rpc DeleteOrder(common.AuthRequest) returns (google.protobuf.Empty);
}

// After - BREAKING CHANGES!
// import "google/api/annotations.proto";    // BREAKING: deleted
// import "google/protobuf/empty.proto";     // BREAKING: deleted  
// import "common/auth.proto";               // BREAKING: deleted

service OrderService {
  rpc GetOrder(GetOrderRequest) returns (Order) {
    option (google.api.http) = {  // ERROR: annotations not available
      get: "/v1/orders/{id}"
    };
  }
  
  rpc DeleteOrder(common.AuthRequest) returns (google.protobuf.Empty);  // ERROR: types not available
}
```

### Good

**Instead of deleting, keep unused imports:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";  // [!code focus] // Keep even if unused
import "common/user.proto";               // [!code focus] // Keep even if unused
import "common/address.proto";            // [!code focus] // Keep even if unused

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  // Removed fields that used Duration, User, Address but kept imports
}
```

**Or replace with equivalent types while keeping imports:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";  // [!code focus] // Keep original import
import "common/user.proto";               // [!code focus] // Keep original import
import "common/address.proto";            // [!code focus] // Keep original import
import "common/v2/user.proto";            // [!code focus] // Add new import

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Duration processing_time = 3 [deprecated = true]; // [!code focus] // Keep old field
  common.User customer = 4 [deprecated = true];                    // [!code focus] // Keep old field
  common.Address shipping_address = 5 [deprecated = true];         // [!code focus] // Keep old field
  
  int32 processing_seconds = 6;             // [!code focus] // New field instead of Duration
  common.v2.UserProfile customer_v2 = 7;    // [!code focus] // New field with updated type
  string shipping_address_text = 8;         // [!code focus] // New field with simpler type
}
```

## Impact

- **Generated Code:** Imported types become unavailable, breaking compilation
- **Wire Format:** Messages using imported types cannot be deserialized
- **Field References:** Any field using imported types becomes invalid
- **Service Definitions:** RPC methods using imported types break
- **Options Usage:** Custom options from imports become unavailable

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
order := &myapi.Order{
    Id:        "order123",
    CreatedAt: timestamppb.New(time.Now()),              // Uses google.protobuf.Timestamp
    Customer: &common.User{                              // ERROR after import deletion
        Id:   "user456",
        Name: "John Doe",
    },
    ShippingAddress: &common.Address{                    // ERROR after import deletion
        Street: "123 Main St",
        City:   "New York",
    },
}

// Generated code compilation fails:
// undefined: common.User
// undefined: common.Address
```

**Server implementation breaks:**
```go
// Before - server uses imported types
func (s *server) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    return &Order{
        Id:        generateOrderId(),
        CreatedAt: timestamppb.New(time.Now()),
        Customer: &common.User{                          // ERROR after import deletion
            Id:   req.CustomerId,
            Name: req.CustomerName,
        },
        ShippingAddress: &common.Address{                // ERROR after import deletion
            Street: req.ShippingStreet,
            City:   req.ShippingCity,
        },
    }, nil
}
```

**Proto compilation breaks:**
```bash
# After deleting imports
$ protoc --go_out=. order.proto
order.proto:8:3: "common.User" is not defined.
order.proto:9:3: "common.Address" is not defined.
order.proto:10:3: "google.protobuf.Duration" is not defined.
```

## Migration Strategy

1. **Keep imports even when unused:**
   ```proto
   // Always keep existing imports
   import "common/user.proto";  // Keep even if no fields use it
   ```

2. **Add new imports alongside old ones:**
   ```proto
   import "common/user.proto";     // Keep old import
   import "common/v2/user.proto";  // Add new import
   ```

3. **Deprecate fields using old imports gradually:**
   ```proto
   message Order {
     common.User customer = 4 [deprecated = true];
     common.v2.UserProfile customer_v2 = 5;  // New field
   }
   ```

4. **Update server code** to handle both old and new fields:
   ```go
   func convertToOrder(req *CreateOrderRequest) *Order {
       order := &Order{Id: generateId()}
       
       // Handle new field first
       if req.CustomerProfileV2 != nil {
           order.CustomerV2 = req.CustomerProfileV2
           return order
       }
       
       // Fall back to old field (still works due to kept import)
       if req.CustomerProfile != nil {
           order.Customer = req.CustomerProfile
       }
       
       return order
   }
   ```

5. **Only remove imports in major version updates** after all fields are migrated

## Common Scenarios

### Replacing Dependencies
```proto
// Instead of deleting old import immediately
// import "old/user.proto";  // Don't delete!

// Keep both during transition
import "old/user.proto";    // Keep for backward compatibility
import "new/user.proto";    // Add new dependency

message Order {
  old.User customer = 4 [deprecated = true];
  new.UserProfile customer_v2 = 5;
}
```

### Removing Unused Features
```proto
// Even if removing features, keep imports
import "feature/analytics.proto";  // Keep even if analytics removed

message Order {
  string id = 1;
  // feature.AnalyticsData analytics = 2;  // Field removed but import kept
}
```

### Protobuf Version Migration
```proto
// When migrating protobuf versions, keep old imports
import "google/protobuf/timestamp.proto";  // Keep standard imports
import "google/protobuf/duration.proto";   // Even if switching to alternatives

// Prefer keeping both old and new approaches
message Event {
  google.protobuf.Timestamp timestamp = 1;    // Keep old
  int64 timestamp_millis = 2;                 // Add alternative
}
```

### Simplifying Dependencies
```proto
// Instead of removing complex type imports
// import "complex/config.proto";  // Don't delete!

// Keep complex import, add simple alternative
import "complex/config.proto";

message Settings {
  complex.ConfigData config = 1 [deprecated = true];  // Keep complex option
  string config_json = 2;                            // Add simple option
}
```

## Import Categories and Impact

### Standard Google Imports
```proto
// Critical - used throughout protobuf ecosystem
import "google/protobuf/timestamp.proto";    // Never delete
import "google/protobuf/duration.proto";     // Never delete
import "google/protobuf/empty.proto";        // Never delete
import "google/api/annotations.proto";       // Never delete if using gRPC-Gateway
```

### Common Library Imports
```proto
// Shared types across services
import "common/user.proto";        // High impact - affects all user references
import "common/address.proto";     // High impact - affects all address fields
import "common/money.proto";       // High impact - affects all monetary fields
```

### Feature-Specific Imports
```proto
// Feature modules that might be deprecated
import "features/analytics.proto";  // Medium impact - specific feature
import "features/logging.proto";    // Medium impact - specific feature
import "internal/debug.proto";      // Low impact - internal use
```

## Wire Format Considerations

### Import Deletion vs Field Usage
```proto
// Import deletion breaks even if types aren't directly used as fields
import "common/enums.proto";

message Order {
  // Even if no fields directly use common.enums types,
  // nested messages or oneof fields might use them
  oneof payment {
    CreditCardPayment card = 1;  // This might internally use common.enums
  }
}
```

### Transitive Dependencies
```proto
// Deleting an import can break transitive dependencies
import "user.proto";  // Don't delete even if not directly used

message Order {
  PaymentInfo payment = 1;  // PaymentInfo might internally reference user.proto types
}
```

### JSON Serialization Impact
```proto
// Imports affect JSON field names and validation
import "google/protobuf/field_mask.proto";

message UpdateRequest {
  google.protobuf.FieldMask update_mask = 1;  // JSON serialization depends on import
}
```

## Prevention Strategies

### Import Auditing
```bash
# Check which imports are actually used
$ grep -r "google.protobuf.Duration" *.proto
$ grep -r "common.User" *.proto

# But still keep imports even if grep shows no usage!
```

### Linting Configuration
```yaml
# Configure linters to allow unused imports
lint:
  ignore:
    - IMPORT_UNUSED  # Allow unused imports for backward compatibility
```

### Documentation
```proto
// Document why imports are kept
import "legacy/types.proto";  // Keep for backward compatibility with v1 clients
import "common/user.proto";   // Required for existing serialized data
```

## Import vs Field Deletion

### Different Breaking Change Types
```proto
// Import deletion (this rule)
// import "user.proto";  // BREAKING: import deleted

// vs Field deletion (different rule)
message Order {
  string id = 1;
  // User customer = 2;  // BREAKING: field deleted (different from import)
}
```

### Compound Breaking Changes
```proto
// Worst case: both import AND field deleted
// import "user.proto";  // BREAKING: import deleted
message Order {
  string id = 1;
  // User customer = 2;  // BREAKING: field also deleted
}
// This creates multiple breaking change violations
```
