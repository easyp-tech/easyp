<!-- TODO: Review translation -->

# SERVICE_NO_DELETE

Категории:

- **WIRE+**

Данное правило проверяет, что ни один service не был удалён из proto‑файлов. Удаление service ломает сгенерированный код и клиентские приложения, которые на него опираются.

## Примеры

### Плохой (Breaking)

**До:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

service OrderService { // [!code --]
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse); // [!code --]
} // [!code --]
```

**После:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// OrderService удалён — BREAKING CHANGE!
```

**Ошибка:**
```
services.proto:8:1: Previously present service "OrderService" was deleted from file. (BREAKING_CHECK)
```

### Хороший (Безопасный)

**Вместо удаления — отметьте service как deprecated:**
```proto
syntax = "proto3";

package myapi.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

service OrderService {
  option deprecated = true; // [!code focus]
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse) {
    option deprecated = true; // [!code focus]
  }
}
```

## Влияние

- **Generated Code:** Классы / интерфейсы сервиса исчезают — клиенты не компилируются
- **Client Applications:** Существующие вызовы к удалённому сервису падают на этапе сборки
- **Runtime:** gRPC клиенты теряют определение метода/сервиса

## Стратегия миграции

1. **Сначала деприкация:**
   ```proto
   service OldService {
     option deprecated = true;
     // ... methods
   }
   ```

2. **Уведомите клиентов** о сроках вывода из эксплуатации

3. **Удалите после периода миграции** в следующей major‑версии