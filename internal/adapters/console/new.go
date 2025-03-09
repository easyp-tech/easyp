package console

import (
	"runtime"

	"go.redsock.ru/protopack/internal/core"
)

// New create new console.
func New() core.Console {
	if runtime.GOOS == "windows" {
		return powershell{}
	}
	return bash{}
}
