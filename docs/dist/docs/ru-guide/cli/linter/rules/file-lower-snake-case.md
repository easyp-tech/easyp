# FILE_LOWER_SNAKE_CASE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы имеют имя в формате lower_snake_case (только строчные буквы, цифры и символ подчеркивания, без пробелов и заглавных букв).

## Examples

### Bad

```proto
// File: bar/FooBaz.proto

syntax = "proto3";
```

### Good

```proto
// File: bar/foo_baz.proto // [!code focus]

syntax = "proto3";
```
