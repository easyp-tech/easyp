# Linter

[[toc]]

## Why Use a Linter for Proto Files?

Linters play a crucial role in modern software development, particularly for proto files. By enforcing style and formatting rules, linters help maintain code quality, reduce potential bugs, and ensure that the codebase is clean and readable. This leads to several benefits:

- **Reduced Development Costs:** Linters catch errors early in the development cycle, saving time and resources that would be spent on debugging and fixing issues later.
- **Improved Team Collaboration:** A standardized codebase makes it easier for team members to understand and work with each other's code, facilitating a smoother collaborative environment.
- **Business Efficiency:** Teams using linters often produce higher quality code, which translates to fewer production issues and maintenance costs. This makes businesses that adopt linting practices more competitive and cost-effective.

## Configuration Reference

EasyP linter provides flexible configuration options to adapt to different project requirements. All configuration is done through the `lint` section in your `easyp.yaml` file.

### Complete Configuration Example

```yaml
version: v1alpha

lint:
  # Rules and categories to use
  use:
    - MINIMAL
    - BASIC
    - COMMENT_SERVICE
    - COMMENT_RPC
  
  # Custom suffixes for naming conventions
  enum_zero_value_suffix: "UNSPECIFIED"
  service_suffix: "Service"
  
  # Directory and rule exclusions
  ignore:
    - "vendor/"
    - "third_party/"
    - "legacy/old_protos/"
  
  except:
    - COMMENT_FIELD
    - COMMENT_MESSAGE
  
  # Comment-based rule ignoring
  allow_comment_ignores: true
  
  # File-specific rule ignoring
  ignore_only:
    COMMENT_SERVICE: ["legacy/", "vendor/"]
    SERVICE_SUFFIX: ["proto/external/"]
```

### Configuration Parameters

#### `use` ([]string)
Specifies which linter rules or rule categories to apply. You can mix individual rules with predefined categories.

**Available categories:**
- **MINIMAL**: Basic package consistency checks
- **BASIC**: Naming conventions and usage patterns  
- **DEFAULT**: Recommended rules for most projects
- **COMMENTS**: Comment presence and formatting
- **UNARY_RPC**: Restrictions on streaming RPCs

**Individual rules**: Any specific rule name (e.g., `ENUM_PASCAL_CASE`, `FIELD_LOWER_SNAKE_CASE`)

**Examples:**
```yaml
# Use predefined categories
use: [MINIMAL, BASIC, DEFAULT]

# Mix categories with individual rules
use:
  - MINIMAL
  - COMMENT_SERVICE
  - COMMENT_RPC
  - ENUM_PASCAL_CASE

# Use only specific rules
use:
  - PACKAGE_DEFINED
  - SERVICE_PASCAL_CASE
  - FIELD_LOWER_SNAKE_CASE
```

#### `enum_zero_value_suffix` (string)
Defines the required suffix for zero-value enum entries. This enforces a consistent naming pattern for the default enum value.

**Default**: Empty (no suffix required)
**Common values**: `"UNSPECIFIED"`, `"UNKNOWN"`, `"DEFAULT"`

**Example:**
```yaml
enum_zero_value_suffix: "UNSPECIFIED"
```

With this setting, enum definitions must follow this pattern:
```protobuf
enum Status {
  STATUS_UNSPECIFIED = 0;  // Required suffix
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}
```

#### `service_suffix` (string)
Specifies the required suffix for service names. This ensures consistent service naming across your project.

**Default**: Empty (no suffix required)
**Common values**: `"Service"`, `"API"`, `"Svc"`

**Example:**
```yaml
service_suffix: "Service"
```

With this setting, service definitions must follow this pattern:
```protobuf
service UserService {     // Required "Service" suffix
  rpc GetUser(...) returns (...);
}
```

#### `ignore` ([]string)
Lists directories or file paths to completely exclude from linting. Supports glob patterns and relative paths from the project root.

**Use cases:**
- Third-party or vendor proto files
- Generated proto files
- Legacy code being phased out
- Test fixtures that intentionally violate rules

**Examples:**
```yaml
ignore:
  - "vendor/"                    # Exclude vendor directory
  - "third_party/"              # Exclude third-party protos
  - "testdata/"                 # Exclude test fixtures
  - "proto/legacy/"             # Exclude legacy protos
  - "**/*_test.proto"           # Exclude test proto files
```

#### `except` ([]string)
Disables specific rules globally across the entire project. Use this when certain rules don't fit your project's conventions.

**When to use:**
- Legacy projects with established naming conventions
- Projects with specific style requirements
- Gradual adoption of linting rules

**Examples:**
```yaml
except:
  - COMMENT_FIELD              # Don't require field comments
  - COMMENT_MESSAGE            # Don't require message comments
  - SERVICE_SUFFIX             # Don't enforce service suffix
  - ENUM_ZERO_VALUE_SUFFIX     # Don't enforce enum zero suffix
```

#### `allow_comment_ignores` (bool)
Enables or disables the ability to ignore specific rules using inline comments in proto files.

**Default**: `false`
**Recommended**: `true` for flexibility during development

**Example:**
```yaml
allow_comment_ignores: true
```

When enabled, you can use comments to ignore rules on specific elements:
```protobuf
// buf:lint:ignore COMMENT_SERVICE
service LegacyUserAPI {
  // nolint:COMMENT_RPC
  rpc GetUser(...) returns (...);
}
```

#### `ignore_only` (map[string][]string)
Allows you to disable specific rules only for certain files or directories, while keeping them active elsewhere.

**Use cases:**
- Legacy code that can't be easily updated
- Third-party protos with different conventions
- Generated code that doesn't follow your style
- Gradual migration strategies

**Format**: `rule_name: [list_of_paths]`

**Examples:**
```yaml
ignore_only:
  # Don't require service comments in legacy directory
  COMMENT_SERVICE: 
    - "proto/legacy/"
    - "vendor/"
  
  # Don't enforce service suffix for external APIs
  SERVICE_SUFFIX:
    - "proto/external/"
    - "third_party/"
  
  # Allow old naming in specific files
  FIELD_LOWER_SNAKE_CASE:
    - "proto/legacy/old_messages.proto"
    - "vendor/external_api.proto"
```

## Comment-Based Rule Ignoring

When `allow_comment_ignores` is enabled, you can use inline comments to ignore specific linter rules for individual proto elements. This provides fine-grained control over rule enforcement.

### Supported Comment Formats

EasyP supports two comment formats for maximum compatibility:

#### Buf-compatible format
```protobuf
// buf:lint:ignore RULE_NAME
```

#### EasyP native format  
```protobuf
// nolint:RULE_NAME
```

### Usage Examples

#### Ignoring Service Rules
```protobuf
// buf:lint:ignore COMMENT_SERVICE
// buf:lint:ignore SERVICE_SUFFIX  
service UserAPI {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

#### Ignoring RPC Rules
```protobuf
service UserService {
  // nolint:COMMENT_RPC
  // nolint:RPC_REQUEST_STANDARD_NAME
  rpc GetUserInfo(UserInfoReq) returns (UserInfoResp);
}
```

#### Ignoring Message and Field Rules
```protobuf
// buf:lint:ignore COMMENT_MESSAGE
message UserData {
  // nolint:COMMENT_FIELD
  // nolint:FIELD_LOWER_SNAKE_CASE
  string userName = 1;
}
```

#### Ignoring Enum Rules
```protobuf
// nolint:COMMENT_ENUM
// nolint:ENUM_ZERO_VALUE_SUFFIX
enum UserType {
  UNKNOWN = 0;  // Would normally require suffix
  ADMIN = 1;
  USER = 2;
}
```

#### Multiple Rules on One Line
```protobuf
// buf:lint:ignore COMMENT_SERVICE,SERVICE_SUFFIX
service LegacyAPI {
  // nolint:COMMENT_RPC,RPC_REQUEST_STANDARD_NAME
  rpc getData(DataReq) returns (DataResp);
}
```

### Best Practices for Comment Ignores

#### Use Sparingly
Comment ignores should be the exception, not the rule. Overuse indicates that your linting configuration may need adjustment.

**Good:**
```protobuf
// Legacy service - buf:lint:ignore SERVICE_SUFFIX
service UserAPI {
  rpc GetUser(...) returns (...);
}
```

**Bad:**
```protobuf
// nolint:COMMENT_SERVICE,SERVICE_SUFFIX,COMMENT_RPC
service UserAPI {
  // nolint:COMMENT_RPC,RPC_REQUEST_STANDARD_NAME
  rpc getUser(...) returns (...);
  // nolint:COMMENT_RPC,RPC_RESPONSE_STANDARD_NAME  
  rpc updateUser(...) returns (...);
}
```

#### Add Explanatory Comments
Always explain why you're ignoring a rule to help future maintainers understand the decision.

```protobuf
// Third-party API compatibility requires non-standard naming
// buf:lint:ignore SERVICE_SUFFIX
service ExternalUserAPI {
  // Legacy method name for backwards compatibility
  // nolint:RPC_REQUEST_STANDARD_NAME
  rpc getUserData(UserReq) returns (UserResp);
}
```

#### Prefer Configuration Over Comments
When multiple files need the same rule ignored, use `ignore_only` configuration instead of individual comments.

**Instead of:**
```protobuf
// In file1.proto
// buf:lint:ignore COMMENT_SERVICE
service API1 { ... }

// In file2.proto  
// buf:lint:ignore COMMENT_SERVICE
service API2 { ... }
```

**Use:**
```yaml
# In easyp.yaml
lint:
  ignore_only:
    COMMENT_SERVICE: ["legacy/"]
```

## Linter Categories

To accommodate different project needs and preferences, EasyP linter provides predefined rule categories. These categories group together various rules, allowing teams to quickly select the level of strictness or areas they want to focus on during linting.

**When to use each category:**

- **MINIMAL:** Essential for any proto project - ensures basic consistency and prevents fundamental issues
- **BASIC:** Recommended for most projects - enforces common naming conventions and best practices  
- **DEFAULT:** Additional quality checks - useful for mature projects with established workflows
- **COMMENTS:** Documentation requirements - important for public APIs and team collaboration
- **UNARY_RPC:** Streaming restrictions - use when your architecture requires only unary RPCs

The available categories are:
- **MINIMAL:** Basic checks to ensure package consistency.
- **BASIC:** Additional checks for naming conventions and usage patterns.
- **DEFAULT:** A set of default rules that most projects should use.
- **COMMENTS:** Ensures that comments are present and properly formatted.
- **UNARY_RPC:** Specific rules for unary RPC services.

### Rule Groupings

Below are the rule groupings under each category:

#### MINIMAL

- `DIRECTORY_SAME_PACKAGE`
- `PACKAGE_DEFINED`
- `PACKAGE_DIRECTORY_MATCH`
- `PACKAGE_SAME_DIRECTORY`

#### BASIC

- `ENUM_FIRST_VALUE_ZERO`
- `ENUM_NO_ALLOW_ALIAS`
- `ENUM_PASCAL_CASE`
- `ENUM_VALUE_UPPER_SNAKE_CASE`
- `FIELD_LOWER_SNAKE_CASE`
- `IMPORT_NO_PUBLIC`
- `IMPORT_NO_WEAK`
- `IMPORT_USED`
- `MESSAGE_PASCAL_CASE`
- `ONEOF_LOWER_SNAKE_CASE`
- `PACKAGE_LOWER_SNAKE_CASE`
- `PACKAGE_SAME_CSHARP_NAMESPACE`
- `PACKAGE_SAME_GO_PACKAGE`
- `PACKAGE_SAME_JAVA_MULTIPLE_FILES`
- `PACKAGE_SAME_JAVA_PACKAGE`
- `PACKAGE_SAME_PHP_NAMESPACE`
- `PACKAGE_SAME_RUBY_PACKAGE`
- `PACKAGE_SAME_SWIFT_PREFIX`
- `RPC_PASCAL_CASE`
- `SERVICE_PASCAL_CASE`

#### DEFAULT

- `ENUM_VALUE_PREFIX`
- `ENUM_ZERO_VALUE_SUFFIX`
- `FILE_LOWER_SNAKE_CASE`
- `RPC_REQUEST_RESPONSE_UNIQUE`
- `RPC_REQUEST_STANDARD_NAME`
- `RPC_RESPONSE_STANDARD_NAME`
- `PACKAGE_VERSION_SUFFIX`
- `SERVICE_SUFFIX`

#### COMMENTS

- `COMMENT_ENUM`
- `COMMENT_ENUM_VALUE`
- `COMMENT_FIELD`
- `COMMENT_MESSAGE`
- `COMMENT_ONEOF`
- `COMMENT_RPC`
- `COMMENT_SERVICE`

#### UNARY_RPC

- `RPC_NO_CLIENT_STREAMING`
- `RPC_NO_SERVER_STREAMING`

## Conclusion

Adopting the EasyP linter for your proto files can significantly enhance your development workflow, code quality, 
and overall project maintainability. With full compatibility with the Buf linter, teams can easily migrate and start benefiting
from our tool's robust features and flexible configuration options.