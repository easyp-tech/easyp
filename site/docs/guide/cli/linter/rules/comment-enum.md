# COMMENT_ENUM

Categories:

- **COMMENT**

This rule checks that enum has a comment.

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

// Foo enum for bar and baz logic // [!code focus]
enum Foo {
    BAR = 0; 
    BAZ = 1; 
}
```