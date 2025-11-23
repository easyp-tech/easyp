# PACKAGE_SAME_CSHARP_NAMESPACE

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы с заданным пакетом используют одно и то же пространство имён C# (`csharp_namespace`).

## Примеры

### Bad

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option csharp_namespace = "Foo.Bar";
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option csharp_namespace = "Foo.Baz";
```

### Good

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option csharp_namespace = "Foo.Bar"; // [!code focus]
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option csharp_namespace = "Foo.Bar"; // [!code focus]
```
