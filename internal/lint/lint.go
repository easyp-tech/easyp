// Package lint provides the core functionality of easyp lint.
package lint

import (
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

// Lint is the core functionality of easyp lint.
type Lint struct {
	rules      []Rule
	ignoreDirs []string
}

// ProtoInfo is the information of a proto file.
type ProtoInfo struct {
	Path string
	Info *unordered.Proto
}

// Rule is an interface for a rule checking.
type Rule interface {
	// Validate validates the proto rule.
	Validate(ProtoInfo) []error
}

// New creates a new Lint.
func New(rules []Rule, ignoreDirs []string) *Lint {
	return &Lint{
		rules:      rules,
		ignoreDirs: ignoreDirs,
	}
}
