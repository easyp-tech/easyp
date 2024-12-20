// Package lint provides the core functionality of easyp lint.
package lint

import (
	"log"
	"reflect"
	"strings"
	"unicode"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/api/factories"
	modulereflect "github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
)

// Lint is the core functionality of easyp lint.
type Lint struct {
	rules         []Rule
	ignore        []string
	deps          []string
	moduleReflect *modulereflect.ModuleReflect
	ignoreOnly    map[string][]string
}

// ImportPath type alias for path import in proto file
type ImportPath string

func ConvertImportPath(source string) ImportPath {
	return ImportPath(strings.Trim(source, "\""))
}

// ProtoInfo is the information of a proto file.
type ProtoInfo struct {
	Path                 string
	Info                 *unordered.Proto
	ProtoFilesFromImport map[ImportPath]*unordered.Proto
}

// GetPackageName return package name of current proto file.
// TODO: write unittest for it
func (p *ProtoInfo) GetPackageName() string {
	if len(p.Info.ProtoBody.Packages) == 0 {
		return ""
	}

	return p.Info.ProtoBody.Packages[0].Name
}

// Rule is an interface for a rule checking.
type Rule interface {
	// Message returns the message of the rule.
	Message() string
	// Validate validates the proto rule.
	Validate(ProtoInfo) ([]Issue, error)
}

// New creates a new Lint.
func New(rules []Rule, ignore []string, ignoreOnly map[string][]string, deps []string) *Lint {
	moduleReflect, err := factories.NewModuleReflect()
	if err != nil {
		log.Fatal(err) // TODO; return error
	}

	return &Lint{
		rules:         rules,
		ignore:        ignore,
		deps:          deps,
		moduleReflect: moduleReflect,
		ignoreOnly:    ignoreOnly,
	}
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
