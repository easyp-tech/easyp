# MESSAGE_PASCAL_CASE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что имена сообщений (message) записаны в стиле PascalCase.

## Примеры

### Bad

```proto
syntax = "proto3";

package foo;

message foo_bar {
    string bar_name = 1;
}
```

### Good

```proto
syntax = "proto3";

package foo;

message FooBar { // [!code focus]
    string bar_name = 1;
}
```
