package core

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/protoast"
)

// Position is a source position carried by issues. It is an alias of
// protoast.Position so lint findings already speak the type the future engine
// emits; go-protoparser values are converted at the AppendIssue boundary via
// fromMetaPos.
type Position = protoast.Position

type (
	// CurrentProjectGitWalker is provider for fs walking for current project
	CurrentProjectGitWalker interface {
		GetDirWalker(workingDir, gitRef, path string) (DirWalker, error)
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

	// ImportPath type alias for path import in proto file
	ImportPath string

	// PackageName type alias for package name `package` section in protofile.
	PackageName string

	Import struct {
		ProtoFilePath string
		PackageName   PackageName
		*parser.Import
	}

	Service struct {
		ProtoFilePath string
		PackageName   PackageName
		*unordered.Service
	}

	Message struct {
		MessagePath   string
		ProtoFilePath string
		PackageName   PackageName
		*unordered.Message
	}

	OneOf struct {
		OneOfPath     string
		ProtoFilePath string
		PackageName   PackageName
		*parser.Oneof
	}

	Enum struct {
		EnumPath      string
		ProtoFilePath string
		PackageName   PackageName
		*unordered.Enum
	}

	Collection struct {
		Imports  map[ImportPath]Import
		Services map[string]Service
		// key message path - for supporting nested messages:
		// message MainMessage {
		// 		message NestedMessage{};
		// };
		// will be: MainMessage.NestedMessage
		Messages map[string]Message
		OneOfs   map[string]OneOf
		Enums    map[string]Enum
	}

	// collects proto data collections
	// packageName -> services,messages etc
	ProtoData map[PackageName]*Collection
)

func GetPackageName(protoFile *unordered.Proto) PackageName {
	if len(protoFile.ProtoBody.Packages) == 0 {
		return ""
	}

	return PackageName(protoFile.ProtoBody.Packages[0].Name)
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

// fromMetaPos converts a go-protoparser source position to the issue position
// type. The field sets are identical in name, type, and order, so the
// conversion copies every field and changes no value.
func fromMetaPos(p meta.Position) Position {
	return Position{
		Filename: p.Filename,
		Offset:   p.Offset,
		Line:     p.Line,
		Column:   p.Column,
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
