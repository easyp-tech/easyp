<!-- TODO: Review translation -->

# ENUM_VALUE_NO_DELETE

Категории:

- **WIRE+**

Правило проверяет, что ни одно значение enum (enum value) не было удалено. Удаление значения ломает совместимость по wire‑формату и сгенерированный код: в уже сохранённых данных могут присутствовать удалённые значения, а клиентский код может ссылаться на соответствующие константы.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
  ORDER_STATUS_CANCELLED = 5;
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  PRIORITY_URGENT = 4;
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
  // ORDER_STATUS_CANCELLED = 5; // [!code --] Удалено значение enum — BREAKING!
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  // PRIORITY_URGENT = 4; // [!code --] Удалено значение enum — BREAKING!
}
```

**Ошибка:**
```
order.proto:9:3: Previously present enum value "5" on enum "OrderStatus" was deleted. (BREAKING_CHECK)
priority.proto:8:3: Previously present enum value "4" on enum "Priority" was deleted. (BREAKING_CHECK)
```

### Дополнительные примеры

**Множественное удаление значений:**
```proto
// Before
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_GUEST = 1;
  USER_ROLE_USER = 2;
  USER_ROLE_MODERATOR = 3;
  USER_ROLE_ADMIN = 4;
  USER_ROLE_SUPERADMIN = 5;
}

// After - BREAKING CHANGES!
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_USER = 2;
  USER_ROLE_ADMIN = 4;
  // USER_ROLE_GUEST = 1;      // BREAKING: deleted
  // USER_ROLE_MODERATOR = 3;  // BREAKING: deleted
  // USER_ROLE_SUPERADMIN = 5; // BREAKING: deleted
}
```

### Хороший (Безопасный)

**Вместо удаления — пометить значения deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
  ORDER_STATUS_CANCELLED = 5 [deprecated = true]; // [!code focus]
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  PRIORITY_URGENT = 4 [deprecated = true]; // [!code focus]
}
```

**Или зарезервировать номер и имя после удаления:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  reserved 5; // [!code focus]
  reserved "ORDER_STATUS_CANCELLED"; // [!code focus]

  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum Priority {
  reserved 4; // [!code focus]
  reserved "PRIORITY_URGENT"; // [!code focus]

  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
}
```

## Влияние

- **Wire Format:** Данные с удалёнными значениями не десериализуются корректно
- **Generated Code:** Константы исчезают — клиентский код не компилируется
- **Client Applications:** Ссылки на удалённые значения вызывают ошибки компиляции
- **Switch Statements:** `switch` по enum ломается (неизвестные case)
- **Default Handling:** Поведение при неизвестных значениях меняется

## Реальный пример

**Код клиента ломается:**
```go
order := &myapi.Order{
    Status: myapi.OrderStatus_ORDER_STATUS_CANCELLED, // ERROR после удаления
}

switch order.Status {
case myapi.OrderStatus_ORDER_STATUS_PENDING:
    // Pending
case myapi.OrderStatus_ORDER_STATUS_CANCELLED: // ERROR: undefined
    // Cancellation
default:
    // Unknown
}
// undefined: myapi.OrderStatus_ORDER_STATUS_CANCELLED
```

**Проблемы с уже сохранёнными данными:**
```json
{
  "id": "order123",
  "status": "ORDER_STATUS_CANCELLED"
}
// После удаления: парсер даёт UNSPECIFIED (0) или ошибку
```

**Валидация на сервере:**
```go
func validateOrderStatus(status OrderStatus) error {
    switch status {
    case OrderStatus_ORDER_STATUS_PENDING,
         OrderStatus_ORDER_STATUS_CONFIRMED,
         OrderStatus_ORDER_STATUS_SHIPPED,
         OrderStatus_ORDER_STATUS_DELIVERED,
         OrderStatus_ORDER_STATUS_CANCELLED: // ERROR после удаления
        return nil
    default:
        return errors.New("invalid order status")
    }
}

func canModifyOrder(order *Order) bool {
    return order.Status != OrderStatus_ORDER_STATUS_CANCELLED // ERROR
}
```

## Стратегия миграции

1. **Сначала пометить deprecated:**
   ```proto
   ORDER_STATUS_CANCELLED = 5 [deprecated = true];
   ```
2. **Перестать использовать deprecated значение в новом коде:**
   ```go
   order.Status = OrderStatus_ORDER_STATUS_DELIVERED
   ```
3. **Обработка legacy в клиенте:**
   ```go
   switch order.Status {
   case OrderStatus_ORDER_STATUS_CANCELLED:
       log.Warn("Legacy cancelled status")
       // Дополнительная логика
   }
   ```
4. **Резервирование после миграции:**
   ```proto
   reserved 5, "ORDER_STATUS_CANCELLED";
   ```
5. **Никогда не переиспользовать номера** — зарезервированы навсегда.

## Типовые сценарии

### Изменения бизнес‑логики
```proto
enum ProductStatus {
  PRODUCT_STATUS_UNSPECIFIED = 0;
  PRODUCT_STATUS_ACTIVE = 1;
  PRODUCT_STATUS_INACTIVE = 2;
  PRODUCT_STATUS_DISCONTINUED = 3 [deprecated = true];
}
```

### Упрощение workflow
```proto
enum TaskStatus {
  TASK_STATUS_UNSPECIFIED = 0;
  TASK_STATUS_TODO = 1;
  TASK_STATUS_IN_PROGRESS = 2;
  TASK_STATUS_IN_REVIEW = 3 [deprecated = true];
  TASK_STATUS_APPROVED = 4 [deprecated = true];
  TASK_STATUS_DONE = 5;
}
```

### Объединение enum
```proto
enum NotificationLevel {
  NOTIFICATION_LEVEL_UNSPECIFIED = 0;
  NOTIFICATION_LEVEL_INFO = 1;
  NOTIFICATION_LEVEL_WARNING = 2;
  NOTIFICATION_LEVEL_ERROR = 3;
  NOTIFICATION_LEVEL_DEBUG = 4 [deprecated = true];
  NOTIFICATION_LEVEL_TRACE = 5 [deprecated = true];
}
```

## Особенности wire‑формата

### Предотвращение переиспользования номера
```proto
enum Status {
  reserved 2;
  reserved "STATUS_DELETED";

  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 3;
  // STATUS_DELETED = 2; // Не переиспользуем
}
```

### JSON совместимость
```proto
enum Color {
  reserved 2;
  reserved "RED";

  COLOR_UNSPECIFIED = 0;
  COLOR_BLUE = 1;
  COLOR_GREEN = 3;
  // COLOR_RED = 2; // Был удалён
}
```

### Proto2 vs Proto3
```proto
enum MyEnum {
  MY_ENUM_UNSPECIFIED = 0;
  MY_ENUM_VALUE_OLD = 1 [deprecated = true];
  MY_ENUM_VALUE_NEW = 2;
}
```
