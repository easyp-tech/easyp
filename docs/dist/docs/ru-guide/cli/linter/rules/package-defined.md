# PACKAGE_DEFINED

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что во всех файлах присутствует директива package.

## Примеры

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
