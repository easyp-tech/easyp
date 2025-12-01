<!-- TODO: Review translation -->

# ONEOF_FIELD_SAME_TYPE

Категории:

- **WIRE+**

Это правило проверяет, что тип поля внутри `oneof` НЕ меняется. Смена типа любого варианта `oneof` — несовместимое (breaking) изменение: существующие данные сериализованы с прежним типом, а клиентский код ожидает конкретные типы-обёртки и методы доступа, сгенерированные для каждого варианта.

## Примеры

### Плохой (Breaking)

**До:**
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

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    int32 category = 2;        // [!code --] Было string -> стало int32 (BREAKING)
    string user_id = 3;        // [!code --] Было int32 -> стало string (BREAKING)
    string is_premium = 4;     // [!code --] Было bool -> стало string (BREAKING)
    string date_range = 5;     // [!code --] Было DateRange -> стало string (BREAKING)
  }
}

message NotificationSettings {
  oneof delivery {
    EmailConfig email = 1;     // [!code --] Было string -> стало EmailConfig (BREAKING)
    string phone_number = 2;   // [!code --] Было int64 -> стало string (BREAKING)
    string webhook = 3;        // [!code --] Было WebhookConfig -> стало string (BREAKING)
  }
}
```

**Ошибка:**
```
search.proto:6:5: Field "2" with name "category" on OneOf "filter" changed type from "string" to "int32". (BREAKING_CHECK)
search.proto:7:5: Field "3" with name "user_id" on OneOf "filter" changed type from "int32" to "string". (BREAKING_CHECK)
search.proto:8:5: Field "4" with name "is_premium" on OneOf "filter" changed type from "bool" to "string". (BREAKING_CHECK)
search.proto:9:5: Field "5" with name "date_range" on OneOf "filter" changed type from "DateRange" to "string". (BREAKING_CHECK)
```

### Дополнительные примеры

**Комплексные изменения типов:**
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

// After - ВСЁ BREAKING
message PaymentRequest {
  oneof payment_info {
    string credit_card = 1;           // CreditCardInfo -> string
    PayPalInfo bank_account = 2;      // BankAccountInfo -> PayPalInfo
    PayPalInfo paypal_email = 3;      // string -> PayPalInfo
    string crypto = 4;                // CryptoWalletInfo -> string
  }
}
```

### Хороший (Безопасный)

**Вместо смены типа — добавьте новое поле в тот же oneof:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    string category = 2 [deprecated = true];         // Старый тип
    int32 user_id = 3 [deprecated = true];
    bool is_premium = 4 [deprecated = true];
    DateRange date_range = 5 [deprecated = true];
    
    int32 category_id = 6;        // Новая версия
    string user_identifier = 7;
    string premium_status = 8;
    string date_filter = 9;
  }
}
```

**Или создайте новый oneof:**
```proto
syntax = "proto3";

package myapi.v1;

message SearchRequest {
  string query = 1;
  
  oneof filter {
    option deprecated = true;
    string category = 2 [deprecated = true];
    int32 user_id = 3 [deprecated = true];
    bool is_premium = 4 [deprecated = true];
    DateRange date_range = 5 [deprecated = true];
  }
  
  oneof filter_v2 {
    int32 category_id = 6;
    string user_identifier = 7;
    PremiumFilter premium_filter = 8;
    TimeFilter time_filter = 9;
  }
}
```

## Влияние

- **Wire Format:** Старые бинарные данные не читаются корректно
- **Generated Code:** Меняются типы-обёртки, код клиента не компилируется
- **Type Safety:** `switch` по варианту oneof ломается
- **Runtime Errors:** Ошибки десериализации и логики при несовпадении типов

## Реальный пример

**Ломается клиентский код:**
```go
// Before
searchReq := &myapi.SearchRequest{
    Query: "golang jobs",
}

// Устанавливаем вариант oneof
searchReq.Filter = &myapi.SearchRequest_Category{
    Category: "engineering",
}

// Обработка вариантов
switch filter := searchReq.Filter.(type) {
case *myapi.SearchRequest_Category:
    return searchByCategory(filter.Category)
case *myapi.SearchRequest_UserId:
    return searchByUserId(filter.UserId)
default:
    return errors.New("unsupported filter")
}

// After — тип Category стал int32
searchReq.Filter = &myapi.SearchRequest_Category{
    Category: "engineering", // ERROR
}

switch filter := searchReq.Filter.(type) {
case *myapi.SearchRequest_Category:
    return searchByCategory(filter.Category) // ERROR
case *myapi.SearchRequest_UserId:
    return searchByUserId(filter.UserId)     // ERROR: UserId теперь string
}
```

**Ломается серверная логика:**
```go
// Before
func handleSearchFilter(req *SearchRequest) (*SearchResult, error) {
    switch filter := req.Filter.(type) {
    case *SearchRequest_Category:
        return searchDatabase("category", filter.Category), nil
    case *SearchRequest_UserId:
        return searchDatabase("user_id", strconv.Itoa(int(filter.UserId))), nil
    default:
        return nil, errors.New("no filter specified")
    }
}

// After — несовпадения типов
func handleSearchFilter(req *SearchRequest) (*SearchResult, error) {
    switch filter := req.Filter.(type) {
    case *SearchRequest_Category:
        // filter.Category теперь int32, а код ожидал string
        return searchDatabase("category", filter.Category), nil // ERROR
    case *SearchRequest_UserId:
        userIdStr := filter.UserId // string вместо int32
        userId, _ := strconv.Atoi(userIdStr)
        return searchDatabase("user_id", strconv.Itoa(userId)), nil
    }
}
```

**Порча данных:**
```json
{
  "query": "golang jobs",
  "category": "engineering"
}
// После смены типа: "engineering" не может быть интерпретировано как int32
```

## Стратегия миграции

1. Добавьте новое поле:
   ```proto
   oneof filter {
     string category = 2 [deprecated = true];
     int32 category_id = 6;
   }
   ```
2. Сервер поддерживает оба:
   ```go
   switch filter := req.Filter.(type) {
   case *SearchRequest_CategoryId:
       return searchByCategoryId(filter.CategoryId)
   case *SearchRequest_Category:
       log.Warn("Deprecated string category used")
       id := convertCategoryNameToId(filter.Category)
       return searchByCategoryId(id)
   }
   ```
3. Клиенты переходят на `category_id`.
4. Скрипты миграции преобразуют старые данные.
5. В следующей major‑версии:
   ```proto
   oneof filter {
     reserved 2, "category";
     int32 category_id = 6;
   }
   ```

## Типовые сценарии

### Повышение типобезопасности
```proto
message ConfigRequest {
  oneof setting {
    string database_url = 1 [deprecated = true];
    DatabaseConfig database_config = 5;
  }
}
```

### Стандартизация числовых типов
```proto
message FilterRequest {
  oneof criteria {
    int32 timestamp = 1 [deprecated = true];
    int64 timestamp_millis = 5;
  }
}
```

### Переход к структурированным типам
```proto
message PaymentRequest {
  oneof payment {
    string card_number = 1 [deprecated = true];
    CreditCardDetails card_details = 5;
  }
}
```

## Матрица совместимости

| From Type | To Type | Результат |
|-----------|---------|-----------|
| `string` | `int32/int64` | ❌ BREAKING |
| `int32` | `string` | ❌ BREAKING |
| `bool` | `string` | ❌ BREAKING |
| `message` | `string` | ❌ BREAKING |
| `string` | `message` | ❌ BREAKING |
| `enum` | `int32` | ❌ BREAKING |

### Влияние на сгенерированный код
```go
// Before:
type SearchRequest_Category struct {
    Category string
}
// After:
type SearchRequest_Category struct {
    Category int32
}
```

### Wire‑формат
```proto
message Request {
  oneof data {
    string text = 1;
    // int32 text = 1; // Тот же номер, другой тип — BREAKING
  }
}
```

## Валидация

**До:**
```go
switch f := req.Filter.(type) {
case *SearchRequest_Category:
    if f.Category == "" { return errors.New("category cannot be empty") }
case *SearchRequest_UserId:
    if f.UserId <= 0 { return errors.New("user_id must be positive") }
}
```

**После (сломано):**
```go
switch f := req.Filter.(type) {
case *SearchRequest_Category:
    if f.Category <= 0 { return errors.New("category must be positive") } // Логика ожидала string
case *SearchRequest_UserId:
    if f.UserId == "" { return errors.New("user_id cannot be empty") }    // Был int32
}
```
