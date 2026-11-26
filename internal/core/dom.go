package core

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/easyp-tech/easyp/internal/protoast"
)

// Position is a source position carried by issues.
type Position = protoast.Position

type (
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

	// ImportPath type alias for path import in proto file
	ImportPath string

	// PackageName type alias for package name `package` section in protofile.
	PackageName string
)

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
