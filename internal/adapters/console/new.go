package console

import (
	"runtime"

	"github.com/easyp-tech/easyp/internal/core"
)

// New create new console.
func New() core.Console {
	if runtime.GOOS == "windows" {
		return powershell{}
	}
	return bash{}
}
