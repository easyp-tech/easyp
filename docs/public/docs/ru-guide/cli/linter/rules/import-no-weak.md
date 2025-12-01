# IMPORT_NO_WEAK

Категории:
- **MINIMAL**
- **BASIC**
- **DEFAULT**

Это правило проверяет, что в proto‑файле не используются weak‑импорты (`import weak "..."`).

Weak import (синтаксис `import weak`) допустим в proto2, но его использование усложняет анализ зависимостей и может приводить к неочевидным ошибкам при генерации кода. Рекомендуется всегда использовать обычный `import`.

## Примеры

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
