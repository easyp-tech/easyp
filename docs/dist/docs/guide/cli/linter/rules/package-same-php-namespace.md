# PACKAGE_SAME_PHP_NAMESPACE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files with a given package are in the same PHP namespace.

## Examples

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
