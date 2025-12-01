# RPC_NO_SERVER_STREAMING

Категории:

- **UNARY_RPC**

Это правило проверяет, что RPC не использует серверный стриминг (ответ не является `stream ...`).

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (stream BarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {} // [!code focus]
}
```
