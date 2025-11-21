# EasyP

**Modern Protocol Buffers toolkit for streamlined development workflows**

EasyP is a comprehensive CLI tool that simplifies working with Protocol Buffers by providing linting, package management, code generation, and compatibility checking in a single, unified interface.

📖 **[Documentation](https://easyp.tech)** | 💬 **[Community](https://t.me/easyptech)**

---

## 🤔 Why EasyP?

### The Problem: Fragmented Proto Workflows

Working with Protocol Buffers across projects creates complexity:

**Tool Chaos**
- Multiple tools for different tasks (linting, generation, dependency management)
- Inconsistent configurations across projects and teams
- Complex setup for new developers joining projects

**Dependency Management Issues**  
- No standardized way to manage proto dependencies
- Version conflicts between different proto libraries
- Difficult to reproduce builds across environments

**Development Friction**
- Manual processes for common proto workflows
- No unified configuration format
- Time-consuming compatibility checks

### The Solution: Unified Proto Toolkit

EasyP consolidates your entire Protocol Buffers workflow into a single, powerful CLI tool with:

- **One Tool, All Tasks**: Linting, generation, dependency management, and compatibility checking
- **Standardized Configuration**: Single `easyp.yaml` file for all proto workflows  
- **Git-Native Dependencies**: Direct integration with Git repositories for proto dependencies
- **Remote Plugin Execution**: Centralized plugin execution for consistent results
- **Built-in Best Practices**: Comprehensive linting rules and breaking change detection

---

## 🚀 Overview

### Key Features

- **🔍 Comprehensive Linting** - Built-in support for buf's linting rules with customizable configurations
- **📦 Smart Package Manager** - Git-based dependency management with lock file support
- **⚡ Code Generation** - Multi-language code generation with local and remote plugin support  
- **🔄 Breaking Change Detection** - Automated API compatibility verification against Git branches
- **🌐 Remote Plugin Support** - Execute plugins via centralized EasyP API service
- **🎯 Developer Experience** - Auto-completion, intuitive commands, and clear error messages

### Supported Plugin Types

- **Local Plugins**: Standard protoc plugins installed on your system
- **Remote Plugins**: Plugins executed via EasyP API service for consistent, isolated execution
- **Custom Plugins**: Support for custom plugin development and distribution

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.20+** (for installation from source)
- **Git** (for dependency management)
- **Protocol Buffers Compiler** (protoc) - for local plugin execution

### Installation

#### Install via Homebrew

```bash
brew install easyp-tech/tap/easyp
```

#### Install from GitHub

```bash
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

#### Build from Source

1. Clone the repository:
```bash
git clone https://github.com/easyp-tech/easyp.git
cd easyp
```

2. Build the binary:
```bash
go build ./cmd/easyp
```

### Initialize Your First Project

1. **Create a new project**:
```bash
mkdir my-proto-project
cd my-proto-project
easyp init
```

2. **Add your proto files**:
```bash
mkdir api
# Add your .proto files to the api/ directory
```

3. **Configure dependencies** (edit `easyp.yaml`):
```yaml
version: v1alpha

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "api"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
```

4. **Download dependencies and generate code**:
```bash
easyp mod download
easyp generate
```

---

## 📋 Commands

### Project Initialization

#### `easyp init`

Creates a new EasyP project with default configuration.

```bash
easyp init [flags]
```

**Flags:**
- `-d, --dir string` - Directory path to initialize (default: ".")

**What it creates:**
- `easyp.yaml` - Project configuration file
- `easyp.lock` - Dependency lock file (created after first `mod download`)

**Example:**
```bash
# Initialize in current directory
easyp init

# Initialize in specific directory  
easyp init --dir ./my-project
```

---

### Linting

#### `easyp lint`

Validates proto files against configurable linting rules.

```bash
easyp lint [flags]
```

**Flags:**
- `-p, --path string` - Path to directory with proto files (default: ".")
- `-f, --format string` - Output format: `text` or `json` (default: "text")
- `--cfg string` - Configuration file path (default: "easyp.yaml")

**Example:**
```bash
# Lint current directory
easyp lint

# Lint specific directory with JSON output
easyp lint --path ./api --format json
```

**Supported Rule Categories:**

**Minimal Rules:**
- `DIRECTORY_SAME_PACKAGE` - Ensures consistent package-directory mapping
- `PACKAGE_DEFINED` - Requires package declaration in all files
- `PACKAGE_DIRECTORY_MATCH` - Package names must match directory structure
- `PACKAGE_SAME_DIRECTORY` - All files in directory must have same package

**Basic Rules:**
- `ENUM_FIRST_VALUE_ZERO` - First enum value must be zero
- `ENUM_PASCAL_CASE` - Enum names in PascalCase
- `FIELD_LOWER_SNAKE_CASE` - Field names in snake_case
- `MESSAGE_PASCAL_CASE` - Message names in PascalCase
- `SERVICE_PASCAL_CASE` - Service names in PascalCase

**Default Rules:**
- `FILE_LOWER_SNAKE_CASE` - File names in snake_case
- `RPC_REQUEST_RESPONSE_UNIQUE` - Unique request/response message types
- `PACKAGE_VERSION_SUFFIX` - Version suffixes in package names
- `SERVICE_SUFFIX` - Consistent service naming suffixes

**Comment Rules:**
- `COMMENT_ENUM` - Documentation for enums
- `COMMENT_FIELD` - Documentation for fields
- `COMMENT_MESSAGE` - Documentation for messages
- `COMMENT_RPC` - Documentation for RPC methods
- `COMMENT_SERVICE` - Documentation for services

---

### Breaking Change Detection

#### `easyp breaking`

Checks API compatibility against a Git reference.

```bash
easyp breaking [flags]
```

**Flags:**
- `--against string` - Git branch/tag to compare against (required)
- `-p, --path string` - Path to directory with proto files (default: ".")  
- `-f, --format string` - Output format: `text` or `json` (default: "text")

**Example:**
```bash
# Check against main branch
easyp breaking --against main

# Check against specific tag
easyp breaking --against v1.2.0 --path ./api

# Get JSON output for CI integration
easyp breaking --against origin/main --format json
```

---

### Package Management

#### `easyp mod download`

Downloads dependencies from lock file or configuration.

```bash
easyp mod download
```

**Behavior:**
- If `easyp.lock` exists: Downloads exact versions from lock file
- If `easyp.lock` missing: Downloads from `easyp.yaml` and creates lock file
- Dependencies cached in `~/.easyp` (or `$EASYPPATH`)

#### `easyp mod update`  

Updates dependencies to latest versions specified in configuration.

```bash
easyp mod update
```

**Behavior:**
- Ignores `easyp.lock` file
- Downloads versions specified in `easyp.yaml`
- Updates `easyp.lock` with new versions

#### `easyp mod vendor`

Copies proto dependencies to local vendor directory.

```bash
easyp mod vendor
```

**Behavior:**
- Creates `vendor/` directory in project root
- Copies all proto files from dependencies
- Useful for offline development or CI environments

**Example Dependency Management Workflow:**
```bash
# Add dependency to easyp.yaml
echo "deps:" > easyp.yaml
echo "  - github.com/googleapis/googleapis@v0.0.0-20230920204549-e6e6cdab5c13" >> easyp.yaml

# Download dependencies
easyp mod download

# Update to latest versions  
easyp mod update

# Create local vendor copy
easyp mod vendor
```

---

### Code Generation

#### `easyp generate`

Generates code from proto files using configured plugins.

```bash
easyp generate [flags]
```

**Flags:**
- `-p, --path string` - Path to directory with proto files (default: ".")

**Input Sources:**

**Local Directory:**
```yaml
generate:
  inputs:
    - directory: "api"
```

**Git Repository:**
```yaml
generate:
  inputs:
    - git_repo:
        url: "github.com/googleapis/googleapis@v0.0.0-20230920204549-e6e6cdab5c13"
        sub_directory: "google/api"
```

**Plugin Types:**

**Local Plugins:**
```yaml
plugins:
  - name: go
    out: .
    opts:
      paths: source_relative
  - name: go-grpc  
    out: .
    opts:
      paths: source_relative
      require_unimplemented_servers: false
```

**Remote Plugins:**
```yaml
plugins:
  - remote: localhost:8080/protoc-gen-python:v3.21.0
    out: ./generated
    opts:
      paths: source_relative
```

**Complete Example:**
```yaml
generate:
  inputs:
    - directory: "api"
    - git_repo:
        url: "github.com/googleapis/googleapis@common-protos-1_3_1"
        sub_directory: "google/api"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative
    - name: grpc-gateway
      out: .
      opts:
        paths: source_relative
    - remote: api.easyp.tech/protoc-gen-typescript:latest
      out: ./web/generated
```

---

## ⚙️ Configuration

### Configuration File

EasyP uses `easyp.yaml` for project configuration:

```yaml
version: v1alpha

# Linting configuration
lint:
  use:
    # Rule categories to enable
    - DIRECTORY_SAME_PACKAGE
    - PACKAGE_DEFINED
    - FIELD_LOWER_SNAKE_CASE
    - MESSAGE_PASCAL_CASE
    
  # Custom rule configuration  
  enum_zero_value_suffix: "UNSPECIFIED"
  service_suffix: "Service"
  
  # Allow comment-based rule ignores
  allow_comment_ignores: true

# Dependencies
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/bufbuild/protoc-gen-validate@v0.9.1

# Code generation
generate:
  inputs:
    - directory: "api"
  plugins:
    - name: go
      out: .
      opts:
        paths: source_relative

# Breaking change detection  
breaking:
  against_git_ref: "main"
  ignore:
    - "deprecated_api/"
```

### Environment Variables

- `EASYP_CFG` - Path to configuration file (default: `easyp.yaml`)
- `EASYP_DEBUG` - Enable debug logging (`true`/`false`)
- `EASYPPATH` - Custom cache directory (default: `~/.easyp`)
- `EASYP_FORMAT` - Default output format (`text`/`json`)

### Dependency Format

Dependencies use Git repository URLs with version specifiers:

```yaml
deps:
  # Git tag
  - github.com/googleapis/googleapis@common-protos-1_3_1
  
  # Commit hash  
  - github.com/bufbuild/protoc-gen-validate@959b9be7b8e44c15b8042e624c7423c0d4f2e5d8
  
  # Latest commit (not recommended for production)
  - github.com/grpc-ecosystem/grpc-gateway
```

### Private Repository Access

#### Using .netrc

Create `~/.netrc`:
```
machine github.com
login your-username  
password your-token
```

#### Using SSH Keys

Configure `~/.gitconfig`:
```ini
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

---

## 🔧 Advanced Usage

### Custom Linting Rules

Create custom rule configurations:

```yaml
lint:
  use:
    - CUSTOM_RULE_SET
    
  # Rule-specific configuration
  enum_zero_value_suffix: "NONE"
  service_suffix: "API"
  
  # Ignore specific rules for files
  ignore_only:
    "legacy/": 
      - "FILE_LOWER_SNAKE_CASE"
      - "PACKAGE_VERSION_SUFFIX"
```

### Multi-Environment Configuration

Different configurations for different environments:

**easyp.dev.yaml:**
```yaml
version: v1alpha
lint:
  use: ["BASIC"]
deps:
  - github.com/googleapis/googleapis@latest
```

**easyp.prod.yaml:**
```yaml  
version: v1alpha
lint:
  use: ["MINIMAL", "BASIC", "DEFAULT", "COMMENTS"]
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

Usage:
```bash
easyp lint --cfg easyp.dev.yaml
easyp lint --cfg easyp.prod.yaml
```

### CI/CD Integration

**CI/CD integration examples are currently in development**

### IDE Integration

IDE plugins are currently **in development** for:
- Visual Studio Code
- GoLand / IntelliJ
- Vim/Neovim

---

## 🎯 Shell Integration

### Auto-completion

#### Zsh

```bash
# Add to ~/.zshrc
source <(easyp completion zsh)

# Or install permanently
easyp completion zsh > ~/.zsh/completions/_easyp
```

#### Bash

```bash
# Add to ~/.bashrc  
source <(easyp completion bash)

# Or install permanently
easyp completion bash > /etc/bash_completion.d/easyp
```

---

## 🐛 Troubleshooting

### Common Issues

#### Config File Not Found
```
Error: config file not found
```

**Solution:** Create `easyp.yaml` or specify path with `--cfg`:
```bash
easyp init
# or
easyp lint --cfg /path/to/config.yaml
```

#### Dependency Download Failures
```
Error: repository does not exist
```

**Solutions:**
1. **Verify URL format**: Use `github.com/user/repo@version`
2. **Check access**: Ensure you have access to private repositories
3. **Configure auth**: Set up `.netrc` or SSH keys for private repos

#### Plugin Not Found
```
Error: plugin 'protoc-gen-go' not found
```

**Solutions:**
1. **Install plugin**: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
2. **Check PATH**: Ensure plugin is in your system PATH
3. **Use remote plugin**: Switch to remote execution if available

#### Breaking Check Issues
```
Error: Cannot find git ref 'main'
```

**Solutions:**
1. **Fetch latest**: `git fetch origin`
2. **Verify branch exists**: `git branch -r | grep main`
3. **Use correct ref**: Check actual branch name (`master` vs `main`)

### Debug Mode

Enable detailed logging:

```bash
export EASYP_DEBUG=true
easyp lint --debug
```

### Cache Issues

Clear EasyP cache:

```bash
rm -rf ~/.easyp
# or
rm -rf $EASYPPATH
```

---

## 🗺️ Roadmap

### Near Term (Next 3 months)

- **🔌 Plugin Registry** - Official plugin registry for easy discovery and installation
- **🎯 IDE Extensions** - Visual Studio Code and IntelliJ plugin support  
- **📊 CI/CD Templates** - Ready-to-use configurations for popular CI/CD platforms
- **📖 Enhanced Documentation** - Interactive tutorials and best practice guides

### Medium Term (3-6 months)

- **🌐 Web Dashboard** - Browser-based project management and visualization
- **📊 Analytics & Metrics** - Proto usage analytics and dependency insights
- **🔄 Migration Tools** - Automated migration from buf and other proto tools
- **🎯 Performance Improvements** - Faster linting and generation for large projects

### Long Term (6+ months)

- **🤖 AI-Powered Suggestions** - Intelligent code generation and optimization hints
- **🏢 Enterprise Features** - Team management, policy enforcement, and audit logs
- **🌍 Multi-Repository Support** - Cross-repository dependency management
- **🔗 API Gateway Integration** - Direct integration with API gateways and service meshes

---

## 🤝 Community

### Get Help & Connect

- **🌐 Official Website**: [https://easyp.tech/](https://easyp.tech/)
- **💬 Telegram Chat**: [https://t.me/easyptech](https://t.me/easyptech)  
- **🐛 Issues & Feature Requests**: [GitHub Issues](https://github.com/easyp-tech/easyp/issues)
- **📚 Documentation**: [https://easyp.tech](https://easyp.tech)

### Contributing

We welcome contributions! Guidelines coming soon.

### Support

- **Community Support**: Telegram chat and GitHub issues
- **Enterprise Support**: Contact us at [support@easyp.tech](mailto:support@easyp.tech)

---

## 📄 License

EasyP is released under the [Apache License 2.0](LICENSE).

---

*Built with ❤️ for the Protocol Buffers community*
