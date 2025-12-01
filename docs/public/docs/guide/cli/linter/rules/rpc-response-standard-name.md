# RPC_RESPONSE_STANDARD_NAME

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all RPC response messages are named `MethodResponse`.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (GetFooRequest) returns (Foo) {}
}
```

### Good

```proto
syntax = "proto3";

service Foo {
    rpc GetFoo (GetFooRequest) returns (GetFooResponse) {} // [!code focus]
}
```
