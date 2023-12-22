// Package core provides the core functionality of easyp.
package core

import (
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

// Core is the core functionality of easyp.
type Core struct {
	rules []Rule
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

// New creates a new Core.
func New(rules []Rule) *Core {
	return &Core{
		rules: rules,
	}
}
