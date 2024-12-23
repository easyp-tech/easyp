# Easyp

`easyp` is a cli tool for workflows with `proto` files.

For now, it's just linter and package manager, but... who knows, who knows...

Just testing

## Community

### Official site

https://easyp.tech/

### Telegram chat

https://t.me/easyptech

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
easyp lint -cfg example.easyp.yaml
```

## Package manager

### Usage

* download

```bash
easyp -cfg example.easyp.yaml mod download
```

Read your dependencies from `easyp.lock` file and install them.


If `easyp.lock` is empty or doesn't exist `easyp` read dependencies from `easyp.yaml` config file (`deps` section).

* update
```bash
easyp -cfg example.easyp.yaml mod update
```

Read dependencies from `easyp.yaml` config file and ignore `easyp.lock` file.

Could be used for update versions: set version in `easyp.yaml` file and run `update` command.

### Configuration

Write list of your dependencies in `easyp.yaml` config with in section `deps`.

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
