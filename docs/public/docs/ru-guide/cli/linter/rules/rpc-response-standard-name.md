# RPC_RESPONSE_STANDARD_NAME

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все сообщения ответа RPC названы в формате `MethodResponse`.

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (GetFooRequest) returns (Foo) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (GetFooRequest) returns (GetFooResponse) {} // [!code focus]
}
```
