# PACKAGE_SAME_DIRECTORY

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы с данным package находятся в одной директории.

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
