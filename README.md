# Easyp

`easyp` is a cli tool for workflows with `proto` files.

For now, it's just linter and package manager, but... who knows, who knows...

Just testing

## Install

### Build from source

1. Clone repository
2. Build
```bash
go build ./cmd/easyp
```

### Install from github

```bash
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

## Linter

`easyp` support `buf's` linter rules.

### Usage

```bash
easyp lint -c example.easyp.yaml
```

## Package manager

### Usage

To usage `easpy` as a package manager use `mod download` command:

```bash
easyp -c example.easyp.yaml mod download
```

Your config file has to contains `deps` section which is list of repositories with proto files and its version (optional).

For example:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

**NOTE:** Use only git tag or full hash of commit version.

By default, `easyp` use `$HOME/.easyp` dir to storage cache and downloaded modules, you could override it with `EASYPPATH` env var.

## Auto-completion

### zsh auto-completion

1. Add the following line to your ~/.zshrc startup script:

```bash
source <(easyp completion zsh)
```

2. Re-launch your shell or run:

```bash
source ~/.zshrc
```

### Bash auto-completion

1. Install [bash-completion](https://github.com/scop/bash-completion#installation) and add the software to your `~/.bashrc`.
2. Add the following line to your ~/.bashrc startup script:

```bash
source <(easyp completion bash)
```

3. Re-launch your shell or run:

```bash
source ~/.bashrc
```

### Roadmap

* [x] Implement support for `buf.work.yaml` config
* [ ] Calc hash sum, store it and compare (i.e go.sum)
* [ ] Code generation
