# PACKAGE_SAME_JAVA_PACKAGE

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all files with a given package are in the same Java package.

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


