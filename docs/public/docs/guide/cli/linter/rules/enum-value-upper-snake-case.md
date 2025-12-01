# ENUM_VALUE_UPPER_SNAKE_CASE

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that enum values are in UPPER_SNAKE_CASE.

## Examples

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

