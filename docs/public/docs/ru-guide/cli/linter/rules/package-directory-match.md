# PACKAGE_DIRECTORY_MATCH

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы находятся в директории, имя которой соответствует имени их пакета.

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
