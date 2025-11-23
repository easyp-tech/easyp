# IMPORT_NO_PUBLIC

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что в proto-файле не используются `public` импорты (оператор `import public "..."`).

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

import public "bar.proto";
```

### Good

```proto
syntax = "proto3";

package foo;

import "bar.proto"; // [!code focus]
```
