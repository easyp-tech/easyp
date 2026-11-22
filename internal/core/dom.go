package core

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
)

type (
	// Rule is an interface for a rule checking.
	Rule interface {
		// Message returns the message of the rule.
		Message() string
		// Validate validates the proto rule.
		Validate(ProtoInfo) ([]Issue, error)
	}

	// IssueInfo contains the information of an issue and the path.
	IssueInfo struct {
		Issue
		Path string
	}

	// Position represents a source location within a proto file.
	Position struct {
		// Filename is the name of the file, if any.
		Filename string
		// Offset is the byte offset, starting at 0.
		Offset int
		// Line is the 1-based line number.
		Line int
		// Column is the 1-based column number.
		Column int
	}

	// Issue contains the information of an issue.
	Issue struct {
		Position   Position
		SourceName string
		Message    string
		RuleName   string
	}

	// ImportPath type alias for path import in proto file
	ImportPath string

	// PackageName type alias for package name `package` section in protofile.
	PackageName string

	// ProtoInfo is the information of a proto file.
	ProtoInfo struct {
		directiveComments []directiveComment
		Path              string
		// ImportPath is relative to a declared source root, independently of the scan root.
		ImportPath           string
		Info                 *unordered.Proto
		ProtoFilesFromImport map[ImportPath]*unordered.Proto
	}
)

// String formats the position as filename:line:column or line:column.
func (p Position) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func GetPackageName(protoFile *unordered.Proto) PackageName {
	if len(protoFile.ProtoBody.Packages) == 0 {
		return ""
	}

	return PackageName(protoFile.ProtoBody.Packages[0].Name)
}

// AppendIssue records a rule finding. The lint engine applies per-file suppressions.
func AppendIssue(
	issues []Issue, lintRule Rule, pos meta.Position, sourceName string, comments []*parser.Comment,
) []Issue {
	return append(issues, buildError(lintRule, Position(pos), sourceName))
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
