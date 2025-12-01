# ONEOF_FIELD_NO_DELETE

Categories:

- **WIRE+**

This rule checks that no fields are deleted from oneOf groups. Deleting a field from a oneOf breaks both wire format compatibility and generated code, as existing data may use the deleted oneOf option and client code depends on the generated field types and accessor methods.

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
    string certificate = 5;
  }
}

message SearchQuery {
  oneof filter {
    string text = 1;
    int32 user_id = 2;
    string category = 3;
    DateRange date_range = 4;
  }
}
```

**After:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    string password = 2;
    string api_key = 3;
    // string oauth_token = 4;  // [!code --] Deleted oneOf field - BREAKING!
    // string certificate = 5;  // [!code --] Deleted oneOf field - BREAKING!
  }
}

message SearchQuery {
  oneof filter {
    string text = 1;
    int32 user_id = 2;
    // string category = 3;     // [!code --] Deleted oneOf field - BREAKING!
    // DateRange date_range = 4; // [!code --] Deleted oneOf field - BREAKING!
  }
}
```

**Error:**
```
login.proto:7:5: Previously present field "4" with name "oauth_token" on OneOf "credentials" was deleted. (BREAKING_CHECK)
login.proto:8:5: Previously present field "5" with name "certificate" on OneOf "credentials" was deleted. (BREAKING_CHECK)
search.proto:6:5: Previously present field "3" with name "category" on OneOf "filter" was deleted. (BREAKING_CHECK)
search.proto:7:5: Previously present field "4" with name "date_range" on OneOf "filter" was deleted. (BREAKING_CHECK)
```

### More Examples

**Nested message oneOf field deletion:**

```proto
// Before
message PaymentRequest {
  string order_id = 1;
  
  oneof payment_method {
    CreditCardPayment credit_card = 2;
    BankTransferPayment bank_transfer = 3;
    CryptoPayment crypto = 4;         // [!code --] 
    GiftCardPayment gift_card = 5;    // [!code --]
  }
}

// After - BREAKING CHANGES!
message PaymentRequest {
  string order_id = 1;
  
  oneof payment_method {
    CreditCardPayment credit_card = 2;
    BankTransferPayment bank_transfer = 3;
    // CryptoPayment crypto = 4;      // BREAKING: deleted
    // GiftCardPayment gift_card = 5; // BREAKING: deleted
  }
}
```

### Good

**Instead of deleting, deprecate the oneOf fields:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    string password = 2;
    string api_key = 3;
    string oauth_token = 4 [deprecated = true]; // [!code focus]
    string certificate = 5 [deprecated = true]; // [!code focus]
  }
}

message SearchQuery {
  oneof filter {
    string text = 1;
    int32 user_id = 2;
    string category = 3 [deprecated = true]; // [!code focus]
    DateRange date_range = 4 [deprecated = true]; // [!code focus]
  }
}
```

**Or add new oneOf fields while keeping old ones:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    string password = 2;
    string api_key = 3;
    string oauth_token = 4 [deprecated = true]; // [!code focus] // Keep old field
    string certificate = 5 [deprecated = true]; // [!code focus] // Keep old field
    
    OAuthCredentials oauth_v2 = 6; // [!code focus] // New structured approach
    CertificateCredentials cert_v2 = 7; // [!code focus] // New structured approach
  }
}
```

## Impact

- **Wire Format:** Existing data with deleted oneOf fields cannot be deserialized properly
- **Generated Code:** OneOf field accessor methods are removed, breaking client compilation
- **Type Safety:** Generated oneOf wrapper types lose options, breaking switch statements
- **Business Logic:** Client code handling specific oneOf cases breaks

## Real-World Example

**Client code breaks:**
```go
// Before - this code works
loginReq := &myapi.LoginRequest{
    Username: "user123",
}

// Set OAuth token (oneOf field)
loginReq.Credentials = &myapi.LoginRequest_OauthToken{
    OauthToken: "token123",  // ERROR after field deletion
}

// Switch on oneOf field types
switch cred := loginReq.Credentials.(type) {
case *myapi.LoginRequest_Password:
    return authenticatePassword(cred.Password)
case *myapi.LoginRequest_OauthToken:  // ERROR: undefined type after deletion
    return authenticateOAuth(cred.OauthToken)
case *myapi.LoginRequest_Certificate: // ERROR: undefined type after deletion
    return authenticateCert(cred.Certificate)
default:
    return errors.New("unsupported credential type")
}

// Generated code compilation fails:
// undefined: myapi.LoginRequest_OauthToken
// undefined: myapi.LoginRequest_Certificate
```

**Server handling breaks:**
```go
// Before - server handles all oneOf field types
func authenticateUser(req *LoginRequest) error {
    switch cred := req.Credentials.(type) {
    case *LoginRequest_Password:
        return validatePassword(cred.Password)
    case *LoginRequest_ApiKey:
        return validateApiKey(cred.ApiKey)
    case *LoginRequest_OauthToken:    // ERROR after deletion
        return validateOAuth(cred.OauthToken)
    case *LoginRequest_Certificate:   // ERROR after deletion
        return validateCertificate(cred.Certificate)
    default:
        return errors.New("no valid credentials provided")
    }
}
```

**Existing data becomes unreadable:**
```json
// Serialized data before deletion
{
  "username": "user123",
  "oauth_token": "abc123xyz"
}

// After oneOf field deletion
// Data deserializes but oauth_token is lost or causes parsing errors
// User authentication fails due to missing credential data
```

## Migration Strategy

1. **Deprecate oneOf fields first:**
   ```proto
   oneof credentials {
     string password = 2;
     string oauth_token = 4 [deprecated = true];
   }
   ```

2. **Add new fields** if replacement is needed:
   ```proto
   oneof credentials {
     string password = 2;
     string oauth_token = 4 [deprecated = true];
     OAuthCredentials oauth_v2 = 6;  // New structured field
   }
   ```

3. **Update server code** to handle both old and new fields:
   ```go
   func authenticateUser(req *LoginRequest) error {
       switch cred := req.Credentials.(type) {
       case *LoginRequest_OauthV2:
           // Handle new structured OAuth
           return validateOAuthV2(cred.OauthV2)
       case *LoginRequest_OauthToken:
           // Handle legacy OAuth (deprecated)
           log.Warn("Using deprecated oauth_token field")
           return validateOAuthLegacy(cred.OauthToken)
       // ... other cases
       }
   }
   ```

4. **Migrate clients** to use new fields:
   ```go
   // Update client to use new structured field
   loginReq.Credentials = &LoginRequest_OauthV2{
       OauthV2: &OAuthCredentials{
           Token:        "token123",
           RefreshToken: "refresh456",
           ExpiresAt:    timestamp,
       },
   }
   ```

5. **Reserve field numbers** after removal in next major version:
   ```proto
   oneof credentials {
     reserved 4, 5;
     reserved "oauth_token", "certificate";
     
     string password = 2;
     string api_key = 3;
     OAuthCredentials oauth_v2 = 6;
   }
   ```

## Common Scenarios

### Removing Unsupported Authentication Methods
```proto
// Instead of deleting discontinued auth methods
message AuthRequest {
  oneof method {
    PasswordAuth password = 1;
    ApiKeyAuth api_key = 2;
    // LdapAuth ldap = 3;  // Don't delete even if LDAP discontinued!
    
    // Better approach:
    LdapAuth ldap = 3 [deprecated = true];  // Mark as deprecated
    OidcAuth oidc = 4;  // Add new method
  }
}
```

### Simplifying OneOf Options
```proto
// Instead of removing complex options
message SearchRequest {
  oneof query_type {
    string simple_text = 1;
    // ComplexQuery complex = 2;  // Don't delete complex queries!
    
    // Better approach:
    ComplexQuery complex = 2 [deprecated = true];
    AdvancedQuery advanced = 3;  // Replacement with better design
  }
}
```

### Business Logic Changes
```proto
// Instead of removing payment methods
message PaymentRequest {
  oneof payment_method {
    CreditCard credit_card = 1;
    // Check check = 2;  // Don't delete even if checks discontinued!
    BankTransfer bank_transfer = 3;
    
    // Better approach:
    Check check = 2 [deprecated = true];  // Keep for existing data
    DigitalWallet digital_wallet = 4;     // Add new methods
  }
}
```

## OneOf Field Deletion vs Other Deletions

### Different from Regular Field Deletion
```proto
// Regular field deletion (also breaking)
message User {
  string name = 1;
  // string email = 2;  // BREAKING: regular field deletion
}

// OneOf field deletion (breaks oneOf semantics)
message User {
  oneof contact {
    string email = 1;
    // string phone = 2;  // BREAKING: oneOf field deletion
  }
}
```

### Impact on Generated Code
```go
// Regular field - simple accessor removed
// user.GetEmail() // Method removed

// OneOf field - wrapper type removed  
// *LoginRequest_OauthToken  // Entire type removed
// switch statement cases break completely
```

### Wire Format Considerations
```proto
// OneOf fields share the same "choice" semantics
// Removing a choice breaks clients that made that choice

message Request {
  oneof data {
    string text = 1;      // Choice A
    bytes binary = 2;     // Choice B - if removed, clients using B break
    JsonData json = 3;    // Choice C
  }
}
```

## Validation and Error Handling

### Before Deletion (Type Safe)
```go
func validateRequest(req *Request) error {
    switch data := req.Data.(type) {
    case *Request_Text:
        return validateText(data.Text)
    case *Request_Binary:      // Type exists
        return validateBinary(data.Binary)
    case *Request_Json:
        return validateJson(data.Json)
    case nil:
        return errors.New("no data provided")
    default:
        return errors.New("unknown data type")
    }
}
```

### After Deletion (Broken)
```go
func validateRequest(req *Request) error {
    switch data := req.Data.(type) {
    case *Request_Text:
        return validateText(data.Text)
    // case *Request_Binary:   // ERROR: type no longer exists
    //     return validateBinary(data.Binary)
    case *Request_Json:
        return validateJson(data.Json)
    case nil:
        return errors.New("no data provided")
    default:
        // Existing binary data falls here now!
        return errors.New("unknown data type")
    }
}
```
