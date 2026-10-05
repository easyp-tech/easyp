package migration

import (
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/rules"
)

type policyOutput struct {
	Version  string                       `yaml:"version"`
	Linters  linterOutput                 `yaml:"linters"`
	Settings map[string]map[string]string `yaml:"linters-settings"`
	Issues   issueOutput                  `yaml:"issues"`
	Breaking breakingOutput               `yaml:"breaking"`
}

type linterOutput struct {
	Default             string   `yaml:"default"`
	Enable              []string `yaml:"enable,omitempty"`
	Disable             []string `yaml:"disable,omitempty"`
	AllowCommentIgnores bool     `yaml:"allow_comment_ignores"`
}

type issueOutput struct {
	ExcludeRules []exclusionOutput `yaml:"exclude-rules,omitempty"`
}
type exclusionOutput struct {
	Path    string   `yaml:"path"`
	Linters []string `yaml:"linters,omitempty"`
}
type breakingOutput struct {
	Baseline   string   `yaml:"baseline,omitempty"`
	Categories []string `yaml:"categories,omitempty"`
	Ignore     []string `yaml:"ignore,omitempty"`
}

type generateOutput struct {
	Version  string             `yaml:"version"`
	Generate targetsOutput      `yaml:"generate"`
	Plugins  []pluginOutput     `yaml:"plugins"`
	Options  v1.GenerateOptions `yaml:"options"`
}
type targetsOutput struct {
	Modules  []string           `yaml:"modules,omitempty"`
	Packages []string           `yaml:"packages,omitempty"`
	Paths    []string           `yaml:"paths,omitempty"`
	Managed  config.ManagedMode `yaml:"managed,omitempty"`
}
type pluginOutput struct {
	Name        string            `yaml:"name,omitempty"`
	Path        string            `yaml:"path,omitempty"`
	Command     []string          `yaml:"command,omitempty"`
	Remote      string            `yaml:"remote,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Out         string            `yaml:"out"`
	Opts        config.PluginOpts `yaml:"opts,omitempty"`
	WithImports bool              `yaml:"with_imports,omitempty"`
}

func convertPolicy(cfg legacyConfig) ([]byte, error) {
	use, err := expandRules(cfg.Lint.Use)
	if err != nil {
		return nil, fmt.Errorf("expandRules: %w", err)
	}
	except, err := expandRules(cfg.Lint.Except)
	if err != nil {
		return nil, fmt.Errorf("expandRules: %w", err)
	}
	effective := slices.DeleteFunc(use, func(name string) bool { return slices.Contains(except, name) })
	minimal, err := expandRules([]string{"MINIMAL"})
	if err != nil {
		return nil, fmt.Errorf("expandRules: %w", err)
	}
	output := policyOutput{Version: "v1", Linters: linterOutput{
		Default: "MINIMAL", Enable: effective, AllowCommentIgnores: cfg.Lint.AllowCommentIgnores,
	}}
	for _, rule := range minimal {
		if !slices.Contains(effective, rule) {
			output.Linters.Disable = append(output.Linters.Disable, rule)
		}
	}
	output.Settings = make(map[string]map[string]string)
	if cfg.Lint.EnumZeroValueSuffix != "" {
		output.Settings["ENUM_ZERO_VALUE_SUFFIX"] = map[string]string{"suffix": cfg.Lint.EnumZeroValueSuffix}
	}
	if cfg.Lint.ServiceSuffix != "" {
		output.Settings["SERVICE_SUFFIX"] = map[string]string{"suffix": cfg.Lint.ServiceSuffix}
	}
	for _, ignore := range cfg.Lint.Ignore {
		paths, err := literalExclusionPaths(ignore)
		if err != nil {
			return nil, fmt.Errorf("literalExclusionPaths: %w", err)
		}
		for _, pattern := range paths {
			output.Issues.ExcludeRules = append(output.Issues.ExcludeRules, exclusionOutput{Path: pattern})
		}
	}
	keys := make([]string, 0, len(cfg.Lint.IgnoreOnly))
	for key := range cfg.Lint.IgnoreOnly {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	ignored := make(map[string][]string)
	for _, key := range keys {
		names, err := expandRules([]string{key})
		if err != nil {
			return nil, fmt.Errorf("expandRules: %w", err)
		}
		paths := cfg.Lint.IgnoreOnly[key]
		for _, name := range names {
			if previous, ok := ignored[name]; ok && !slices.Equal(previous, paths) {
				return nil, fmt.Errorf("lint.ignore_only overlaps for %s with different paths; make the legacy selection unambiguous before migration", name)
			}
			ignored[name] = paths
		}
		for _, ignore := range paths {
			patterns, err := prefixExclusionPaths(ignore)
			if err != nil {
				return nil, fmt.Errorf("prefixExclusionPaths: %w", err)
			}
			for _, pattern := range patterns {
				output.Issues.ExcludeRules = append(output.Issues.ExcludeRules, exclusionOutput{Path: pattern, Linters: names})
			}
		}
	}
	output.Breaking = breakingOutput{Categories: cfg.Breaking.Use, Ignore: cfg.Breaking.Ignore}
	for _, category := range cfg.Breaking.Use {
		if category != "FILE" {
			return nil, fmt.Errorf("unsupported legacy breaking.use %q; migrate breaking policy manually", category)
		}
	}
	for _, ignore := range cfg.Breaking.Ignore {
		if err := literalPolicyPath(ignore); err != nil {
			return nil, fmt.Errorf("literalPolicyPath: %w", err)
		}
	}
	if cfg.Breaking.AgainstGitRef != "" {
		output.Breaking.Baseline = "git:" + cfg.Breaking.AgainstGitRef
	}
	raw, err := yaml.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("Marshal: %w", err)
	}
	return raw, nil
}

func expandRules(selection []string) ([]string, error) {
	for _, name := range selection {
		if strings.Contains(name, "$") {
			return nil, fmt.Errorf("rule placeholder %q prevents safe lint-set analysis; migrate manually", name)
		}
	}
	if err := rules.ValidateNames(selection); err != nil {
		return nil, fmt.Errorf("ValidateNames: %w", err)
	}
	groups := make(map[string][]string)
	for _, group := range rules.AllGroups() {
		groups[group.Key] = group.Rules
	}
	var result []string
	for _, name := range selection {
		names, ok := groups[name]
		if !ok {
			names = []string{name}
		}
		for _, rule := range names {
			if !slices.Contains(result, rule) {
				result = append(result, rule)
			}
		}
	}
	return result, nil
}

func literalPolicyPath(value string) error {
	if strings.Contains(value, "$") {
		return fmt.Errorf("exclusion path placeholder %q prevents safe literal/glob analysis; migrate manually", value)
	}
	if path.IsAbs(value) || strings.ContainsAny(value, "\\:\x00") {
		return fmt.Errorf("exclusion path %q must be relative to easyp.yaml; migrate this path manually", value)
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == ".." {
			return fmt.Errorf("exclusion path %q leaves the policy directory; manual migration required", value)
		}
	}
	return nil
}

func escapeGlobLiteral(value string) string {
	return strings.NewReplacer("*", "[*]", "?", "[?]", "[", "[[]").Replace(value)
}

func literalExclusionPaths(value string) ([]string, error) {
	if err := literalPolicyPath(value); err != nil {
		return nil, fmt.Errorf("literalPolicyPath: %w", err)
	}
	value = filepath.ToSlash(filepath.Clean(value))
	if value == "." {
		return []string{"**"}, nil
	}
	escaped := escapeGlobLiteral(value)
	if escaped != value || strings.HasSuffix(value, ".proto") {
		return []string{escaped, escaped + "/**"}, nil
	}
	return []string{value}, nil
}

func prefixExclusionPaths(value string) ([]string, error) {
	if err := literalPolicyPath(value); err != nil {
		return nil, fmt.Errorf("literalPolicyPath: %w", err)
	}
	if value == "" {
		return []string{"**"}, nil
	}
	if strings.TrimSuffix(value, "/") != filepath.ToSlash(filepath.Clean(value)) {
		return nil, fmt.Errorf("ignore_only raw prefix %q has noncanonical path segments; manual migration required", value)
	}
	// v0 used strings.HasPrefix, not directory boundaries or glob matching.
	// The second pattern covers descendants; the first covers the prefix's own
	// segment. Escaping preserves metacharacters as literal filename bytes.
	prefix := escapeGlobLiteral(value) + "*"
	return []string{prefix, prefix + "/**"}, nil
}

func convertGenerate(cfg legacyConfig, selected, packages, paths []string) ([]byte, error) {
	// Empty explicitly disables v1 inheritance of a parent's Go package prefix.
	emptyPrefix := ""
	output := generateOutput{
		Version: "v1", Generate: targetsOutput{Modules: selected, Packages: packages, Paths: paths, Managed: cfg.Generate.Managed},
		Plugins: []pluginOutput{}, Options: v1.GenerateOptions{Go: v1.GoOptions{PackagePrefix: &emptyPrefix}},
	}
	for i, plugin := range cfg.Generate.Plugins {
		converted := pluginOutput{Name: plugin.Name, Path: plugin.Path, Command: plugin.Command,
			Remote: plugin.Remote, Out: plugin.Out, Opts: plugin.Opts, WithImports: plugin.WithImports}
		if plugin.Remote != "" {
			colon := strings.LastIndexByte(plugin.Remote, ':')
			if colon <= strings.LastIndexByte(plugin.Remote, '/') || !semver.IsValid(plugin.Remote[colon+1:]) {
				return nil, fmt.Errorf("generate.plugins[%d].remote needs a pinned semantic version (remote:version); a placeholder hiding the version requires manual migration", i)
			}
			converted.Remote, converted.Version = plugin.Remote[:colon], plugin.Remote[colon+1:]
		}
		candidate := v1.Plugin{Name: converted.Name, Path: converted.Path, Command: converted.Command,
			Remote: converted.Remote, Version: converted.Version, Out: converted.Out, Opts: v1.PluginOptions(converted.Opts), WithImports: converted.WithImports}
		if err := candidate.Validate(); err != nil {
			return nil, fmt.Errorf("Validate: %w", err)
		}
		output.Plugins = append(output.Plugins, converted)
	}
	raw, err := yaml.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("Marshal: %w", err)
	}
	return raw, nil
}
