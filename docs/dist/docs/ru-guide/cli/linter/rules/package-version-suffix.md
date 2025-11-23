# PACKAGE_VERSION_SUFFIX

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что суффикс версии пакета имеет формат `_vX`, где `X` — числовая версия.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;
```

### Good

```proto
syntax = "proto3";

package foo.v1; // [!code focus]
```
