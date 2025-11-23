# ENUM_VALUE_PREFIX

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all enum values are prefixed with the enum name.

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
