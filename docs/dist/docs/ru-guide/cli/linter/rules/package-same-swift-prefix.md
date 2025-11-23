# PACKAGE_SAME_SWIFT_PREFIX

Категории:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что все файлы с одним и тем же package используют один и тот же `swift_prefix`.

## Examples

### Bad

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option swift_prefix = "Foo";
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option swift_prefix = "Bar";
```

### Good

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option swift_prefix = "Foo"; // [!code focus]
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option swift_prefix = "Foo"; // [!code focus]
```
