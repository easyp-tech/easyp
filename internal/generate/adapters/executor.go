package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
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

// RunCmd shell command. Running git mod, for example.
// inspired by cmd/go/internal/modfetch/codehost/codehost.go:318:Run from go mod
// But go mod function looks too complicated
// so for PoC/MVP I decided to implement simpler solution
func RunCmd(ctx context.Context, dir string, command string, commandParams ...string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	fullCommand := append([]string{command}, commandParams...)
	cmd := exec.CommandContext(ctx, "bash", "-c", strings.Join(fullCommand, " "))
	cmd.Dir = dir
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
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
