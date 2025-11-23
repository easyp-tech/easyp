# SERVICE_SUFFIX

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все сервисы имеют суффикс `Service` или ваш пользовательский суффикс.

## Примеры

### Bad

```proto
syntax = "proto3";

service Foo {
    rpc Bar(BarRequest) returns (BarResponse);
}
```

### Good

```proto
syntax = "proto3";

service FooService { // [!code focus]
    rpc Bar(BarRequest) returns (BarResponse); 
}
```
