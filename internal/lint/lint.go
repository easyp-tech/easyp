// Package lint provides the core functionality of easyp lint.
package lint

import (
	"log"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/api/factories"
	modulereflect "github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
)

// Lint is the core functionality of easyp lint.
type Lint struct {
	rules         []Rule
	ignoreDirs    []string
	deps          []string
	moduleReflect *modulereflect.ModuleReflect
}

// ImportPath type alias for path import in proto file
type ImportPath string

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
func New(rules []Rule, ignoreDirs []string, deps []string) *Lint {
	moduleReflect, err := factories.NewModuleReflect()
	if err != nil {
		log.Fatal(err) // TODO; return error
	}

	return &Lint{
		rules:         rules,
		ignoreDirs:    ignoreDirs,
		deps:          deps,
		moduleReflect: moduleReflect,
	}
}
