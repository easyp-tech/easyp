# PACKAGE_SAME_GO_PACKAGE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы с одним и тем же package используют одно и то же значение опции `go_package`.

## Примеры

### Bad

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option go_package = "example.com/foo/bar";
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option go_package = "example.com/foo/baz";
```

### Good

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option go_package = "example.com/foo/bar"; // [!code focus]
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option go_package = "example.com/foo/bar"; // [!code focus]
```
