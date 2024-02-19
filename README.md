# Easyp

`easyp` is a cli tool for workflows with `proto` files.

For now, it's just linter and package manager, but... who knows, who knows...

## Package manager

### Usage

To usage `easpy` as a package manager use `mod` command:

```bash
easyp mod -c example.easyp.yaml
```

Your config file has to contains `deps` section which is list of repositories with proto files and its version (optional).

For example:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

**NOTE:** Use only git tag or full hash of commit version.

By default, `easyp` use `$HOME/.easyp` dir to storage cache and downloaded modules, you could override it with `EASYPPATH` env var.

### Roadmap

* [ ] Implement support for `buf.work.yaml` config
* [ ] Calc hash sum, store it and compare (i.e go.sum)
* [ ] Buf's blob support (for IDE plugins)
