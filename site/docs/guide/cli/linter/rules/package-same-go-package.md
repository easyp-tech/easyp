# PACKAGE_SAME_GO_PACKAGE

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files with a given package are in the same Go package.

## Examples

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
