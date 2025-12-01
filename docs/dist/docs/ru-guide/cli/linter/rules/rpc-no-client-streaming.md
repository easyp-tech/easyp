# RPC_NO_CLIENT_STREAMING

Категории:

- **UNARY_RPC**

Это правило проверяет, что RPC не использует клиентский стриминг (запрос не является `stream ...`).

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (stream BarRequest) returns (BarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {} // [!code focus]
}
```
