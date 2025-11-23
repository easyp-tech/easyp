# FIELD_LOWER_SNAKE_CASE

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that field names of messages are in lower_snake_case.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;

message Foo {
    string BarName = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

message Foo {
    string bar_name = 1; // [!code focus]
}
```
