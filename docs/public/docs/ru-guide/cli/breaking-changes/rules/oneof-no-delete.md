<!-- TODO: Review translation -->

# ONEOF_NO_DELETE

Категории:

- **WIRE+**

Это правило проверяет, что ни один `oneof` блок не был удалён из сообщений. Удаление `oneof` ломает совместимость по wire‑формату и сгенерированный код: существующие данные могут содержать ранее выбранный вариант, а клиентский код зависит от типов-обёрток и методов доступа, генерируемых для `oneof`.

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

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  // credentials oneof удалён — BREAKING CHANGE!
  string password = 2;     // Теперь обычные поля
  string api_key = 3;      // Потеряна взаимная исключаемость
  string oauth_token = 4;
}

message PaymentMethod {
  // method oneof удалён — BREAKING CHANGE!
  CreditCard credit_card = 1;  // Теперь можно установить все поля одновременно
  BankAccount bank_account = 2;
  PayPal paypal = 3;
}

message Order {
  string id = 1;
  PaymentMethod payment = 2;
}
```

**Ошибка:**
```
login.proto:5:3: Previously present oneof "credentials" was deleted. (BREAKING_CHECK)
payment.proto:2:3: Previously present oneof "method" was deleted. (BREAKING_CHECK)
```

### Дополнительные примеры

**Удаление вложенного oneof:**
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
    // contact_method oneof удалён
    string email = 1;    // Больше не эксклюзивно
    string phone = 2;    // Можно задать всё сразу
    string slack_id = 3;
  }
  
  ContactInfo contact = 2;
}
```

### Хороший (Безопасный)

**Вместо удаления — депрецируйте oneof:**
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

**Или заменить новой структурой:**
```proto
syntax = "proto3";

package myapi.v1;

message LoginRequest {
  string username = 1;
  
  oneof credentials {
    option deprecated = true; // [!code focus] // Старый oneof
    string password = 2 [deprecated = true]; // [!code focus]
    string api_key = 3 [deprecated = true]; // [!code focus]
    string oauth_token = 4 [deprecated = true]; // [!code focus]
  }
  
  oneof auth_method { // [!code focus] // Новый дизайн
    PasswordAuth password_auth = 5; // [!code focus]
    ApiKeyAuth api_auth = 6; // [!code focus]
    OAuthAuth oauth_auth = 7; // [!code focus]
  } // [!code focus]
}
```

## Влияние

- **Wire Format:** Теряются семантика взаимной исключаемости — больше одного поля может быть установлено
- **Generated Code:** Удаляются типы-обёртки и accessor методы — код клиента не компилируется
- **Business Logic:** Нарушаются инварианты «только один из»
- **Validation:** Логика проверок должна быть переписана вручную

## Реальный пример

**Код клиента ломается:**
```go
// Before - oneOf даёт взаимную исключаемость
loginReq := &myapi.LoginRequest{
    Username: "user123",
}

loginReq.Credentials = &myapi.LoginRequest_Password{
    Password: "secret123",
}

switch cred := loginReq.Credentials.(type) {
case *myapi.LoginRequest_Password:  // ERROR после удаления
    return authenticatePassword(cred.Password)
case *myapi.LoginRequest_ApiKey:    // ERROR после удаления
    return authenticateApiKey(cred.ApiKey)
default:
    return errors.New("no credentials provided")
}

// Compilation errors:
// undefined: myapi.LoginRequest_Password
// undefined field: loginReq.Credentials
```

**Валидация на сервере ломается:**
```go
// Before
func validateLoginRequest(req *LoginRequest) error {
    if req.Credentials == nil {
        return errors.New("credentials required")
    }
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

// After
func validateLoginRequest(req *LoginRequest) error {
    credCount := 0
    if req.Password != "" { credCount++ }
    if req.ApiKey != "" { credCount++ }
    if req.OauthToken != "" { credCount++ }

    if credCount == 0 {
        return errors.New("no credentials provided")
    }
    if credCount > 1 {
        return errors.New("multiple credentials provided")
    }

    if req.Password != "" {
        return validatePassword(req.Password)
    }
    // Дополнительные проверки вручную...
    return nil
}
```

**Проблемы бизнес‑логики:**
```go
// Before
payment := &PaymentMethod{
    Method: &PaymentMethod_CreditCard{
        CreditCard: &CreditCard{Number: "1234"},
    },
}

// After — можно задать несколько способов сразу
payment := &PaymentMethod{
    CreditCard:  &CreditCard{Number: "1234"},
    BankAccount: &BankAccount{Number: "5678"}, // Ломает бизнес-логику
}
```

## Стратегия миграции

1. **Депрецируйте старый oneof и его поля:**
   ```proto
   oneof old_credentials {
     option deprecated = true;
     string password = 2 [deprecated = true];
     string api_key = 3 [deprecated = true];
   }
   ```

2. **Создайте новый блок при необходимости:**
   ```proto
   oneof new_auth_method {
     PasswordAuth password_auth = 5;
     ApiKeyAuth api_auth = 6;
   }
   ```

3. **Сервер: поддерживайте оба до миграции:**
   ```go
   func handleAuth(req *LoginRequest) error {
       if req.NewAuthMethod != nil {
           return handleNewAuth(req.NewAuthMethod)
       }
       if req.Credentials != nil {
           return handleOldAuth(req.Credentials)
       }
       return errors.New("no auth method provided")
   }
   ```

4. **Клиенты: переходите на новый oneof:**
   ```go
   loginReq.NewAuthMethod = &LoginRequest_PasswordAuth{
       PasswordAuth: &PasswordAuth{
           Username: "user123",
           Password: "secret123",
       },
   }
   ```

5. **Удалите старый oneof** в следующей major‑версии.

## Типовые сценарии

### Замена oneof обычными полями
```proto
message SearchRequest {
  // НЕЛЬЗЯ просто убрать oneof — ломает семантику
  string text_query = 1;    // BREAKING
  int32 user_id = 2;        // BREAKING
  string category = 3;      // BREAKING
}

message SearchRequest {
  oneof query {
    option deprecated = true;
    string text_query = 1 [deprecated = true];
    int32 user_id = 2 [deprecated = true];
    string category = 3 [deprecated = true];
  }
  string search_text = 4;
  repeated string filters = 5;
}
```

### Упрощение структуры сообщения
```proto
message NotificationSettings {
  string email = 1;       // BREAKING: был в oneof
  string phone = 2;       // BREAKING
  string webhook_url = 3; // BREAKING
}

message NotificationSettings {
  oneof delivery_method {
    EmailNotification email = 1;
    SmsNotification phone = 2;
    WebhookNotification webhook = 3;
  }
}
```

### Миграция с proto2 на proto3
```proto
// proto2
message Request {
  optional string name = 1;
  oneof data {
    string text = 2;
    int32 number = 3;
  }
}

// proto3 — сохраняем oneof
message Request {
  string name = 1;
  oneof data {
    string text = 2;
    int32 number = 3;
  }
}
```

## Семантика OneOf

### Что даёт oneof
- Взаимная исключаемость
- Типобезопасность в коде
- Ясность намерения API («выбери один»)
- Эффективная сериализация — кодируется только выбранный вариант

### Что теряется при удалении
- Инвариант «только один»
- Генерируемые wrapper-типы и методы
- Автоматическая валидация выбора
- Понятность контракта API

### Влияние на код генерации
```go
// С oneof
type LoginRequest struct {
    Username string
    Credentials isLoginRequest_Credentials
}

// Без oneof
type LoginRequest struct {
    Username   string
    Password   string
    ApiKey     string
    OauthToken string
    // Возможны многозначные конфликтующие состояния
}
```

