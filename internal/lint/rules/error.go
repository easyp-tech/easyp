package rules

import (
	"encoding/json"
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
)

var _ error = (*Error)(nil)

// Error is an error with meta information.
type Error struct {
	Path       string        `json:"path"`
	Position   meta.Position `json:"position"`
	SourceName string        `json:"source_name"`
	Err        error         `json:"err"`
}

// Error implements error.
func (e Error) Error() string {
	return fmt.Errorf("%s:%d:%d:%w", e.Path, e.Position.Line, e.Position.Column, e.Err).Error()
}

// Unwrap implements error.
func (e Error) Unwrap() error {
	return e.Err
}

// MarshalJSON implements json.Marshaler.
func (e Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"path":        e.Path,
		"line":        e.Position.Line,
		"column":      e.Position.Column,
		"source_name": e.SourceName,
		"err":         e.Err.Error(),
	})
}

// BuildError creates an Error.
func BuildError(path string, pos meta.Position, sourceName string, err error) error {
	return &Error{
		Path:       path,
		Position:   pos,
		SourceName: sourceName,
		Err:        err,
	}
}
