# RPC_PASCAL_CASE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all RPC names are in PascalCase.

## Examples

### Bad

```proto
syntax = "proto3";

service foo_bar {
    rpc get_foo_bar (FooBarRequest) returns (FooBarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service FooBar {
    rpc GetFooBar (FooBarRequest) returns (FooBarResponse) {} // [!code focus]
}
```
