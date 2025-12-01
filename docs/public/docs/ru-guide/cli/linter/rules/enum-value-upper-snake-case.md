# ENUM_VALUE_UPPER_SNAKE_CASE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что значения enum записаны в стиле UPPER_SNAKE_CASE.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

enum Foo {
    barName = 0;
    bazName = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

enum Foo {
    BAR_NAME = 0; // [!code focus]
    BAZ_NAME = 1; // [!code focus]
}
```
