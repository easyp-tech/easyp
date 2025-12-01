# SERVICE_SUFFIX

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all services are suffixed with `Service` or your custom suffix.

## Examples

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar(BarRequest) returns (BarResponse);
}
```

### Good

```proto
syntax = "proto3";

service FooService { // [!code focus]
    rpc Bar(BarRequest) returns (BarResponse); 
}
```

