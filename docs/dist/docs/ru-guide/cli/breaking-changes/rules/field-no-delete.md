<!-- TODO: Review translation -->

# FIELD_NO_DELETE

Категории:

- **WIRE+**

Это правило проверяет, что ни одно поле сообщения (message field) не было удалено. Удаление поля ломает совместимость по wire‑формату и сгенерированный код: в существующих данных может присутствовать удалённое поле, а клиентский код может ссылаться на него.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  int32 age = 3;
  string phone = 4; // [!code --]
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  int32 age = 3;
  // phone поле удалено — BREAKING CHANGE!
}
```

**Ошибка:**
```
user.proto:6:3: Previously present field "4" with name "phone" on message "User" was deleted. (BREAKING_CHECK)
```

### Хороший (Безопасный)

**Вместо удаления — пометить поле deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  int32 age = 3;
  string phone = 4 [deprecated = true]; // [!code focus]
}
```

**Или зарезервировать номер и имя поля после удаления:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  reserved 4; // [!code focus]
  reserved "phone"; // [!code focus]
  
  string name = 1;
  string email = 2;
  int32 age = 3;
}
```

## Влияние

- **Wire Format:** Старые сообщения с удалённым полем некорректно десериализуются
- **Generated Code:** Удаляются аксессоры (get/set), компиляция клиента ломается
- **Data Loss:** Потеря сохранённых данных, связанных с полем
- **JSON Compatibility:** JSON парсеры ожидают наличие поля и могут выдавать ошибки

## Стратегия миграции

1. **Сначала пометить deprecated:**
   ```proto
   string old_field = 5 [deprecated = true];
   ```

2. **Прекратить запись в поле** в прикладном коде

3. **Зарезервировать поле** в следующей версии, чтобы избежать повторного использования:
   ```proto
   reserved 5;
   reserved "old_field";
   ```

4. **Никогда не переиспользовать номер поля** — должен оставаться в reserved навсегда