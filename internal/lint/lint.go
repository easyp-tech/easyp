// Package lint provides the core functionality of easyp lint.
package lint

import (
	"log"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/api/factories"
	modulereflect "github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
)

// LintParams stores params for linters
type LintParams struct {
	Rules               []Rule
	IgnoreDirs          []string
	Deps                []string
	AllowCommentIgnores bool
}

// Lint is the core functionality of easyp lint.
type Lint struct {
	rules         []Rule
	ignoreDirs    []string
	deps          []string
	moduleReflect *modulereflect.ModuleReflect
}

var lintParams *LintParams

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

// Rule is an interface for a rule checking.
type Rule interface {
	// Validate validates the proto rule.
	Validate(ProtoInfo) []error
}

// New creates a new Lint.
func New(lp *LintParams) *Lint {
	lintParams = lp

	moduleReflect, err := factories.NewModuleReflect()
	if err != nil {
		log.Fatal(err) // TODO: return error
	}

	return &Lint{
		rules:         lp.Rules,
		ignoreDirs:    lp.IgnoreDirs,
		deps:          lp.Deps,
		moduleReflect: moduleReflect,
	}
}

func SetLintParams(p *LintParams) {
	lintParams = p
}

func GetLintParams() *LintParams {
	return lintParams
}
