<!-- TODO: Review translation -->

# RPC_SAME_RESPONSE_TYPE

Категории:

- **WIRE+**

Это правило проверяет, что RPC‑методы сохраняют тот же тип ответного сообщения (response message type). Изменение типа ответа RPC ломает совместимость по wire‑формату и сгенерированный код: клиенты ожидают конкретную структуру при получении ответа.

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

message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
}

message UpdateUserResponse {
  bool success = 1;
  string message = 2;
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponseV2); // [!code --] Изменён тип ответа
  rpc UpdateUser(UpdateUserRequest) returns (UserUpdateResponse); // [!code --] Изменён тип ответа
}

message GetUserResponseV2 { // [!code --] Другой тип сообщения
  string id = 1;
  UserProfile profile = 2;
  repeated string permissions = 3;
}

message UserUpdateResponse { // [!code --] Другой тип сообщения
  UpdateResult result = 1;
  UserProfile updated_profile = 2;
}
```

**Ошибка:**
```
user_service.proto:6:3: RPC "GetUser" on service "UserService" changed response type from "GetUserResponse" to "GetUserResponseV2". (BREAKING_CHECK)
user_service.proto:7:3: RPC "UpdateUser" on service "UserService" changed response type from "UpdateUserResponse" to "UserUpdateResponse". (BREAKING_CHECK)
```

### Дополнительные примеры

**Смена типа из другого пакета:**
```proto
// Before
import "common/responses.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// After - BREAKING CHANGE!
import "v2/responses.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (v2.CreateUserResponse); // Другой package
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
  rpc GetUserV2(GetUserRequest) returns (GetUserResponseV2); // [!code focus] // Новый метод
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option deprecated = true; // [!code focus]
  }
  rpc UpdateUserV2(UpdateUserRequest) returns (UserUpdateResponse); // [!code focus] // Новый метод
}

// Старые сообщения сохраняем для обратной совместимости
message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
}

// Новая расширенная версия
message GetUserResponseV2 { // [!code focus]
  string id = 1; // [!code focus]
  UserProfile profile = 2; // [!code focus]
  repeated string permissions = 3; // [!code focus]
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

message GetUserResponse { // [!code focus]
  string id = 1; // [!code focus]
  UserProfile profile = 2; // [!code focus]
  repeated string permissions = 3; // [!code focus]
} // [!code focus]
```

## Влияние

- **Wire Format:** Ответы сервера не десериализуются у старых клиентов
- **Generated Code:** Сигнатуры ожидают старый тип — компиляция падает
- **gRPC Calls:** Меняются типы в сгенерированных stub'ах
- **Runtime Errors:** Несоответствие типов ломает обработку ответа

## Реальный пример

**Код клиента ломается:**
```go
// Before
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{
    UserId: "user123",
})
if err != nil {
    return err
}

userID := resp.UserId
name := resp.Name

// After (тип ответа изменён)
client := myapi.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &myapi.GetUserRequest{
    UserId: "user123",
})
if err != nil {
    return err
}

// Ошибки компиляции:
userID := resp.UserId  // ERROR
name := resp.Name      // ERROR

// Новый способ:
userID := resp.Id
profile := resp.Profile
permissions := resp.Permissions
```

**Имплементация сервера также должна быть переписана:**
```go
// Before
func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    return &GetUserResponse{
        UserId: "user123",
        Name:   "John Doe",
        Email:  "john@example.com",
    }, nil
}

// After
func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponseV2, error) {
    return &GetUserResponseV2{
        Id: "user123",
        Profile: &UserProfile{
            Name:  "John Doe",
            Email: "john@example.com",
        },
        Permissions: []string{"read"},
    }, nil
}
```

## Стратегия миграции

1. **Добавьте новый RPC c новым типом ответа:**
   ```proto
   rpc GetUser(GetUserRequest) returns (GetUserResponse) {
     option deprecated = true;
   }
   rpc GetUserV2(GetUserRequest) returns (GetUserResponseV2);
   ```

2. **Поддерживайте оба на сервере:**
   ```go
   func (s *server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
       user := s.getUserInternal(req.UserId)
       return &GetUserResponse{
           UserId: user.ID,
           Name:   user.Name,
           Email:  user.Email,
       }, nil
   }

   func (s *server) GetUserV2(ctx context.Context, req *GetUserRequest) (*GetUserResponseV2, error) {
       user := s.getUserInternal(req.UserId)
       return &GetUserResponseV2{
           Id: user.ID,
           Profile: &UserProfile{
               Name:  user.Name,
               Email: user.Email,
           },
           Permissions: user.Permissions,
       }, nil
   }
   ```

3. **Мигрируйте клиентов на новый RPC:**
   ```go
   resp, err := client.GetUserV2(ctx, &myapi.GetUserRequest{
       UserId: userID,
   })
   if err != nil {
       return err
   }
   profile := resp.Profile
   permissions := resp.Permissions
   ```

4. **Удалите старый RPC** в следующей major‑версии:
   ```proto
   rpc GetUserV2(GetUserRequest) returns (GetUserResponseV2);
   ```

## Типовые сценарии

### Добавление новых обязательных полей
```proto
// ПЛОХО: меняем существующий тип
message GetUserResponse {
  string user_id = 1;
  string name = 2;
  UserProfile profile = 3; // BREAKING
}

// Хорошо: новая версия
message GetUserResponseV2 {
  string user_id = 1;
  string name = 2;
  UserProfile profile = 3;
}
```

### Реструктуризация ответа
```proto
// ПЛОХО: меняем структуру прямо
message UpdateUserResponse {
  UserDetails details = 1; // BREAKING
}

// Хорошо: версионирование
rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
  option deprecated = true;
}
rpc UpdateUserV2(UpdateUserRequest) returns (UpdateUserResponseV2);
```

### Изменение формата ошибок
```proto
// ПЛОХО: меняем формат напрямую
message CreateUserResponse {
  ErrorDetails error = 1; // BREAKING
}

// Хорошо: новая версия
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
  option deprecated = true;
}
rpc CreateUserV2(CreateUserRequest) returns (CreateUserResponseV2);
```

### Миграция формата
```proto
// ПЛОХО: перевод простой структуры в сложную
message GetUserResponse {
  UserData data = 1; // BREAKING
}

// Хорошо: предоставляем обе версии
rpc GetUser(GetUserRequest) returns (GetUserResponse) {
  option deprecated = true;
}
rpc GetUserDetailed(GetUserRequest) returns (GetUserDetailedResponse);
```
