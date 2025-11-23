# RPC_REQUEST_STANDARD_NAME

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all RPC request messages are named `MethodRequest`.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (FooRequest) returns (FooResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (GetFooRequest) returns (FooResponse) {} // [!code focus]
}
```
