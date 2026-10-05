# MCP reference for all v1 configuration formats

The stable easyp_config_describe tool accepts easyp.yaml, easyp.gen.yaml, protobuf.mod and protobuf.lock. File is a known format selector, never a filename to open. Tool calls only construct in-memory reference data; they do not read project files, access Git/the network, expand environment variables, change files or execute plugins.

For YAML, schema and examples[].yaml remain compatible with existing clients. protobuf.lock obtains its actual schema from v1.SchemaJSON("protobuf.lock"); documentation distinguishes structural rules from ParseLock semantic checks. Synthetic sample commits/hashes are labelled and are not assertions about real content.

protobuf.mod is text: format=text, grammar={dialect,syntax,notes}, examples containing format=text and text, no schema and no examples[].yaml. Directive/field selection covers module, roots, require, replace and their relevant child fields. Grammar follows the parser's whitespace tokens, block boundaries, comments/BOM, duplicate-source rejection and major-version identity rules. It does not invent a YAML schema for a text format.

include_schema controls JSON Schema for YAML and grammar for text. Other detail/child/example flags keep their existing meaning. Empty file selects easyp.yaml; empty path, root and $ select the root; indices normalize to []. examples_limit defaults to 10, accepts 1 through 50, and now rejects invalid bounds rather than clamping. This validation tightening is explicit. Unknown file/path/arguments fail; valid previous YAML inputs and existing YAML example fields remain supported.

Responses own all mutable maps/slices. Parallel calls and mutation tests verify isolation. The MCP input schema lists the four formats and matches the typed argument names, while output continues using the SDK's typed schema inference.

See ../../mcp/easypconfig/README.md for requests. Unit tests cover parser/schema parity, returned-state isolation and in-memory SDK transport. easyp-test/tests/e2e/v1/mcp_module_reference_test.go exercises actual stdio initialization/list/call, invalid inputs, no project reads, example CLI validation and exact generated-schema equality. EASYP_MCP_BIN selects the freshly built server; the fallback builds it from explicit EASYP_SOURCE, never a global installed binary.
