# COMMENT_ENUM

Категории:

- **COMMENT**

Это правило проверяет, что enum имеет комментарий.

## Примеры

### Bad

```proto
syntax = "proto3";

enum Foo {
    BAR = 0;
    BAZ = 1;
}
```

### Good

```proto
syntax = "proto3";

// Foo enum for bar and baz logic // [!code focus]
enum Foo {
    BAR = 0;
    BAZ = 1;
}
```
