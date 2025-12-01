# RPC_NO_CLIENT_STREAMING

Categories:

- **UNARY_RPC**

This rule checks that rpc has no client streaming.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar (stream BarRequest) returns (BarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc Bar (BarRequest) returns (BarResponse) {} // [!code focus]
}
```
