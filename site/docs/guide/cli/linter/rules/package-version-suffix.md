# PACKAGE_VERSION_SUFFIX

Categories:

- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that the package version suffix is `_vX` where `X` is a number of version.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;   
```

### Good

```proto
syntax = "proto3";

package foo.v1; // [!code focus]
```


