package dependency

import (
	"strings"
)

// Dependency contains dependency name and its version
type Dependency struct {
	Name    string
	Version string
}

func ParseDependency(module string) Dependency {
	parts := strings.Split(module, "@")
	name := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}
	return Dependency{
		Name:    name,
		Version: version,
	}
}
