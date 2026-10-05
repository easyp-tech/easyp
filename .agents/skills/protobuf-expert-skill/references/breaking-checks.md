# EasyP Breaking Change Checks Reference

EasyP detects breaking API changes by comparing the current proto files against a previous git ref (branch, tag, or commit).

## Usage

```sh
easyp breaking --path proto --against main
```

## Configuration

```yaml
breaking:
  ignore:
    - proto/internal          # Paths to exclude from checks
  against_git_ref: main       # Default ref (overridden by --against flag)
```

## Check Categories

### Import Changes

| Check | Description |
|-------|-------------|
| Import deleted | A previously existing import was removed |

### Service Changes

| Check | Description |
|-------|-------------|
| Service deleted | An entire service definition was removed |
| RPC deleted | An RPC method was removed from a service |
| RPC request type changed | The request message type of an RPC was changed |
| RPC response type changed | The response message type of an RPC was changed |

### Message Changes

| Check | Description |
|-------|-------------|
| Message deleted | A message definition was removed |
| Field deleted | A field was removed (reports field number and name) |
| Field type changed | A field's type was changed |
| Field became optional | A required/default field was changed to optional |
| Field became not optional | An optional field was changed to required |

### OneOf Changes

| Check | Description |
|-------|-------------|
| OneOf deleted | A oneof group was removed |
| OneOf field deleted | A field within a oneof was removed |
| OneOf field type changed | A field type within a oneof was changed |

### Enum Changes

| Check | Description |
|-------|-------------|
| Enum deleted | An enum definition was removed |
| Enum value deleted | An enum value was removed |
| Enum value name changed | An enum value was renamed |

## Backward-Compatible Alternatives

Instead of making a breaking change, use these patterns:

| Breaking Change | Safe Alternative |
|----------------|-----------------|
| Remove a field | Mark as `reserved` and deprecate: `reserved 3; reserved "old_field";` |
| Change field type | Add a new field with the new type, deprecate the old one |
| Remove an enum value | Reserve the number and name, add replacement value |
| Rename an enum value | Add new value, deprecate old (both keep same number) |
| Remove an RPC | Deprecate with `option deprecated = true;` first |
| Remove a service | Deprecate first, remove in next major version |
| Change RPC request/response | Create a new RPC with new types |

## Ignoring Breaking Changes

For intentional breaks, add paths to `breaking.ignore`:

```yaml
breaking:
  ignore:
    - proto/internal/experimental
```

Or change the comparison ref to start fresh:

```yaml
breaking:
  against_git_ref: v2.0.0    # Compare against a specific release tag
```
