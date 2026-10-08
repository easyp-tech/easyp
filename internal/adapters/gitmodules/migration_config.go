package gitmodules

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func readMigrationLegacyRoots(checkout string, files []string) ([]string, error) {
	for _, name := range files {
		if path.Base(name) != v1.ModuleFile {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("ReadFile: %w", err)
		}
		if v1.IsModuleManifest(raw) {
			return nil, fmt.Errorf("cannot reproduce legacy roots with native protobuf.mod %q", name)
		}
	}
	roots, err := readMigrationArchiveRoots(checkout, files)
	if err != nil {
		return nil, fmt.Errorf("readMigrationArchiveRoots: %w", err)
	}
	return roots, nil
}

// Native protobuf.mod owns the v1 namespace, but released v0 installers read
// these producer configs independently when renaming their archive entries.
func readMigrationArchiveRoots(checkout string, files []string) ([]string, error) {
	tracked := make(map[string]bool, len(files))
	for _, name := range files {
		tracked[name] = true
	}
	// Legacy ReadFromRepo preferred buf.work.yaml, then buf.yaml, then easyp.yaml.
	// Malformed higher-priority files are refused rather than silently falling
	// back to roots whose provenance can no longer be established reliably.
	for _, name := range []string{"buf.work.yaml", "buf.yaml", "easyp.yaml"} {
		if !tracked[name] {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(checkout, name))
		if err != nil {
			return nil, fmt.Errorf("ReadFile: %w", err)
		}
		roots, err := parseMigrationLegacyRoots(name, raw)
		if err != nil {
			return nil, fmt.Errorf("parseMigrationLegacyRoots: %s: %w", name, err)
		}
		for _, root := range roots {
			if root == "" || root == "." {
				continue
			}
			if !filepath.IsLocal(filepath.FromSlash(root)) || path.Clean(root) != root || strings.Contains(root, "\\") {
				return nil, fmt.Errorf("unsupported legacy root %q in %s", root, name)
			}
		}
		return roots, nil
	}
	return nil, nil
}

func parseMigrationLegacyRoots(name string, raw []byte) ([]string, error) {
	// ParseConfig historically expanded environment variables before parsing.
	// The environment that produced an existing lock cannot be reconstructed.
	if bytes.ContainsRune(raw, '$') {
		return nil, fmt.Errorf("legacy config contains environment placeholders")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document yaml.Node
	err := decoder.Decode(&document)
	if err != nil {
		return nil, fmt.Errorf("Decode: %w", err)
	}
	var extra yaml.Node
	err = decoder.Decode(&extra)
	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("legacy config must contain exactly one YAML document")
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
		return nil, fmt.Errorf("legacy config must contain a mapping")
	}
	if err := validateMigrationYAML(&document); err != nil {
		return nil, fmt.Errorf("validateMigrationYAML: %w", err)
	}
	root := document.Content[0]
	switch name {
	case "buf.work.yaml":
		fields, err := migrationMapping(root, "version", "directories")
		if err != nil {
			return nil, fmt.Errorf("migrationMapping: %w", err)
		}
		if fields["directories"] == nil {
			return nil, fmt.Errorf("legacy Buf workspace is missing directories")
		}
		return migrationStringList(fields["directories"])
	case "buf.yaml":
		return migrationBufRoots(root)
	case "easyp.yaml":
		return migrationEasyPRoots(root)
	default:
		return nil, fmt.Errorf("unsupported legacy config %q", name)
	}
}

func migrationBufRoots(root *yaml.Node) ([]string, error) {
	fields, err := migrationMapping(root, "version", "modules", "name", "deps", "lint", "breaking", "build")
	if err != nil {
		return nil, fmt.Errorf("migrationMapping: %w", err)
	}
	version, err := migrationString(fields["version"])
	if err != nil {
		return nil, fmt.Errorf("migrationString: %w", err)
	}
	modules := fields["modules"]
	if modules == nil {
		// The old reader used only modules[].path, including for older Buf
		// versions. In particular it did not strip v1beta1 build.roots.
		if version == "v1" || version == "v1beta1" {
			return nil, nil
		}
		return nil, fmt.Errorf("legacy Buf config is missing modules")
	}
	if version != "v2" || modules.Kind != yaml.SequenceNode || len(modules.Content) == 0 {
		return nil, fmt.Errorf("legacy Buf modules must be a nonempty v2 list")
	}
	roots := make([]string, 0, len(modules.Content))
	for _, module := range modules.Content {
		fields, err := migrationMapping(module, "path", "name", "lint", "breaking")
		if err != nil {
			return nil, fmt.Errorf("migrationMapping: %w", err)
		}
		root, err := migrationString(fields["path"])
		if err != nil {
			return nil, fmt.Errorf("migrationString: %w", err)
		}
		if root == "" {
			return nil, fmt.Errorf("legacy Buf module path is empty")
		}
		roots = append(roots, root)
	}
	return roots, nil
}

func migrationEasyPRoots(root *yaml.Node) ([]string, error) {
	fields, err := migrationMapping(root, "deps", "generate", "lint", "breaking")
	if err != nil {
		return nil, fmt.Errorf("migrationMapping: %w", err)
	}
	if fields["generate"] == nil {
		return nil, nil
	}
	generate, err := migrationMapping(fields["generate"], "inputs", "plugins", "managed")
	if err != nil {
		return nil, fmt.Errorf("migrationMapping: %w", err)
	}
	inputs := generate["inputs"]
	if inputs == nil {
		return nil, nil
	}
	if inputs.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("legacy generate.inputs must be a list")
	}
	roots := make([]string, 0, len(inputs.Content))
	for _, input := range inputs.Content {
		fields, err := migrationMapping(input, "directory", "git_repo")
		if err != nil {
			return nil, fmt.Errorf("migrationMapping: %w", err)
		}
		if len(fields) != 1 {
			return nil, fmt.Errorf("legacy input must have exactly one source")
		}
		if gitRepo := fields["git_repo"]; gitRepo != nil {
			if _, err := migrationMapping(gitRepo, "url", "sub_directory", "root"); err != nil {
				return nil, fmt.Errorf("migrationMapping: %w", err)
			}
			// The old readEasyp appended the zero directory root for Git inputs.
			roots = append(roots, "")
			continue
		}
		directory := fields["directory"]
		if directory.Kind == yaml.ScalarNode {
			if _, err := migrationString(directory); err != nil {
				return nil, fmt.Errorf("migrationString: %w", err)
			}
			roots = append(roots, ".")
			continue
		}
		fields, err = migrationMapping(directory, "path", "root")
		if err != nil {
			return nil, fmt.Errorf("migrationMapping: %w", err)
		}
		if field := fields["path"]; field != nil {
			if _, err := migrationString(field); err != nil {
				return nil, fmt.Errorf("migrationString: %w", err)
			}
		}
		root := "."
		if field := fields["root"]; field != nil {
			root, err = migrationString(field)
			if err != nil {
				return nil, fmt.Errorf("migrationString: %w", err)
			}
			if root == "" {
				root = "."
			}
		}
		roots = append(roots, root)
	}
	return roots, nil
}

func validateMigrationYAML(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode || node.Anchor != "" || node.Tag == "!!null" || node.Tag == "!!merge" {
		return fmt.Errorf("legacy config contains an alias, anchor, merge, or null value")
	}
	if node.Kind == yaml.MappingNode {
		seen := make(map[string]bool)
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || seen[key.Value] {
				return fmt.Errorf("legacy config contains a duplicate or non-string key %q", key.Value)
			}
			seen[key.Value] = true
		}
	}
	for _, child := range node.Content {
		if err := validateMigrationYAML(child); err != nil {
			return fmt.Errorf("validateMigrationYAML: %w", err)
		}
	}
	return nil
}

func migrationMapping(node *yaml.Node, allowed ...string) (map[string]*yaml.Node, error) {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("legacy config field must be a mapping")
	}
	fields := make(map[string]*yaml.Node, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if !slices.Contains(allowed, key) {
			return nil, fmt.Errorf("unsupported legacy config field %q", key)
		}
		fields[key] = node.Content[i+1]
	}
	return fields, nil
}

func migrationString(node *yaml.Node) (string, error) {
	if node == nil || node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return "", fmt.Errorf("legacy config field must be a string")
	}
	return node.Value, nil
}

func migrationStringList(node *yaml.Node) ([]string, error) {
	if node.Kind != yaml.SequenceNode || len(node.Content) == 0 {
		return nil, fmt.Errorf("legacy roots must be a nonempty list")
	}
	roots := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		root, err := migrationString(item)
		if err != nil {
			return nil, fmt.Errorf("migrationString: %w", err)
		}
		if root == "" {
			return nil, fmt.Errorf("legacy root is empty")
		}
		roots = append(roots, root)
	}
	return roots, nil
}
