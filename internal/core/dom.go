package core

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/easyp-tech/easyp/internal/protoast"
)

// Position is a location in a proto file: 1-based line, 1-based rune column
// and 0-based byte offset. Its JSON form matches earlier releases.
type Position = protoast.Position

type (
	// Rule is an interface for a rule checking.
	Rule interface {
		// Message returns the message of the rule.
		Message() string
		// Validate validates the proto rule.
		Validate(ProtoInfo) ([]Issue, error)
	}

	// ModuleRule needs every file before it can report, such as rules that
	// compare files of one package or directory. Lint passes each file to
	// Collect and asks for findings once, after the last file; it does not
	// call Validate. Suppressions and ignore_only apply to the findings by path.
	ModuleRule interface {
		Rule
		Collect(ProtoInfo) error
		Finish() []IssueInfo
	}

	// IssueInfo contains the information of an issue and the path.
	IssueInfo struct {
		Issue
		Path string
	}

	// Issue contains the information of an issue.
	Issue struct {
		Position   Position
		SourceName string
		Message    string
		RuleName   string
	}

	// Resolution is what linking adds to a file.
	Resolution struct {
		// UsedImports holds the direct imports that define, or publicly
		// re-export a file that defines, a name this file refers to.
		UsedImports map[ImportPath]bool
		// RPCTypes maps an rpc's byte offset to its request and response
		// types, fully qualified without a leading dot.
		RPCTypes map[int][2]string
	}

	// ImportPath type alias for path import in proto file
	ImportPath string

	// PackageName type alias for package name `package` section in protofile.
	PackageName string

	// ProtoInfo is the information of a proto file.
	ProtoInfo struct {
		Path string
		// ImportPath is relative to a declared source root, independently of the scan root.
		ImportPath string
		// File is the lint view of the file; ImportedFiles holds its direct imports.
		File          *protoast.File
		ImportedFiles map[ImportPath]*protoast.File
		// Resolved holds name resolution across files; nil when the files
		// could not be linked, in which case rules fall back to the syntax.
		Resolved *Resolution
	}
)

// AppendIssue records a rule finding. The lint engine applies per-file suppressions.
func AppendIssue(issues []Issue, lintRule Rule, pos Position, sourceName string) []Issue {
	return append(issues, buildError(lintRule, pos, sourceName))
}

// GetRuleName returns rule name
func GetRuleName(rule Rule) string {
	return toUpperSnakeCase(reflect.TypeOf(rule).Elem().Name())
}

// toUpperSnakeCase converts a string from PascalCase or camelCase to UPPER_SNEAK_CASE.
func toUpperSnakeCase(s string) string {
	var result []rune

	for i, r := range s {
		if unicode.IsUpper(r) {
			// Добавляем подчеркивание, когда:
			// 1. Не первый символ.
			// 2. Предыдущий символ не был заглавной буквой, либо следующий является прописной буквой.
			if i > 0 && (unicode.IsLower(rune(s[i-1])) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1])))) {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToUpper(r))
	}

	return string(result)
}

// buildError creates an Issue.
func buildError(lintRule Rule, pos Position, sourceName string) Issue {
	return Issue{
		Position:   pos,
		SourceName: sourceName,
		Message:    lintRule.Message(),
		RuleName:   GetRuleName(lintRule),
	}
}

type OpenImportFileError struct {
	FileName string
}

func (e *OpenImportFileError) Error() string {
	return fmt.Sprintf("open import file `%s`", e.FileName)
}

type GitRefNotFoundError struct {
	GitRef string
}

func (e *GitRefNotFoundError) Error() string {
	return fmt.Sprintf("git ref `%s` not found", e.GitRef)
}

func ConvertImportPath(source string) ImportPath {
	return ImportPath(strings.Trim(source, "\""))
}

type (
	// PluginSource is the source of the plugin.
	PluginSource struct {
		Name    string
		Remote  string
		Path    string
		Command []string
	}
	// Plugin is a plugin for gRPC generator.
	Plugin struct {
		Source      PluginSource
		Out         string
		Options     map[string][]string
		WithImports bool
	}
	// InputFilesDir is the configuration of the directory with additional functionality.
	InputFilesDir struct {
		Path string
		Root string
	}
	// Inputs is the source for generating code.
	Inputs struct {
		InputFilesDir []InputFilesDir
	}
)
