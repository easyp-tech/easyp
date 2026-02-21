# Quickstart

EasyP is a modern toolkit for working with Protobuf files. In this quickstart, you'll learn to:

1. **Initialize** your project with EasyP configuration
2. **Set up linting** to catch errors and ensure best practices
3. **Manage dependencies** from Git repositories
4. **Generate code** from your Protobuf files

## Prerequisites

Before you begin, make sure you have:

- **EasyP CLI installed** - See [Installation guide](/docs/guide/introduction/install) if you haven't already
- **Git** installed and in your `$PATH`
- **Protobuf files** ready to work with (or create some as you go)

You can verify your installation:

```bash
easyp --version
```

## Initialize your project

Start by creating an `easyp.yaml` configuration file in your project root:

```bash
easyp init
```

This command generates a commented template with default lint and breaking settings.
`easyp init` is interactive: if `buf.yml`/`buf.yaml` exists in the target root, EasyP asks whether to migrate it; if `easyp.yaml` already exists, it asks before overwriting.

## Configure linting

Linting helps catch errors and ensures your Protobuf files follow best practices. EasyP provides default rules compatible with Buf standards.

`easyp init` already generates a lint section. Adjust it if needed, for example:

```yaml
lint:
  use:
    - DEFAULT
```

Now you can run the linter:

```bash
easyp lint
```

The linter will check all your `.proto` files and report any issues found.

::: tip
EasyP's linting rules are 100% compatible with Buf, making migration easy if you're switching tools.
:::

## Manage dependencies

If your project uses third-party Protobuf packages (like Google APIs or gRPC Gateway), EasyP makes it easy to manage them.

### Dependency format

Dependencies use a simple format: `$GIT_LINK@$VERSION`

- **`$GIT_LINK`** - URL to any Git repository (GitHub, GitLab, etc.)
- **`$VERSION`** - Git tag or commit hash (optional)

If you omit the version, EasyP downloads the latest commit from the default branch.

### Add dependencies

Update your `easyp.yaml` to include dependencies:

```yaml
lint:
  use:
    - DEFAULT
deps:
  - github.com/googleapis/googleapis                          # Latest commit
  - github.com/grpc-ecosystem/grpc-gateway@v2.20.0           # Specific version
```

### Download dependencies

Download the packages you specified:

```bash
easyp mod download
```

This command:
- Downloads dependencies from `easyp.lock` if it exists
- Otherwise, downloads from `easyp.yaml` and creates `easyp.lock`

::: info
**Tip:** Use `easyp mod update` to ignore the lock file and fetch the latest versions from `easyp.yaml`.
:::

## Generate code

Now let's configure code generation. EasyP supports all standard Protobuf plugins.

### Configure generation

Add plugin configuration to your `easyp.yaml`:

```yaml
lint:
  use:
    - DEFAULT
deps:
  - github.com/googleapis/googleapis
  - github.com/grpc-ecosystem/grpc-gateway@v2.20.0
generate:
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

### Run code generation

Generate your code stubs:

```bash
easyp generate
```

EasyP will:
1. Resolve all dependencies
2. Connect to the EasyP service (or run locally)
3. Execute the configured plugins
4. Output generated code to the specified directories

::: tip
EasyP uses a hybrid runtime with WASM for speed and Docker for heavy plugins, giving you the best of both worlds.
:::

## Next steps

Congratulations! 🎉 You've successfully:
- ✅ Initialized an EasyP project
- ✅ Configured linting rules
- ✅ Managed third-party dependencies
- ✅ Generated code from your Protobuf files

### Learn more

- **[CLI Reference](/docs/guide/cli/linter/linter)** - Deep dive into all CLI commands
- **[API Service](/docs/guide/api-service/overview)** - Set up remote code generation
- **[Configuration](/docs)** - Explore all configuration options

---

*Enjoy working with Protobuf without any pain! 🚀*
