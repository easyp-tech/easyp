# Quickstart

## Initial setup

To start using EasyP, you need to initialize the project.

To do this, run the following command in the terminal:

```bash
easyp init 
```

## Setup linter rules

For example, you can use default linter rules.

```yaml
lint:
  use:
    - DEFAULT
```

After setting up the linter rules, you can run the linter command:

```bash
easyp lint
```

## Setup package dependencies

If you ara using third-party packages,
you need to add them to the `dependencies` section in the `easyp.yaml` file.

```yaml
lint:
  use:
    - DEFAULT
deps:
  - github.com/googleapis/googleapis
  - github.com/grpc-ecosystem/grpc-gateway@v2.20.0
```

And now you can download packages by command

```bash
easyp mod download
```

## Setup proto generation

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

On this step, you can generate proto files by command:

```bash
easyp generate
```

*Enjoy working with proto without any pain! 🚀*
