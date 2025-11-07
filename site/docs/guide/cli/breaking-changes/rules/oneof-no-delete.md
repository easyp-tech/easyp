# ONEOF_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no oneOf fields are deleted from messages. Deleting a oneOf breaks both wire format compatibility and generated code, as existing data may use the oneOf structure and client code depends on the generated oneOf types and accessor methods.

## Examples

### Bad

**Before:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    string password = 2;
    string api_key = 3;
    string oauth_token = 4;
  }
}

message PaymentMethod { // [!code --]
  oneof method { // [!code --]
    CreditCard credit_card = 1; // [!code --]
    BankAccount bank_account = 2; // [!code --]
    PayPal paypal = 3; // [!code --]
  } // [!code --]
} // [!code --]

message Order {
  string id = 1;
  PaymentMethod payment = 2;
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  // credentials oneof was deleted - BREAKING CHANGE!
  string password = 2;     // Now regular fields
  string api_key = 3;      // Lost exclusive semantics
  string oauth_token = 4;
}

message PaymentMethod {
  // method oneof was deleted - BREAKING CHANGE!
  CreditCard credit_card = 1;  // Now all can be set simultaneously
  BankAccount bank_account = 2;
  PayPal paypal = 3;
}

message Order {
  string id = 1;
  PaymentMethod payment = 2;
}
```

**Error:**
```
login.proto:5:3: Previously present oneof "credentials" was deleted. (BREAKING_CHECK)
payment.proto:2:3: Previously present oneof "method" was deleted. (BREAKING_CHECK)
```

### More Examples

**Nested oneOf deletion:**

```proto
// Before
message UserProfile {
  string name = 1;
  
  message ContactInfo {
    oneof contact_method { // [!code --]
      string email = 1; // [!code --]
      string phone = 2; // [!code --]
      string slack_id = 3; // [!code --]
    } // [!code --]
  }
  
  ContactInfo contact = 2;
}

// After - BREAKING CHANGE!
message UserProfile {
  string name = 1;
  
  message ContactInfo {
    // contact_method oneof was deleted
    string email = 1;    // No longer mutually exclusive
    string phone = 2;    // Can set all fields now
    string slack_id = 3;
  }
  
  ContactInfo contact = 2;
}
```

### Good

**Instead of deleting, deprecate the oneOf:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    option deprecated = true; // [!code focus]
    string password = 2 [deprecated = true]; // [!code focus]
    string api_key = 3 [deprecated = true]; // [!code focus]
    string oauth_token = 4 [deprecated = true]; // [!code focus]
  }
}

message PaymentMethod {
  oneof method {
    option deprecated = true; // [!code focus]
    CreditCard credit_card = 1 [deprecated = true]; // [!code focus]
    BankAccount bank_account = 2 [deprecated = true]; // [!code focus]
    PayPal paypal = 3 [deprecated = true]; // [!code focus]
  }
}
```

**Or replace with new structure:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    option deprecated = true; // [!code focus] // Old oneOf
    string password = 2 [deprecated = true]; // [!code focus]
    string api_key = 3 [deprecated = true]; // [!code focus]
    string oauth_token = 4 [deprecated = true]; // [!code focus]
  }
  
  oneof auth_method { // [!code focus] // New oneOf with better design
    PasswordAuth password_auth = 5; // [!code focus]
    ApiKeyAuth api_auth = 6; // [!code focus]
    OAuthAuth oauth_auth = 7; // [!code focus]
  } // [!code focus]
}
```

## Impact

- **Wire Format:** OneOf semantics are lost - fields are no longer mutually exclusive
- **Generated Code:** OneOf accessor methods are removed, breaking client compilation
- **Business Logic:** Mutual exclusivity constraints are violated
- **Validation:** Client code expecting only one field to be set may break

## Real-World Example

**Client code breaks:**
```go
// Before - oneOf provides mutual exclusivity
loginReq := &myapi.LoginRequest{
    Username: "user123",
}

// Set one of the credential options (oneOf semantics)
loginReq.Credentials = &myapi.LoginRequest_Password{
    Password: "secret123",
}

// Check which credential type is set
switch cred := loginReq.Credentials.(type) {
case *myapi.LoginRequest_Password:  // ERROR after oneOf deletion
    return authenticatePassword(cred.Password)
case *myapi.LoginRequest_ApiKey:    // ERROR after oneOf deletion  
    return authenticateApiKey(cred.ApiKey)
default:
    return errors.New("no credentials provided")
}

// Generated code compilation fails:
// undefined: myapi.LoginRequest_Password
// undefined field: loginReq.Credentials
```

**Server validation breaks:**
```go
// Before - server validates oneOf semantics
func validateLoginRequest(req *LoginRequest) error {
    if req.Credentials == nil {
        return errors.New("credentials required")
    }
    
    // OneOf ensures only one credential type is set
    switch req.Credentials.(type) {
    case *LoginRequest_Password:
        return validatePassword(req.GetPassword())
    case *LoginRequest_ApiKey:
        return validateApiKey(req.GetApiKey())
    case *LoginRequest_OauthToken:
        return validateOAuthToken(req.GetOauthToken())
    default:
        return errors.New("unknown credential type")
    }
}

// After oneOf deletion - validation logic breaks
func validateLoginRequest(req *LoginRequest) error {
    // No more oneOf - multiple fields can be set!
    credCount := 0
    if req.Password != "" { credCount++ }
    if req.ApiKey != "" { credCount++ }      
    if req.OauthToken != "" { credCount++ }
    
    if credCount == 0 {
        return errors.New("no credentials provided")
    }
    if credCount > 1 {
        return errors.New("multiple credentials provided") // Manual check needed!
    }
    
    // Lost type safety - need manual checks
    if req.Password != "" {
        return validatePassword(req.Password)
    }
    // ... more manual validation
}
```

**Business logic problems:**
```go
// Before - oneOf guarantees mutual exclusivity
payment := &PaymentMethod{
    Method: &PaymentMethod_CreditCard{
        CreditCard: &CreditCard{Number: "1234"},
    },
}
// Impossible to set multiple payment methods simultaneously

// After oneOf deletion - business logic can break
payment := &PaymentMethod{
    CreditCard:  &CreditCard{Number: "1234"},
    BankAccount: &BankAccount{Number: "5678"}, // Both set! Logic breaks!
}
```

## Migration Strategy

1. **Deprecate the oneOf and all its fields:**
   ```proto
   oneof old_credentials {
     option deprecated = true;
     string password = 2 [deprecated = true];
     string api_key = 3 [deprecated = true];
   }
   ```

2. **Create new structure** if needed:
   ```proto
   oneof new_auth_method {
     PasswordAuth password_auth = 5;
     ApiKeyAuth api_auth = 6;
   }
   ```

3. **Update server code** to handle both during transition:
   ```go
   func handleAuth(req *LoginRequest) error {
       // Handle new structure first
       if req.NewAuthMethod != nil {
           return handleNewAuth(req.NewAuthMethod)
       }
       
       // Fall back to old oneOf (deprecated)
       if req.Credentials != nil {
           return handleOldAuth(req.Credentials)
       }
       
       return errors.New("no auth method provided")
   }
   ```

4. **Migrate clients** to use new structure:
   ```go
   // Update client to use new oneOf
   loginReq.NewAuthMethod = &LoginRequest_PasswordAuth{
       PasswordAuth: &PasswordAuth{
           Username: "user123",
           Password: "secret123",
       },
   }
   ```

5. **Remove old oneOf** in next major version

## Common Scenarios

### Replacing OneOf with Regular Fields
```proto
// Instead of removing oneOf structure entirely
message SearchRequest {
  // Don't delete oneOf - breaks mutual exclusivity
  string text_query = 1;    // BREAKING: was in oneOf
  int32 user_id = 2;        // BREAKING: was in oneOf  
  string category = 3;      // BREAKING: was in oneOf
}

// Keep oneOf, add new fields if needed
message SearchRequest {
  oneof query {
    option deprecated = true;
    string text_query = 1 [deprecated = true];
    int32 user_id = 2 [deprecated = true];
    string category = 3 [deprecated = true];
  }
  
  // Add new non-exclusive fields if that's the intent
  string search_text = 4;
  repeated string filters = 5;
}
```

### Simplifying Message Structure
```proto
// Instead of flattening oneOf structure
message NotificationSettings {
  // Don't delete oneOf - loses type safety
  string email = 1;         // BREAKING: was in oneOf
  string phone = 2;         // BREAKING: was in oneOf
  string webhook_url = 3;   // BREAKING: was in oneOf
}

// Keep oneOf for type safety
message NotificationSettings {
  oneof delivery_method {
    EmailNotification email = 1;
    SmsNotification phone = 2; 
    WebhookNotification webhook = 3;
  }
}
```

### Proto Syntax Migration
```proto
// When migrating from proto2 to proto3
// Don't lose oneOf semantics during migration

// proto2
message Request {
  optional string name = 1;
  oneof data {
    string text = 2;
    int32 number = 3;
  }
}

// proto3 - keep oneOf structure
message Request {
  string name = 1;
  oneof data {  // Don't delete this!
    string text = 2;
    int32 number = 3;
  }
}
```

## OneOf Semantics

### What OneOf Provides
- **Mutual Exclusivity**: Only one field can be set at a time
- **Type Safety**: Generated code enforces single choice
- **Clear Intent**: API design expresses "choose one" semantics
- **Efficient Wire Format**: Only selected field is serialized

### What's Lost When Deleted
- **Business Logic**: No enforcement of mutual exclusivity
- **Generated Code**: No oneOf wrapper types or methods
- **Validation**: Manual checks needed for exclusivity
- **API Clarity**: Intent of "choose one" is lost

### Generated Code Impact
```go
// With oneOf - type safe
type LoginRequest struct {
    Username string
    // Types that are assignable to Credentials:
    //   *LoginRequest_Password
    //   *LoginRequest_ApiKey
    Credentials isLoginRequest_Credentials
}

// Without oneOf - no type safety
type LoginRequest struct {
    Username    string
    Password    string  // All can be set simultaneously
    ApiKey      string  // No mutual exclusivity
    OauthToken  string
}
```
