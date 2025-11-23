# RPC_NO_SERVER_STREAMING

Categories:

- **UNARY_RPC**

This rule checks that rpc has no server streaming.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (stream BarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {} // [!code focus]
}
```
