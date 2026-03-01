package easypconfig

func docsByPath() map[string]nodeDoc {
	return map[string]nodeDoc{
		"$": {
			Fields: []FieldDoc{
				{Path: "version", Type: "string", Required: false, Description: "Legacy compatibility field.", DefaultValue: "omitted", Examples: []string{"v1alpha"}},
				{Path: "lint", Type: "object", Required: false, Description: "Linter configuration and rule selection."},
				{Path: "deps", Type: "array<string>", Required: false, Description: "Dependency repositories in format <repo>@<version>."},
				{Path: "generate", Type: "object", Required: false, Description: "Code generation configuration."},
				{Path: "breaking", Type: "object", Required: false, Description: "Breaking changes check configuration."},
			},
			Examples: []Example{
				{
					Title:       "minimal_config",
					Description: "Small valid configuration with local input and one plugin.",
					YAML:        "lint:\n  use:\n    - DIRECTORY_SAME_PACKAGE\ngenerate:\n  inputs:\n    - directory: proto\n  plugins:\n    - name: go\n      out: .\n      opts:\n        paths: source_relative\n",
					Paths:       []string{"$", "lint", "generate"},
				},
				{
					Title:       "full_config_reference",
					Description: "Reference config touching all top-level sections.",
					YAML:        "version: v1alpha\nlint:\n  use:\n    - DEFAULT\n  enum_zero_value_suffix: _UNSPECIFIED\n  service_suffix: Service\n  ignore:\n    - vendor\n  ignore_only:\n    RPC_REQUEST_STANDARD_NAME:\n      - proto/legacy\ndeps:\n  - github.com/googleapis/googleapis@common-protos-1_3_1\ngenerate:\n  inputs:\n    - directory:\n        path: api\n        root: .\n    - git_repo:\n        url: github.com/acme/contracts@v1.2.3\n        sub_directory: proto\n        root: .\n  plugins:\n    - name: go\n      out: gen/go\n      opts:\n        paths: source_relative\n    - remote: api.easyp.tech/grpc/go:v1.5.1\n      out: gen/go\n      with_imports: true\n  managed:\n    enabled: true\n    disable:\n      - module: buf.build/googleapis/googleapis\n    override:\n      - file_option: go_package_prefix\n        value: github.com/acme/contracts/gen/go\nbreaking:\n  against_git_ref: main\n  ignore:\n    - proto/legacy\n",
					Paths:       []string{"$", "lint", "deps", "generate", "breaking"},
				},
			},
		},
		"lint": {
			Fields: []FieldDoc{
				{Path: "lint.use", Type: "array<string>", Required: false, Description: "Rule groups and/or individual lint rule names.", AllowedValues: lintUseAllowedValues(), DefaultValue: "[]"},
				{Path: "lint.enum_zero_value_suffix", Type: "string", Required: false, Description: "Required suffix for enum zero value.", DefaultValue: "UNSPECIFIED (runtime default)"},
				{Path: "lint.service_suffix", Type: "string", Required: false, Description: "Required suffix for service names.", DefaultValue: "Service (runtime default)"},
				{Path: "lint.ignore", Type: "array<string>", Required: false, Description: "Paths to exclude from linting.", DefaultValue: "[]"},
				{Path: "lint.except", Type: "array<string>", Required: false, Description: "Rules to disable globally.", DefaultValue: "[]"},
				{Path: "lint.allow_comment_ignores", Type: "boolean", Required: false, Description: "Allow inline ignore comments in proto files.", DefaultValue: "false"},
				{Path: "lint.ignore_only", Type: "map<string, array<string>>", Required: false, Description: "Disable specific rules only for selected paths.", DefaultValue: "{}"},
			},
			Examples: []Example{
				{
					Title:       "lint_groups_and_exceptions",
					Description: "Rule groups with selected exceptions and comment ignores enabled.",
					YAML:        "lint:\n  use:\n    - DEFAULT\n    - RPC_NO_CLIENT_STREAMING\n  except:\n    - COMMENT_RPC\n    - COMMENT_SERVICE\n  allow_comment_ignores: true\n",
					Paths:       []string{"lint"},
				},
				{
					Title:       "lint_ignore_only_rules",
					Description: "Disable specific rules only for selected paths.",
					YAML:        "lint:\n  use:\n    - DEFAULT\n  ignore_only:\n    PACKAGE_VERSION_SUFFIX:\n      - proto/legacy\n    RPC_REQUEST_STANDARD_NAME:\n      - proto/public\n",
					Paths:       []string{"lint"},
				},
			},
		},
		"deps": {
			Fields: []FieldDoc{
				{Path: "deps[]", Type: "string", Required: false, Description: "Dependency in format <repo>@<version>.", Examples: []string{"github.com/googleapis/googleapis@v1.0.0", "github.com/bufbuild/protoc-gen-validate"}},
			},
			Examples: []Example{
				{
					Title:       "deps_with_and_without_revision",
					Description: "Dependencies can be pinned or float to repository default revision.",
					YAML:        "deps:\n  - github.com/googleapis/googleapis@common-protos-1_3_1\n  - github.com/bufbuild/protoc-gen-validate\n",
					Paths:       []string{"deps"},
				},
			},
		},
		"generate": {
			Fields: []FieldDoc{
				{Path: "generate.inputs", Type: "array<object>", Required: true, Description: "Input sources for proto files.", DefaultValue: "must be provided"},
				{Path: "generate.plugins", Type: "array<object>", Required: true, Description: "Plugin definitions for generation.", DefaultValue: "must be provided"},
				{Path: "generate.managed", Type: "object", Required: false, Description: "Managed mode rules for file/field options.", DefaultValue: "{}"},
			},
			Examples: []Example{
				{
					Title:       "generate_local_and_remote_plugin",
					Description: "Local directory input with remote plugin execution.",
					YAML:        "generate:\n  inputs:\n    - directory:\n        path: api\n        root: .\n  plugins:\n    - remote: api.easyp.tech/protobuf/go:v1.36.10\n      out: .\n      opts:\n        paths: source_relative\n",
					Paths:       []string{"generate", "generate.inputs", "generate.plugins"},
				},
				{
					Title:       "generate_all_sections",
					Description: "Generate section with local+git inputs, multiple plugin styles, and managed mode.",
					YAML:        "generate:\n  inputs:\n    - directory: proto\n    - git_repo:\n        url: github.com/acme/contracts@v1.2.3\n        sub_directory: api\n  plugins:\n    - name: go\n      out: gen/go\n    - command: [\"go\", \"run\", \"example.com/protoc-gen-custom@latest\"]\n      out: gen/custom\n      with_imports: true\n  managed:\n    enabled: true\n    disable:\n      - module: buf.build/googleapis/googleapis\n    override:\n      - file_option: go_package_prefix\n        value: github.com/acme/contracts/gen/go\n",
					Paths:       []string{"generate", "generate.inputs", "generate.plugins", "generate.managed"},
				},
			},
		},
		"generate.inputs": {
			Fields: []FieldDoc{
				{Path: "generate.inputs[].directory", Type: "string | object", Required: false, Description: "Local input directory. Shorthand string or object with path/root."},
				{Path: "generate.inputs[].git_repo", Type: "object", Required: false, Description: "Remote git repository input."},
			},
			Examples: []Example{
				{
					Title:       "inputs_directory_and_git_repo",
					Description: "Each item uses exactly one source: directory or git_repo.",
					YAML:        "generate:\n  inputs:\n    - directory: proto\n    - git_repo:\n        url: github.com/acme/contracts@v1.0.0\n        sub_directory: api\n  plugins:\n    - name: go\n      out: .\n",
					Paths:       []string{"generate.inputs"},
				},
			},
			Notes: []string{
				"Each input item must contain exactly one of `directory` or `git_repo`.",
			},
		},
		"generate.inputs[].directory": {
			Fields: []FieldDoc{
				{Path: "generate.inputs[].directory.path", Type: "string", Required: true, Description: "Directory with .proto files (relative to config root unless absolute).", Examples: []string{"proto", "api/proto"}},
				{Path: "generate.inputs[].directory.root", Type: "string", Required: false, Description: "Import root for path normalization.", DefaultValue: "."},
			},
			Examples: []Example{
				{
					Title:       "input_directory_shorthand",
					Description: "Shorthand string form for directory input.",
					YAML:        "generate:\n  inputs:\n    - directory: proto\n  plugins:\n    - name: go\n      out: .\n",
					Paths:       []string{"generate.inputs[].directory"},
				},
				{
					Title:       "input_directory_object_with_root",
					Description: "Object form with explicit path and root.",
					YAML:        "generate:\n  inputs:\n    - directory:\n        path: api/proto\n        root: api\n  plugins:\n    - name: go\n      out: .\n",
					Paths:       []string{"generate.inputs[].directory"},
				},
			},
		},
		"generate.inputs[].git_repo": {
			Fields: []FieldDoc{
				{Path: "generate.inputs[].git_repo.url", Type: "string", Required: true, Description: "Git repo URL with optional revision.", Examples: []string{"github.com/acme/common@v1.0.0"}},
				{Path: "generate.inputs[].git_repo.sub_directory", Type: "string", Required: false, Description: "Subdirectory inside checked-out repository."},
				{Path: "generate.inputs[].git_repo.root", Type: "string", Required: false, Description: "Import root under repository contents.", DefaultValue: "\"\""},
			},
			Examples: []Example{
				{
					Title:       "input_git_repo_full",
					Description: "Git input with revision, subdirectory and import root.",
					YAML:        "generate:\n  inputs:\n    - git_repo:\n        url: github.com/acme/contracts@v1.3.0\n        sub_directory: proto/public\n        root: proto\n  plugins:\n    - name: go\n      out: .\n",
					Paths:       []string{"generate.inputs[].git_repo"},
				},
			},
		},
		"generate.plugins": {
			Fields: []FieldDoc{
				{Path: "generate.plugins[]", Type: "object", Required: true, Description: "Plugin item with exactly one source and required output directory."},
			},
			Examples: []Example{
				{
					Title:       "plugins_all_source_variants",
					Description: "Plugins can use name, remote, path, or command as a source.",
					YAML:        "generate:\n  plugins:\n    - name: go\n      out: gen/go\n    - remote: api.easyp.tech/protobuf/go:v1.36.10\n      out: gen/go\n    - path: ./bin/protoc-gen-custom\n      out: gen/custom\n    - command: [\"go\", \"run\", \"example.com/protoc-gen-alt@latest\"]\n      out: gen/alt\n",
					Paths:       []string{"generate.plugins"},
				},
			},
		},
		"generate.plugins[]": {
			Fields: []FieldDoc{
				{Path: "generate.plugins[].name", Type: "string", Required: false, Description: "Built-in/local plugin name (one source option).", Examples: []string{"go", "go-grpc"}},
				{Path: "generate.plugins[].remote", Type: "string", Required: false, Description: "Remote plugin endpoint (one source option).", Examples: []string{"api.easyp.tech/protobuf/go:v1.36.10"}},
				{Path: "generate.plugins[].path", Type: "string", Required: false, Description: "Explicit path to plugin binary (one source option)."},
				{Path: "generate.plugins[].command", Type: "array<string>", Required: false, Description: "Command invocation for plugin (one source option).", Examples: []string{`["go","run","github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1"]`}},
				{Path: "generate.plugins[].out", Type: "string", Required: true, Description: "Output directory for generated files.", Examples: []string{".", "gen/go"}},
				{Path: "generate.plugins[].opts", Type: "map<string, string | []string>", Required: false, Description: "Plugin options; value can be scalar or array of scalars."},
				{Path: "generate.plugins[].with_imports", Type: "boolean", Required: false, Description: "Include dependency protos in generation.", DefaultValue: "false"},
			},
			Examples: []Example{
				{
					Title:       "plugin_remote",
					Description: "Remote plugin source.",
					YAML:        "generate:\n  plugins:\n    - remote: api.easyp.tech/grpc/go:v1.5.1\n      out: .\n      opts:\n        paths: source_relative\n",
					Paths:       []string{"generate.plugins[]"},
				},
				{
					Title:       "plugin_command",
					Description: "Command-based plugin source.",
					YAML:        "generate:\n  plugins:\n    - command: [\"go\", \"run\", \"github.com/bufbuild/protoc-gen-validate@v0.10.1\"]\n      out: gen/go\n",
					Paths:       []string{"generate.plugins[]"},
				},
				{
					Title:       "plugin_name",
					Description: "Built-in plugin selected by name.",
					YAML:        "generate:\n  plugins:\n    - name: go-grpc\n      out: gen/go\n",
					Paths:       []string{"generate.plugins[]"},
				},
				{
					Title:       "plugin_path",
					Description: "Plugin binary loaded from explicit local path.",
					YAML:        "generate:\n  plugins:\n    - path: ./bin/protoc-gen-openapi\n      out: gen/openapi\n",
					Paths:       []string{"generate.plugins[]"},
				},
				{
					Title:       "plugin_opts_scalar_and_array",
					Description: "Plugin opts values can be scalar or arrays of scalars.",
					YAML:        "generate:\n  plugins:\n    - remote: api.easyp.tech/community/stephenh-ts-proto:v1.178.0\n      out: gen/ts\n      opts:\n        env: node\n        outputServices:\n          - grpc-js\n          - generic-definitions\n        useExactTypes: false\n",
					Paths:       []string{"generate.plugins[]"},
				},
				{
					Title:       "plugin_with_imports_enabled",
					Description: "Enable generation for dependency protos as well.",
					YAML:        "generate:\n  plugins:\n    - name: go\n      out: gen/go\n      with_imports: true\n",
					Paths:       []string{"generate.plugins[]"},
				},
			},
		},
		"generate.managed": {
			Fields: []FieldDoc{
				{Path: "generate.managed.enabled", Type: "boolean", Required: false, Description: "Enable managed mode option rewriting.", DefaultValue: "false"},
				{Path: "generate.managed.disable", Type: "array<object>", Required: false, Description: "Disable managed mode per module/path/option."},
				{Path: "generate.managed.override", Type: "array<object>", Required: false, Description: "Override file/field options with values."},
			},
			Examples: []Example{
				{
					Title:       "managed_mode_full",
					Description: "Managed mode with both disable and override rules.",
					YAML:        "generate:\n  managed:\n    enabled: true\n    disable:\n      - module: buf.build/googleapis/googleapis\n      - field_option: jstype\n        field: acme.v1.Message.count\n    override:\n      - file_option: go_package_prefix\n        value: github.com/acme/contracts/gen/go\n      - field_option: jstype\n        field: acme.v1.Message.count\n        value: JS_STRING\n",
					Paths:       []string{"generate.managed"},
				},
			},
		},
		"generate.managed.disable": {
			Fields: []FieldDoc{
				{Path: "generate.managed.disable[].module", Type: "string", Required: false, Description: "Apply disable to module."},
				{Path: "generate.managed.disable[].path", Type: "string", Required: false, Description: "Apply disable to path."},
				{Path: "generate.managed.disable[].file_option", Type: "string", Required: false, Description: "Disable this file option."},
				{Path: "generate.managed.disable[].field_option", Type: "string", Required: false, Description: "Disable this field option."},
				{Path: "generate.managed.disable[].field", Type: "string", Required: false, Description: "Field selector for field_option."},
			},
			Examples: []Example{
				{
					Title:       "managed_disable_variants",
					Description: "Disable rules by path, file option, and field option.",
					YAML:        "generate:\n  managed:\n    disable:\n      - path: proto/third_party\n      - file_option: java_package\n      - field_option: jstype\n        field: acme.v1.Message.count\n",
					Paths:       []string{"generate.managed.disable"},
				},
			},
			Notes: []string{
				"At least one key in each disable item is required.",
				"`file_option` and `field_option` cannot be used together.",
				"`field` requires `field_option`.",
			},
		},
		"generate.managed.override": {
			Fields: []FieldDoc{
				{Path: "generate.managed.override[].file_option", Type: "string", Required: false, Description: "Target file option to override."},
				{Path: "generate.managed.override[].field_option", Type: "string", Required: false, Description: "Target field option to override."},
				{Path: "generate.managed.override[].value", Type: "any", Required: true, Description: "Override value."},
				{Path: "generate.managed.override[].module", Type: "string", Required: false, Description: "Optional module selector."},
				{Path: "generate.managed.override[].path", Type: "string", Required: false, Description: "Optional path selector."},
				{Path: "generate.managed.override[].field", Type: "string", Required: false, Description: "Optional field selector (for field_option)."},
			},
			Examples: []Example{
				{
					Title:       "managed_override_file_option",
					Description: "Override a file option with a custom value.",
					YAML:        "generate:\n  managed:\n    override:\n      - file_option: java_package_prefix\n        value: com.acme.generated\n",
					Paths:       []string{"generate.managed.override"},
				},
				{
					Title:       "managed_override_field_option",
					Description: "Override a field option for a selected field.",
					YAML:        "generate:\n  managed:\n    override:\n      - field_option: jstype\n        field: acme.v1.Message.count\n        value: JS_NUMBER\n",
					Paths:       []string{"generate.managed.override"},
				},
			},
			Notes: []string{
				"Each override item requires exactly one of file_option or field_option.",
				"`field` can only be used with `field_option`.",
			},
		},
		"breaking": {
			Fields: []FieldDoc{
				{Path: "breaking.ignore", Type: "array<string>", Required: false, Description: "Paths excluded from breaking-change checks.", DefaultValue: "[]"},
				{Path: "breaking.against_git_ref", Type: "string", Required: false, Description: "Branch/tag/commit used for comparison."},
			},
			Examples: []Example{
				{
					Title:       "breaking_with_ref_and_ignore",
					Description: "Compare against selected git ref and ignore selected directories.",
					YAML:        "breaking:\n  against_git_ref: origin/main\n  ignore:\n    - proto/experimental\n",
					Paths:       []string{"breaking"},
				},
			},
		},
	}
}
