# ENUM_VALUE_PREFIX

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все значения enum имеют префикс, совпадающий с именем enum.

## Examples

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
    FOO_BAR = 0; // [!code focus]
}
```
