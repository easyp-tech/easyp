# PACKAGE_SAME_SWIFT_PREFIX

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files with a given package are in the same Swift prefix.

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
