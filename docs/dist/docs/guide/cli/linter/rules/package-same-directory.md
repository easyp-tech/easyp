# PACKAGE_SAME_DIRECTORY

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files with a given package are in the same directory.

## Examples

### Bad

```proto
// File: bar/foo.proto

syntax = "proto3";

package foo;

message Foo {
    string bar = 1;
}

```

### Good

```proto
// File: bar/foo.proto // [!code focus]

syntax = "proto3";

package bar; // [!code focus]

message Foo {
    string bar = 1;
}

```
