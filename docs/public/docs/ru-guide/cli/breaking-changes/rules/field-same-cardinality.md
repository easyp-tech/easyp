<!-- TODO: Review translation -->

# FIELD_SAME_CARDINALITY

Категории:

- **WIRE+**

Правило проверяет, что поля сообщений сохраняют ту же кардинальность (обязательность / optional). Изменение кардинальности поля ломает совместимость по wire‑формату и сгенерированный код: семантика присутствия и ожидания клиентского кода отличаются для optional и required (implicit) полей.

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
}

message CreateUserRequest {
  string name = 1;
  optional string email = 2;
  string phone = 3;
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  optional string email = 2; // [!code --] Было required (implicit), стало optional
  int32 age = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2; // [!code --] Было optional, стало требуемым (implicit required)
  string phone = 3;
}
```

**Ошибка:**
```
user.proto:6:3: Field "2" with name "email" on message "User" became optional. (BREAKING_CHECK)
request.proto:7:3: Field "2" with name "email" on message "CreateUserRequest" became not optional. (BREAKING_CHECK)
```

### Дополнительные примеры

**Проблемы миграции Proto2 → Proto3:**

```proto
// Before (proto2)
syntax = "proto2";

message Order {
  required string id = 1;
  optional string notes = 2;
  repeated string tags = 3;
}

// After (proto3) - BREAKING CHANGES!
syntax = "proto3";

message Order {
  string id = 1;           // BREAKING: required -> implicit optional
  string notes = 2;        // BREAKING: explicit optional -> implicit optional
  repeated string tags = 3; // OK
}
```

### Хороший (Безопасный)

**Вместо изменения кардинальности — добавьте новое поле:**
```proto
syntax = "proto3";

package myapi.v1;

message User {
  string name = 1;
  string email = 2 [deprecated = true]; // [!code focus] Сохраняем старую кардинальность
  int32 age = 3;
  optional string email_optional = 4; // [!code focus] Новое optional поле
}

message CreateUserRequest {
  string name = 1;
  optional string email = 2 [deprecated = true]; // [!code focus]
  string phone = 3;
  string email_required = 4; // [!code focus] Новое требуемое поле
}
```

**Или создайте новую версию сообщения:**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus]

message User {
  string name = 1;
  optional string email = 2; // [!code focus] Чистая модель в v2
  int32 age = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2; // [!code focus] Требуемое в v2
  string phone = 3;
}
```

## Влияние

- **Wire Format:** Меняется семантика присутствия — клиенты неправильно ожидают поля
- **Generated Code:** Меняются методы доступа / проверки (has, clear)
- **Validation:** Правила проверки обязательности переписываются
- **Default Values:** Optional и required по‑разному обрабатывают отсутствие значения

## Реальный пример

**Код клиента ломается при смене кардинальности:**
```go
// Before - поле фактически требуется (implicit)
user := &myapi.User{
    Name:  "John",
    Email: "john@example.com",
    Age:   30,
}

if user.Email != "" {
    // Есть значение
}

// After - поле стало optional (pointer)
user := &myapi.User{
    Name: "John",
    Age:  30,
    // Email может отсутствовать
}

if user.Email != nil && *user.Email != "" {
    // Есть значение (новый способ)
}

user.Email = &emailValue // Присваивание указателя вместо строки
```

**Валидация на сервере:**
```go
// Before
func validateUser(user *User) error {
    if user.Email == "" {
        return errors.New("email is required")
    }
    return nil
}

// After
func validateUser(user *User) error {
    if user.Email == nil {
        return errors.New("email is required")
    }
    if *user.Email == "" {
        return errors.New("email cannot be empty")
    }
    return nil
}
```

## Стратегия миграции

1. **Добавьте новое поле** с нужной кардинальностью:
   ```proto
   string old_email = 2 [deprecated = true];
   optional string new_email = 5;
   ```

2. **Период двойной записи** — заполняйте оба поля:
   ```go
   user := &User{
       Name:     "John",
       OldEmail: email,
       NewEmail: &email,
   }
   ```

3. **Клиенты постепенно переходят**:
   ```go
   email := user.NewEmail
   if email == nil && user.OldEmail != "" {
       v := user.OldEmail
       email = &v
   }
   ```

4. **Удаление старого поля** в следующей major‑версии:
   ```proto
   reserved 2, "old_email";
   optional string new_email = 5;
   ```

## Типовые сценарии

### Миграция Proto2 → Proto3
```proto
syntax = "proto3";
message Order {
  string id = 1; // BREAKING если был required
}

// Сохранить семантику:
syntax = "proto3";
message Order {
  string id = 1;
  optional string notes = 2;
}
```

### Делать optional поле required
```proto
message CreateUserRequest {
  string name = 1;
  string email = 2; // BREAKING (раньше optional)
}

message CreateUserRequest {
  string name = 1;
  optional string email = 2 [deprecated = true];
  string required_email = 3;
}
```

### Делать required поле optional
```proto
message User {
  optional string phone = 3; // BREAKING (было implicit required)
}

message User {
  string phone = 3 [deprecated = true];
  optional string phone_optional = 4;
}
```

## Типы кардинальности в Protobuf

### Proto3
- **Implicit optional**: `string name = 1;`
- **Explicit optional**: `optional string email = 2;`
- **Repeated**: `repeated string tags = 3;`
- **Map**: `map<string, string> metadata = 4;`

### Proto2
- **Required**: `required string id = 1;`
- **Optional**: `optional string email = 2;`
- **Repeated**: `repeated string tags = 3;`

### Матрица Breaking изменений

| From | To | Result |
|------|----|--------|
| Required | Optional | ❌ BREAKING |
| Optional | Required | ❌ BREAKING |
| Implicit | Explicit | ❌ BREAKING |
| Explicit | Implicit | ❌ BREAKING |
| Single | Repeated | ❌ BREAKING |
| Repeated | Single | ❌ BREAKING |
| Field | Map | ❌ BREAKING |
| Map | Field | ❌ BREAKING |