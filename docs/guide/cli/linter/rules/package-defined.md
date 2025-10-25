
# PACKAGE_DEFINED

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files have a package declaration.

## Examples

### Bad

```proto
syntax = "proto3";

message Foo {
    string bar = 1;
}

```

### Good

```proto
syntax = "proto3";

package foo; // [!code focus]

message Foo {
    string bar = 1;
}
```

