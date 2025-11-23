# FILE_LOWER_SNAKE_CASE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files are in lower_snake_case.

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
