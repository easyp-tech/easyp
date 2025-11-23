# RPC_REQUEST_RESPONSE_UNIQUE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все сообщения запроса и ответа RPC уникальны: для каждого метода используются собственные (не переиспользуются одинаковые пары `Request` / `Response` у разных RPC).

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (FooRequest) returns (FooResponse) {}
    rpc GetBar (FooRequest) returns (FooResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (FooRequest) returns (FooResponse) {} // [!code focus]
    rpc GetBar (BarRequest) returns (BarResponse) {} // [!code focus]
}
```
