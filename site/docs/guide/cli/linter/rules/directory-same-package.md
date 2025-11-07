# DIRECTORY_SAME_PACKAGE

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files in a given directory are in the same package.

## Examples

### Bad

```proto
// File: dir/foo.proto

syntax = "proto3";
   
package foo;

message Foo {
    string bar = 1;
}
```

### Good

```proto
// File: dir/foo/foo.proto // [!code focus]

syntax = "proto3";

package dir.foo; // [!code focus]

message Foo {
    string bar = 1;
}
```
