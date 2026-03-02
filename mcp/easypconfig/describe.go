package easypconfig

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/easyp-tech/easyp/internal/rules"
)

const (
	ToolName      = "easyp_config_describe"
	SchemaVersion = "easyp-config-v1"
)

type DescribeInput struct {
	Path            string `json:"path,omitempty"`
	IncludeSchema   *bool  `json:"include_schema,omitempty"`
	IncludeFields   *bool  `json:"include_fields,omitempty"`
	IncludeExamples *bool  `json:"include_examples,omitempty"`
	IncludeChildren *bool  `json:"include_children,omitempty"`
	ExamplesLimit   *int   `json:"examples_limit,omitempty"`
}

type FieldDoc struct {
	Path          string   `json:"path"`
	Type          string   `json:"type"`
	Required      bool     `json:"required"`
	Description   string   `json:"description"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	DefaultValue  string   `json:"default_value,omitempty"`
	Examples      []string `json:"examples,omitempty"`
	Notes         []string `json:"notes,omitempty"`
}

type Example struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	YAML        string   `json:"yaml"`
	Paths       []string `json:"paths,omitempty"`
}

type DescribeOutput struct {
	SchemaVersion string         `json:"schema_version"`
	SelectedPath  string         `json:"selected_path"`
	Schema        map[string]any `json:"schema,omitempty"`
	Fields        []FieldDoc     `json:"fields,omitempty"`
	Examples      []Example      `json:"examples,omitempty"`
	Notes         []string       `json:"notes,omitempty"`
}

type nodeDoc struct {
	Fields   []FieldDoc
	Examples []Example
	Notes    []string
}

type spec struct {
	SchemaVersion string
	SchemaByPath  map[string]map[string]any
	DocsByPath    map[string]nodeDoc
}

var (
	specOnce sync.Once
	specData spec

	arrayIndexPattern = regexp.MustCompile(`\[\d+\]`)
)

func Describe(input DescribeInput) (DescribeOutput, error) {
	s := getSpec()
	return s.describe(input)
}

func getSpec() spec {
	specOnce.Do(func() {
		specData = newSpec()
	})
	return specData
}

func (s spec) describe(input DescribeInput) (DescribeOutput, error) {
	selectedPath, ok := s.resolvePath(input.Path)
	if !ok {
		return DescribeOutput{}, fmt.Errorf("unknown path %q", input.Path)
	}

	includeSchema := boolOrDefault(input.IncludeSchema, true)
	includeFields := boolOrDefault(input.IncludeFields, true)
	includeExamples := boolOrDefault(input.IncludeExamples, true)
	includeChildren := boolOrDefault(input.IncludeChildren, true)
	examplesLimit := intOrDefault(input.ExamplesLimit, 10)
	if examplesLimit < 1 {
		examplesLimit = 1
	}
	if examplesLimit > 50 {
		examplesLimit = 50
	}

	paths := s.pathsFor(selectedPath, includeChildren)

	out := DescribeOutput{
		SchemaVersion: s.SchemaVersion,
		SelectedPath:  selectedPath,
	}

	if includeSchema {
		out.Schema = cloneSchemaMap(s.SchemaByPath[selectedPath])
	}
	if includeFields {
		out.Fields = s.collectFields(paths)
	}
	if includeExamples {
		out.Examples = s.collectExamples(paths, examplesLimit)
	}
	out.Notes = s.collectNotes(paths)

	return out, nil
}

func (s spec) resolvePath(rawPath string) (string, bool) {
	path := normalizePath(rawPath)
	if s.hasPath(path) {
		return path, true
	}

	normPath := removeArrayMarkers(path)
	for _, candidate := range s.allPaths() {
		if removeArrayMarkers(candidate) == normPath {
			return candidate, true
		}
	}

	return "", false
}

func (s spec) pathsFor(selectedPath string, includeChildren bool) []string {
	if !includeChildren {
		return []string{selectedPath}
	}

	allPaths := s.allPaths()
	paths := make([]string, 0, len(allPaths))
	for _, p := range allPaths {
		if isPathWithin(selectedPath, p) {
			paths = append(paths, p)
		}
	}
	return paths
}

func (s spec) collectFields(paths []string) []FieldDoc {
	seen := make(map[string]struct{})
	out := make([]FieldDoc, 0)
	for _, p := range paths {
		doc, ok := s.DocsByPath[p]
		if !ok {
			continue
		}
		for _, f := range doc.Fields {
			if _, exists := seen[f.Path]; exists {
				continue
			}
			seen[f.Path] = struct{}{}
			out = append(out, cloneFieldDoc(f))
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

func (s spec) collectExamples(paths []string, limit int) []Example {
	out := make([]Example, 0, limit)
	seen := make(map[string]struct{})
	for _, p := range paths {
		doc, ok := s.DocsByPath[p]
		if !ok {
			continue
		}
		for _, ex := range doc.Examples {
			key := exampleKey(ex)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, cloneExample(ex))
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}

func (s spec) collectNotes(paths []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, p := range paths {
		doc, ok := s.DocsByPath[p]
		if !ok {
			continue
		}
		for _, note := range doc.Notes {
			if _, exists := seen[note]; exists {
				continue
			}
			seen[note] = struct{}{}
			out = append(out, note)
		}
	}
	return out
}

func (s spec) hasPath(path string) bool {
	if _, ok := s.SchemaByPath[path]; ok {
		return true
	}
	if _, ok := s.DocsByPath[path]; ok {
		return true
	}
	return false
}

func (s spec) allPaths() []string {
	seen := make(map[string]struct{})
	paths := make([]string, 0, len(s.SchemaByPath)+len(s.DocsByPath))

	for p := range s.SchemaByPath {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}
	for p := range s.DocsByPath {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	sort.Strings(paths)
	return paths
}

func boolOrDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func intOrDefault(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}

func lintUseAllowedValues() []string {
	return append([]string(nil), rules.AllLintUseValues()...)
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" || strings.EqualFold(path, "root") {
		return "$"
	}

	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, ".")
	path = strings.ReplaceAll(path, "[*]", "[]")
	path = arrayIndexPattern.ReplaceAllString(path, "[]")
	path = strings.TrimSuffix(path, ".")

	return path
}

func removeArrayMarkers(path string) string {
	return strings.ReplaceAll(path, "[]", "")
}

func isPathWithin(base, candidate string) bool {
	if base == "$" {
		return true
	}
	if base == candidate {
		return true
	}
	if strings.HasPrefix(candidate, base+".") {
		return true
	}
	if strings.HasPrefix(candidate, base+"[].") {
		return true
	}
	if candidate == base+"[]" {
		return true
	}
	return false
}

func cloneFieldDoc(in FieldDoc) FieldDoc {
	out := in
	out.AllowedValues = append([]string(nil), in.AllowedValues...)
	out.Examples = append([]string(nil), in.Examples...)
	out.Notes = append([]string(nil), in.Notes...)
	return out
}

func cloneExample(in Example) Example {
	out := in
	out.Paths = append([]string(nil), in.Paths...)
	return out
}

func exampleKey(ex Example) string {
	return strings.Join([]string{
		ex.Title,
		ex.Description,
		ex.YAML,
		strings.Join(ex.Paths, "\x1f"),
	}, "\x1e")
}

func newSpec() spec {
	return spec{
		SchemaVersion: SchemaVersion,
		SchemaByPath:  SchemaByPath(),
		DocsByPath:    docsByPath(),
	}
}
