package models

import (
	"errors"
	"strings"
)

const (
	// If version was omitted
	Omitted RequestedVersion = ""

	// generated version prefix
	generatedVersionPrefix = "v0.0.0"
	// generated version looks like: v0.0.0-20240222234643-814bf88cf225
	// prefix + datetime + commit
	generatedVersionPartsCount = 3
	generatedVersionSep        = "-"
)

type (
	// ModuleHash alias for module's hash
	// used in lock file for verification
	ModuleHash string

	// RequestedVersion for installing
	RequestedVersion string
)

var (
	ErrRequestedVersionNotGenerated = errors.New("requested version is not generated")
)

// Module contain requested dependency name and its version
type Module struct {
	Name    string           // Full path on remote repository
	Version RequestedVersion // Version obtained from config (Omitted if version was omitted)
}

type GeneratedVersionParts struct {
	Datetime   string
	CommitHash string
}

// NewModule create Module struct from raw dependency string: remote@version
// dependency string format: origin@version: github.com/company/repository@v1.2.3
func NewModule(dependency string) Module {
	parts := strings.Split(dependency, "@")
	name := parts[0]
	version := Omitted // by default set version as Omitted
	if len(parts) > 1 {
		version = RequestedVersion(parts[1])
	}
	return Module{
		Name:    name,
		Version: version,
	}
}

// GetParts return parts of GeneratedVersion
// if RequestedVersion is not generated return error
func (v RequestedVersion) GetParts() (GeneratedVersionParts, error) {
	parts := strings.Split(string(v), generatedVersionSep)
	if len(parts) != generatedVersionPartsCount {
		return GeneratedVersionParts{}, ErrRequestedVersionNotGenerated
	}

	if parts[0] != generatedVersionPrefix {
		return GeneratedVersionParts{}, ErrRequestedVersionNotGenerated
	}

	ver := GeneratedVersionParts{
		Datetime:   parts[1],
		CommitHash: parts[2],
	}
	return ver, nil
}

// IsGenerated check if requested was generated and it's not a commit's tag
// like v0.0.0-20240222234643-814bf88cf225 in go mod
func (v RequestedVersion) IsGenerated() bool {
	_, err := v.GetParts()
	if err != nil {
		return false
	}

	return true
}

// IsOmitted check if requested version is omitted
func (v RequestedVersion) IsOmitted() bool {
	return v == Omitted
}

func (g GeneratedVersionParts) GetVersionString() string {
	return generatedVersionPrefix + generatedVersionSep + g.Datetime + generatedVersionSep + g.CommitHash
}
