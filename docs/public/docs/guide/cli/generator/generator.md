# Generator

[[toc]]

EasyP includes a powerful generator that simplifies the process of generating code from proto files. By using YAML configurations, EasyP makes this process more user-friendly and intuitive compared to using the protoc command directly.

## Key Features of the Generator

1. **Simplified Code Generation**:
    - Generate code from proto files using a `YAML` configuration.
    - Avoid the need to write long and complex protoc commands.

2. **Wrapper around protoc**:
    - EasyP functions as a wrapper around protoc, providing a more convenient API through configuration files.
    - Supports all options and plugins available in protoc.

3. **Flexibility and Customization**:
    - Use the same parameters as protoc plugins, directly in the configuration file.
    - Support for multiple plugins and their parameters in a single configuration.

4. **Generate Code from Multiple Sources**:
    - Generate code from local directories or remote repositories.
    - Easily integrate with existing projects and repositories.

5. **Remote Generation**:
    - Generate code from remote Git repositories without local checkout.

6. **Package Manager Integration**:
    - Seamless integration with EasyP's package manager for dependency management.
    - Automatic resolution and inclusion of proto dependencies.

## Configuration Reference

### Complete Configuration Example

```yaml
version: v1alpha

# Package manager dependencies
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

# Code generation configuration
generate:
  inputs:
    # Local directory input
    - directory: 
        path: "proto"
        root: "."
    
    # Remote Git repository input
    - git_repo:
        url: "github.com/acme/weather@v1.2.3"
        sub_directory: "proto/api"
        out: "external"
    
    # Another remote repository
    - git_repo:
        url: "https://github.com/company/internal-protos.git"
        sub_directory: "definitions"
        out: "internal"

  plugins:
    # Local plugin execution
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/mycompany/myproject
      with_imports: true
    
    # Local plugin with custom options
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false

  # Managed mode - automatically set file and field options
  managed:
    enabled: true
    disable:
      - module: github.com/googleapis/googleapis  # Disable for specific module
    override:
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
      - file_option: java_package_prefix
        value: com.mycompany
      - file_option: csharp_namespace_prefix
        value: MyCompany
      - field_option: jstype
        value: JS_STRING
        path: api/v1/  # Apply to specific path

```

### Input Sources

#### Local Directory Input

Local directory input is the most common and straightforward way to specify proto files for generation. Use this when your proto files are already present in your project directory structure.

**When to use:**
- Proto files are part of your project repository
- You need full control over proto file organization
- Working with a single service or application
- Proto files don't change frequently

```yaml
inputs:
  - directory: "proto"                    # Simple string format
  
  # OR detailed format
  - directory:
      path: "proto"                       # Path to proto files
      root: "."                          # Root directory for import resolution
```

The `root` parameter is particularly useful in monorepo setups where you need to control import path resolution. When `root` is set to a parent directory, import paths in your proto files will be resolved relative to that root, not the current working directory.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `path` | string | ✅ | - | Path to the directory containing proto files |
| `root` | string | ❌ | `"."` | Root directory used for import path resolution |

**Examples:**

These examples show the difference between basic directory specification and advanced configuration with custom import resolution:

```yaml
# Basic usage - Simple path specification
inputs:
  - directory: "api/proto"

# Advanced usage with custom root - Controls import path resolution
inputs:
  - directory:
      path: "services/auth/proto" 
      root: "services/auth"        # Imports will be relative to this path
```

#### Remote Git Repository Input

Remote Git repository input allows you to generate code from proto files hosted in external repositories without requiring a local checkout. This is particularly powerful for consuming APIs from other teams or external vendors.

**When to use:**
- Consuming proto definitions from other teams or services
- Integrating with vendor APIs that provide proto definitions
- Working with shared proto libraries across multiple projects
- You want to ensure you're always using the correct version of external APIs

**Recommended approach:**
- Always pin to specific versions in production (`@v1.0.0`, not latest)
- Use semantic versions when available for easier dependency management
- Prefer public tags over commit hashes for better traceability

```yaml
inputs:
  - git_repo:
      url: "github.com/company/protos@v1.0.0"    # Required: Repository URL with optional version
      sub_directory: "api"                        # Optional: Subdirectory within repo
      out: "external"                            # Optional: Local output directory
```

The `out` parameter controls where proto files are extracted locally. This is useful for organizing different remote sources and avoiding naming conflicts.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `url` | string | ✅ | - | Git repository URL with optional version/tag/commit |
| `sub_directory` | string | ❌ | `""` | Subdirectory within the repository containing proto files |
| `out` | string | ❌ | `""` | Local directory where proto files will be extracted |

**URL Format Options:**

EasyP supports multiple URL formats to give you flexibility in how you reference remote repositories. Each format serves different use cases depending on your stability and versioning requirements:

- **Tagged versions** are recommended for production as they provide stable, immutable references
- **Semantic versions** offer better readability and dependency management
- **Commit hashes** give you access to specific commits when tags aren't available
- **Latest** should only be used in development environments due to unpredictability
- **Full HTTPS URLs** are useful for private repositories or non-GitHub hosting

```yaml
# Tag version - Best for production stability
url: "github.com/googleapis/googleapis@common-protos-1_3_1"

# Semantic version - Easier dependency management
url: "github.com/grpc-ecosystem/grpc-gateway@v2.19.1"  

# Commit hash - When you need a specific commit
url: "github.com/company/protos@abc123def456"

# Latest - Development only, not recommended for production
url: "github.com/company/protos"

# Full HTTPS URL - For private repos or custom Git hosts
url: "https://github.com/company/private-protos.git"
```

**Examples:**

These examples demonstrate common patterns for consuming remote proto definitions from different types of repositories:

```yaml
# Public repository with specific version - Most common for external APIs
inputs:
  - git_repo:
      url: "github.com/googleapis/googleapis@common-protos-1_3_1"
      sub_directory: "google"
      out: "googleapis"

# Private repository with authentication - For internal company APIs
inputs:
  - git_repo:
      url: "github.com/mycompany/internal-protos@v2.1.0"
      sub_directory: "api/definitions"
      out: "internal"

# Multiple remote sources - Common in microservices architectures
inputs:
  - git_repo:
      url: "github.com/grpc-ecosystem/grpc-gateway@v2.19.1"
      sub_directory: "protoc-gen-openapiv2/options"
      out: "gateway"
  - git_repo:
      url: "github.com/bufbuild/protoc-gen-validate@v0.10.1"  
      sub_directory: "validate"
      out: "validate"
```

### Plugin Configuration

Plugin configuration is where you specify which code generators to run and how they should behave. EasyP supports any protoc plugin, making it extremely flexible for different language ecosystems and use cases.

At a high level, there are **four ways to specify how a plugin should be executed**:

- **`name`** – run plugin by name from `PATH` or use a builtin plugin.
- **`path`** – run plugin by absolute/relative path to an executable file.
- **`remote`** – run plugin via remote URL (EasyP remote executor).
- **`command`** – run plugin via arbitrary command (for example, `go run ...`).

Only **one** of `name`, `path`, `remote`, or `command` must be specified for each plugin.

#### Plugin by Name (`name`)

Local plugin execution by name is the standard approach where plugins are installed on your system and executed directly by EasyP.

**When to use `name`:**
- Standard language support (Go, Python, TypeScript, etc.)
- You have control over the build environment
- Performance is critical (no network overhead)
- You want to rely on PATH or builtin plugins

**Installation requirements:**
- Plugins must be installed and available in your PATH (if not using builtin plugins).
- Plugin names follow the `protoc-gen-{name}` convention.
- Use package managers (go install, npm install, pip install) for installation.

```yaml
plugins:
  - name: go                              # Plugin name (must be installed locally or builtin)
    out: ./generated                      # Output directory
    opts:                                 # Plugin-specific options
      paths: source_relative
      module: github.com/mycompany/project
    with_imports: true                    # Include dependency imports
```

The `with_imports` parameter is crucial when you're using dependencies from the package manager. Set it to `true` to include proto files from your `deps` section in the generation process.

#### Plugin by Path (`path`)

Sometimes you need to run a plugin from a specific binary without putting it into `PATH` (for example, a binary in your repo or in a build directory). In this case you can specify an explicit path:

```yaml
plugins:
  - path: ./bin/protoc-gen-my-custom
    out: ./gen/custom
    opts:
      foo: bar
```

**When to use `path`:**
- You keep plugin binaries in the repository (for reproducible builds).
- You use different plugin versions side-by-side.
- You don't want to pollute the global `PATH`.

#### Remote Plugin (`remote`)

Remote plugins are executed via a remote URL. EasyP will send a `CodeGeneratorRequest` to the remote endpoint and receive a `CodeGeneratorResponse` back.

Below is a real example of using the EasyP API Service as a remote plugin executor (same format as in the API Service docs):

```yaml
generate:
  plugins:
    # Remote plugin execution via EasyP API Service
    - remote: api.easyp.tech/protobuf/go:v1.36.10
      out: .
      opts:
        paths: source_relative

    - remote: api.easyp.tech/grpc/go:v1.5.1
      out: .
      opts:
        paths: source_relative
```

**Typical use cases for `remote`:**
- Centralized plugin service inside your organization (e.g. EasyP API Service).
- Running heavy plugins on a dedicated server instead of CI agents.
- Sharing the same plugin implementation across multiple teams.

#### Executing Plugin via Command (`command`)

You can specify a plugin as an array of commands to execute. This is useful for running plugins via `go run` or any other tool without prior installation of the plugin binary:

```yaml
plugins:
  - command: ["go", "run", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]
    out: ./gen/go
    opts:
      paths: source_relative
```

In this mode EasyP:
- **builds and runs the command** you provide as a child process;
- **writes `CodeGeneratorRequest` to stdin** of the process;
- **reads `CodeGeneratorResponse` from stdout**, just like with regular protoc plugins.

**Plugin Source Priority:**
1. `command` — execute via specified command (highest priority)
2. `remote` — remote plugin via URL
3. `name` — local plugin from PATH or builtin plugin
4. `path` — path to plugin executable file

**Parameters (plugin sources and common options):**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | string | ❌ | - | Plugin name or identifier (e.g. `go`, `go-grpc`, `grpc-gateway`) |
| `command` | []string | ❌ | - | Command to execute plugin (e.g., `["go", "run", "package"]`) |
| `remote` | string | ❌ | - | Remote plugin URL |
| `path` | string | ❌ | - | Path to plugin executable file |
| `out` | string | ✅ | - | Output directory for generated files |
| `opts` | map[string]string | ❌ | `{}` | Plugin-specific options (mapped to `--opt=value`) |
| `with_imports` | bool | ❌ | `false` | Include proto files from dependencies |

**Note:** Only one plugin source (`name`, `command`, `remote`, or `path`) must be specified for each plugin.

**Command source examples:**

```yaml
generate:
  inputs:
    - directory: "proto"

  plugins:
    # 1) gRPC-Gateway via go run (no pre-installed binary required)
    - command: ["go", "run", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]
      out: ./gen/go
      opts:
        paths: source_relative
        generate_unbound_methods: true

    # 2) protoc-gen-validate via go run
    - command: ["go", "run", "github.com/bufbuild/protoc-gen-validate@v0.10.1"]
      out: ./gen/go
      opts:
        paths: source_relative
        lang: go

    # 3) Any custom wrapper script
    - command: ["bash", "./scripts/custom-protoc-plugin.sh"]
      out: ./gen/custom
      opts:
        foo: bar
```

#### Builtin Plugins

EasyP includes builtin plugins for basic protobuf and gRPC languages. These plugins are embedded in the binary as WASM modules and do not require installation of external dependencies.

**Benefits of builtin plugins:**
- **Portability**: Single binary with all necessary plugins
- **Convenience**: No need to install external dependencies
- **Stability**: Plugin versions are fixed in the binary
- **Isolation**: Independent of system plugin installations

**Supported builtin plugins:**

#### Protobuf Base Plugins

The following plugins are builtin for generating base protobuf code:

| Plugin Name | Description | Corresponding protoc Plugin |
|-------------|-------------|----------------------------|
| `cpp` | Generate C++ code from proto files | `protoc-gen-cpp` |
| `csharp` | Generate C# code from proto files | `protoc-gen-csharp` |
| `java` | Generate Java code from proto files | `protoc-gen-java` |
| `kotlin` | Generate Kotlin code from proto files | `protoc-gen-kotlin` |
| `objc` | Generate Objective-C code from proto files | `protoc-gen-objc` |
| `php` | Generate PHP code from proto files | `protoc-gen-php` |
| `python` | Generate Python code from proto files | `protoc-gen-python` |
| `ruby` | Generate Ruby code from proto files | `protoc-gen-ruby` |

#### gRPC Plugins

The following plugins are builtin for generating gRPC code:

| Plugin Name | Description | Corresponding protoc Plugin |
|-------------|-------------|----------------------------|
| `grpc_cpp` | Generate gRPC code for C++ | `grpc_cpp_plugin` |
| `grpc_csharp` | Generate gRPC code for C# | `grpc_csharp_plugin` |
| `grpc_java` | Generate gRPC code for Java | `grpc_java_plugin` |
| `grpc_node` | Generate gRPC code for Node.js | `grpc_node_plugin` |
| `grpc_objc` | Generate gRPC code for Objective-C | `grpc_objective_c_plugin` |
| `grpc_php` | Generate gRPC code for PHP | `grpc_php_plugin` |
| `grpc_python` | Generate gRPC code for Python | `grpc_python_plugin` |
| `grpc_ruby` | Generate gRPC code for Ruby | `grpc_ruby_plugin` |

**Plugin Selection Logic:**

EasyP uses the following priority when selecting an executor for a plugin:

1. **Remote plugin** (if `url` is specified) — always has the highest priority
2. **Builtin plugin** (if plugin is builtin and not found in PATH) — used automatically
3. **Local plugin** (from PATH) — used by default for backward compatibility

**Usage Example:**

```yaml
generate:
  inputs:
    - directory: "proto"
  plugins:
    # Builtin Python plugin (automatically used if protoc-gen-python is not found in PATH)
    - name: python
      out: ./gen/python
      opts:
        pyi_out: ./gen/python
    
    # Builtin gRPC plugin for Python
    - name: grpc_python
      out: ./gen/python
    
    # Builtin C++ plugin (automatically used if protoc-gen-cpp is not found in PATH)
    - name: cpp
      out: ./gen/cpp
      opts:
        dllexport_decl: EXPORT
    
    # Builtin gRPC plugin for C++
    - name: grpc_cpp
      out: ./gen/cpp
```

**Requirements:**

Builtin plugins are included in the EasyP binary:

```bash
# Build
go build ./cmd/easyp

# Install
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

**Backward Compatibility:**

Builtin plugins are fully compatible with existing configurations. If a plugin is found in PATH, it will be used instead of the builtin one. This ensures that:

- Existing configurations continue to work without changes
- You can override a builtin plugin by installing it in your system
- Priority is given to local installations for flexibility

### Plugin Options Reference

This section covers the most commonly used plugins and their configuration options. Each plugin has specific parameters that control how code is generated, and understanding these options is crucial for getting the output you need.

#### Go Plugins

Go plugins are the most mature and widely used protoc plugins. The `paths` option controls how import paths are resolved, while other options provide fine-grained control over the generated code:

```yaml
plugins:
  # protoc-gen-go - Generates Go structs and basic protobuf functionality
  - name: go
    out: ./gen/go
    opts:
      paths: source_relative              # source_relative | import
      module: github.com/company/project  # Go module path for import generation
      
  # protoc-gen-go-grpc - Generates gRPC service stubs and clients
  - name: go-grpc
    out: ./gen/go
    opts:
      paths: source_relative
      require_unimplemented_servers: false  # Generate UnimplementedServer embedding
```

#### gRPC-Gateway Plugins

gRPC-Gateway plugins enable you to serve gRPC services as REST APIs and generate OpenAPI documentation. These are essential for building HTTP/JSON APIs from gRPC services:

```yaml
plugins:
  # protoc-gen-grpc-gateway - Generates REST-to-gRPC reverse proxy
  - name: grpc-gateway
    out: ./gen/go
    opts:
      paths: source_relative
      generate_unbound_methods: true      # Include methods without HTTP bindings
      
  # protoc-gen-openapiv2 - Generates OpenAPI/Swagger documentation
  - name: openapiv2  
    out: ./gen/openapi
    opts:
      simple_operation_ids: true          # Use simple names for operation IDs
      generate_unbound_methods: false     # Exclude methods without HTTP bindings
      json_names_for_fields: true         # Use JSON names instead of proto names
```

#### Validation Plugins

Validation plugins generate code that automatically validates proto message fields based on constraints defined in your proto files. This eliminates the need for manual validation code:

```yaml
plugins:
  # protoc-gen-validate - Generates field validation code
  - name: validate-go
    out: ./gen/go
    opts:
      paths: source_relative
      lang: go                           # Target language for validation code
```

#### TypeScript/JavaScript Plugins

TypeScript plugins are essential for frontend development, providing type-safe interfaces for your proto definitions and gRPC services in web applications:

```yaml
plugins:
  # protoc-gen-ts - Generates TypeScript definitions and serialization
  - name: ts
    out: ./gen/typescript  
    opts:
      declaration: true                   # Generate .d.ts type definition files
      target: es2017                     # ECMAScript target for compatibility
      
  # protoc-gen-grpc-web - Generates gRPC-Web clients for browsers
  - name: grpc-web
    out: ./gen/web
    opts:
      import_style: typescript           # Module system for generated code
      mode: grpcweb                      # Transport mode for gRPC-Web
```

## Managed Mode

Managed mode automatically sets file and field options in your protobuf descriptors during code generation without modifying the original `.proto` files. This feature is compatible with `buf`'s managed mode and provides a consistent way to manage language-specific options across your codebase.

**Key benefits:**
- **No proto file modifications**: Options are applied at generation time, keeping your proto files clean
- **Consistent defaults**: Automatic application of language-specific naming conventions
- **Centralized configuration**: Manage all options in one place (`easyp.yaml`)
- **Module-specific rules**: Apply different options to different modules or paths
- **buf compatibility**: Works the same way as `buf` managed mode

### How It Works

When managed mode is enabled, EasyP automatically applies file and field options to your protobuf descriptors before code generation. This happens in memory, so your original `.proto` files remain unchanged.

**Default values** are applied for certain options based on language conventions:
- Java: `java_package_prefix` defaults to `"com"`, `java_multiple_files` defaults to `true`
- C#: `csharp_namespace` defaults to PascalCase of package name
- Ruby: `ruby_package` defaults to PascalCase with `::` separator
- PHP: `php_namespace` defaults to PascalCase with `\` separator
- Objective-C: `objc_class_prefix` defaults to first letters of package parts
- C++: `cc_enable_arenas` defaults to `true`

**Overrides** allow you to set specific values for options, with support for filtering by module, path, or field.

**Disables** allow you to prevent managed mode from modifying specific options or files.

### Configuration

```yaml
generate:
  managed:
    enabled: true
    disable:
      # Disable managed mode for specific module
      - module: github.com/googleapis/googleapis
      
      # Disable specific option globally
      - file_option: java_package_prefix
      
      # Disable for specific path
      - path: legacy/
        file_option: go_package
      
      # Disable field option for specific field
      - field_option: jstype
        field: com.example.User.id
    
    override:
      # Override go_package_prefix for all files
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
      
      # Override for specific module
      - file_option: java_package_prefix
        value: com.mycompany
        module: github.com/mycompany/internal-protos
      
      # Override for specific path
      - file_option: csharp_namespace_prefix
        value: MyCompany
        path: api/v1/
      
      # Override for multiple files with same value using prefix path
      # This matches both internal/cms/bmi.proto and internal/cms/bmi_service.proto
      - file_option: go_package
        value: spec/cms/bmi
        path: "internal/cms/bmi"
      
      # Override field option for specific path
      - field_option: jstype
        value: JS_STRING
        path: api/v1/
      
      # Override for specific field
      - field_option: jstype
        value: JS_NUMBER
        field: com.example.User.big_id
```

### Supported File Options

| Option | Description | Has Default? |
|--------|-------------|--------------|
| `go_package` | Go import path | ❌ |
| `go_package_prefix` | Prefix for Go import paths | ❌ |
| `java_package` | Java package name | ❌ |
| `java_package_prefix` | Prefix for Java packages | ✅ (`"com"`) |
| `java_package_suffix` | Suffix for Java packages | ❌ |
| `java_multiple_files` | Generate multiple Java files | ✅ (`true`) |
| `java_outer_classname` | Outer class name | ✅ (PascalCase + "Proto") |
| `java_string_check_utf8` | UTF-8 validation | ❌ |
| `csharp_namespace` | C# namespace | ✅ (PascalCase) |
| `csharp_namespace_prefix` | Prefix for C# namespaces | ❌ |
| `ruby_package` | Ruby module name | ✅ (PascalCase with `::`) |
| `ruby_package_suffix` | Suffix for Ruby packages | ❌ |
| `php_namespace` | PHP namespace | ✅ (PascalCase with `\`) |
| `php_metadata_namespace` | PHP metadata namespace | ❌ |
| `php_metadata_namespace_suffix` | Suffix for PHP metadata | ❌ |
| `objc_class_prefix` | Objective-C class prefix | ✅ (First letters) |
| `swift_prefix` | Swift prefix | ❌ |
| `optimize_for` | Code generation optimization | ❌ |
| `cc_enable_arenas` | C++ arena allocation | ✅ (`true`) |

### Supported Field Options

| Option | Description | Applies To |
|--------|-------------|------------|
| `jstype` | JavaScript type for 64-bit integers | `int64`, `uint64`, `sint64`, `fixed64`, `sfixed64` |

### Examples

#### Basic Setup with Defaults

Enable managed mode to get automatic defaults for all supported languages:

```yaml
generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
    - name: java
      out: ./gen/java
    - name: csharp
      out: ./gen/csharp
  managed:
    enabled: true
```

This will automatically:
- Set `java_package` to `com.<package>` for all files
- Set `java_multiple_files` to `true`
- Set `csharp_namespace` to PascalCase of package name
- Set `ruby_package` to PascalCase with `::` separator
- And more...

#### Custom Go Package Prefix

Override the Go package prefix for your project:

```yaml
generate:
  managed:
    enabled: true
    override:
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
```

This will set `go_package` to `github.com/mycompany/myproject/gen/go/<package>` for all files.

#### Dynamic Go Package Paths with Markers

For more complex path generation, you can use markers in `go_package_prefix` or `go_package` values:

```yaml
generate:
  managed:
    enabled: true
    override:
      # Use file path directly (without .proto extension)
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_path}}
      
      # Use only directory path
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_dir}}
      
      # Remove prefix from directory path
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_dir_without:internal/}}
      
      # Remove prefix from full file path
      - file_option: go_package_prefix
        value: github.com/mycompany/{{file_path_without:internal/}}
```

**Available markers:**
- `{{file_path}}` - Full file path without `.proto` extension
  - Example: `internal/cms/as.proto` → `internal/cms/as`
- `{{file_dir}}` - Directory path only, without filename
  - Example: `internal/cms/as.proto` → `internal/cms`
- `{{file_dir_without:prefix/}}` - Directory path with prefix removed, and base filename without `_service`/`_grpc` suffixes
  - Example: `{{file_dir_without:internal/}}` for `internal/cms/as_service.proto` → `cms/as`
- `{{file_path_without:prefix/}}` - Full file path with prefix removed
  - Example: `{{file_path_without:internal/}}` for `internal/cms/as.proto` → `cms/as`

#### Module-Specific Overrides

Apply different options to different modules:

```yaml
generate:
  managed:
    enabled: true
    override:
      # Default for all files
      - file_option: go_package_prefix
        value: github.com/mycompany/myproject/gen/go
      
      # Specific override for internal module
      - file_option: go_package_prefix
        value: github.com/mycompany/internal/gen/go
        module: github.com/mycompany/internal-protos
```

#### Disabling for External Dependencies

Disable managed mode for external dependencies that already have their options set:

```yaml
generate:
  managed:
    enabled: true
    disable:
      - module: github.com/googleapis/googleapis
      - module: github.com/grpc-ecosystem/grpc-gateway
```

#### JavaScript Type Safety

Set `jstype` to `JS_STRING` for all 64-bit integer fields to prevent precision loss in JavaScript:

```yaml
generate:
  managed:
    enabled: true
    override:
      - field_option: jstype
        value: JS_STRING
        path: api/v1/  # Apply to specific path
```

### Path Matching

Path matching in managed mode uses prefix-based matching (same as `buf`):

- **Directory path** (ending with `/`): Matches all files in that directory and subdirectories
  - Example: `path: "internal/cms/"` matches `internal/cms/as.proto`, `internal/cms/node.proto`, `internal/cms/v1/service.proto`
- **Exact file path** (ending with `.proto`): Matches only that specific file
  - Example: `path: "internal/cms/as.proto"` matches only `internal/cms/as.proto`
- **Prefix path** (no trailing `/` or `.proto`): Uses prefix matching (not directory-aware)
  - Example: `path: "internal/cms"` matches `internal/cms/as.proto` but also `internal/cmsv2/file.proto`

### Rule Precedence

When multiple rules match the same file or field, the following precedence applies:

1. **Disable rules** take precedence - if an option is disabled, it won't be applied
2. **Override rules** are applied in order - the last matching rule wins
3. **Default values** are applied only if no override matches and the option isn't disabled

### Compatibility with buf

EasyP's managed mode is compatible with `buf`'s managed mode. The same configuration format and behavior apply, making it easy to migrate between tools or use both in the same workflow.

## Descriptor Set Generation

**https://protobuf.dev/programming-guides/techniques/#self-description**

EasyP supports generating binary FileDescriptorSet files using the `--descriptor_set_out` flag. This allows you to create self-describing protobuf messages that include schema information alongside the data.

**CLI flags:**

- `--descriptor_set_out <path>` - Output path for the binary FileDescriptorSet
- `--include_imports` - Include all transitive dependencies in the FileDescriptorSet

**Example:**

```bash
# Generate descriptor set with only target files
easyp generate --descriptor_set_out=./schema.pb

# Generate descriptor set with all dependencies
easyp generate --descriptor_set_out=./schema.pb --include_imports
```

Self-describing messages are useful for dynamic message parsing, runtime schema validation, schema registries, and building generic gRPC clients. For more information, see the [Protocol Buffers documentation on self-description](https://protobuf.dev/programming-guides/techniques/#self-description).

## Package Manager Integration

One of EasyP's most powerful features is the seamless integration between the package manager and code generator. This integration eliminates the common problem of managing proto dependencies manually and ensures that your generated code always has access to the correct versions of imported proto files.

**Key benefits:**
- **Automatic dependency resolution**: No need to manually manage proto import paths
- **Version consistency**: Dependencies are locked to specific versions via `easyp.lock`
- **Transitive dependencies**: EasyP handles dependencies of dependencies automatically
- **Performance**: Local caching means dependencies are downloaded once and reused

### Automatic Dependency Resolution

When you define dependencies in the `deps` section, the generator automatically includes them in the proto path. This means your proto files can import from these dependencies without any additional configuration.

**How it works:**
1. EasyP downloads and caches dependencies based on your `deps` configuration
2. During generation, these cached proto files are automatically added to the protoc import path
3. Your proto files can import from dependencies using standard import statements
4. Generated code includes both your local protos and dependency protos when `with_imports: true`

Here's a simple example showing how dependency resolution works automatically. Notice that you only need to specify the dependencies once in the `deps` section:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true    # This automatically includes googleapis and grpc-gateway protos
```

### Dependency Usage Examples

The following examples demonstrate common patterns for integrating external proto dependencies into your code generation workflow.

#### Using Google APIs

Google APIs are among the most commonly used proto dependencies, providing standard types and annotations for REST APIs, field validation, and common data structures.

**When to use Google APIs:**
- Building REST APIs with gRPC-Gateway
- Need standard types like `Timestamp`, `Duration`, `Any`
- Want to use Google's field behavior annotations
- Building services that integrate with Google Cloud

This configuration shows the minimal setup needed to use Google APIs in your proto files:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "api/proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```

**Important:** Always use a pinned version (like `common-protos-1_3_1`) rather than latest to ensure build reproducibility.

Once configured, your proto files can import and use Google API definitions. Here's an example of a service using HTTP annotations:

```proto
// api/proto/service.proto
syntax = "proto3";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

service MyService {
  rpc GetData(GetDataRequest) returns (GetDataResponse) {
    option (google.api.http) = {
      get: "/v1/data"
    };
  }
}
```

#### Using Validation Rules

Protoc-gen-validate provides powerful field validation capabilities that can be embedded directly in your proto definitions, eliminating the need for separate validation logic in your application code.

**When to use validation:**
- Input validation for API endpoints
- Database model constraints
- Configuration file validation
- Any scenario where data integrity is critical

**Benefits:**
- Validation rules are part of the proto definition (single source of truth)
- Code generation creates validation functions automatically
- Consistent validation across different languages
- Better performance than runtime reflection-based validation

```yaml
deps:
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
    - name: validate-go
      out: ./gen/go  
      opts:
        paths: source_relative
```

**Note:** You need both the dependency (for proto imports) and the plugin (for code generation) to get full validation support.

Here's how validation rules look in your proto files. The generated code will automatically validate these constraints:

```proto
// proto/user.proto
syntax = "proto3";

import "validate/validate.proto";

message User {
  string email = 1 [(validate.rules).string.email = true];
  int32 age = 2 [(validate.rules).int32.gte = 0];
}
```

#### Complex Multi-Dependency Setup

This example demonstrates a production-ready configuration that combines multiple dependencies and plugins for a complete API development workflow:

```yaml
deps:
  # Core Google APIs - Standard types and HTTP annotations
  - github.com/googleapis/googleapis@common-protos-1_3_1
  
  # gRPC Gateway for REST APIs - Enables HTTP/JSON interfaces
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  
  # Validation rules - Field-level validation constraints
  - github.com/bufbuild/protoc-gen-validate@v0.10.1
  
  # Company internal shared types - Common business objects
  - github.com/mycompany/shared-protos@v1.5.0

generate:
  inputs:
    - directory: "api/proto"
  plugins:
    # Go code generation - Core protobuf structures
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/mycompany/myservice
      with_imports: true
      
    # gRPC service stubs - Server and client interfaces
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false
        
    # REST Gateway - HTTP-to-gRPC proxy code
    - name: grpc-gateway
      out: ./gen/go
      opts:
        paths: source_relative
        
    # OpenAPI documentation - API specification
    - name: openapiv2
      out: ./gen/openapi
      opts:
        simple_operation_ids: true
        
    # Validation code - Input validation functions
    - name: validate-go
      out: ./gen/go
      opts:
        paths: source_relative
```

### Dependency Cache Integration

The generator leverages EasyP's module cache for fast builds:

```bash
# Download dependencies once
easyp mod download

# Generate code (uses cached dependencies)
easyp generate

# Dependencies are cached in ~/.easyp/mod/
ls ~/.easyp/mod/github.com/googleapis/googleapis/
```

## Remote Generation

Remote generation is a powerful feature that allows you to generate code from proto files hosted in remote Git repositories without requiring a local checkout. This enables true microservices architecture where teams can consume each other's APIs without tight coupling.

**Key advantages:**
- **Decoupled development**: Teams can work independently while consuming each other's APIs
- **Version control**: Pin to specific versions of external APIs for stability
- **Reduced repository size**: No need to vendor or submodule external proto files
- **Automatic updates**: Easy to update to newer versions when ready

**Best practices:**
- Always use tagged versions in production environments
- Test with latest versions in development, but pin in production
- Use semantic versioning when available for easier dependency management
- Consider the network implications for CI/CD systems

### Remote Proto Sources

Generate from remote repositories directly. This is particularly useful in microservices architectures where different teams own different proto definitions.

**Typical workflow:**
1. Team A publishes proto definitions in a Git repository with proper versioning
2. Team B references these protos in their `easyp.yaml` configuration
3. During generation, EasyP automatically fetches and uses the remote protos
4. Generated code includes client libraries for Team A's services

Here's a practical example showing how to combine local and remote proto sources in a single generation configuration:

```yaml
generate:
  inputs:
    # Local protos - Your service's own API definitions
    - directory: "proto"
    
    # Remote public repository - External vendor API
    - git_repo:
        url: "github.com/acme/weather-api@v2.1.0"
        sub_directory: "proto/weather/v1"
        out: "external/weather"
    
    # Remote private repository - Internal company API
    - git_repo:
        url: "github.com/mycompany/internal-apis@main"
        sub_directory: "user-service/proto"
        out: "internal/user"
        
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
      with_imports: true
```



### Remote Generation Use Cases

These examples demonstrate real-world scenarios where remote generation provides significant value in distributed development environments.

#### Multi-Team Development

Multi-team development is where remote generation truly shines. Instead of coordinating shared repositories or complex dependency management, teams can independently evolve their APIs while consumers automatically get updates through versioned dependencies.

This pattern is especially valuable in large organizations where:
- Teams have different release cycles and development velocities
- API ownership is clearly defined but consumption is widespread  
- You want to avoid the overhead of coordinating shared proto repositories
- Different teams use different technology stacks but need to communicate

```yaml
# Team A (Order Service) generates from Team B's proto definitions
generate:
  inputs:
    # Local service definitions - APIs owned by this team
    - directory: "proto/orders"
    
    # User service protos from another team - Stable, versioned API
    - git_repo:
        url: "github.com/company/user-service@v1.8.0"  
        sub_directory: "api/proto"
        out: "external/users"
        
    # Payment service protos - Different team, different version
    - git_repo:
        url: "github.com/company/payment-service@v2.3.1"
        sub_directory: "proto/payment/v2"  
        out: "external/payments"
        
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/order-service
      with_imports: true
```

#### Vendor API Integration

Many vendors now provide proto definitions for their APIs, enabling strongly-typed client generation instead of relying on hand-written HTTP clients. This approach provides better type safety, automatic serialization, and often better performance.

Benefits of using vendor proto definitions:
- **Type safety**: Compile-time checking of API calls and data structures
- **Automatic updates**: New API features become available through version updates
- **Consistency**: Same interface patterns across different vendor integrations
- **Performance**: Binary serialization is often faster than JSON
- **Documentation**: Proto files serve as authoritative API documentation

```yaml
# Generate clients for external vendor APIs
generate:
  inputs:
    # Vendor's public proto definitions - Financial services
    - git_repo:
        url: "github.com/stripe/stripe-proto@v1.0.0"
        sub_directory: "proto"
        out: "vendor/stripe"
        
    # Communication service APIs - SMS/Voice integration
    - git_repo:  
        url: "github.com/twilio/twilio-protos@v2.1.0"
        sub_directory: "definitions"
        out: "vendor/twilio"
        
  plugins:
    - name: go
      out: ./clients/go
      opts:
        paths: source_relative
        module: github.com/mycompany/integrations
```



## Commands

The EasyP command-line interface provides flexible options for running code generation with different configurations and environments.

### Basic Generation

These are the most commonly used command patterns for everyday development and production use:

```bash
# Use default easyp.yaml configuration - Most common for development
easyp generate

# Use custom configuration file - Essential for multi-environment setups  
easyp -cfg production.easyp.yaml generate

# Generate with verbose output - Helpful for debugging and CI/CD
easyp -v generate

# Generate with custom cache location - Useful for CI systems or shared environments
EASYPPATH=/tmp/easyp-cache easyp generate
```

### Integration with Package Manager

EasyP's package manager integration means you can either explicitly manage dependencies or let the generator handle them automatically. The explicit approach gives you more control, while the automatic approach is more convenient:

```bash
# Explicit workflow - Better for CI/CD and when you want to cache dependencies
easyp mod download    # Download and cache dependencies first
easyp generate        # Generate code using cached dependencies

# Automatic workflow - Convenient for development (generate downloads dependencies automatically)
easyp generate
```

### Advanced Usage

These advanced usage patterns are useful for specific deployment scenarios, debugging, or when you need fine-grained control over the generation process:

```bash
# Generate from specific input directory - Override config file settings
easyp generate --input-dir=./api/proto

# Generate using vendored dependencies - For offline builds or Docker containers
easyp mod vendor
easyp -I easyp_vendor generate

# Generate with custom protoc path - When using custom or newer protoc versions
PROTOC_PATH=/usr/local/bin/protoc easyp generate
```

## Common Patterns

These patterns represent real-world scenarios and best practices for organizing code generation in different project structures.



### Multi-Language Generation

Multi-language generation is essential for organizations using different technologies across their stack. EasyP makes it easy to generate consistent client libraries and types for multiple programming languages from the same proto definitions.

**Common scenarios:**
- **Full-stack applications**: Go/Java backend with TypeScript frontend for web apps
- **Data platforms**: Go services with Python data science tools and analysis scripts  
- **Microservices**: Different services implemented in optimal languages for their domain
- **Client libraries**: Providing SDKs in multiple languages for external developers
- **Legacy integration**: Modern gRPC services with legacy systems using different languages

**Performance considerations:**
- Each plugin runs independently, so generation time scales linearly with plugin count
- Consider using parallel execution (`make -j4`) for large numbers of plugins
- Output directories should be organized hierarchically to avoid file conflicts
- Some plugins are faster than others - profile your build to identify bottlenecks

**Maintenance benefits:**
- Single source of truth for API definitions prevents schema drift
- Consistent types across all languages reduce integration bugs
- Automatic synchronization when proto definitions change eliminates manual updates
- Reduced chance of API drift between different language implementations
- Easier refactoring since changes propagate to all generated code

This example shows a typical multi-language setup for a full-stack application with backend services, web frontend, data analysis, and documentation:

```yaml
generate:
  inputs:
    - directory: "proto"
    
  plugins:
    # Go backend services - Primary implementation language
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/backend
      with_imports: true
      
    - name: go-grpc  # gRPC server and client stubs
      out: ./gen/go
      opts:
        paths: source_relative
        
    # TypeScript frontend - Web application client code
    - name: ts  
      out: ./gen/typescript
      opts:
        declaration: true       # Generate type definitions
        target: es2020         # Modern JavaScript for browsers
        
    # Python data science - Analytics and ML workflows
    - name: python
      out: ./gen/python
      opts:
        mypy_out: ./gen/python-stubs  # Type checking support
        
    # Documentation - API reference for developers
    - name: doc
      out: ./docs/api
      opts:
        markdown: true         # Generate markdown documentation
```

**Organization tip:** Use separate output directories for each language to avoid file conflicts and make it easier to integrate with language-specific build systems.





The EasyP generator provides a comprehensive solution for protocol buffer code generation, supporting everything from simple local development to complex enterprise multi-language workflows with remote dependencies.