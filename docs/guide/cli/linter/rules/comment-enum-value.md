# COMMENT_ENUM_VALUE

Categories:

- **COMMENT**

This rule checks that enum values have a comment.

## Examples

### Bad

```proto
syntax = "proto3";

enum Foo {
    BAR = 0;
    BAZ = 1;
}
```

### Good

```proto

syntax = "proto3";

enum Foo {
    // BAR value for bar logic // [!code focus]
    BAR = 0; 
    // BAZ value for baz logic // [!code focus]
    BAZ = 1; 
}
```
