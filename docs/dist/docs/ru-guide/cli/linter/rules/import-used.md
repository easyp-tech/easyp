# IMPORT_USED

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все импорты используются в proto-файле.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

import "bar.proto";
```

### Good

```proto
syntax = "proto3";

package foo;

import "bar.proto"; // [!code focus]

message Foo {
    bar.Bar bar = 1; // [!code focus]
}
```
