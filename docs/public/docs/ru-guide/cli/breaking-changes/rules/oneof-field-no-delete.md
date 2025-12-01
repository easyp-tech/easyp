<!-- TODO: Review translation -->

# ONEOF_FIELD_NO_DELETE

Категории:

- **WIRE+**

Правило проверяет, что ни одно поле внутри `oneof` группы не было удалено. Удаление поля из `oneof` ломает совместимость по wire‑формату и сгенерированный код: в сохранённых данных может присутствовать выбранный вариант, а клиентский код опирается на сгенерированные типы‑обёртки и методы доступа.

## Примеры

### Плохой (Breaking)

**До:**
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

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    string password = 2;
    string api_key = 3;
    // string oauth_token = 4;  // [!code --] Удалено поле oneof — BREAKING!
    // string certificate = 5;  // [!code --] Удалено поле oneof — BREAKING!
  }
}

message SearchQuery {
  oneof filter {
    string text = 1;
    int32 user_id = 2;
    // string category = 3;     // [!code --] Удалено поле oneof — BREAKING!
    // DateRange date_range = 4; // [!code --] Удалено поле oneof — BREAKING!
  }
}
```

**Ошибка:**
```
login.proto:7:5: Previously present field "4" with name "oauth_token" on OneOf "credentials" was deleted. (BREAKING_CHECK)
login.proto:8:5: Previously present field "5" with name "certificate" on OneOf "credentials" was deleted. (BREAKING_CHECK)
search.proto:6:5: Previously present field "3" with name "category" on OneOf "filter" was deleted. (BREAKING_CHECK)
search.proto:7:5: Previously present field "4" with name "date_range" on OneOf "filter" was deleted. (BREAKING_CHECK)
```

### Дополнительные примеры

**Удаление вложенных oneof полей:**
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

### Хороший (Безопасный)

**Вместо удаления — пометить поля deprecated:**
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

**Или добавить новые поля, оставив старые:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    string password = 2;
    string api_key = 3;
    string oauth_token = 4 [deprecated = true]; // [!code focus] // Старое поле
    string certificate = 5 [deprecated = true]; // [!code focus]
    
    OAuthCredentials oauth_v2 = 6; // [!code focus] // Новый структурированный вариант
    CertificateCredentials cert_v2 = 7; // [!code focus]
  }
}
```

## Влияние

- **Wire Format:** Данные, содержащие удалённый вариант, не десериализуются корректно
- **Generated Code:** Удаляются типы‑обёртки и accessor методы — компиляция клиентов ломается
- **Type Safety:** Логика `switch` по типам oneof разваливается
- **Business Logic:** Код, ожидающий конкретный вариант, перестаёт работать

## Реальный пример

**Код клиента ломается:**
```go
// Before
loginReq := &myapi.LoginRequest{
    Username: "user123",
}

loginReq.Credentials = &myapi.LoginRequest_OauthToken{
    OauthToken: "token123", // ERROR после удаления поля
}

switch cred := loginReq.Credentials.(type) {
case *myapi.LoginRequest_Password:
    return authenticatePassword(cred.Password)
case *myapi.LoginRequest_OauthToken:  // ERROR: тип удалён
    return authenticateOAuth(cred.OauthToken)
case *myapi.LoginRequest_Certificate: // ERROR: тип удалён
    return authenticateCert(cred.Certificate)
default:
    return errors.New("unsupported credential type")
}
// undefined: myapi.LoginRequest_OauthToken
// undefined: myapi.LoginRequest_Certificate
```

**Серверная логика ломается:**
```go
func authenticateUser(req *LoginRequest) error {
    switch cred := req.Credentials.(type) {
    case *LoginRequest_Password:
        return validatePassword(cred.Password)
    case *LoginRequest_ApiKey:
        return validateApiKey(cred.ApiKey)
    case *LoginRequest_OauthToken:    // ERROR после удаления
        return validateOAuth(cred.OauthToken)
    case *LoginRequest_Certificate:   // ERROR после удаления
        return validateCertificate(cred.Certificate)
    default:
        return errors.New("no valid credentials provided")
    }
}
```

**Старые данные становятся нечитаемыми:**
```json
{
  "username": "user123",
  "oauth_token": "abc123xyz"
}
// После удаления: токен теряется или вызывает ошибку парсинга
```

## Стратегия миграции

1. **Депрецируйте сначала:**
   ```proto
   oneof credentials {
     string password = 2;
     string oauth_token = 4 [deprecated = true];
   }
   ```
2. **Добавьте новые структурированные поля:**
   ```proto
   oneof credentials {
     string password = 2;
     string oauth_token = 4 [deprecated = true];
     OAuthCredentials oauth_v2 = 6;
   }
   ```
3. **Сервер: поддерживайте оба варианта:**
   ```go
   switch cred := req.Credentials.(type) {
   case *LoginRequest_OauthV2:
       return validateOAuthV2(cred.OauthV2)
   case *LoginRequest_OauthToken:
       log.Warn("Deprecated oauth_token used")
       return validateOAuthLegacy(cred.OauthToken)
   }
   ```
4. **Мигрируйте клиентов на новое поле:**
   ```go
   loginReq.Credentials = &LoginRequest_OauthV2{
       OauthV2: &OAuthCredentials{
           Token: "token123",
           RefreshToken: "refresh456",
           ExpiresAt: timestamp,
       },
   }
   ```
5. **Зарезервируйте номера удалённых полей в следующей major‑версии:**
   ```proto
   oneof credentials {
     reserved 4, 5;
     reserved "oauth_token", "certificate";
     string password = 2;
     string api_key = 3;
     OAuthCredentials oauth_v2 = 6;
   }
   ```

## Типовые сценарии

### Удаление неподдерживаемых методов аутентификации
```proto
message AuthRequest {
  oneof method {
    PasswordAuth password = 1;
    ApiKeyAuth api_key = 2;
    // LdapAuth ldap = 3;  // НЕЛЬЗЯ просто удалить
    LdapAuth ldap = 3 [deprecated = true];
    OidcAuth oidc = 4;
  }
}
```

### Упрощение вариантов oneof
```proto
message SearchRequest {
  oneof query_type {
    string simple_text = 1;
    // ComplexQuery complex = 2; // Не удалять напрямую
    ComplexQuery complex = 2 [deprecated = true];
    AdvancedQuery advanced = 3;
  }
}
```

### Изменения бизнес‑логики платежей
```proto
message PaymentRequest {
  oneof payment_method {
    CreditCard credit_card = 1;
    BankTransfer bank_transfer = 3;
    // Check check = 2; // Не удалять
    Check check = 2 [deprecated = true];
    DigitalWallet digital_wallet = 4;
  }
}
```

## OneOf Field Deletion vs обычное удаление

### Отличия от удаления обычного поля
```proto
message User {
  string name = 1;
  // string email = 2; // BREAKING: обычное поле
}

message User {
  oneof contact {
    string email = 1;
    // string phone = 2; // BREAKING: поле в oneof
  }
}
```

### Влияние на сгенерированный код
```go
// Обычное поле: удаляется простой геттер
// user.GetEmail()

// Поле в oneof: пропадает тип-обёртка, кейсы switch ломаются
// *LoginRequest_OauthToken
```

### Wire‑особенности
```proto
message Request {
  oneof data {
    string text = 1;
    bytes binary = 2;  // Если удалить — клиенты, использующие binary, ломаются
    JsonData json = 3;
  }
}
```

## Валидация и обработка ошибок

### До удаления (типобезопасно)
```go
switch data := req.Data.(type) {
case *Request_Text:
    return validateText(data.Text)
case *Request_Binary:
    return validateBinary(data.Binary)
case *Request_Json:
    return validateJson(data.Json)
case nil:
    return errors.New("no data provided")
default:
    return errors.New("unknown data type")
}
```

### После удаления (сломано)
```go
switch data := req.Data.(type) {
case *Request_Text:
    return validateText(data.Text)
// case *Request_Binary: // Тип исчез — кейс не компилируется
case *Request_Json:
    return validateJson(data.Json)
case nil:
    return errors.New("no data provided")
default:
    return errors.New("unknown data type") // Старые бинарные данные попадают сюда
}
```
