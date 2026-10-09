# EasyP v1 Breaking Change Checks Reference

`easyp breaking` compiles the current sources and a Git baseline. Each side uses its own `protobuf.mod`/`protobuf.lock` and imports. It then compares the linked descriptors using the selected **profiles**.

## Usage

```bash
easyp breaking                              # baseline from easyp.yaml (breaking.baseline), else master
easyp breaking --against main               # raw ref on the CLI — no git: prefix
easyp breaking --path proto --against origin/main
easyp --format json breaking --against v1.4.0
```

- **CI:** needs the baseline in history (`fetch-depth: 0` on GitHub Actions).
- **Exit codes:**
  - 0: compatible;
  - 1: breaking changes, or any error;
  - 2: ref or repository not found, or missing import.
- **Baseline still on v0 config** (right after migrating): this works only when the project has no dependencies. Otherwise the baseline compile cannot resolve dependency imports, because the v0 revision has no `protobuf.lock`. Use a v0 `easyp breaking` for that one PR, see [migration-v0-to-v1.md](./migration-v0-to-v1.md#step-7--verify-on-v1).

## Configuration

```yaml
version: v1
breaking:
  baseline: git:main          # default ref; must be empty or git:<ref>
  categories: [FILE]          # FILE (default) | PACKAGE | WIRE_JSON | WIRE — several = union
  ignore_unstable: true       # skip packages whose last component is unstable (v1alpha1, v1beta1, v1test)
  ignore:
    - proto/internal          # paths relative to this easyp.yaml
  extends: ./.policies/breaking.yaml   # optional shared base for this section
```

## Choosing a profile

| Profile | Protects | Choose when |
|---|---|---|
| `FILE` (default) | Generated source code, including which **file** declares each type | Consumers compile generated code, and files/packages are part of the contract (Go, Java, …) |
| `PACKAGE` | Generated source code per **package**; declarations may move between files of a package | Same as FILE, but you reorganise files inside packages |
| `WIRE_JSON` | Binary wire format **and** protobuf JSON (field/enum names matter) | Clients may be generated from older schemas; JSON/REST gateways are used |
| `WIRE` | Binary wire format only | Only binary protobuf on the wire; names are free to change |

Strictness: FILE > PACKAGE > WIRE_JSON > WIRE.

| Change | FILE | PACKAGE | WIRE_JSON | WIRE |
|---|---|---|---|---|
| Move a declaration within its package | ✗ | ✓ | ✓ | ✓ |
| Rename a field (even keeping `json_name`) | ✗ | ✗ | ✗ | ✓ |
| `int32` → `int64` | ✗ | ✗ | ✗ | ✓ |
| `int32` → `uint32` | ✗ | ✗ | ✓ | ✓ |
| `string` → `bytes` | ✗ | ✗ | ✗ | ✓ |
| `bytes` → `string`, `int32` → `sint32` | ✗ | ✗ | ✗ | ✗ |
| Delete a non-required field | ✗ | ✗ | only if number **and** name reserved | only if number reserved |
| Add/remove a required field, change a real oneof | ✗ | ✗ | ✗ | ✗ |
| Add explicit `optional` presence | ✗ | ✗ | ✓ | ✓ |
| Repeated entry messages → compatible `map` | ✗ | ✗ | ✗ | ✓ |

Always checked, in every profile:
- deleting a service or RPC;
- changing an RPC's request/response type, streaming or idempotency;
- changing a default value;
- releasing reserved names or ranges.

The source profiles (FILE, PACKAGE) also check file and field code-generation options, for example:
- `go_package`, `java_package`, `java_multiple_files`;
- `csharp_namespace`, `objc_class_prefix`;
- `json_name`, `jstype`, `ctype`.

## Rule names in findings

Findings name the rule, e.g. `FIELD_SAME_TYPE`, `FIELD_NO_DELETE_UNLESS_NUMBER_RESERVED`.

| Area | Rules |
|---|---|
| Deletion | `PACKAGE_NO_DELETE`, `FILE_NO_DELETE`, `MESSAGE_NO_DELETE`, `ENUM_NO_DELETE`, `SERVICE_NO_DELETE`, `RPC_NO_DELETE`, `ONEOF_NO_DELETE`, `FIELD_NO_DELETE`, `ENUM_VALUE_NO_DELETE`, `EXTENSION_NO_DELETE` |
| Deletion with reservation (WIRE/WIRE_JSON) | `FIELD_NO_DELETE_UNLESS_NUMBER_RESERVED`, `FIELD_NO_DELETE_UNLESS_NAME_RESERVED`, `ENUM_VALUE_NO_DELETE_UNLESS_NUMBER_RESERVED`, `ENUM_VALUE_NO_DELETE_UNLESS_NAME_RESERVED`, `RESERVED_MESSAGE_NO_DELETE`, `RESERVED_ENUM_NO_DELETE` |
| Moves (FILE only) | `MESSAGE_MOVED`, `ENUM_MOVED`, `SERVICE_MOVED`, `ONEOF_MOVED`, `EXTENSION_MOVED` |
| Fields | `FIELD_SAME_NAME`, `FIELD_SAME_TYPE`, `FIELD_SAME_CARDINALITY`, `FIELD_SAME_JSON_NAME`, `FIELD_SAME_ONEOF`, `FIELD_SAME_DEFAULT`, `FIELD_SAME_JSTYPE`, `FIELD_SAME_CPP_STRING_TYPE`, `FIELD_SAME_UTF8_VALIDATION`, `FIELD_WIRE_COMPATIBLE_TYPE`, `FIELD_WIRE_JSON_COMPATIBLE_TYPE`, `FIELD_WIRE_(JSON_)COMPATIBLE_CARDINALITY` |
| Enums | `ENUM_VALUE_SAME_NAME`, `ENUM_SAME_TYPE`, `ENUM_SAME_JSON_FORMAT` |
| Messages | `MESSAGE_SAME_REQUIRED_FIELDS`, `MESSAGE_SAME_JSON_FORMAT`, `MESSAGE_SAME_MESSAGE_SET_WIRE_FORMAT`, `MESSAGE_NO_REMOVE_STANDARD_DESCRIPTOR_ACCESSOR` |
| RPCs | `RPC_SAME_REQUEST_TYPE`, `RPC_SAME_RESPONSE_TYPE`, `RPC_SAME_CLIENT_STREAMING`, `RPC_SAME_SERVER_STREAMING`, `RPC_SAME_IDEMPOTENCY_LEVEL` |
| File options (FILE/PACKAGE) | `FILE_SAME_PACKAGE`, `FILE_SAME_SYNTAX`, `FILE_SAME_GO_PACKAGE`, `FILE_SAME_JAVA_PACKAGE`, `FILE_SAME_JAVA_MULTIPLE_FILES`, `FILE_SAME_CSHARP_NAMESPACE`, … (one per language option) |
| Extensions | `EXTENSION_SAME_NUMBER`, `EXTENSION_SAME_EXTENDEE`, `EXTENSION_MESSAGE_NO_DELETE` |

## Backward-compatible alternatives

| Breaking change | Safe alternative |
|---|---|
| Remove a field | `reserved 3; reserved "old_field";` (WIRE_JSON needs both, WIRE only the number; FILE/PACKAGE still reject the delete; deprecate instead: `[deprecated = true]`) |
| Change a field type | Add a new field with the new type, deprecate the old one |
| Remove or rename an enum value | Keep the number, add the new value, deprecate the old one (`option allow_alias = true` to alias) |
| Remove an RPC or service | `option deprecated = true;`, remove in the next major package (`v2`) |
| Change RPC request/response | Add a new RPC with new types |
| Move a message to another file | Use `PACKAGE` (or looser) if file moves are acceptable for your consumers |
| Rename a field | Only acceptable with `WIRE`; otherwise add a new field |

## Handling intentional breaks

- Pre-release packages: put them in `v1alpha1`/`v1beta1` and set `breaking.ignore_unstable: true`.
- Internal or experimental paths: `breaking.ignore: [proto/internal]`.
- A new major version: create a new package (`acme.orders.v2`) instead of breaking `v1`.
- A one-off accepted break: run against a newer baseline, e.g. `easyp breaking --against v2.0.0`.

## v0 → v1 changes

| v0 | v1 |
|---|---|
| `breaking.against_git_ref: main` | `breaking.baseline: git:main` |
| `breaking.use: [FILE]` | `breaking.categories: [FILE]` (also the default) |
| AST checker with a few categories (import, service, message, oneof, enum) | Descriptor profiles FILE / PACKAGE / WIRE_JSON / WIRE with the rule catalog above |
| — | `ignore_unstable`, `extends`, union of several categories |

`easyp migrate` converts only `breaking.use: [FILE]`. For other values, remove `breaking.use` before migrating and set `categories` afterwards.

The v1 FILE profile is a descriptor-based checker. It flags more than the v0 checker did, for example moving a type between files or changing `go_package`. Expect new findings on the first run after migrating.
