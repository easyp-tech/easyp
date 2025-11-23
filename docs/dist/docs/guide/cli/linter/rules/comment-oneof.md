# COMMENT_ONEOF

Categories:

- **COMMENT**

This rule checks that oneof has a comment.

## Examples

### Bad

```proto
syntax = "proto3";

message Foo {
    oneof bar {
        string baz = 1;
        string qux = 2;
    }
}
```

### Good

```proto
syntax = "proto3";

message Foo {
    // bar oneof for baz and qux logic // [!code focus]
    oneof bar {
        string baz = 1; 
        string qux = 2; 
    }
}
```
