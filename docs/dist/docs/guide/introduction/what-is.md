# What is EasyP?

EasyP is a modern Protocol Buffers toolkit that consolidates your entire protobuf workflow into a single, powerful CLI tool. It combines linting, package management, code generation, and compatibility checking in a unified interface designed to eliminate the complexity of traditional protobuf development.

<div class="tip custom-block" style="padding-top: 8px">

Just want to try it out? Skip to the [Quickstart](./guide/introduction/quickstart).

</div>

## The Problem: Fragmented Proto Workflows

Working with Protocol Buffers across projects creates unnecessary complexity and friction:

### Tool Chaos
- **Multiple tools for different tasks**: Separate tools for linting, generation, and dependency management
- **Inconsistent configurations**: Different config formats across projects and teams
- **Complex setup**: Time-consuming onboarding for new developers joining projects

### Dependency Management Issues  
- **No standardized approach**: Each project handles proto dependencies differently
- **Version conflicts**: Difficult to resolve conflicts between different proto libraries
- **Build reproducibility**: Hard to ensure consistent builds across environments

### Development Friction
- **Manual processes**: Repetitive workflows for common proto tasks
- **Configuration sprawl**: Multiple config files with different formats
- **Time-consuming compatibility checks**: Manual verification of API changes

## The Solution: Unified Proto Toolkit

EasyP addresses these challenges by providing a comprehensive, opinionated toolkit that standardizes protobuf workflows:

### One Tool, All Tasks
Instead of juggling multiple tools, EasyP provides everything you need:
- **Linting** with comprehensive rule sets
- **Package management** with Git-native dependencies
- **Code generation** with local and remote plugin support
- **Breaking change detection** for API compatibility

### Standardized Configuration
A single `easyp.yaml` file configures your entire protobuf workflow, eliminating configuration sprawl and ensuring consistency across projects.

### Git-Native Dependencies
Direct integration with Git repositories for proto dependencies eliminates the need for centralized registries while providing version control and reproducibility.

## Key Features

| Feature | Description |
|---------|-------------|
| 🔍 **Comprehensive Linting** | Built-in support for buf's linting rules with customizable configurations to enforce API design best practices. |
| 📦 **Smart Package Manager** | Git-based dependency management with lock file support for reproducible builds across environments. |
| ⚡ **Code Generation** | Multi-language code generation with support for both local protoc plugins and remote plugin execution. |
| 🔄 **Breaking Change Detection** | Automated API compatibility verification against Git branches to prevent accidental breaking changes. |
| 🌐 **Remote Plugin Support** | Execute plugins via centralized EasyP API service for consistent, isolated execution without local dependencies. |
| 🎯 **Developer Experience** | Auto-completion, intuitive commands, clear error messages, and comprehensive documentation. |

## Supported Plugin Types

EasyP provides flexibility in how you execute code generation plugins:

### Local Plugins
Standard protoc plugins installed on your system:
```yaml
plugins:
  - name: go
    out: .
    opts:
      paths: source_relative
```

### Remote Plugins
Plugins executed via EasyP API service for consistent results:
```yaml
plugins:
  - remote: api.easyp.tech/protoc-gen-typescript:latest
    out: ./web/generated
```

### Custom Plugins
Support for custom plugin development and distribution through the ecosystem.

## Decentralized Package Management

Unlike other tools, EasyP doesn't rely on a centralized server for package distribution. Instead, any Git repository can serve as a package source, giving you:

- **Flexibility and Control**: Use any Git hosting service (GitHub, GitLab, self-hosted)
- **No Vendor Lock-in**: Your dependencies aren't tied to a specific registry
- **Version Control**: Full Git history and branching for your proto dependencies
- **Enterprise-Friendly**: Works with private repositories and corporate Git infrastructure

## Seamless Migration from Buf

EasyP is designed with buf compatibility in mind:

- **Compatible rule sets**: Uses the same linting rules as buf
- **Familiar configuration**: Similar YAML structure with enhanced features
- **Drop-in replacement**: Can often replace buf with minimal configuration changes
- **Gradual migration**: Adopt EasyP features incrementally without disrupting existing workflows

### EasyP vs buf.build Comparison

| Feature | EasyP | buf.build |
|---------|--------|-----------|
| **Dependency Management** | Git-based repositories | Buf Schema Registry (BSR) |
| **Vendor Lock-in** | None - works with any Git hosting | Tied to BSR for full features |
| **Plugin Execution** | Local + Remote plugins | Local + BSR plugins |
| **Private Dependencies** | Any Git provider (GitHub, GitLab, etc.) | BSR or manual management |
| **Offline Development** | Full support with `mod vendor` | Limited without BSR access |
| **Enterprise Integration** | Works with existing Git infrastructure | Requires BSR setup |
| **Breaking Change Detection** | Against any Git reference | Against any Git reference |
| **Package Distribution** | Any Git repository | BSR required for publishing |
| **License** | Apache 2.0 | Apache 2.0 |
| **Community** | Growing | Established |

**Migration Benefits:**
- **No infrastructure changes**: Continue using your existing Git repositories
- **Gradual adoption**: Start with EasyP while keeping existing buf configurations
- **Enhanced flexibility**: Access to both local and remote plugin execution
- **Simplified workflows**: Single configuration file for all protobuf operations

## Our Goals for Protobuf

EasyP's mission is to accelerate the adoption of **schema-driven API development** by making Protocol Buffers more accessible and reliable:

### Modern Protobuf Ecosystem
We're building on Protobuf's proven foundation to create a modern development experience that rivals the simplicity of REST/JSON while providing superior type safety and performance.

### Developer Experience First
Every feature is designed with developer productivity in mind, from intuitive CLI commands to comprehensive error messages and documentation.

### Enterprise Ready
Built for teams and organizations with features like reproducible builds, private repository support, and comprehensive CI/CD integration.

## Why Choose EasyP?

### For Individual Developers
- **Quick Setup**: Get started with protobuf projects in minutes, not hours
- **Unified Workflow**: One tool for all your protobuf needs
- **Clear Documentation**: Comprehensive guides and examples

### For Teams
- **Consistent Standards**: Enforced linting rules and formatting across all projects
- **Reproducible Builds**: Lock files ensure everyone builds the same way
- **Easy Onboarding**: New team members can be productive immediately

### For Organizations
- **No Vendor Lock-in**: Git-based dependencies work with your existing infrastructure
- **Enterprise Security**: Support for private repositories and custom authentication
- **CI/CD Ready**: Designed for automated workflows and continuous integration

## What's Next?

EasyP simplifies protobuf development so you can focus on building great APIs instead of managing toolchain complexity. Whether you're starting a new project or migrating from existing tools, EasyP provides a smooth path to modern protobuf development.

Ready to get started? Check out our [Installation Guide](./install) and [Quickstart Tutorial](./quickstart).

## Stargazers over time
[![Stargazers over time](https://starchart.cc/easyp-tech/easyp.svg?variant=adaptive)](https://starchart.cc/easyp-tech/easyp)
