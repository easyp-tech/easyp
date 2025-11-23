# PACKAGE_SAME_PHP_NAMESPACE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы с одним и тем же `package` используют одно и то же значение опции `php_namespace`.

## Примеры

### Bad

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option php_namespace = "Foo\\Bar";
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option php_namespace = "Foo\\Baz";
```

### Good

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option php_namespace = "Foo\\Bar"; // [!code focus]
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option php_namespace = "Foo\\Bar"; // [!code focus]
```
