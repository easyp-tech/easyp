# SERVICE_PASCAL_CASE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все имена сервисов написаны в стиле PascalCase.

## Примеры

### Bad

```proto
syntax = "proto3";

service foo_bar {
    rpc get_foo_bar (FooBarRequest) returns (FooBarResponse) {}
}
```

### Good

```proto
syntax = "proto3";

service FooBar { // [!code focus]
    rpc GetFooBar (FooBarRequest) returns (FooBarResponse) {} 
}
```
