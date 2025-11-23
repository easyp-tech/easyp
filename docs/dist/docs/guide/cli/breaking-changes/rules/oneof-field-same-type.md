# ONEOF_FIELD_SAME_TYPE

Categories:

- **WIRE+**

This rule checks that oneOf field types cannot be changed. Changing the type of a field within a oneOf breaks both wire format compatibility and generated code, as existing data contains the old type and client code expects specific field types in oneOf wrapper structures.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    string category = 2;
    int32 user_id = 3;
    bool is_premium = 4;
    DateRange date_range = 5;
  }
}

message NotificationSettings {
  oneof delivery {
    string email = 1;
    int64 phone_number = 2;
    WebhookConfig webhook = 3;
  }
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    int32 category = 2;        // [!code --] Changed from string to int32 - BREAKING!
    string user_id = 3;        // [!code --] Changed from int32 to string - BREAKING!
    string is_premium = 4;     // [!code --] Changed from bool to string - BREAKING!
    string date_range = 5;     // [!code --] Changed from DateRange to string - BREAKING!
  }
}

message NotificationSettings {
  oneof delivery {
    EmailConfig email = 1;     // [!code --] Changed from string to EmailConfig - BREAKING!
    string phone_number = 2;   // [!code --] Changed from int64 to string - BREAKING!
    string webhook = 3;        // [!code --] Changed from WebhookConfig to string - BREAKING!
  }
}
```

**Error:**
```
search.proto:6:5: Field "2" with name "category" on OneOf "filter" changed type from "string" to "int32". (BREAKING_CHECK)
search.proto:7:5: Field "3" with name "user_id" on OneOf "filter" changed type from "int32" to "string". (BREAKING_CHECK)
search.proto:8:5: Field "4" with name "is_premium" on OneOf "filter" changed type from "bool" to "string". (BREAKING_CHECK)
search.proto:9:5: Field "5" with name "date_range" on OneOf "filter" changed type from "DateRange" to "string". (BREAKING_CHECK)
```

### More Examples

**Complex type changes:**

```proto
// Before
message PaymentRequest {
  oneof payment_info {
    CreditCardInfo credit_card = 1;
    BankAccountInfo bank_account = 2;
    string paypal_email = 3;
    CryptoWalletInfo crypto = 4;
  }
}

// After - ALL BREAKING CHANGES!
message PaymentRequest {
  oneof payment_info {
    string credit_card = 1;           // BREAKING: CreditCardInfo -> string
    PayPalInfo bank_account = 2;      // BREAKING: BankAccountInfo -> PayPalInfo  
    PayPalInfo paypal_email = 3;      // BREAKING: string -> PayPalInfo
    string crypto = 4;                // BREAKING: CryptoWalletInfo -> string
  }
}
```

### Good

**Instead of changing type, add new oneOf field:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    string category = 2 [deprecated = true];     // [!code focus] // Keep original type
    int32 user_id = 3 [deprecated = true];       // [!code focus] // Keep original type
    bool is_premium = 4 [deprecated = true];     // [!code focus] // Keep original type
    DateRange date_range = 5 [deprecated = true]; // [!code focus] // Keep original type
    
    int32 category_id = 6;                       // [!code focus] // New field with desired type
    string user_identifier = 7;                 // [!code focus] // New field with desired type
    string premium_status = 8;                  // [!code focus] // New field with desired type
    string date_filter = 9;                    // [!code focus] // New field with desired type
  }
}
```

**Or create a new oneOf group:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    option deprecated = true;                   // [!code focus] // Deprecate old oneOf
    string category = 2 [deprecated = true];    // [!code focus]
    int32 user_id = 3 [deprecated = true];      // [!code focus]
    bool is_premium = 4 [deprecated = true];    // [!code focus]
    DateRange date_range = 5 [deprecated = true]; // [!code focus]
  }
  
  oneof filter_v2 {                            // [!code focus] // New oneOf with correct types
    int32 category_id = 6;                     // [!code focus]
    string user_identifier = 7;               // [!code focus]
    PremiumFilter premium_filter = 8;          // [!code focus]
    TimeFilter time_filter = 9;               // [!code focus]
  }                                           // [!code focus]
}
```

## Impact

- **Wire Format:** Binary data cannot be deserialized correctly between versions
- **Generated Code:** OneOf wrapper types change, breaking client compilation  
- **Type Safety:** Switch statements on oneOf types break with compilation errors
- **Runtime Errors:** Type mismatches cause parsing failures for existing data

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
searchReq := &myapi.SearchRequest{
    Query: "golang jobs",
}

// Set filter using oneOf (original types)
searchReq.Filter = &myapi.SearchRequest_Category{
    Category: "engineering",  // string type
}

// Handle filter types
switch filter := searchReq.Filter.(type) {
case *myapi.SearchRequest_Category:
    return searchByCategory(filter.Category)  // string parameter
case *myapi.SearchRequest_UserId:
    return searchByUserId(filter.UserId)      // int32 parameter
default:
    return errors.New("unsupported filter")
}

// After type change - compilation fails
searchReq.Filter = &myapi.SearchRequest_Category{
    Category: "engineering",  // ERROR: expects int32, not string
}

// Switch statement breaks
switch filter := searchReq.Filter.(type) {
case *myapi.SearchRequest_Category:
    return searchByCategory(filter.Category)  // ERROR: Category is now int32
case *myapi.SearchRequest_UserId:   
    return searchByUserId(filter.UserId)      // ERROR: UserId is now string
}
```

**Server implementation breaks:**
```go
// Before - server expects original types
func handleSearchFilter(req *SearchRequest) (*SearchResult, error) {
    switch filter := req.Filter.(type) {
    case *SearchRequest_Category:
        categoryName := filter.Category  // string
        return searchDatabase("category", categoryName), nil
    case *SearchRequest_UserId:
        userId := filter.UserId  // int32
        return searchDatabase("user_id", strconv.Itoa(int(userId))), nil
    default:
        return nil, errors.New("no filter specified")
    }
}

// After type change - runtime errors
func handleSearchFilter(req *SearchRequest) (*SearchResult, error) {
    switch filter := req.Filter.(type) {
    case *SearchRequest_Category:
        categoryId := filter.Category  // now int32, but code expects string
        return searchDatabase("category", categoryId), nil  // ERROR: type mismatch
    case *SearchRequest_UserId:
        userIdStr := filter.UserId  // now string, but conversion logic expects int32
        userId, _ := strconv.Atoi(userIdStr)  // Works but semantics changed
        return searchDatabase("user_id", strconv.Itoa(userId)), nil
    }
}
```

**Existing data corruption:**
```json
// Serialized data before type change
{
  "query": "golang jobs",
  "category": "engineering"
}

// After changing category from string to int32
// Data cannot be deserialized - "engineering" is not a valid int32
// Results in parsing errors or data loss
```

## Migration Strategy

1. **Add new oneOf fields** with correct types:
   ```proto
   oneof filter {
     string category = 2 [deprecated = true];
     int32 category_id = 6;  // New field with desired type
   }
   ```

2. **Update server code** to handle both old and new fields:
   ```go
   func handleSearchFilter(req *SearchRequest) error {
       switch filter := req.Filter.(type) {
       case *SearchRequest_CategoryId:
           // Handle new int32 field
           return searchByCategoryId(filter.CategoryId)
       case *SearchRequest_Category:
           // Handle legacy string field (deprecated)
           log.Warn("Using deprecated string category field")
           categoryId := convertCategoryNameToId(filter.Category)
           return searchByCategoryId(categoryId)
       }
   }
   ```

3. **Migrate clients** to use new fields:
   ```go
   // Update client to use new typed field
   searchReq.Filter = &SearchRequest_CategoryId{
       CategoryId: 123,  // int32 instead of string
   }
   ```

4. **Provide migration utilities** for data conversion:
   ```go
   func MigrateSearchRequest(old *SearchRequestV1) *SearchRequestV2 {
       new := &SearchRequestV2{Query: old.Query}
       
       switch filter := old.Filter.(type) {
       case *SearchRequestV1_Category:
           // Convert string category to int32 ID
           categoryId := lookupCategoryId(filter.Category)
           new.Filter = &SearchRequestV2_CategoryId{CategoryId: categoryId}
       }
       
       return new
   }
   ```

5. **Remove old fields** in next major version:
   ```proto
   oneof filter {
     reserved 2, "category";
     int32 category_id = 6;
   }
   ```

## Common Scenarios

### Improving Type Safety
```proto
// Instead of changing loose string types to structured types
message ConfigRequest {
  oneof setting {
    // string database_url = 1;  // Don't change type directly!
    
    // Add new structured field instead
    string database_url = 1 [deprecated = true];
    DatabaseConfig database_config = 5;  // New structured field
  }
}
```

### Standardizing Field Types  
```proto
// Instead of converting between numeric types
message FilterRequest {
  oneof criteria {
    // int32 timestamp = 1;  // Don't change to int64 directly!
    
    // Add new field with correct type
    int32 timestamp = 1 [deprecated = true];
    int64 timestamp_millis = 5;  // New field with proper precision
  }
}
```

### Message Structure Changes
```proto
// Instead of flattening/expanding message types
message PaymentRequest {
  oneof payment {
    // string card_number = 1;  // Don't change to structured type!
    
    // Provide both options during transition
    string card_number = 1 [deprecated = true];
    CreditCardDetails card_details = 5;  // New structured approach
  }
}
```

## Type Change Compatibility Matrix

### Never Compatible (Always Breaking)
| From Type | To Type | Result |
|-----------|---------|---------|
| `string` | `int32/int64` | ❌ BREAKING |
| `int32` | `string` | ❌ BREAKING |  
| `bool` | `string` | ❌ BREAKING |
| `message` | `string` | ❌ BREAKING |
| `string` | `message` | ❌ BREAKING |
| `enum` | `int32` | ❌ BREAKING |

### Generated Code Impact
```go
// String to int32 change breaks wrapper types
// Before:
type SearchRequest_Category struct {
    Category string
}

// After:  
type SearchRequest_Category struct {
    Category int32  // Different type - compilation fails
}

// All client code using SearchRequest_Category breaks
```

### Wire Format Considerations
```proto
// OneOf fields in wire format are identified by field number
// Type changes break deserialization even if field number stays same

message Request {
  oneof data {
    string text = 1;    // Wire tag: 1, type: string
    // int32 text = 1;   // Same wire tag, different type - BREAKING!
  }
}
```

## Validation and Error Handling

### Type-Safe Validation (Before Change)
```go
func validateSearchFilter(req *SearchRequest) error {
    switch filter := req.Filter.(type) {
    case *SearchRequest_Category:
        if filter.Category == "" {  // String validation
            return errors.New("category cannot be empty")
        }
    case *SearchRequest_UserId:
        if filter.UserId <= 0 {  // Integer validation
            return errors.New("user_id must be positive")
        }
    }
    return nil
}
```

### Broken Validation (After Type Change)
```go
func validateSearchFilter(req *SearchRequest) error {
    switch filter := req.Filter.(type) {
    case *SearchRequest_Category:
        if filter.Category <= 0 {  // ERROR: Category is now int32, validation logic wrong
            return errors.New("category must be positive")
        }
    case *SearchRequest_UserId:
        if filter.UserId == "" {  // ERROR: UserId is now string, validation logic wrong  
            return errors.New("user_id cannot be empty")
        }
    }
    return nil
}
```
