package core

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/samber/lo"
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

	// Console is provide to terminal command in console.
	Console interface {
		RunCmd(ctx context.Context, dir string, command string, commandParams ...string) (string, error)
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

	// ProtoInfo is the information of a proto file.
	ProtoInfo struct {
		Path                 string
		Info                 *unordered.Proto
		ProtoFilesFromImport map[ImportPath]*unordered.Proto
	}
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

func ConvertImportPath(source string) ImportPath {
	return ImportPath(strings.Trim(source, "\""))
}

type (
	// Plugin is a plugin for gRPC generator.
	Plugin struct {
		Name    string
		Out     string
		Options map[string]string
	}
	// Inputs is the source for generating code.
	Inputs struct {
		Dirs []string
	}
	// Config is the configuration for EasyP generate.
	Config struct {
		Deps    []string
		Plugins []Plugin
		Inputs  Inputs
	}
	// Query is a query for making sh command.
	Query struct {
		Compiler string
		Imports  []string
		Plugins  []Plugin
		Files    []string
	}
)

func (q Query) build() string {
	var buf bytes.Buffer

	buf.WriteString(q.Compiler)
	buf.WriteString(" \\\n")

	for _, imp := range q.Imports {
		buf.WriteString(" -I ")
		buf.WriteString(imp)
		buf.WriteString(" \\\n")
	}

	for _, plugin := range q.Plugins {
		buf.WriteString(" --")
		buf.WriteString(plugin.Name)
		buf.WriteString("_out=")
		buf.WriteString(plugin.Out)
		buf.WriteString(" \\\n")
		buf.WriteString(" --")
		buf.WriteString(plugin.Name)
		buf.WriteString("_opt=")

		options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
			return k + "=" + v
		})
		buf.WriteString(strings.Join(options, ","))
		buf.WriteString(" \\\n")
	}

	for i, file := range q.Files {
		buf.WriteString(" ")
		buf.WriteString(file)

		if i != len(q.Files)-1 {
			buf.WriteString(" \\\n")
		}
	}

	return buf.String()
}
