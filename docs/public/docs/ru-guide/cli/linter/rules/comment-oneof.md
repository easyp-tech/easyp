# COMMENT_ONEOF

Категории:

- **COMMENT**

Это правило проверяет, что oneof имеет комментарий.

## Примеры

### Bad

```proto
syntax = "proto3";

message Foo {
    oneof bar {
        string baz = 1;
        string qux = 2;
    }
}
```

### Good

```proto
syntax = "proto3";

message Foo {
    // bar oneof for baz and qux logic // [!code focus]
    oneof bar {
        string baz = 1;
        string qux = 2;
    }
}
```
