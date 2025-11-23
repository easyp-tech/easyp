# COMMENT_MESSAGE

Категории:

- **COMMENT**

Это правило проверяет, что message имеет комментарий.

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

// Foo message for bar and baz logic // [!code focus]
message Foo {
    // bar field for bar logic 
    string bar = 1; 
    // baz field for baz logic
    string baz = 2; 
}
```
