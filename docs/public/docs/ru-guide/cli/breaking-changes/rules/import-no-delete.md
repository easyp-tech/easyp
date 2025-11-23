<!-- TODO: Review translation -->

# IMPORT_NO_DELETE

Категории:

- **WIRE+**

Это правило проверяет, что ни одна инструкция `import` не была удалена из proto‑файла. Удаление import ломает совместимость по wire‑формату и сгенерированный код: типы из импортируемых файлов могли использоваться в текущем файле, и их удаление делает эти типы недоступными.

## Примеры

### Плохой (Breaking)

**До:**
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

**После:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
// import "google/protobuf/duration.proto";  // [!code --] Удалён import — BREAKING
// import "common/user.proto";               // [!code --] Удалён import — BREAKING
// import "common/address.proto";            // [!code --] Удалён import — BREAKING

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Duration processing_time = 3;  // ERROR: Duration недоступен
  common.User customer = 4;                      // ERROR: User недоступен
  common.Address shipping_address = 5;           // ERROR: Address недоступен
}
```

**Ошибка:**
```
order.proto:5:1: Previously import "google/protobuf/duration.proto" was deleted. (BREAKING_CHECK)
order.proto:6:1: Previously import "common/user.proto" was deleted. (BREAKING_CHECK)
order.proto:7:1: Previously import "common/address.proto" was deleted. (BREAKING_CHECK)
```

### Дополнительные примеры

**Сервис ломается при удалённых import:**
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
    option (google.api.http) = {  // ERROR: annotations недоступен
      get: "/v1/orders/{id}"
    };
  }
  
  rpc DeleteOrder(common.AuthRequest) returns (google.protobuf.Empty);  // ERROR: типы недоступны
}
```

### Хороший (Безопасный)

**Вместо удаления оставьте неиспользуемые import:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";  // [!code focus] // Оставляем даже если не используется
import "common/user.proto";               // [!code focus]
import "common/address.proto";            // [!code focus]

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  // Поля, использовавшие Duration/User/Address, удалены — import оставлены
}
```

**Или добавьте новые типы, сохранив старые import:**
```proto
syntax = "proto3";

package myapi.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";  // [!code focus]
import "common/user.proto";               // [!code focus]
import "common/address.proto";            // [!code focus]
import "common/v2/user.proto";            // [!code focus] // Новый import

message Order {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Duration processing_time = 3 [deprecated = true]; // [!code focus]
  common.User customer = 4 [deprecated = true];                     // [!code focus]
  common.Address shipping_address = 5 [deprecated = true];          // [!code focus]
  
  int32 processing_seconds = 6;
  common.v2.UserProfile customer_v2 = 7;
  string shipping_address_text = 8;
}
```

## Влияние

- **Generated Code:** Типы пропадают → ошибка компиляции
- **Wire Format:** Сообщения с удалёнными типами не десериализуются
- **Field References:** Поля с этими типами становятся невалидны
- **Service Definitions:** RPC с типами из удалённых файлов ломаются
- **Options Usage:** Пользовательские опции из импортов недоступны

## Реальный пример

**Код клиента ломается:**
```go
order := &myapi.Order{
    Id:        "order123",
    CreatedAt: timestamppb.New(time.Now()),
    Customer: &common.User{          // ERROR после удаления import
        Id:   "user456",
        Name: "John Doe",
    },
    ShippingAddress: &common.Address{ // ERROR после удаления import
        Street: "123 Main St",
        City:   "New York",
    },
}
// undefined: common.User
// undefined: common.Address
```

**Серверная логика ломается:**
```go
func (s *server) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    return &Order{
        Id:        generateOrderId(),
        CreatedAt: timestamppb.New(time.Now()),
        Customer: &common.User{          // ERROR
            Id:   req.CustomerId,
            Name: req.CustomerName,
        },
        ShippingAddress: &common.Address{ // ERROR
            Street: req.ShippingStreet,
            City:   req.ShippingCity,
        },
    }, nil
}
```

**Компиляция proto ломается:**
```bash
$ protoc --go_out=. order.proto
order.proto:8:3: "common.User" is not defined.
order.proto:9:3: "common.Address" is not defined.
order.proto:10:3: "google.protobuf.Duration" is not defined.
```

## Стратегия миграции

1. Сохраняйте import даже если типы более не используются:
   ```proto
   import "common/user.proto"; // Оставить
   ```
2. Добавляйте новые import рядом со старыми:
   ```proto
   import "common/user.proto";
   import "common/v2/user.proto";
   ```
3. Депрецируйте поля со старыми типами постепенно:
   ```proto
   message Order {
     common.User customer = 4 [deprecated = true];
     common.v2.UserProfile customer_v2 = 5;
   }
   ```
4. Сервер: поддерживайте оба варианта:
   ```go
   if req.CustomerProfileV2 != nil { order.CustomerV2 = req.CustomerProfileV2; return order }
   if req.CustomerProfile != nil { order.Customer = req.CustomerProfile }
   ```
5. Удаляйте import только в следующей major‑версии после полной миграции.

## Типовые сценарии

### Замена зависимости
```proto
import "old/user.proto";    // Старый
import "new/user.proto";    // Новый

message Order {
  old.User customer = 4 [deprecated = true];
  new.UserProfile customer_v2 = 5;
}
```

### Удаление функционала
```proto
import "feature/analytics.proto"; // Оставить даже если поле удалено

message Order {
  string id = 1;
  // feature.AnalyticsData analytics = 2; // Поле удалено, import остаётся
}
```

### Миграция версии protobuf
```proto
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";

message Event {
  google.protobuf.Timestamp timestamp = 1;
  int64 timestamp_millis = 2;
}
```

### Упрощение сложных зависимостей
```proto
import "complex/config.proto";

message Settings {
  complex.ConfigData config = 1 [deprecated = true];
  string config_json = 2;
}
```

## Категории import и влияние

### Стандартные Google imports
```proto
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/empty.proto";
import "google/api/annotations.proto";
```

### Общие библиотечные imports
```proto
import "common/user.proto";
import "common/address.proto";
import "common/money.proto";
```

### Фичевые imports
```proto
import "features/analytics.proto";
import "features/logging.proto";
import "internal/debug.proto";
```

## Особенности wire‑формата

### Удаление import vs использование поля
```proto
import "common/enums.proto";
message Order {
  oneof payment {
    CreditCardPayment card = 1;
  }
}
```

### Транзитивные зависимости
```proto
import "user.proto";
message Order {
  PaymentInfo payment = 1;
}
```

### JSON сериализация
```proto
import "google/protobuf/field_mask.proto";
message UpdateRequest {
  google.protobuf.FieldMask update_mask = 1;
}
```

## Профилактика

### Аудит import
```bash
grep -r "google.protobuf.Duration" *.proto
grep -r "common.User" *.proto
```
(Даже если не найдено — не удаляем.)

### Настройка линтера
```yaml
lint:
  ignore:
    - IMPORT_UNUSED
```

### Документация
```proto
import "legacy/types.proto";  // Для обратной совместимости
import "common/user.proto";   // Требуется для уже сохранённых данных
```

## Import vs удаление поля

### Разные типы breaking
```proto
// import "user.proto"; // BREAKING: удалён import
message Order {
  string id = 1;
  // User customer = 2; // BREAKING: удалено поле
}
```

### Составные нарушения
```proto
// import "user.proto"; // BREAKING
message Order {
  string id = 1;
  // User customer = 2; // BREAKING
}
```
