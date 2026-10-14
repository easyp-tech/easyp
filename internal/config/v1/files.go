package v1

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Standard v1 configuration filenames shared by readers and commands.
const (
	ModuleFile   = "protobuf.mod"
	LockFile     = "protobuf.lock"
	PolicyFile   = "easyp.yaml"
	GenerateFile = "easyp.gen.yaml"
)

// FileKind identifies which v1 configuration format a filename denotes.
type FileKind int

// File kinds recognized from canonical names and named profiles.
const (
	UnknownFileKind FileKind = iota
	PolicyFileKind
	GenerateFileKind
	ModuleFileKind
	LockFileKind
)

// String returns a human-readable name used in diagnostics.
func (k FileKind) String() string {
	switch k {
	case PolicyFileKind:
		return "policy config"
	case GenerateFileKind:
		return "generation config"
	case ModuleFileKind:
		return "module manifest"
	case LockFileKind:
		return "module lock"
	default:
		return "unknown file"
	}
}

// KindOf classifies a path by its base name. Besides the canonical names,
// named profiles use the suffixes ".easyp.yaml" (policy) and ".easyp.gen.yaml"
// (generation), for example private.easyp.gen.yaml.
func KindOf(path string) FileKind {
	name := filepath.Base(path)
	switch {
	case name == ModuleFile:
		return ModuleFileKind
	case name == LockFile:
		return LockFileKind
	case name == GenerateFile || strings.HasSuffix(name, "."+GenerateFile):
		return GenerateFileKind
	case name == PolicyFile || strings.HasSuffix(name, "."+PolicyFile):
		return PolicyFileKind
	default:
		return UnknownFileKind
	}
}

// RequireKind rejects a recognized configuration file of another format.
// Unrecognized names are accepted as the expected kind, so arbitrarily named
// files remain selectable explicitly.
func RequireKind(path string, want FileKind) error {
	got := KindOf(path)
	if got == UnknownFileKind || got == want {
		return nil
	}
	return fmt.Errorf("%s is a %s; this command expects a %s", path, got, want)
}
