# PACKAGE_SAME_RUBY_PACKAGE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files with a given package are in the same Ruby package.

## Examples

### Bad

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option ruby_package = "Foo";
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option ruby_package = "Bar";
```

### Good

```proto
// File: pkg/foo.proto

syntax = "proto3";

package pkg;

option ruby_package = "Foo"; // [!code focus]
```

```proto
// File: pkg/bar.proto

syntax = "proto3";

package pkg;

option ruby_package = "Foo"; // [!code focus]
```
