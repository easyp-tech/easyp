<!-- TODO: Review translation -->

# MESSAGE_NO_DELETE

Категории:

- **WIRE+**

Это правило проверяет, что ни одно сообщение (message) не было удалено из proto‑файлов. Удаление message ломает совместимость по wire‑формату и сгенерированный код: в существующих данных могут присутствовать экземпляры удалённого типа, а клиентский код зависит от сгенерированных структур и классов.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3;
}

message Address { // [!code --]
  string street = 1; // [!code --]
  string city = 2; // [!code --]
  string country = 3; // [!code --]
} // [!code --]

message Order {
  string id = 1;
  User user = 2;
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3;  // ERROR: Address больше не существует
}

// Address удалён — BREAKING CHANGE!

message Order {
  string id = 1;
  User user = 2;
}
```

**Ошибка:**
```
user.proto:8:1: Previously present message "Address" was deleted from file. (BREAKING_CHECK)
```

### Дополнительные примеры

**Удаление вложенного сообщения:**

```proto
// Before
message User {
  string name = 1;
  Profile profile = 2;
  
  message Profile { // [!code --]
    string bio = 1; // [!code --]
    string avatar_url = 2; // [!code --]
  } // [!code --]
}

// After - BREAKING CHANGE!
message User {
  string name = 1;
  Profile profile = 2;  // ERROR: Profile удалён
  
  // Вложенное сообщение Profile удалено
}
```

### Хороший (Безопасный)

**Вместо удаления — пометить как deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3 [deprecated = true]; // [!code focus]
}

message Address {
  option deprecated = true; // [!code focus]
  string street = 1;
  string city = 2;
  string country = 3;
}

message Order {
  string id = 1;
  User user = 2;
}
```

**Или заменить новой версией сообщения:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2;
  Address address = 3 [deprecated = true]; // [!code focus] // Старое поле
  AddressV2 address_v2 = 4; // [!code focus] // Новая улучшенная структура
}

message Address {
  option deprecated = true; // [!code focus]
  string street = 1;
  string city = 2;
  string country = 3;
}

message AddressV2 { // [!code focus]
  string street_address = 1; // [!code focus]
  string city = 2; // [!code focus]
  string state = 3; // [!code focus]
  string postal_code = 4; // [!code focus]
  string country = 5; // [!code focus]
} // [!code focus]

message Order {
  string id = 1;
  User user = 2;
}
```

## Влияние

- **Wire Format:** Старые данные с удалённым сообщением не десериализуются
- **Generated Code:** Классы/структуры удаляются — компиляция клиента ломается
- **Field References:** Поля, использующие удалённый тип, становятся невалидны
- **Nested Dependencies:** Все сообщения, ссылающиеся на удалённый тип, ломаются

## Реальный пример

**Код клиента ломается:**
```go
// Before
user := &myapi.User{
    Name:  "John Doe",
    Email: "john@example.com",
    Address: &myapi.Address{  // ERROR после удаления
        Street:  "123 Main St",
        City:    "New York",
        Country: "USA",
    },
}

// undefined: myapi.Address
```

**Старые данные становятся нечитаемыми:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "address": {
    "street": "123 Main St",
    "city": "New York",
    "country": "USA"
  }
}
// После удаления Address: парсер не знает поле address
```

## Стратегия миграции

1. **Пометить как deprecated:**
   ```proto
   message OldMessage {
     option deprecated = true;
     // ... fields
   }
   ```
2. **Пометить поля, использующие его:**
   ```proto
   OldMessage old_data = 5 [deprecated = true];
   ```
3. **Добавить замену при необходимости:**
   ```proto
   NewMessage new_data = 6;
   ```
4. **Удалить в следующей major‑версии после периода миграции**

## Типовые сценарии

### Рефакторинг структуры
```proto
message UserInfo {
  option deprecated = true;
  string name = 1;
  string email = 2;
}

message UserProfile {
  string full_name = 1;
  string email_address = 2;
  string phone = 3;
}
```

### Удаление «неиспользуемых» сообщений
```proto
message OldConfig {
  option deprecated = true;
  // Не удалять сразу — могут быть сохранённые данные
}
```

### Миграция версии
```proto
package myapi.v2;

message User {
  // Новая структура без устаревших ссылок
}
```
