# ONEOF_LOWER_SNAKE_CASE

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that oneof names are in lower_snake_case.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;

message Foo {
    oneof BarName {
        string bar_name = 1;
    }
}
```

### Good

```proto
syntax = "proto3";

package foo;

message Foo {
    oneof bar_name { // [!code focus]
        string bar_name = 1;
    }
}
```
