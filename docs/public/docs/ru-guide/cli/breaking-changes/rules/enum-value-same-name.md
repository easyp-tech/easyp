<!-- TODO: Review translation -->

# ENUM_VALUE_SAME_NAME

Категории:

- **WIRE+**

Это правило проверяет, что имя enum значения для каждого номера (number) не изменилось. Переименование enum значения при сохранении его номера ломает совместимость JSON и сгенерированный код: клиенты ожидают конкретные имена констант.

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
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_APPROVED = 2; // [!code --] Было ORDER_STATUS_CONFIRMED
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_COMPLETED = 4; // [!code --] Было ORDER_STATUS_DELIVERED
}
```

**Ошибка:**
```
order.proto:7:3: Enum value "2" on enum "OrderStatus" changed name from "ORDER_STATUS_CONFIRMED" to "ORDER_STATUS_APPROVED". (BREAKING_CHECK)
order.proto:9:3: Enum value "4" on enum "OrderStatus" changed name from "ORDER_STATUS_DELIVERED" to "ORDER_STATUS_COMPLETED". (BREAKING_CHECK)
```

### Дополнительные примеры

**Множественные переименования:**
```proto
// Before
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMIN = 1;
  USER_ROLE_MODERATOR = 2;
  USER_ROLE_USER = 3;
}

// After - ВСЁ BREAKING
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_ADMINISTRATOR = 1;  // BREAKING: ADMIN -> ADMINISTRATOR
  USER_ROLE_MOD = 2;            // BREAKING: MODERATOR -> MOD
  USER_ROLE_MEMBER = 3;         // BREAKING: USER -> MEMBER
}
```

### Хороший (Безопасный)

**Вместо переименования — добавить новые значения и пометить старые deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2 [deprecated = true]; // [!code focus]
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4 [deprecated = true]; // [!code focus]
  ORDER_STATUS_APPROVED = 5; // [!code focus] // Новое значение вместо переименования
  ORDER_STATUS_COMPLETED = 6; // [!code focus]
}
```

**Или создать новую версию enum:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus]

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_APPROVED = 2; // [!code focus] // «Чистые» имена в v2
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_COMPLETED = 4; // [!code focus]
}
```

## Влияние

- **Generated Code:** Меняются имена констант — компиляция клиентов ломается
- **JSON Compatibility:** JSON использует имена enum — старые документы не парсятся
- **Client Applications:** Код со старым именем константы не компилируется
- **Documentation:** Документация устаревает с неправильными именами

## Реальный пример

**Ломается код клиента:**
```go
// Before
status := myapi.OrderStatus_ORDER_STATUS_CONFIRMED
if order.Status == myapi.OrderStatus_ORDER_STATUS_DELIVERED {
    // handle delivery
}

// After (переименовано) — ошибки компиляции
status := myapi.OrderStatus_ORDER_STATUS_CONFIRMED  // ERROR: undefined
if order.Status == myapi.OrderStatus_ORDER_STATUS_DELIVERED { // ERROR: undefined
    // handle delivery
}
```

**JSON совместимость ломается:**
```json
// Раньше
{
  "status": "ORDER_STATUS_CONFIRMED"
}

// После
{
  "status": "ORDER_STATUS_APPROVED" // Старые JSON больше не парсятся
}
```

## Стратегия миграции

1. **Добавьте новое значение вместо переименования:**
   ```proto
   ORDER_STATUS_CONFIRMED = 2 [deprecated = true];
   ORDER_STATUS_APPROVED = 5;
   ```

2. **Сервер обрабатывает оба значения на переходном этапе**
3. **Мигрируйте клиентов** на новые значения
4. **Удалите (reserve) старое имя и номер** в следующей major‑версии:
   ```proto
   reserved 2, "ORDER_STATUS_CONFIRMED";
   ORDER_STATUS_APPROVED = 5;
   ```

## Allow Alias (исключение)

С `allow_alias = true` можно временно иметь два имени для одного номера:

```proto
enum Status {
  option allow_alias = true;
  STATUS_UNSPECIFIED = 0;
  STATUS_OLD_NAME = 1 [deprecated = true];
  STATUS_NEW_NAME = 1; // Один номер, два имени
}
```

Примечание: Сейчас EasyP всё равно считает это breaking (может измениться в будущем).
