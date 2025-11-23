# ENUM_PASCAL_CASE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что имена enum написаны в стиле PascalCase.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

enum foo_bar {
    BAR = 0;
    BAZ = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

enum FooBar { // [!code focus]
    BAR = 0;
    BAZ = 1;
}
```
