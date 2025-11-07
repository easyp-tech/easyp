# MESSAGE_PASCAL_CASE

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that message names are in PascalCase.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;

message foo_bar {
    string bar_name = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

message FooBar { // [!code focus]
    string bar_name = 1;
}
```