package migration

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// These types intentionally describe v0, rather than the runtime config, whose
// legacy fields may be removed independently of the explicit migration tool.
type legacyConfig struct {
	Version  string            `yaml:"version"`
	Lint     config.LintConfig `yaml:"lint"`
	Deps     []string          `yaml:"deps"`
	Generate legacyGenerate    `yaml:"generate"`
	Breaking legacyBreaking    `yaml:"breaking"`
}

type legacyBreaking struct {
	Ignore        []string `yaml:"ignore"`
	AgainstGitRef string   `yaml:"against_git_ref"`
	Use           []string `yaml:"use"`
}

type legacyGenerate struct {
	Inputs  []legacyInput      `yaml:"inputs"`
	Plugins []config.Plugin    `yaml:"plugins"`
	Managed config.ManagedMode `yaml:"managed"`
}

type legacyInput struct {
	Directory *legacyDirectory `yaml:"directory"`
	GitRepo   *legacyGit       `yaml:"git_repo"`
}

type legacyDirectory struct {
	Path string `yaml:"path"`
	Root string `yaml:"root"`
}

type legacyGit struct {
	URL          string `yaml:"url"`
	SubDirectory string `yaml:"sub_directory"`
	Root         string `yaml:"root"`
}

func (d *legacyDirectory) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		*d = legacyDirectory{Path: node.Value, Root: "."}
		return nil
	}
	type rawDirectory legacyDirectory
	var raw rawDirectory
	if err := node.Decode(&raw); err != nil {
		return fmt.Errorf("Decode: %w", err)
	}
	*d = legacyDirectory(raw)
	if d.Root == "" {
		d.Root = "."
	}
	return nil
}

func parseLegacy(raw []byte) (legacyConfig, []string, error) {
	var cfg legacyConfig
	warnings, err := decodeLegacy(raw, &cfg)
	if err != nil {
		return legacyConfig{}, nil, fmt.Errorf("decodeLegacy: %w", err)
	}
	if cfg.Version != "" {
		warnings = append(warnings, fmt.Sprintf("legacy version %q is compatibility metadata and is omitted from v1 files", cfg.Version))
	}
	if err := cfg.Generate.Managed.Validate(); err != nil {
		return legacyConfig{}, nil, fmt.Errorf("Validate: %w", err)
	}
	for i, input := range cfg.Generate.Inputs {
		if (input.Directory == nil) == (input.GitRepo == nil) {
			return legacyConfig{}, nil, fmt.Errorf("generate.inputs[%d] must select exactly one directory or git_repo", i)
		}
	}
	return cfg, warnings, nil
}

// decodeLegacy validates values that v0 actually consumed before custom
// UnmarshalYAML methods can coerce them. Unknown legacy keys are preserved only
// in the byte-identical backup: v0 ignored them as well, so migration reports a
// warning rather than inventing v1 semantics for them.
func decodeLegacy(raw []byte, out any) ([]string, error) {
	node, err := document(raw)
	if err != nil {
		return nil, fmt.Errorf("document: %w", err)
	}
	var warnings []string
	if err := typedNode(node, reflect.TypeOf(out).Elem(), "config", &warnings); err != nil {
		return nil, fmt.Errorf("typedNode: %w", err)
	}
	if err := node.Decode(out); err != nil {
		return nil, fmt.Errorf("Decode: %w", err)
	}
	return warnings, nil
}

func decodeStrict(raw []byte, out any) error {
	node, err := document(raw)
	if err != nil {
		return fmt.Errorf("document: %w", err)
	}
	var warnings []string
	if err := typedNode(node, reflect.TypeOf(out).Elem(), "config", &warnings); err != nil {
		return fmt.Errorf("typedNode: %w", err)
	}
	if len(warnings) > 0 {
		return fmt.Errorf("%s", warnings[0])
	}
	if err := node.Decode(out); err != nil {
		return fmt.Errorf("Decode: %w", err)
	}
	return nil
}

func document(raw []byte) (*yaml.Node, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var root yaml.Node
	if err := decoder.Decode(&root); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty YAML input")
		}
		return nil, fmt.Errorf("Decode: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("expected exactly one YAML document")
	}
	if len(root.Content) != 1 {
		return nil, fmt.Errorf("empty YAML input")
	}
	if root.Content[0].Tag == "!!null" {
		return nil, fmt.Errorf("null YAML value at line %d", root.Content[0].Line)
	}
	if err := unambiguousNode(root.Content[0]); err != nil {
		return nil, fmt.Errorf("unambiguousNode: %w", err)
	}
	if root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected a YAML mapping")
	}
	return root.Content[0], nil
}

func unambiguousNode(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		return fmt.Errorf("YAML aliases and anchors require manual migration (line %d)", node.Line)
	}
	switch node.Tag {
	case "!!map", "!!seq", "!!str", "!!int", "!!float", "!!bool", "!!null":
	default:
		return fmt.Errorf("unsupported YAML tag %q at line %d", node.Tag, node.Line)
	}
	if node.Kind == yaml.MappingNode {
		seen := make(map[string]bool)
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Value == "<<" {
				return fmt.Errorf("mapping keys must be strings; YAML merges are not supported (line %d)", key.Line)
			}
			if seen[key.Value] {
				return fmt.Errorf("duplicate YAML key %q at line %d", key.Value, key.Line)
			}
			seen[key.Value] = true
		}
	}
	for _, child := range node.Content {
		if err := unambiguousNode(child); err != nil {
			return err
		}
	}
	return nil
}

func typedNode(node *yaml.Node, typ reflect.Type, path string, warnings *[]string) error {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if node.Tag == "!!null" {
		if path == "config.deps" && typ.Kind() == reflect.Slice {
			*warnings = append(*warnings, "legacy deps: null is treated as an empty dependency list")
			return nil
		}
		return fmt.Errorf("null YAML value at line %d", node.Line)
	}
	if typ == reflect.TypeFor[legacyDirectory]() && node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		return nil
	}
	if typ == reflect.TypeFor[config.PluginOpts]() || typ == reflect.TypeFor[v1.PluginOptions]() {
		if typ == reflect.TypeFor[v1.PluginOptions]() && node.Kind == yaml.SequenceNode {
			for _, item := range node.Content {
				if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
					return fmt.Errorf("%s must contain option strings", path)
				}
			}
			return nil
		}
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("%s must be an options mapping", path)
		}
		for i := 1; i < len(node.Content); i += 2 {
			value := node.Content[i]
			if value.Kind == yaml.SequenceNode {
				for _, child := range value.Content {
					if child.Kind != yaml.ScalarNode {
						return fmt.Errorf("%s option values must be scalars", path)
					}
				}
			} else if value.Kind != yaml.ScalarNode {
				return fmt.Errorf("%s option values must be scalars or lists", path)
			}
		}
		return nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("%s must be a mapping", path)
		}
		fields := make(map[string]reflect.Type)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			key := strings.Split(field.Tag.Get("yaml"), ",")[0]
			if key != "-" && field.IsExported() {
				fields[key] = field.Type
			}
		}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			field, ok := fields[key]
			if !ok {
				*warnings = append(*warnings, fmt.Sprintf("legacy key %s.%s at line %d was ignored by v0 and is omitted from v1 files", path, key, node.Content[i].Line))
				continue
			}
			if err := typedNode(node.Content[i+1], field, path+"."+key, warnings); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if node.Kind != yaml.SequenceNode {
			return fmt.Errorf("%s must be a sequence", path)
		}
		for i, child := range node.Content {
			if err := typedNode(child, typ.Elem(), fmt.Sprintf("%s[%d]", path, i), warnings); err != nil {
				return err
			}
		}
	case reflect.Map:
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("%s must be a mapping", path)
		}
		for i := 1; i < len(node.Content); i += 2 {
			if err := typedNode(node.Content[i], typ.Elem(), path+"."+node.Content[i-1].Value, warnings); err != nil {
				return err
			}
		}
	case reflect.String:
		if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
			return fmt.Errorf("%s must be a string", path)
		}
	case reflect.Bool:
		if node.Tag != "!!bool" {
			return fmt.Errorf("%s must be a boolean; placeholders preventing analysis require manual migration", path)
		}
	case reflect.Int:
		if node.Tag != "!!int" {
			return fmt.Errorf("%s must be an integer", path)
		}
	case reflect.Interface:
		if node.Kind != yaml.ScalarNode {
			return fmt.Errorf("%s must be a scalar", path)
		}
	default:
		return fmt.Errorf("unsupported value at %s", path)
	}
	return nil
}
