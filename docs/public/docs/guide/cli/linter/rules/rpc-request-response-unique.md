# RPC_REQUEST_RESPONSE_UNIQUE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all RPC request and response messages are unique.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (FooRequest) returns (FooResponse) {}
    rpc GetBar (FooRequest) returns (FooResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (FooRequest) returns (FooResponse) {} // [!code focus]
    rpc GetBar (BarRequest) returns (BarResponse) {} // [!code focus]
}
```


