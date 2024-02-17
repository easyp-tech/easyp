package models

import (
	"strings"
)

// Package contain package name and its version
type Package struct {
	Name    string
	Version string
}

// NewPackage create Package struct from raw module string:
func NewPackage(module string) Package {
	parts := strings.Split(module, "@")
	name := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}
	return Package{
		Name:    name,
		Version: version,
	}
}
