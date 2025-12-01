package core

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/core/models"
)

type (
	// Rule is an interface for a rule checking.
	Rule interface {
		// Message returns the message of the rule.
		Message() string
		// Validate validates the proto rule.
		Validate(ProtoInfo) ([]Issue, error)
	}

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
		Position   meta.Position
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
		Path                 string
		Info                 *unordered.Proto
		ProtoFilesFromImport map[ImportPath]*unordered.Proto
	}

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

type Repo interface {
	// GetFiles returns list of all files in repository
	GetFiles(ctx context.Context, revision models.Revision, dirs ...string) ([]string, error)

	// ReadFile returns file's content from repository
	ReadFile(ctx context.Context, revision models.Revision, fileName string) (string, error)

	// Archive passed storage to archive and return full path to archive
	Archive(
		ctx context.Context, revision models.Revision, cacheDownloadPaths models.CacheDownloadPaths,
	) error

	// ReadRevision reads commit's revision by passed version
	// or return the latest commit if version is empty
	ReadRevision(ctx context.Context, requestedVersion models.RequestedVersion) (models.Revision, error)

	// Fetch from remote repository specified version
	Fetch(ctx context.Context, revision models.Revision) error
}

func GetPackageName(protoFile *unordered.Proto) PackageName {
	if len(protoFile.ProtoBody.Packages) == 0 {
		return ""
	}

	return PackageName(protoFile.ProtoBody.Packages[0].Name)
}

// AppendIssue check if lint error is ignored -> add new error to slice
// otherwise ignore appending
func AppendIssue(
	issues []Issue, lintRule Rule, pos meta.Position, sourceName string, comments []*parser.Comment,
) []Issue {
	if CheckIsIgnored(comments, GetRuleName(lintRule)) {
		return issues
	}

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
func buildError(lintRule Rule, pos meta.Position, sourceName string) Issue {
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
		Options     map[string]string
		WithImports bool
	}
	// InputGitRepo is the configuration of the git repository.
	InputGitRepo struct {
		URL          string
		SubDirectory string
		Out          string
		Root         string
	}
	// InputFilesDir is the configuration of the directory with additional functionality.
	InputFilesDir struct {
		Path string
		Root string
	}
	// Inputs is the source for generating code.
	Inputs struct {
		InputFilesDir []InputFilesDir
		InputGitRepos []InputGitRepo
	}
	// Config is the configuration for EasyP generate.
	Config struct {
		Deps    []string
		Plugins []Plugin
		Inputs  Inputs
	}
	// Query is a query for making sh command.
	Query struct {
		Imports []string
		Plugins []Plugin
		Files   []string
	}
)
