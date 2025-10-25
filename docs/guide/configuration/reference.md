# Configuration Reference

[[toc]]

## Overview

EasyP uses a YAML configuration file (typically named `easyp.yaml`) to manage all aspects of your protobuf workflow: linting, dependency management, code generation, and breaking change detection. This reference documents every configuration option available in EasyP.

## Configuration File Structure

The configuration file uses YAML format and consists of several top-level sections:

```yaml
version: v1alpha           # Configuration format version (required)
lint: {}                   # Linting configuration
deps: []                   # Package dependencies
generate: {}               # Code generation settings
breaking: {}               # Breaking changes detection settings
```

## Complete Configuration Example

Here's a comprehensive example showing all available options:

```yaml
version: v1alpha

# Linting configuration
lint:
  use:
    - DEFAULT
    - COMMENTS
  enum_zero_value_suffix: "UNSPECIFIED"
  service_suffix: "Service"
  ignore:
    - "vendor/"
    - "third_party/"
  except:
    - FIELD_LOWER_SNAKE_CASE
  allow_comment_ignores: true
  ignore_only:
    SERVICE_SUFFIX: ["legacy/"]
    COMMENT_RPC: ["generated/"]

# Package dependencies
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate

# Code generation configuration
generate:
  inputs:
    - directory: "proto"
    - directory:
        path: "api"
        root: "services"
    - git_repo:
        url: "github.com/company/shared-protos@v1.0.0"
        sub_directory: "proto"
        out: "vendor"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/myproject
      with_imports: true
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false
    - name: grpc-gateway
      url: "remote-plugin.company.com:8080/grpc-gateway:v2"
      out: ./gen/go
      opts:
        paths: source_relative

# Breaking changes detection
breaking:
  against_git_ref: "main"
  ignore:
    - "experimental/"
    - "internal/"
```

## Version Field

### `version` (string, required)

Specifies the configuration file format version. This ensures compatibility between your configuration and the EasyP version you're using.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `version` | string | ✅ Yes | - | Configuration format version |

**Supported values:**
- `v1alpha` - Current configuration format

**Example:**
```yaml
version: v1alpha
```

**Notes:**
- This field is required and must be the first field in your configuration file
- Future versions of EasyP may introduce new configuration formats
- EasyP will provide migration tools when new versions are released

## Lint Configuration

The `lint` section controls how EasyP checks your protobuf files for style, consistency, and best practices violations.

### Lint Fields Reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `use` | []string | No | `[]` | List of linting rules or rule categories to enable |
| `enum_zero_value_suffix` | string | No | `""` | Required suffix for zero-value enum entries |
| `service_suffix` | string | No | `""` | Required suffix for service names |
| `ignore` | []string | No | `[]` | Directories or files to exclude from linting |
| `except` | []string | No | `[]` | Specific rules to disable globally |
| `allow_comment_ignores` | bool | No | `false` | Allow inline comment-based rule ignoring |
| `ignore_only` | map[string][]string | No | `{}` | Disable specific rules for certain paths |

### `lint.use` ([]string)

Specifies which linting rules or rule categories to apply to your protobuf files.

**Available rule categories:**
- `MINIMAL` - Essential package consistency checks
- `BASIC` - Common naming conventions and patterns
- `DEFAULT` - Recommended rules for most projects
- `COMMENTS` - Documentation and comment requirements
- `UNARY_RPC` - Restrictions on streaming RPCs

**Individual rules:**
You can also specify individual rule names instead of categories. See the [Linter Rules Reference](../cli/linter/linter.md) for a complete list.

**Examples:**
```yaml
lint:
  # Use predefined categories
  use: [DEFAULT, COMMENTS]
  
  # Mix categories with individual rules
  use:
    - MINIMAL
    - BASIC
    - COMMENT_SERVICE
    - ENUM_PASCAL_CASE
  
  # Use only specific rules (maximum control)
  use:
    - PACKAGE_DEFINED
    - MESSAGE_PASCAL_CASE
    - FIELD_LOWER_SNAKE_CASE
```

### `lint.enum_zero_value_suffix` (string)

Defines the required suffix for zero-value enum entries (the entry with value 0).

**Common values:**
- `"UNSPECIFIED"` (Google style guide recommendation)
- `"UNKNOWN"`
- `"INVALID"`
- `"DEFAULT"`

**Example:**
```yaml
lint:
  enum_zero_value_suffix: "UNSPECIFIED"
```

**Effect on your protos:**
```protobuf
// With enum_zero_value_suffix: "UNSPECIFIED"
enum Status {
  STATUS_UNSPECIFIED = 0;  // ✅ Correct - has required suffix
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

enum Priority {
  PRIORITY_UNKNOWN = 0;   // ❌ Wrong - should be PRIORITY_UNSPECIFIED
  PRIORITY_HIGH = 1;
}
```

### `lint.service_suffix` (string)

Specifies the required suffix for all service names.

**Common values:**
- `"Service"` (most common)
- `"API"`
- `"Svc"`
- `"Server"`

**Example:**
```yaml
lint:
  service_suffix: "Service"
```

**Effect on your protos:**
```protobuf
// With service_suffix: "Service"
service UserService {      // ✅ Correct - has required suffix
  rpc GetUser(...) returns (...);
}

service AuthAPI {          // ❌ Wrong - should be AuthService
  rpc Login(...) returns (...);
}
```

### `lint.ignore` ([]string)

Lists directories or file paths to completely exclude from linting. Supports:
- Relative paths from project root
- Directory names (must end with `/`)
- Glob patterns

**Use cases:**
- Third-party/vendor proto files
- Generated proto files
- Legacy code being phased out
- Test fixtures

**Example:**
```yaml
lint:
  ignore:
    - "vendor/"              # Ignore entire vendor directory
    - "third_party/"        # Ignore third-party protos
    - "proto/legacy/"       # Ignore specific subdirectory
    - "**/*_pb.proto"       # Ignore generated files
    - "testdata/"           # Ignore test fixtures
```

### `lint.except` ([]string)

Disables specific linting rules globally across the entire project.

**When to use:**
- Your project has established conventions that differ from the rules
- Gradual adoption of linting (start permissive, tighten over time)
- Legacy projects that can't be easily refactored

**Example:**
```yaml
lint:
  except:
    - COMMENT_FIELD           # Don't require field comments
    - COMMENT_MESSAGE         # Don't require message comments  
    - SERVICE_SUFFIX          # Don't enforce service suffix
    - ENUM_ZERO_VALUE_SUFFIX  # Don't enforce enum zero suffix
```

### `lint.allow_comment_ignores` (bool)

Enables or disables the ability to ignore specific rules using inline comments in proto files.

**Example:**
```yaml
lint:
  allow_comment_ignores: true
```

**When enabled, you can use in proto files:**
```protobuf
// buf:lint:ignore COMMENT_SERVICE
service LegacyAPI {
  // nolint:COMMENT_RPC
  rpc GetData(...) returns (...);
}
```

**Supported comment formats:**
- `// buf:lint:ignore RULE_NAME` - Buf-compatible format
- `// nolint:RULE_NAME` - EasyP native format

### `lint.ignore_only` (map[string][]string)

Allows you to disable specific rules only for certain files or directories, while keeping them active elsewhere.

**Format:** Map where keys are rule names and values are lists of paths to ignore.

**Example:**
```yaml
lint:
  ignore_only:
    # Don't require service comments in legacy code
    COMMENT_SERVICE:
      - "proto/legacy/"
      - "vendor/"
    
    # Don't enforce suffix for external APIs
    SERVICE_SUFFIX:
      - "proto/external/"
      - "third_party/"
    
    # Allow old naming in specific files
    FIELD_LOWER_SNAKE_CASE:
      - "proto/v1/old_api.proto"
      - "migrations/2023/data.proto"
```

## Dependencies Configuration

The `deps` section lists external protobuf packages your project depends on.

### `deps` ([]string)

A list of Git repositories containing protobuf files that your project needs.

**Format:** `repository@version`
- `repository` - Git repository path (without https://)
- `version` - Git tag, commit hash, or omit for latest

**Examples:**
```yaml
deps:
  # Specific version (recommended for production)
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  
  # Latest version (useful during development)
  - github.com/bufbuild/protoc-gen-validate
  
  # Specific commit (for unreleased features)
  - github.com/company/internal-protos@abc123def456789
  
  # Private repositories (configure Git authentication)
  - github.com/company/private-protos@v1.0.0
  - gitlab.company.com/platform/shared-types@v2.1.0
```

**Version formats:**
- **Git tags:** `@v1.2.3`, `@common-protos-1_3_1`
- **Commit hash:** `@abc123def456789` (full or short)
- **Latest:** Omit version to use latest available tag
- **Pseudo-versions:** Auto-generated for commits without tags

**Notes:**
- Dependencies are downloaded to `~/.easyp/mod/` (or `$EASYPPATH/mod/`)
- Use `easyp mod download` to fetch dependencies
- Use `easyp mod update` to update to latest versions
- Lock file (`easyp.lock`) ensures reproducible builds

## Generate Configuration

The `generate` section configures code generation from your protobuf files.

### Generate Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `inputs` | []Input | Yes | Sources of proto files to generate from |
| `plugins` | []Plugin | Yes | Code generation plugins to run |

### `generate.inputs` ([]Input)

Specifies where to find the protobuf files for code generation. Each input can be either a local directory or a remote Git repository.

#### Local Directory Input

**Simple format (string):**
```yaml
generate:
  inputs:
    - directory: "proto"
    - directory: "api/v1"
```

**Advanced format (object):**
```yaml
generate:
  inputs:
    - directory:
        path: "proto"      # Directory with proto files
        root: "."          # Root directory for imports (default: ".")
    - directory:
        path: "api"
        root: "services"   # Imports will be resolved from services/api/
```

**Fields:**
- `path` (string, required) - Directory containing proto files
- `root` (string, optional) - Root directory for import resolution (default: ".")

#### Remote Git Repository Input

Fetch proto files from external Git repositories:

```yaml
generate:
  inputs:
    - git_repo:
        url: "github.com/company/shared-protos@v1.0.0"
        sub_directory: "proto"     # Directory within the repo
        out: "vendor/shared"        # Where to place generated code
    
    - git_repo:
        url: "gitlab.company.com/platform/api-definitions"
        sub_directory: "public"
        out: "generated/platform"
```

**Fields:**
- `url` (string, required) - Git repository URL with optional version
- `sub_directory` (string, optional) - Directory within repo containing protos
- `out` (string, optional) - Output directory for generated code

### `generate.plugins` ([]Plugin)

Defines which code generation plugins to run and how to configure them.

#### Plugin Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Plugin name (e.g., "go", "python", "grpc-gateway") |
| `out` | string | Yes | Output directory for generated files |
| `opts` | map[string]string | No | Plugin-specific options |
| `url` | string | No | Remote plugin URL for distributed execution |
| `with_imports` | bool | No | Include imported proto files in generation |

#### Local Plugin Execution

```yaml
generate:
  plugins:
    # Go code generation
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/myproject
      with_imports: true
    
    # Go gRPC generation
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false
    
    # gRPC-Gateway generation
    - name: grpc-gateway
      out: ./gen/go
      opts:
        paths: source_relative
        generate_unbound_methods: true
    
    # OpenAPI generation
    - name: openapiv2
      out: ./docs/api
      opts:
        simple_operation_ids: true
        json_names_for_fields: true
        output_format: yaml
```

#### Remote Plugin Execution

For distributed builds or custom plugins:

```yaml
generate:
  plugins:
    - name: custom-plugin
      url: "plugin-server.company.com:8080/custom-plugin:v1.0"
      out: ./gen/custom
      opts:
        custom_option: "value"
```

**URL format:** `host:port/plugin_name:version`

#### Common Plugin Options

**Go plugins (`go`, `go-grpc`):**
- `paths` - Import path mode: `source_relative` or `import`
- `module` - Go module path
- `require_unimplemented_servers` - Generate unimplemented server stubs

**gRPC-Gateway plugin:**
- `paths` - Import path mode
- `generate_unbound_methods` - Generate standalone HTTP handlers
- `logtostderr` - Log to stderr
- `allow_repeated_fields_in_body` - Allow repeated fields in HTTP body

**OpenAPI v2 plugin:**
- `simple_operation_ids` - Use simple operation IDs
- `json_names_for_fields` - Use JSON field names
- `output_format` - Output format: `json` or `yaml`
- `allow_repeated_fields_in_body` - Allow repeated fields

**Validation plugin (`validate-go`):**
- `paths` - Import path mode
- `lang` - Target language

### `generate.plugins[].with_imports` (bool)

Controls whether to also generate code for imported proto files.

**When to use:**
- `true` - Generate code for all dependencies (monolithic approach)
- `false` - Only generate for your proto files (modular approach)

**Example:**
```yaml
generate:
  plugins:
    - name: go
      out: ./gen
      with_imports: true   # Also generate googleapis, etc.
```

## Breaking Configuration

The `breaking` section configures backward compatibility checking.

### Breaking Fields Reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `against_git_ref` | string | No | `"master"` | Git reference to compare against |
| `ignore` | []string | No | `[]` | Directories to exclude from analysis |

### `breaking.against_git_ref` (string)

Specifies which Git reference (branch, tag, or commit) to compare against when checking for breaking changes.

**Examples:**
```yaml
breaking:
  # Compare against main branch (default)
  against_git_ref: "main"
  
  # Compare against specific tag
  against_git_ref: "v1.0.0"
  
  # Compare against specific commit
  against_git_ref: "abc123def"
  
  # Compare against previous commit
  against_git_ref: "HEAD~1"
```

**Common patterns:**
- `main` or `master` - Compare against main branch
- `production` - Compare against production branch
- `v1.0.0` - Compare against specific release
- `HEAD~1` - Compare against previous commit
- `origin/main` - Compare against remote branch

### `breaking.ignore` ([]string)

Lists directories or files to exclude from breaking change analysis.

**Use cases:**
- Experimental APIs under development
- Internal/private APIs not exposed to clients
- Test fixtures that intentionally break compatibility
- Legacy code being deprecated

**Example:**
```yaml
breaking:
  ignore:
    - "experimental/"      # Ignore experimental features
    - "internal/"         # Ignore internal APIs
    - "proto/v2alpha/"    # Ignore alpha version
    - "testdata/"         # Ignore test fixtures
```

## Environment Variables

EasyP respects several environment variables that override or supplement configuration:

| Variable | Description | Example |
|----------|-------------|---------|
| `EASYPPATH` | Custom cache directory location | `/opt/easyp-cache` |
| `EASYP_CONFIG` | Default config file path | `./custom-easyp.yaml` |
| `NO_COLOR` | Disable colored output | `1` or `true` |
| `EASYP_DEBUG` | Enable debug logging | `1` or `true` |

## Configuration Best Practices

### 1. Version Control

Always commit your `easyp.yaml` and `easyp.lock` files:
```bash
git add easyp.yaml easyp.lock
git commit -m "Add EasyP configuration"
```

### 2. Use Specific Versions

Pin dependencies to specific versions in production:
```yaml
# Good for production
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  
# Avoid in production
deps:
  - github.com/googleapis/googleapis  # Uses latest
```

### 3. Progressive Linting

Start with minimal rules and gradually add more:
```yaml
# Start with
lint:
  use: [MINIMAL]

# Then add
lint:
  use: [MINIMAL, BASIC]

# Finally
lint:
  use: [DEFAULT, COMMENTS]
```

### 4. Document Exceptions

Always explain why rules are disabled:
```yaml
lint:
  # Legacy API compatibility requires non-standard naming
  except:
    - FIELD_LOWER_SNAKE_CASE  # Old API uses camelCase
    - SERVICE_SUFFIX          # External services don't use suffix
```

### 5. Separate Environments

Use different configs for different environments:
```yaml
# development.easyp.yaml - Permissive for rapid development
lint:
  use: [BASIC]
  allow_comment_ignores: true

# production.easyp.yaml - Strict for quality
lint:
  use: [DEFAULT, COMMENTS]
  allow_comment_ignores: false
```

### 6. Organize Inputs

Group related proto files logically:
```yaml
generate:
  inputs:
    # Public API
    - directory: "api/public"
    
    # Internal services
    - directory: "api/internal"
    
    # Shared types
    - directory: "api/common"
```

## Migration from Other Tools

### From Buf

EasyP is largely compatible with Buf configuration. Key differences:

| Buf | EasyP | Notes |
|-----|-------|-------|
| `buf.yaml` | `easyp.yaml` | Different file name |
| `version: v1` | `version: v1alpha` | Different version scheme |
| `lint.use: [DEFAULT]` | `lint.use: [DEFAULT]` | Same syntax |
| `deps` in `buf.yaml` | `deps` in `easyp.yaml` | Same concept, different sources |
| `buf.gen.yaml` | `generate` in `easyp.yaml` | Unified configuration |

### From Protoc

Replace complex Protoc commands with simple configuration:

**Before (Makefile):**
```makefile
protoc \
  -I proto \
  -I third_party/googleapis \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  proto/**/*.proto
```

**After (easyp.yaml):**
```yaml
deps:
  - github.com/googleapis/googleapis

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
    - name: go-grpc
      out: .
      opts:
        paths: source_relative
```

## Troubleshooting

### Common Configuration Issues

#### "Configuration file not found"

**Problem:** EasyP can't find `easyp.yaml`

**Solution:**
```bash
# Specify config file explicitly
easyp -cfg custom-config.yaml lint

# Or create default config
easyp init
```

#### "Invalid configuration format"

**Problem:** YAML syntax error or missing required fields

**Solution:**
1. Check YAML syntax with a validator
2. Ensure `version: v1alpha` is present
3. Verify proper indentation (spaces, not tabs)

#### "Unknown rule in use"

**Problem:** Specified linting rule doesn't exist

**Solution:**
```yaml
# Check rule name spelling
lint:
  use:
    - DEFAULT        # Correct
    - DEFAULTS       # Wrong - no 'S'
```

#### "Plugin not found"

**Problem:** Code generation plugin isn't installed

**Solution:**
```bash
# Install required plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Configuration Schema

For tooling integration, here's the JSON Schema for `easyp.yaml`:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version"],
  "properties": {
    "version": {
      "type": "string",
      "enum": ["v1alpha"]
    },
    "lint": {
      "type": "object",
      "properties": {
        "use": {
          "type": "array",
          "items": {"type": "string"}
        },
        "enum_zero_value_suffix": {"type": "string"},
        "service_suffix": {"type": "string"},
        "ignore": {
          "type": "array",
          "items": {"type": "string"}
        },
        "except": {
          "type": "array",
          "items": {"type": "string"}
        },
        "allow_comment_ignores": {"type": "boolean"},
        "ignore_only": {
          "type": "object",
          "additionalProperties": {
            "type": "array",
            "items": {"type": "string"}
          }
        }
      }
    },
    "deps": {
      "type": "array",
      "items": {"type": "string"}
    },
    "generate": {
      "type": "object",
      "properties": {
        "inputs": {
          "type": "array",
          "items": {
            "type": "object"
          }
        },
        "plugins": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "out"],
            "properties": {
              "name": {"type": "string"},
              "out": {"type": "string"},
              "opts": {
                "type": "object",
                "additionalProperties": {"type": "string"}
              },
              "url": {"type": "string"},
              "with_imports": {"type": "boolean"}
            }
          }
        }
      }
    },
    "breaking": {
      "type": "object",
      "properties": {
        "against_git_ref": {"type": "string"},
        "ignore": {
          "type": "array",
          "items": {"type": "string"}
        }
      }
    }
  }
}
```

This configuration reference provides complete documentation for all EasyP configuration options. For specific use cases and examples, refer to the individual feature documentation sections.