package rules

import (
	"fmt"
)

var _ error = (*Error)(nil)

// Error is an error with meta information.
type Error struct {
	Err error
}

// Error implements error.
func (e Error) Error() string {
	return fmt.Errorf("%w", e.Err).Error()
}

// Unwrap implements error.
func (e Error) Unwrap() error {
	return e.Err
}
