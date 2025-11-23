# COMMENT_SERVICE

Categories:

- **COMMENT**

This rule checks that service has a comment.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

// Foo service for bar logic // [!code focus]
service Foo {
    // Bar rpc for bar logic 
    rpc Bar (BarRequest) returns (BarResponse) {}
}
```


