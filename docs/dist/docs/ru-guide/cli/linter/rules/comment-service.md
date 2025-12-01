# COMMENT_SERVICE

Категории:

- **COMMENT**

Это правило проверяет, что сервис имеет комментарий.

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

// Foo service for bar logic // [!code focus]
service Foo {
    // Bar rpc for bar logic 
    rpc Bar (BarRequest) returns (BarResponse) {}
}
```
