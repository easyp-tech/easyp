# ENUM_NO_ALLOW_ALIAS

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that the `allow_alias` option is not set to `true` in an enum.

## Examples

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
