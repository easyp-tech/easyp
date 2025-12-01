# PACKAGE_SAME_JAVA_PACKAGE

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы с одним и тем же `package` используют одно и то же значение опции `java_package`.

## Examples

### Bad

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option java_package = "com.example.foo";
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option java_package = "com.example.bar";
```

### Good

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option java_package = "com.example.foo"; // [!code focus]
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option java_package = "com.example.foo"; // [!code focus]
```
