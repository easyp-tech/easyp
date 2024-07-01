package lint

import (
	"fmt"
)

type LinterError struct {
	path string
	err  error
}

func (e *LinterError) String() string {
	return fmt.Sprintf("%s:%s", e.path, e.err.Error())
}

func NewLinterError(path string, err error) *LinterError {
	return &LinterError{
		path: path,
		err:  err,
	}
}
