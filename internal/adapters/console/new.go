package console

import (
	"github.com/easyp-tech/easyp/internal/core"
)

// New create new console.
func New() core.Console {
	return bash{}
}
