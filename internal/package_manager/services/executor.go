package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type RunError struct {
	Command       string
	CommandParams []string
	Dir           string
	Err           error
	Stderr        string
}

func (e RunError) Error() string {
	// TODO: extend error output
	return fmt.Sprintf("Command: %s; Err: %v; Stderr: %s", e.Command, e.Err, e.Stderr)
}

// RunCmd shell command. Running git package_manager, for example.
// inpsired by cmd/go/internal/modfetch/codehost/codehost.go:318:Run from go package_manager
// But go package_manager function looks too complicated
// so for PoC/MVP I decided to implement simpler solution
func RunCmd(ctx context.Context, dir string, command string, commandParams ...string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, command, commandParams...)

	cmd.Dir = dir
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", &RunError{
			Command:       command,
			CommandParams: commandParams,
			Dir:           dir,
			Err:           err,
			Stderr:        stderr.String(),
		}
	}

	return stdout.String(), nil
}
