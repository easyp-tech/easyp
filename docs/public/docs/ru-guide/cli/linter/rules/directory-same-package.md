# DIRECTORY_SAME_PACKAGE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы в заданной директории находятся в одном и том же package.

## Примеры

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
