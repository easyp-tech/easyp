# PACKAGE_LOWER_SNAKE_CASE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что имена package записаны в формате lower_snake_case (строчные буквы, при необходимости цифры и символ подчеркивания, без пробелов и заглавных букв).

## Примеры

### Bad

```proto
// File: bar/foo.proto
syntax = "proto3";

package FooBar;
```

### Good

```proto
// File: bar/foo.proto
syntax = "proto3";

package foo_bar; // [!code focus]
```
