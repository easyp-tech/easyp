# IMPORT_USED

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that all imports are used in a proto file.

## Examples

### Bad

```proto
syntax = "proto3";

package foo;

import "bar.proto";
```

### Good

```proto
syntax = "proto3";

package foo;

import "bar.proto"; // [!code focus]

message Foo {
    bar.Bar bar = 1; // [!code focus]
}
```
