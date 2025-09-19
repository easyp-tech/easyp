package console

import (
	"context"
	"io"
	"runtime"
)

// Console is provide to terminal command in console.
type Console interface {
	RunCmd(ctx context.Context, dir string, command string, commandParams ...string) (string, error)
	RunCmdWithStdin(ctx context.Context, dir string, stdin io.Reader, command string, commandParams ...string) (string, error)
}

// New create new console.
func New() Console {
	if runtime.GOOS == "windows" {
		return powershell{}
	}
	return bash{}
}
