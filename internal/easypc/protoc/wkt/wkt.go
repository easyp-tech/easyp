// Package wkt provides embedded Well-Known Type proto definitions for standard imports.
package wkt

import (
	"embed"
	"io/fs"
)

//go:embed google/protobuf/*.proto google/protobuf/compiler/*.proto
var FS embed.FS

// ReadFile reads an embedded WKT file by canonical path (e.g. "google/protobuf/descriptor.proto").
func ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(FS, path)
}
