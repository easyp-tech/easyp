<!-- TODO: Review translation -->

# ENUM_NO_DELETE

Категории:

- **WIRE+**

Это правило проверяет, что ни один enum не был удалён из proto‑файлов. Удаление enum ломает совместимость по wire‑формату и сгенерированный код: сохранённые данные могут содержать удалённые значения, а клиентский код опирается на сгенерированные типы и константы enum.

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
}

enum UserRole { // [!code --]
  USER_ROLE_UNSPECIFIED = 0; // [!code --]
  USER_ROLE_ADMIN = 1; // [!code --]
  USER_ROLE_USER = 2; // [!code --]
} // [!code --]

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3;
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
}

// UserRole enum удалён — BREAKING CHANGE!

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3; // ERROR: UserRole больше не существует
}
```

**Ошибка:**
```
user.proto:8:1: Previously present enum "UserRole" was deleted from file. (BREAKING_CHECK)
```

### Дополнительные примеры

**Удаление вложенного enum:**
```proto
// Before
message User {
  string name = 1;
  Status status = 2;
  
  enum Status { // [!code --]
    STATUS_UNSPECIFIED = 0; // [!code --]
    STATUS_ACTIVE = 1; // [!code --]
    STATUS_INACTIVE = 2; // [!code --]
  } // [!code --]
}

// After - BREAKING CHANGE!
message User {
  string name = 1;
  Status status = 2;  // ERROR: Status удалён
  
  // Вложенный enum Status удалён
}
```

### Хороший (Безопасный)

**Вместо удаления — пометить enum deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum UserRole {
  option deprecated = true; // [!code focus]
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
  USER_ROLE_USER = 2;
}

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3 [deprecated = true]; // [!code focus]
}
```

**Или заменить новым enum:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
}

enum UserRole {
  option deprecated = true; // [!code focus] // Старый enum
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
  USER_ROLE_USER = 2;
}

enum UserPermission { // [!code focus] // Новый enum
  USER_PERMISSION_UNSPECIFIED = 0; // [!code focus]
  USER_PERMISSION_READ = 1; // [!code focus]
  USER_PERMISSION_WRITE = 2; // [!code focus]
  USER_PERMISSION_ADMIN = 3; // [!code focus]
} // [!code focus]

message Order {
  string id = 1;
  OrderStatus status = 2;
  UserRole created_by_role = 3 [deprecated = true]; // [!code focus]
  UserPermission created_by_permission = 4; // [!code focus]
}
```

## Влияние

- **Wire Format:** Данные с удалённым enum не десериализуются корректно
- **Generated Code:** Тип и константы удаляются — компиляция клиентов ломается
- **Field References:** Поля с типом удалённого enum становятся невалидны
- **Switch Statements:** `switch` по enum ломается (неизвестные кейсы)

## Реальный пример

**Код клиента ломается:**
```go
// Before
order := &myapi.Order{
    Id:            "order123",
    Status:        myapi.OrderStatus_ORDER_STATUS_PENDING,
    CreatedByRole: myapi.UserRole_USER_ROLE_ADMIN,  // ERROR после удаления
}

switch order.CreatedByRole {
case myapi.UserRole_USER_ROLE_ADMIN:  // ERROR
    // Admin logic
case myapi.UserRole_USER_ROLE_USER:   // ERROR
    // User logic
}

// undefined: myapi.UserRole
// undefined: myapi.UserRole_USER_ROLE_ADMIN
```

**Старые данные становятся нечитаемыми:**
```json
{
  "id": "order123",
  "status": "ORDER_STATUS_PENDING",
  "created_by_role": "USER_ROLE_ADMIN"
}
// После удаления UserRole — парсер не знает поле created_by_role
```

**Серверная логика ломается:**
```go
func validateOrder(order *Order) error {
    switch order.CreatedByRole {
    case UserRole_USER_ROLE_ADMIN:  // ERROR
        return nil
    case UserRole_USER_ROLE_USER:   // ERROR
        return validateUserOrder(order)
    default:
        return errors.New("invalid user role")
    }
}
```

## Стратегия миграции

1. **Депрецируйте enum:**
   ```proto
   enum OldEnum {
     option deprecated = true;
     // ...
   }
   ```
2. **Пометьте поля со старым enum deprecated:**
   ```proto
   OldEnum old_field = 5 [deprecated = true];
   ```
3. **Создайте новый enum:**
   ```proto
   enum NewEnum {
     // Улучшенный набор значений
   }
   NewEnum new_field = 6;
   ```
4. **Сервер обрабатывает оба:**
   ```go
   func handleRole(oldRole OldEnum, newRole NewEnum) {
       if newRole != NEW_ENUM_UNSPECIFIED {
           return handleNewRole(newRole)
       }
       return handleOldRole(oldRole)
   }
   ```
5. **Удалите старый enum в следующей major‑версии.**

## Типовые сценарии

### Редизайн enum
```proto
enum Priority {
  option deprecated = true;
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_HIGH = 2;
}

enum TaskPriority {
  TASK_PRIORITY_UNSPECIFIED = 0;
  TASK_PRIORITY_LOW = 1;
  TASK_PRIORITY_MEDIUM = 2;
  TASK_PRIORITY_HIGH = 3;
  TASK_PRIORITY_URGENT = 4;
}
```

### Консолидация enum
```proto
enum Status {
  option deprecated = true;
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}

enum State {
  option deprecated = true;
  STATE_UNSPECIFIED = 0;
  STATE_ENABLED = 1;
}

enum EntityStatus {
  ENTITY_STATUS_UNSPECIFIED = 0;
  ENTITY_STATUS_ACTIVE = 1;
  ENTITY_STATUS_INACTIVE = 2;
  ENTITY_STATUS_ENABLED = 3;
  ENTITY_STATUS_DISABLED = 4;
}
```

### Перемещение enum в новый пакет
```proto
enum UserRole {
  option deprecated = true;
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
}

import "common/roles.proto";

message Order {
  UserRole old_role = 3 [deprecated = true];
  common.UserRole new_role = 4;
}
```

### Версионирование / новая версия пакета
```proto
package myapi.v2;

enum UserPermission {
  USER_PERMISSION_UNSPECIFIED = 0;
  USER_PERMISSION_READ = 1;
  USER_PERMISSION_WRITE = 2;
  USER_PERMISSION_ADMIN = 3;
}

message Order {
  string id = 1;
  UserPermission created_by_permission = 2;
}
```
