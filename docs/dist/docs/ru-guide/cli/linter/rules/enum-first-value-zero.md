# ENUM_FIRST_VALUE_ZERO

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что первое значение enum равно нулю.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;

enum Foo {
    BAR = 1;
    BAZ = 2;
}
```

### Good

```proto
syntax = "proto3";

package foo;

enum Foo {
    BAR = 0; // [!code focus]
    BAZ = 1;
}
```
