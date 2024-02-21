package models

import (
	"strings"
)

// Module contain dependency name and its version
type Module struct {
	Name    string
	Version string
}

// NewModule create Module struct from raw dependency string: remote@version
func NewModule(dependency string) Module {
	parts := strings.Split(dependency, "@")
	name := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}
	return Module{
		Name:    name,
		Version: version,
	}
}
