# ONEOF_LOWER_SNAKE_CASE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что имена oneof записаны в стиле lower_snake_case.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

message Foo {
    oneof BarName {
        string bar_name = 1;
    }
}
```

### Good

```proto
syntax = "proto3";

package foo;

message Foo {
    oneof bar_name { // [!code focus]
        string bar_name = 1;
    }
}
```
