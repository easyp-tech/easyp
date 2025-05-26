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

## Init

Creates empty `easyp` project.

Creates `easyp.yaml` (by default) and `easyp.lock` files.

### Usage

```bash
easyp init
```

## Linter

`easyp` support `buf's` linter rules.

### Usage

```bash
easyp lint -cfg example.easyp.yaml
```
## Breaking check

Checking your current API on backward compatibility with API from another branch.

### Usage

```bash
easyp breaking --against $BRANCH_TO_COMPARE_WITH
```

## Generate

Generate proto files. 

### Usage

There are several ways to get proto files to generate:
1. from current local repository:
```yaml
generate:
  inputs:
    - directory: WHERE YOUR PROTO FILES ARE
```
2. from remote git repository:
```yaml
generate:
  inputs:
    - git_repo:
        url: "URL TO REMOTE REPO"
        sub_directory: DIR WITH PROTO FILES ON REMOTE REPO
```
**NOTE:** format `url` the same as in `deps` section.

`plugins` section: config for `protoc`

Example:
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

## Package manager

Install dependence from `easyp` config (or lock file).

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

* vendor
```bash
easyp -cfg example.easyp.yaml mod vendor
```

Copy all your proto files dependencies to local dir (like `go mod vendor` command).


### Configuration

Write list of your dependencies in `easyp.yaml` config with in section `deps`.

For example:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

**NOTE:** Use only git tag or full hash of commit version.

By default, `easyp` use `$HOME/.easyp` dir to storage cache and downloaded modules, you could override it with `EASYPPATH` env var.

### Install from private repositories

There are two ways to install from private repository.

1. Use `.netrc`

Create `.netrc` in your home dir:
```
machine $GIT_HOSTING
login $YOUR_LOGIN
password $YOUR_API_TOKEN
```

In that case you have to create API token on git hosting

2. Use ssh keys

* Configure your `ssh` config (`~/.ssh/config`) with path to private key and git hosting's params

* Configure your git config (`~/.gitconfig`):
```
[url "ssh://git@$GIT_HOSTING/"]
    insteadOf = https://$GIT_HOSTING/
```

for example:
```
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

### Use mirrors

For use your company private mirrors (e.g artifactory) you can use `mirrors` settings.

```yaml
mirrors:
  - origin: onprem-vcs.loc
    use: github.com
```

In that case dependency from your config with host `onprem-vcs.loc` will be replaced with `github.com`.

So you are able to download deps from your mirror without modificate easyp config file.

In case when you have to use private 

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
