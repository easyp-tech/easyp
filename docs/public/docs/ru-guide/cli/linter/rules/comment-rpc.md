# COMMENT_RPC

Categories:

- **COMMENT**

This rule checks that rpc has a comment.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {}
    rpc Baz (BazRequest) returns (BazResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    // Bar rpc for bar logic // [!code focus]
    rpc Bar (BarRequest) returns (BarResponse) {}
    // Baz rpc for baz logic // [!code focus]
    rpc Baz (BazRequest) returns (BazResponse) {}
}
```
