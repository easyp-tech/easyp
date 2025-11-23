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

If you are using third-party packages,
you need to add them to the `deps` section in the `easyp.yaml` file.

### Dependency Format

Format is simple: `$GIT_LINK@$VERSION` where:
- `$GIT_LINK`: is just a link to git repo (github, gitlab etc)
- `$VERSION`: git tag or FULL hash of commit

If version is omitted then easyp will download the latest commit from default branch.

```yaml
lint:
  use:
    - DEFAULT
deps:
  - github.com/googleapis/googleapis                           # Latest commit
  - github.com/grpc-ecosystem/grpc-gateway@v2.20.0            # Specific tag
```

### Download Dependencies

Now you can download packages:

```bash
easyp mod download
```

**Note:** 
- `download` command will download deps from your `easyp.lock` file. If lock file is missing then easyp downloads packages with versions from `easyp.yaml` file and creates lock file
- `update` command ignores lock file and downloads packages with versions from `easyp.yaml` file and creates/updates lock file

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
