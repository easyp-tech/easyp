# ENUM_ZERO_VALUE_SUFFIX

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все нулевые значения enum имеют суффикс `_NONE` или ваш кастомный (пользовательский) суффикс.

## Примеры

### Bad

```proto
syntax = "proto3";

enum Foo {
    BAR = 0;
}
```

### Good

```proto
syntax = "proto3";

enum Foo {
    FOO_BAR_NONE = 0; // [!code focus]
}
```
