package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
)

var _ error = (*Error)(nil)

// Error is an error with meta information.
type Error struct {
	position   meta.Position
	sourceName string
	comments   []parser.Comment
	err        error
}

// Error implements error.
func (e Error) Error() string {
	return fmt.Errorf("%d:%d:%s: %w", e.position.Line, e.position.Column, e.sourceName, e.err).Error()
}

// Unwrap implements error.
func (e Error) Unwrap() error {
	return e.err
}

// BuildError creates an Error.
func BuildError(pos meta.Position, sourceName string, err error) error {
	return &Error{
		position:   pos,
		sourceName: sourceName,
		err:        err,
	}
}
