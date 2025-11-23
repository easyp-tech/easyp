# PACKAGE_LOWER_SNAKE_CASE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that package names are in lower_snake_case.

## Examples

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
