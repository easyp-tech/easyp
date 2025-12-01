# COMMENT_FIELD

Категории:

- **COMMENT**

Это правило проверяет, что все поля сообщений имеют комментарий.

## Примеры

### Bad

```proto
syntax = "proto3";

message Foo {
    string bar = 1;
    string baz = 2;
}
```

### Good

```proto
syntax = "proto3";

message Foo {
    // bar field for bar logic // [!code focus]
    string bar = 1;
    // baz field for baz logic // [!code focus]
    string baz = 2;
}
```
