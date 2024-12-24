package console

import (
	"fmt"
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
