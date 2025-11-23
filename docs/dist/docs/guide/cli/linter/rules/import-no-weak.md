# IMPORT_NO_WEAK

Categories:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

This rule checks that no weak imports are used in a proto file.


## Examples

### Bad

```proto
syntax = "proto2";

package foo;

import weak "bar.proto";
```

### Good

```proto
syntax = "proto2";

package foo;

import "bar.proto"; // [!code focus]
```

