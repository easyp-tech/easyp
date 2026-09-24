package easypconfig

import v1 "github.com/easyp-tech/easyp/internal/config/v1"

var descriptions = map[string]map[string]string{
	v1.PolicyFile: {
		"version":          "Configuration version; v1 is the only supported value.",
		"linters":          "Lint rule selection for this module.",
		"linters.default":  "Preset used before enable and disable rules.",
		"linters.enable":   "Additional lint rules to enable.",
		"linters.disable":  "Lint rules to disable.",
		"linters.extends":  "Reserved reference to another lint policy.",
		"linters-settings": "Settings for supported lint rules.",
		"linters-settings.ENUM_ZERO_VALUE_SUFFIX": "Suffix for zero enum values.",
		"linters-settings.SERVICE_SUFFIX":         "Suffix for service names.",
		"issues":                                  "Rules for suppressing lint findings.",
		"issues.exclude-rules":                    "Suppress lint findings matching a rule.",
		"issues.exclude-rules[].path":             "Reserved path matcher for excluded findings.",
		"issues.exclude-rules[].linters":          "Lint rules suppressed by this exclusion.",
		"breaking":                                "Compatibility checks against a Git baseline.",
		"breaking.baseline":                       "Baseline in git:<ref> form.",
		"breaking.ignore":                         "Paths ignored by breaking checks.",
		"breaking.categories":                     "Reserved breaking check category filter.",
		"breaking.extends":                        "Reserved reference to another breaking policy.",
		"breaking.ignore_unstable":                "Reserved flag for ignoring unstable declarations.",
	},
	v1.GenerateFile: {
		"version":                                  "Configuration version; v1 is the only supported value.",
		"generate":                                 "Select modules for this generation run.",
		"generate.modules":                         "Module identities selected from protobuf.mod dependencies or the local workspace.",
		"generate.packages":                        "Reserved protobuf package filter; execution is not implemented yet.",
		"generate.managed":                         "Managed file and field option rules.",
		"plugins":                                  "Generators executed for selected modules.",
		"plugins[].name":                           "Local or built-in plugin name.",
		"plugins[].path":                           "Explicit plugin binary path, relative to the generation working directory when not absolute.",
		"plugins[].command":                        "Custom plugin executable and arguments, run from the generation working directory.",
		"plugins[].remote":                         "Remote plugin endpoint; requires a pinned version.",
		"plugins[].version":                        "Pinned version for a remote plugin.",
		"plugins[].out":                            "Output directory for generated files.",
		"plugins[].opts":                           "Plugin options as a list or map of scalar values.",
		"options":                                  "Language-specific generator options.",
		"options.go":                               "Go generator options.",
		"options.go.package_prefix":                "Go package prefix; may be inherited from an ancestor generator file.",
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
		"generate.managed.override[].file_option":  "File option to set.",
		"generate.managed.override[].field_option": "Field option to set.",
		"generate.managed.override[].field":        "Fully qualified field selected with field_option.",
		"generate.managed.override[].value":        "Value assigned to the selected option.",
	},
}

func examplesFor(file string) []Example {
	switch file {
	case v1.PolicyFile:
		return []Example{
			{Title: "lint_policy", YAML: "version: v1\nlinters:\n  default: STANDARD\n  enable: [FILE_LOWER_SNAKE_CASE]\n", Paths: []string{"linters", "linters.default", "linters.enable"}},
			{Title: "breaking_policy", YAML: "version: v1\nbreaking:\n  baseline: git:main\n", Paths: []string{"breaking", "breaking.baseline"}},
		}
	case v1.GenerateFile:
		return []Example{
			{Title: "local_plugin", YAML: "version: v1\nplugins:\n  - name: go\n    out: gen/go\n    opts: [paths=source_relative]\n", Paths: []string{"plugins", "plugins[]", "plugins[].name", "plugins[].out", "plugins[].opts"}},
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
	case file == v1.GenerateFile && within("generate.packages", path):
		return []string{"generate.packages is reserved and generation fails when it is set."}
	case file == v1.PolicyFile && path == "linters.extends":
		return []string{"linters.extends is not implemented; lint fails when it is set."}
	case file == v1.PolicyFile && path == "issues.exclude-rules[].path":
		return []string{"issues.exclude-rules.path matching is not implemented; lint fails when it is set."}
	case file == v1.PolicyFile && path == "breaking.categories":
		return []string{"breaking.categories is not supported by the current checker."}
	case file == v1.PolicyFile && path == "breaking.ignore_unstable":
		return []string{"breaking.ignore_unstable is not supported by the current checker."}
	case file == v1.PolicyFile && path == "breaking.extends":
		return []string{"breaking.extends is not implemented; breaking checks fail when it is set."}
	default:
		return nil
	}
}
