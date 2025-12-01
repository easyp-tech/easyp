<!-- TODO: Review translation -->

# RPC_SAME_REQUEST_TYPE

Категории:

- **WIRE+**

Это правило проверяет, что RPC‑методы сохраняют прежний тип сообщения запроса (request message type). Изменение типа запроса RPC ломает совместимость по wire‑формату и сгенерированный код: клиенты ожидают конкретную структуру при вызове метода.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message UpdateUserRequest {
  string user_id = 1;
  string name = 2;
  string email = 3;
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequestV2) returns (GetUserResponse); // [!code --] Изменён тип запроса
  rpc UpdateUser(UserUpdateRequest) returns (UpdateUserResponse); // [!code --] Изменён тип запроса
}

message GetUserRequestV2 { // [!code --] Новый тип запроса
  string id = 1;
  bool include_profile = 2;
}

message UserUpdateRequest { // [!code --] Новый тип запроса
  string id = 1;
  UserProfile profile = 2;
}
```

**Ошибка:**
```
user_service.proto:6:3: RPC "GetUser" on service "UserService" changed request type from "GetUserRequest" to "GetUserRequestV2". (BREAKING_CHECK)
user_service.proto:7:3: RPC "UpdateUser" on service "UserService" changed request type from "UpdateUserRequest" to "UserUpdateRequest". (BREAKING_CHECK)
```

### Дополнительные примеры

**Смена типа из другого пакета:**
```proto
// Before
import "common/user.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// After - BREAKING CHANGE!
import "v2/user.proto";

service UserService {
  rpc CreateUser(v2.CreateUserRequest) returns (CreateUserResponse); // Другой package
}
```

### Хороший (Безопасный)

**Вместо изменения типа — версионируйте RPC:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc GetUserV2(GetUserRequestV2) returns (GetUserResponse); // [!code focus] // Новый метод
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc UpdateUserV2(UserUpdateRequest) returns (UpdateUserResponse); // [!code focus] // Новый метод
}

// Старые сообщения — для обратной совместимости
message GetUserRequest {
  string user_id = 1;
}

// Новая версия сообщения
message GetUserRequestV2 { // [!code focus]
  string id = 1; // [!code focus]
  bool include_profile = 2; // [!code focus]
} // [!code focus]
```

**Или создайте новую версию сервиса (package v2):**
```proto
syntax = "proto3";

package myapi.v2; // [!code focus]

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse); // [!code focus]
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse); // [!code focus]
}

message GetUserRequest { // [!code focus]
  string id = 1; // [!code focus]
  bool include_profile = 2; // [!code focus]
} // [!code focus]
```

## Влияние

- **Wire Format:** Старые запросы клиентов не десериализуются новым сервером
- **Generated Code:** Сигнатуры методов в stubs меняются — компиляция клиентов падает
- **gRPC Calls:** Требуется использовать другую структуру запроса
- **Runtime Errors:** Несоответствие типов вызывает ошибки выполнения

## Реальный пример

**Код клиента ломается:**
```go
// Before
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{
    UserId: "user123",
})

// After (тип запроса изменён)
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{  // ERROR: тип не найден
    UserId: "user123",  // ERROR: поле не найдено
})

// Новый корректный вызов:
resp, err := client.GetUser(ctx, &myapi.GetUserRequestV2{
    Id: "user123",
    IncludeProfile: true,
})
```

**Серверная реализация ломается:**
```go
// Before
func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    userID := req.UserId
    // ...
}

// After
func (s *server) GetUser(ctx context.Context, req *GetUserRequestV2) (*GetUserResponse, error) {
    userID := req.Id
    include := req.IncludeProfile
    // ... новая логика
}
```

## Стратегия миграции

1. **Добавьте новый RPC с новым типом запроса:**
   ```proto
   rpc GetUser(GetUserRequest) returns (GetUserResponse) {
     option deprecated = true;
   }
   rpc GetUserV2(GetUserRequestV2) returns (GetUserResponse);
   ```

2. **Поддерживайте оба метода на сервере:**
   ```go
   func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
       return s.getUserInternal(req.UserId, false)
   }

   func (s *server) GetUserV2(ctx context.Context, req *GetUserRequestV2) (*GetUserResponse, error) {
       return s.getUserInternal(req.Id, req.IncludeProfile)
   }
   ```

3. **Мигрируйте клиентов на V2:**
   ```go
   resp, err := client.GetUserV2(ctx, &myapi.GetUserRequestV2{
       Id: userID,
       IncludeProfile: true,
   })
   ```

4. **Удалите старый метод в следующей major‑версии:**
   ```proto
   rpc GetUserV2(GetUserRequestV2) returns (GetUserResponse);
   ```

## Типовые сценарии

### Добавление обязательных полей
```proto
// ПЛОХО: меняем существующий запрос
message GetUserRequest {
  string user_id = 1;
  bool include_deleted = 2; // BREAKING
}

// Хорошо: новая версия
message GetUserRequestV2 {
  string user_id = 1;
  bool include_deleted = 2;
}
```

### Реструктуризация данных запроса
```proto
// ПЛОХО: меняем структуру прямо
message UpdateUserRequest {
  UserProfile profile = 1; // BREAKING
}

// Хорошо: версионирование
rpc UpdateUser(UpdateUserRequest) returns (...) {
  option deprecated = true;
}
rpc UpdateUserV2(UpdateUserRequestV2) returns (...);
```

### Миграция пакета
```proto
// ПЛОХО: заменяем тип на пакет v2 прямо
import "v2/messages.proto"; // BREAKING

// Хорошо: добавляем новый RPC
rpc OldMethod(v1.Request) returns (...) {
  option deprecated = true;
}
rpc NewMethod(v2.Request) returns (...);
```
