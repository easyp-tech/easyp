# ENUM_FIRST_VALUE_ZERO

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that the first value of an enum is zero.

## Examples

### Bad

```proto

syntax = "proto3";

package foo;

enum Foo {
    BAR = 1;
    BAZ = 2;
}
```

### Good

```proto
syntax = "proto3";

package foo;

enum Foo {
    BAR = 0; // [!code focus]
    BAZ = 1;
}
```
