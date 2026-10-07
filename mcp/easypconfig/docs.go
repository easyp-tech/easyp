package easypconfig

import v1 "github.com/easyp-tech/easyp/internal/config/v1"

var descriptions = map[string]map[string]string{
	v1.LockFile: {
		"version":                              "Lock format version: integer 1, not the YAML configuration string v1.",
		"modules":                              "Published dependency graph, with one verified revision per module identity. An empty list is valid for a dependency-free module.",
		"modules[].source":                     "Exact module identity; separate v2+ lines use their matching /vN suffix.",
		"modules[].version":                    "Selected semantic version or a full commit. A commit-valued version must equal commit; major suffix and legacy +incompatible provenance are checked.",
		"modules[].commit":                     "Full 40- or 64-character hexadecimal Git commit. Tags and short SHAs are not commit pins.",
		"modules[].hash":                       "h1: followed by standard base64 encoding of a 32-byte SHA-256 content digest. Fetch/install verify the materialized logical snapshot from the pinned Git tree, including resolved internal aliases; a sample hash is not proof of a real repository.",
		"modules[].roots":                      "Verified fallback import-root directories relative to a dependency without root metadata. Inferred only from imports inside that dependency or supplied by checked get --import-root/migration hints, then replayed on cold-cache and frozen operations. Consumer imports validate the resulting namespace but never select roots. Authoritative native/Buf/legacy roots retain priority; roots are independent of generation paths/packages.",
		"modules[].bsr":                        "BSR requests declared by this Git dependency and their recorded Git targets; frozen commands replay these bindings without calling a resolver.",
		"modules[].bsr[].dependency.module":    "Original BSR identity, separate from the selected Git source.",
		"modules[].bsr[].dependency.reference": "Requested Buf label or BSR commit; preserved even when a compatibility snapshot cannot prove equivalence.",
		"modules[].bsr[].dependency.commit":    "Optional 32-character hexadecimal BSR commit from buf.lock; not a Git commit.",
		"modules[].bsr[].dependency.digest":    "Optional opaque BSR digest from buf.lock. Compatibility snapshots preserve but do not verify it.",
		"modules[].bsr[].dependency.config":    "Declaring buf.yaml path relative to the verified Git checkout, including v1 workspace directories.",
		"modules[].bsr[].git.module":           "Git module selected by the BSR resolver; it participates in the existing dependency graph.",
		"modules[].bsr[].git.version":          "Pinned Git semantic version or full commit. Moving references and implicit latest are not accepted.",
		"modules[].bsr[].resolution":           "compatibility_snapshot does not guarantee BSR revision equivalence or verify its digest. exact is reserved for a backend that proves the correspondence.",
	},
	v1.PolicyFile: {
		"version":                       "Configuration version; v1 is the only supported value.",
		"linters":                       "Lint rule selection for this module.",
		"linters.default":               "Preset used before enable and disable rules.",
		"linters.enable":                "Implemented lint rules or groups to enable; repeated selections are idempotent.",
		"linters.disable":               "Implemented lint rules or groups to disable; disabling takes precedence over enabling.",
		"linters.allow_comment_ignores": "Allow scoped comment suppressions (default true). easyp:disable RULE annotates a declaration or a block; paired easyp:enable ends an explicit region. Legacy nolint:RULE and buf:lint:ignore RULE are also recognized.",
		"linters.extends":               "Base lint policy from a workspace-relative file or a declared, locked module; local adjustments override it.",
		"linters-settings":              "Settings for supported lint rules.",
		"linters-settings.ENUM_ZERO_VALUE_SUFFIX": "Suffix for zero enum values.",
		"linters-settings.SERVICE_SUFFIX":         "Suffix for service names.",
		"issues":                                  "Rules for suppressing lint findings.",
		"issues.exclude-rules":                    "Suppress lint findings matching a rule.",
		"issues.exclude-rules[].path":             "Policy-source-relative path or glob; literal directories include descendants, ** matches directory segments.",
		"issues.exclude-rules[].linters":          "Implemented lint rules or groups suppressed by this exclusion; unknown or unimplemented names are rejected.",
		"breaking":                                "Compatibility checks against a Git baseline.",
		"breaking.baseline":                       "Empty or git:<ref> baseline; an explicit --against overrides it, while the CLI default does not. Runtime and config validation enforce the same syntax.",
		"breaking.ignore":                         "Paths ignored by breaking checks.",
		"breaking.categories":                     "Compatibility profiles: FILE, PACKAGE, WIRE_JSON and WIRE. Omitted or empty defaults to FILE for v1 policies; profiles use compiled descriptors.",
		"breaking.extends":                        "Base breaking policy from a local file or declared, locked module; evaluated separately for every checked module.",
		"breaking.ignore_unstable":                "Ignore only declarations in packages with unstable version suffixes; stable package checks remain active.",
	},
	v1.GenerateFile: {
		"version":                                  "Configuration version; v1 is the only supported value.",
		"generate":                                 "Select modules for this generation run.",
		"generate.modules":                         "Module identities or workspace module paths; strings select whole modules, objects add their own paths/packages intersected with global filters. Omitted selects the local module containing the generator.",
		"generate.modules[].module":                "Module identity or workspace-relative module directory.",
		"generate.modules[].paths":                 "Literal module-directory-relative files or subtrees, intersected with global filters; every selector must match in this module.",
		"generate.modules[].packages":              "Exact protobuf packages in this module, intersected with global filters; every selector must match in this module.",
		"generate.packages":                        "Exact protobuf package names among selected modules, intersected with generate.paths. Empty means all packages; imports remain available for compilation.",
		"generate.paths":                           "Literal module-relative files or directory subtrees across selected modules, intersected with generate.packages and module filters. Available without generate.modules. Empty means all sources. Distinct from plugins.opts.paths, which controls plugin output layout.",
		"generate.managed":                         "Managed file and field option rules.",
		"plugins":                                  "Generators executed for selected modules.",
		"plugins[].name":                           "Local or built-in plugin name.",
		"plugins[].path":                           "Explicit plugin binary path, relative to the generation working directory when not absolute.",
		"plugins[].command":                        "Custom plugin executable and arguments, run from the generation working directory.",
		"plugins[].remote":                         "Remote plugin endpoint; requires a pinned version.",
		"plugins[].version":                        "Pinned version for a remote plugin.",
		"plugins[].out":                            "Output directory for generated files.",
		"plugins[].with_imports":                   "Generate transitive imports with this plugin only; false by default. Not the descriptor include_imports flag.",
		"plugins[].opts":                           "Plugin options as a list or map of scalar values.",
		"options":                                  "Language-specific generator options.",
		"options.go":                               "Go generator options.",
		"options.go.package_prefix":                "Go package prefix; may be inherited. Changes only Go options unless generate.managed.enabled explicitly enables full managed mode.",
		"generate.managed.enabled":                 "Enable managed descriptor options.",
		"generate.managed.disable":                 "Rules that prevent managed mode from changing matching options.",
		"generate.managed.override":                "Rules that set file or field options.",
		"generate.managed.disable[].module":        "Limit the disable rule to a module.",
		"generate.managed.disable[].package":       "Limit the disable rule to a protobuf package.",
		"generate.managed.disable[].path":          "Limit the disable rule to a file or directory path.",
		"generate.managed.disable[].file_option":   "File option that managed mode must leave unchanged.",
		"generate.managed.disable[].field_option":  "Field option that managed mode must leave unchanged.",
		"generate.managed.disable[].field":         "Fully qualified field selected with field_option.",
		"generate.managed.override[].module":       "Limit the override to a module.",
		"generate.managed.override[].package":      "Limit the override to a protobuf package.",
		"generate.managed.override[].path":         "Limit the override to a file or directory path.",
		"generate.managed.override[].file_option":  "Supported managed file option to set; unknown names are rejected before generation.",
		"generate.managed.override[].field_option": "Supported managed field option to set; currently jstype.",
		"generate.managed.override[].field":        "Fully qualified field selected with field_option.",
		"generate.managed.override[].value":        "Typed value for the selected option: a string, boolean, or supported enum name. Wrong types and enum values are rejected.",
	},
}

func examplesFor(file string) []Example {
	switch file {
	case v1.ModuleFile:
		return moduleExamples()
	case v1.LockFile:
		return lockExamples()
	case v1.PolicyFile:
		return []Example{
			{Title: "shared_lint", Description: "Create .policies/lint.yaml with a linters section in this workspace before execution.", YAML: "version: v1\nlinters:\n  extends: ./.policies/lint.yaml\n  disable: [PACKAGE_VERSION_SUFFIX]\n", Paths: []string{"linters.extends"}},
			{Title: "lint_policy", YAML: "version: v1\nlinters:\n  default: STANDARD\n  enable: [FILE_LOWER_SNAKE_CASE]\n", Paths: []string{"linters", "linters.default", "linters.enable"}},
			{Title: "breaking_policy", YAML: "version: v1\nbreaking:\n  baseline: git:main\n  categories: [FILE]\n", Paths: []string{"breaking", "breaking.baseline", "breaking.categories"}},
		}
	case v1.GenerateFile:
		return []Example{
			{Title: "package_selection", Description: "Requires an api.v1 package in the selected module sources.", YAML: "version: v1\ngenerate:\n  packages: [api.v1]\nplugins:\n  - name: go\n    out: gen\n    with_imports: true\n", Paths: []string{"generate.packages"}},
			{Title: "path_selection", Description: "Requires a source in the mcp import subtree of a selected module; plugin paths controls generated output layout.", YAML: "version: v1\ngenerate:\n  paths: [mcp]\nplugins:\n  - name: go\n    out: gen\n    opts: [paths=source_relative]\n", Paths: []string{"generate.paths"}},
			{Title: "local_plugin", YAML: "version: v1\nplugins:\n  - name: go\n    out: gen/go\n    opts: [paths=source_relative]\n", Paths: []string{"plugins", "plugins[]", "plugins[].name", "plugins[].out", "plugins[].opts"}},
			{Title: "remote_plugin", YAML: "version: v1\nplugins:\n  - remote: plugins.beta.easyp.tech/protocolbuffers/go\n    version: v1.36.11\n    out: gen/go\n    opts: [paths=source_relative]\n", Paths: []string{"plugins", "plugins[]", "plugins[].remote", "plugins[].version"}},
			{Title: "binary_path", YAML: "version: v1\nplugins:\n  - path: ./tools/protoc-gen-custom\n    out: gen/custom\n", Paths: []string{"plugins", "plugins[]", "plugins[].path"}},
			{Title: "custom_command", YAML: "version: v1\nplugins:\n  - command: [sh, ./tools/protoc-plugin.sh]\n    out: gen/script\n", Paths: []string{"plugins", "plugins[]", "plugins[].command"}},
			{Title: "managed_mode", YAML: "version: v1\ngenerate:\n  managed:\n    enabled: true\n    override:\n      - file_option: go_package_prefix\n        value: example.com/gen\nplugins:\n  - name: go\n    out: gen/go\n", Paths: []string{"generate", "generate.managed", "generate.managed.override"}},
		}
	default:
		return nil
	}
}

func selectExamples(file, path string, limit int) []Example {
	var selected []Example
	for _, example := range examplesFor(file) {
		if path != "$" {
			matches := false
			for _, examplePath := range example.Paths {
				if within(examplePath, path) || within(path, examplePath) {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}
		selected = append(selected, example)
		if len(selected) == limit {
			break
		}
	}
	return selected
}

func notesFor(file, path string) []string {
	switch {
	case file == v1.LockFile:
		return []string{"Schema comes directly from v1.SchemaJSON. ParseLock adds semantic checks for duplicate sources, matching module majors and commit-valued versions. Examples contain synthetic commits/hashes, not fetched data.", "Local replace never rewrites the published lock. --frozen requires a valid complete manifest/lock graph and rejects replacements; it may download exact pinned contents but never resolves a new version. Preserve the lock when investigating a cache mismatch.", "BSR compatibility_snapshot bindings preserve the original request and Buf lock pin, but do not prove BSR revision equivalence or verify the BSR digest. Frozen commands validate metadata and reuse recorded Git targets; older locks with BSR dependencies require easyp mod tidy."}
	case file == v1.GenerateFile && within("generate.packages", path):
		return []string{"Names match exact protobuf packages, not prefixes or file paths. Package and path filters intersect. Every selector must match an output source across the selected modules of each project, even without plugins. Unknown names fail before plugins and descriptor writes. Generation filters are not inherited. with_imports is independent per plugin; descriptor include_imports controls exported dependencies."}
	case file == v1.GenerateFile && within("generate.paths", path):
		return []string{"Selectors match literal module-relative file names or component-bounded directory subtrees. The base is the selected module directory, not its protobuf import roots or the generator file. Empty or omitted paths select all sources. Canonical portable relative paths cannot contain whitespace, dot segments, absolute paths, backslashes or globs. Global and module-local package/path filters intersect; global selectors must match across selected modules, scoped selectors must match their own module, even without plugins. Unknown selectors fail before plugins and descriptor writes. Module roots, source boundaries and required imports remain unchanged; no gitignore filtering is added. Global filters can be used without generate.modules. Generation filters are not inherited. plugins.opts.paths is a plugin output-layout option; with_imports and descriptor include_imports retain their separate meanings."}
	case file == v1.PolicyFile && path == "linters.extends":
		return []string{"Use ./ or ../ for local files, or declared-module#policy-path. Versions belong only in protobuf.mod/protobuf.lock. Local adjustments override the base; issues are not inherited. validate-config requires verified cached content and never downloads it."}
	case file == v1.PolicyFile && path == "issues.exclude-rules[].path":
		return []string{"Paths are relative to the easyp.yaml providing the effective issues section, not the invocation directory. Supports *, ?, character classes and whole-segment **. A matching exclusion without linters suppresses every rule for that file."}
	case file == v1.PolicyFile && path == "breaking.categories":
		return []string{"FILE checks generated-source and file identity; PACKAGE permits moves within a package; WIRE_JSON protects binary and JSON; WIRE protects binary encoding. Multiple profiles select their union without duplicate findings. Omitted or empty categories default to FILE for v1 policies. Profile catalog and limits are documented in .spec/config/breaking-profiles.md."}
	case file == v1.PolicyFile && path == "breaking.ignore_unstable":
		return []string{"Filters unstable version-suffixed packages in both comparison revisions. Stable package declarations, imports and references remain checked."}
	case file == v1.PolicyFile && path == "breaking.extends":
		return []string{"Use ./ or ../ for local files, or declared-module#policy-path. Baseline and ignore paths apply at the consuming policy. Explicit false and empty lists override the base. Frozen mode uses verified pins and rejects local replacements."}
	default:
		return nil
	}
}
