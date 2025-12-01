# RPC_REQUEST_STANDARD_NAME

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все сообщения запроса RPC названы в формате `MethodRequest`.

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (FooRequest) returns (FooResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (GetFooRequest) returns (FooResponse) {} // [!code focus]
}
```
