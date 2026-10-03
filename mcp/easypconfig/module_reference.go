package easypconfig

import (
	"fmt"
	"sort"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// GrammarReference describes textual directives, never an executable config or
// a substitute JSON Schema for the non-YAML protobuf.mod format.
type GrammarReference struct {
	Dialect string   `json:"dialect"`
	Syntax  []string `json:"syntax"`
	Notes   []string `json:"notes,omitempty"`
}

func moduleFields() map[string]FieldDoc {
	return map[string]FieldDoc{
		"module":             {Path: "module", Type: "directive", Required: true, Description: "Exactly one module identity. Native v0/v1 share an unsuffixed identity; v2+ uses /vN. Identity does not rewrite proto package or import names."},
		"roots":              {Path: "roots", Type: "directive", DefaultValue: ".", Description: "One or more module-relative import roots. Absolute paths and paths escaping the module are rejected."},
		"roots[]":            {Path: "roots[]", Type: "relative-path", Description: "Physical source directory below the manifest directory; the root prefix is removed from proto import names."},
		"require":            {Path: "require", Type: "directive", Description: "Direct or derived module requirements. Duplicate sources are rejected; versions are declared here, not in generation or extends settings."},
		"require[]":          {Path: "require[]", Type: "requirement", Description: "Module identity followed by an optional semantic version or full commit."},
		"require[].module":   {Path: "require[].module", Type: "module-identity", Required: true, Description: "A known Git module identity, including a matching /vN for native v2+. Unknown proto imports do not infer a Git repository."},
		"require[].version":  {Path: "require[].version", Type: "version-or-commit", Description: "Semantic version, a full 40/64-character hexadecimal commit, or omitted. Versionless first use resolves HEAD; ordinary use retains its lock pin. +incompatible requires verified pre-native published metadata."},
		"require[].indirect": {Path: "require[].indirect", Type: "comment-marker", Description: "The comment // indirect marks a derived requirement. A direct proto import promotes it; it is not a separate grammar directive or a runtime version constraint."},
		"replace":            {Path: "replace", Type: "directive", Description: "Main-module local overrides. A replacement does not add a require; duplicate replacement sources are rejected."},
		"replace[]":          {Path: "replace[]", Type: "replacement", Description: "A required module identity, =>, and an unversioned local directory. Only the main module's replacements apply throughout its graph."},
		"replace[].module":   {Path: "replace[].module", Type: "module-identity", Required: true, Description: "Exact identity being replaced, including any /vN suffix."},
		"replace[].target":   {Path: "replace[].target", Type: "local-path", Required: true, Description: "Absolute path or path relative to this manifest. A local overlay can change the effective graph without rewriting published protobuf.lock; --frozen rejects replace."},
	}
}

func moduleGrammar(path string) GrammarReference {
	syntax := []string{"module <identity>", "roots <relative-path> [<relative-path> ...]", "require <module> [<semantic-version-or-full-commit>] [// indirect]", "replace <module> => <local-directory>", "roots (\n    <relative-path>\n)", "require (\n    <module> [<version-or-commit>]\n)", "replace (\n    <module> => <local-directory>\n)"}
	if path != "$" {
		directive := strings.Split(strings.TrimSuffix(path, "[]"), ".")[0]
		directive = strings.TrimSuffix(directive, "[]")
		selected := make([]string, 0, len(syntax))
		for _, line := range syntax {
			if strings.HasPrefix(line, directive+" ") {
				selected = append(selected, line)
			}
		}
		syntax = selected
	}
	return GrammarReference{Dialect: "protobuf.mod/v1", Syntax: syntax, Notes: []string{
		"Whitespace-delimited, unquoted tokens. One module directive; block entries and closing parentheses occupy separate lines. Inline and nested blocks are rejected.",
		"BOM, blank lines, and // or # comments at token boundaries are accepted. Environment variables are not expanded by the text manifest parser.",
		"roots defaults to dot. require and replace reject duplicate source identities. Legacy direct/indirect blocks require explicit easyp migrate; no automatic conversion occurs here.",
	}}
}

func describeModule(input DescribeInput) (DescribeOutput, error) {
	path := normalizePath(input.Path)
	fields := moduleFields()
	if _, ok := fields[path]; !ok && path != "$" {
		return DescribeOutput{}, fmt.Errorf("unknown path %q in %s", input.Path, v1.ModuleFile)
	}
	out := DescribeOutput{SchemaVersion: SchemaVersion, File: v1.ModuleFile, SelectedPath: path, Format: "text", Notes: []string{
		"protobuf.mod is a text manifest, not YAML. No JSON Schema is fabricated; grammar describes its actual directives. This tool returns reference data and never reads the supplied file as a path.",
		"Module identities, roots and versions remain separate from protobuf package names. Policy extends uses declared, locked modules; exact generate.packages selectors never acquire new dependencies. BSR-to-Git discovery is not implemented.",
	}}
	if enabled(input.IncludeSchema) {
		grammar := moduleGrammar(path)
		out.Grammar = &grammar
	}
	if enabled(input.IncludeFields) {
		paths := make([]string, 0, len(fields))
		for candidate := range fields {
			if candidate == path || (enabled(input.IncludeChildren) && within(path, candidate)) {
				paths = append(paths, candidate)
			}
		}
		sort.Strings(paths)
		for _, candidate := range paths {
			out.Fields = append(out.Fields, fields[candidate])
		}
	}
	if enabled(input.IncludeExamples) {
		out.Examples = selectExamples(v1.ModuleFile, path, exampleLimit(input.ExamplesLimit))
	}
	return out, nil
}

func moduleExamples() []Example {
	return []Example{
		{Title: "module_and_roots", Format: "text", Text: "module example.com/acme/contracts\nroots proto\n", Paths: []string{"module", "roots"}},
		{Title: "versioned_requirements", Format: "text", Text: "module example.com/acme/service\nrequire (\n    example.com/acme/common v1.2.3\n    example.com/acme/types/v2 v2.1.0 // indirect\n)\n", Paths: []string{"require"}},
		{Title: "explicit_commit", Description: "Synthetic SHA for syntax demonstration; not a fetched revision.", Format: "text", Text: "module example.com/acme/service\nrequire example.com/acme/common 1111111111111111111111111111111111111111\n", Paths: []string{"require[].version"}},
		{Title: "local_overlay", Description: "The local directory must exist to run commands; it is not included in the published lock.", Format: "text", Text: "module example.com/acme/service\nrequire example.com/acme/common v1.2.3\nreplace example.com/acme/common => ../common\n", Paths: []string{"replace"}},
	}
}

func lockExamples() []Example {
	return []Example{
		{Title: "empty_lock", YAML: "version: 1\nmodules: []\n", Paths: []string{"version", "modules"}},
		{Title: "version_pin", Description: "Synthetic revision and content hash; demonstrates the format, not a real downloaded module.", YAML: "version: 1\nmodules:\n  - source: example.com/acme/common\n    version: v1.2.3\n    commit: \"1111111111111111111111111111111111111111\"\n    hash: h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n", Paths: []string{"modules"}},
		{Title: "commit_pin", Description: "Synthetic 64-character Git commit; version equals commit. Not fetched content.", YAML: "version: 1\nmodules:\n  - source: example.com/acme/types/v2\n    version: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n    commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n    hash: h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n", Paths: []string{"modules[].version", "modules[].commit", "modules[].hash"}},
	}
}
