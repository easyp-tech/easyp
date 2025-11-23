# IMPORT_NO_PUBLIC

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that no public imports are used in a proto file.


## Examples

### Bad

```proto
syntax = "proto3";

package foo;

import public "bar.proto";
```

### Good

```proto
syntax = "proto3";

package foo;

import "bar.proto"; // [!code focus]
```

