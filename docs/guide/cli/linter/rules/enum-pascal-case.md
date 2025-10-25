# ENUM_PASCAL_CASE


Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that enum names are in PascalCase.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;

enum foo_bar {
    BAR = 0;
    BAZ = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

enum FooBar { // [!code focus]
    BAR = 0;
    BAZ = 1;
}
```
