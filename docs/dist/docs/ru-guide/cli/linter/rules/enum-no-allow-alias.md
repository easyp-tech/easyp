# ENUM_NO_ALLOW_ALIAS

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что опция `allow_alias` не установлена в значение `true` внутри enum.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

enum Foo {
    option allow_alias = true; // [!code focus]
    BAR = 0;
    BAZ = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

enum Foo {
    BAR = 0;
    BAZ = 1;
}
```
