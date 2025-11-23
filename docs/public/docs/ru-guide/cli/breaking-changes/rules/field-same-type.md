<!-- TODO: Review translation -->

# FIELD_SAME_TYPE

Категории:

- **WIRE+**

Правило проверяет, что типы полей сообщений (message fields) не изменяются. Смена типа поля ломает совместимость по wire‑формату и сгенерированный код: бинарное представление и ожидания клиентского кода различаются для разных типов.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

message Product {
  string name = 1;
  int32 price = 2;
  bool available = 3;
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message Product {
  string name = 1;
  string price = 2; // [!code --] Changed from int32 to string (изменён тип)
  bool available = 3;
}
```

**Ошибка:**
```
product.proto:5:3: Field "2" with name "price" on message "Product" changed type from "int32" to "string". (BREAKING_CHECK)
```

### Дополнительные примеры

**Несовместимые изменения типов:**
```proto
// Before
message Order {
  string id = 1;
  int64 timestamp = 2;
  repeated string tags = 3;
  OrderStatus status = 4;
}

// After - ВСЁ BREAKING
message Order {
  int32 id = 1;           // string -> int32: BREAKING
  string timestamp = 2;   // int64 -> string: BREAKING
  string tags = 3;        // repeated -> singular: BREAKING
  int32 status = 4;       // enum -> int32: BREAKING
}
```

### Хороший (Безопасный)

**Вместо смены типа — добавьте новое поле:**
```proto
syntax = "proto3";

package myapi.v1;

message Product {
  string name = 1;
  int32 price = 2 [deprecated = true]; // [!code focus] // Старое поле помечено deprecated
  bool available = 3;
  string price_formatted = 4; // [!code focus] // Новое поле со строковым представлением
}
```

**Или создайте новую версию сообщения:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus] // Новая версия package

message Product {
  string name = 1;
  string price = 2; // [!code focus] // В v2 уже string
  bool available = 3;
}
```

## Влияние

- **Wire Format:** Старые бинарные данные не десериализуются корректно
- **Generated Code:** Меняются типы в сгенерированном коде — клиенты не компилируются
- **Runtime Errors:** Ошибки парсинга при несовпадении ожидаемого типа
- **Data Corruption:** Неверная интерпретация байтов данных

## Типовые проблемы изменений типов

### Числовые типы
```proto
// BREAKING: различное кодирование/семантика
int32 -> int64    // Иное кодирование
int32 -> uint32   // Меняется знак
int32 -> string   // Совсем другой формат
```

### Коллекции
```proto
// BREAKING: изменяется структура
string -> repeated string              // Одиночное -> коллекция
repeated int32 -> map<string, int32>   // Массив -> map
```

### Message типы
```proto
// BREAKING: другая структура сообщения
UserInfo -> UserProfile  // Совершенно разные сообщения
string -> UserInfo       // Скаляр -> сообщение
```

## Стратегия миграции

1. **Добавьте новое поле** с нужным типом:
   ```proto
   int32 old_price = 2 [deprecated = true];
   string new_price = 5; // Новое поле
   ```

2. **Период двойной записи** — заполняйте оба поля для совместимости

3. **Мигрируйте клиентов** на новое поле постепенно

4. **Удалите старое поле** в следующей major‑версии:
   ```proto
   reserved 2, "old_price";
   string new_price = 5;
   ```

## Безопасные изменения (Теоретически wire‑совместимые)

Примечание: EasyP сейчас считает ЛЮБУЮ смену типа breaking. Однако технически некоторые случаи совместимы:

- `int32` ↔ `uint32` (одно кодирование)
- `int64` ↔ `uint64` (одно кодирование)
- `string` → `bytes` (если все строки валидный UTF‑8)

Эти исключения могут получить отдельную обработку в будущих версиях EasyP с более гибкими уровнями строгости.