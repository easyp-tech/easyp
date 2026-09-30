package api

import (
	"context"
	"fmt"
)

// terminalInitializationPrompter never opens a controlling terminal behind
// redirected CLI streams. Explicit identities still work without a terminal.
type terminalInitializationPrompter struct {
	terminal bool
	prompt   initializationPrompter
}

func (p terminalInitializationPrompter) Input(ctx context.Context, message, fallback string) (string, error) {
	if !p.terminal {
		return "", fmt.Errorf("module identity is required in noninteractive mode; pass --module")
	}
	return p.prompt.Input(ctx, message, fallback)
}

func (p terminalInitializationPrompter) Confirm(ctx context.Context, message string, fallback bool) (bool, error) {
	if !p.terminal {
		return false, fmt.Errorf("cannot confirm %q without terminal input and output; existing files were not changed", message)
	}
	return p.prompt.Confirm(ctx, message, fallback)
}
