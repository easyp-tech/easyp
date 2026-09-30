# EasyP configuration MCP

<code>easyp_config_describe</code> describes four v1 configuration formats without reading a project, fetching dependencies or executing plugins. The tool name and existing YAML response fields are preserved.

| File selector | Output format | Reference |
|---|---|---|
| <code>easyp.yaml</code> | <code>yaml</code> | Policy fields, examples and the actual JSON Schema |
| <code>easyp.gen.yaml</code> | <code>yaml</code> | Generation fields, examples and the actual JSON Schema |
| <code>protobuf.lock</code> | <code>yaml</code> | Locked dependency fields and the actual lock JSON Schema |
| <code>protobuf.mod</code> | <code>text</code> | Directive grammar, fields and text examples; no fabricated JSON Schema |

## Run

~~~sh
go run ./cmd/easyp-mcp
~~~

Start this command from the EasyP source repository or run a built <code>easyp-mcp</code> binary from any directory. Protocol messages use stdout and process diagnostics use stderr.

Go programs register the same reference with <code>easypconfig.RegisterTool(server)</code> from <code>github.com/easyp-tech/easyp/mcp/easypconfig</code>.

## Requests

The default <code>file</code> is <code>easyp.yaml</code>; empty <code>path</code>, <code>$</code> and <code>root</code> select the root. Array indices normalize to <code>[]</code>. The file argument is an exact format selector, never a filesystem path or URL.

~~~json
{"file":"protobuf.mod","path":"require[].version"}
~~~

~~~json
{"file":"protobuf.lock","path":"modules[0].commit"}
~~~

<code>include_schema</code>, <code>include_fields</code>, <code>include_examples</code> and <code>include_children</code> default to true. For a text manifest, <code>include_schema</code> controls the grammar instead; <code>schema</code> is always omitted. With <code>include_children=false</code>, only the selected field/directive is described. <code>examples_limit</code> defaults to 10 and must be an integer from 1 through 50. Out-of-range values now return a validation error instead of silently clamping.

Output always includes <code>format</code>. YAML examples retain <code>examples[].yaml</code>; text examples contain <code>format: text</code> and <code>text</code>, with no misleading YAML value. Manifest syntax is described by <code>grammar.dialect</code>, <code>grammar.syntax</code> and <code>grammar.notes</code>. Unknown files, paths and invalid arguments return tool errors.

## Contracts

YAML schemas come directly from <code>internal/config/v1.SchemaJSON</code>, the same source used by <code>schema-gen</code>. Lock descriptions distinguish structural constraints from semantic checks: unique source identities, valid module major suffixes, matching commit-valued versions and verified content. Example commits and hashes are explicitly synthetic and do not certify fetched content.

Manifest directives are <code>module</code>, <code>roots</code>, <code>require</code> and <code>replace</code>. Examples are checked by <code>ParseModule</code>; lock examples by <code>ParseLock</code> and the generated schema. Main-module local replacements do not alter the published lock. Frozen mode rejects replacements and requires a consistent pinned graph. The reference does not resolve unknown imports, query BSR, load policy dependencies or expand environment variables.

## Verification

~~~sh
go test -race -count=1 ./mcp/easypconfig
~~~

The external <code>easyp-test</code> suite invokes the built stdio server using <code>EASYP_MCP_BIN</code>. When omitted, the test builds that server from an explicit absolute <code>EASYP_SOURCE</code>. It verifies initialization, tool listing, real calls, valid examples, schema equality, legacy YAML compatibility, input errors and lack of project-file access. The full audit builds both binaries from the same source revision before testing.
