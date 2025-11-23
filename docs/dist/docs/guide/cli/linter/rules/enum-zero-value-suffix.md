# ENUM_ZERO_VALUE_SUFFIX

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all enum zero values are suffixed with `_NONE` or your custom suffix.

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
    FOO_BAR_NONE = 0; // [!code focus]
}
```


