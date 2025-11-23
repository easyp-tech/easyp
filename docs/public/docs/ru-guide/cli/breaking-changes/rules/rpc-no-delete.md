<!-- TODO: Review translation -->

# RPC_NO_DELETE

Категории:

- **WIRE+**

Правило проверяет, что ни один RPC‑метод не был удалён из service. Удаление RPC ломает сгенерированный код и клиентские приложения, которые этот метод вызывают.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse); // [!code --]
}
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  // DeleteUser RPC удалён — BREAKING CHANGE!
}
```

**Ошибка:**
```
services.proto:7:1: Previously present RPC "DeleteUser" on service "UserService" was deleted. (BREAKING_CHECK)
```

### Хороший (Безопасный)

**Вместо удаления — пометить RPC deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option deprecated = true; // [!code focus]
  }
}
```

## Влияние

- **Generated Code:** Stub/метод исчезает — клиенты не компилируются
- **Client Applications:** Существующие вызовы метода падают при сборке
- **Runtime:** gRPC клиенты теряют определение метода

## Стратегия миграции

1. **Сначала деприкация:**
   ```proto
   rpc OldMethod(OldRequest) returns (OldResponse) {
     option deprecated = true;
   }
   ```

2. **Имплементация может возвращать ошибку**, указывая что метод устарел (например, INTERNAL / UNIMPLEMENTED)

3. **Удалить в следующей major‑версии** после миграции всех клиентов