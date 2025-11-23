# FIELD_LOWER_SNAKE_CASE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что имена полей сообщений записаны в стиле lower_snake_case.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

message Foo {
    string BarName = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

message Foo {
    string bar_name = 1; // [!code focus]
}
```
