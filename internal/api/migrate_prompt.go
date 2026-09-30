package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/term"
)

// migrationPrompt keeps normal terminal line editing and needs no raw mode or
// alternate screen. It never opens /dev/tty to bypass redirected application IO.
type migrationPrompt struct {
	reader *bufio.Reader
	writer io.Writer
}

func newMigrationPrompt(reader io.Reader, writer io.Writer) *migrationPrompt {
	return &migrationPrompt{reader: bufio.NewReaderSize(reader, 16*1024), writer: writer}
}

func migrationHasTerminal(reader io.Reader, writer io.Writer) bool {
	type descriptor interface{ Fd() uintptr }
	input, inputOK := reader.(descriptor)
	output, outputOK := writer.(descriptor)
	return inputOK && outputOK && term.IsTerminal(input.Fd()) && term.IsTerminal(output.Fd())
}

func migrationInteractiveMode(moduleSet, flagSet, requested, terminal bool) (bool, error) {
	if flagSet {
		if requested && !terminal {
			return false, fmt.Errorf("--interactive requires terminal input and output; use --module with explicit --write/--resolve-lock flags for scripts")
		}
		return requested, nil
	}
	return !moduleSet && terminal, nil
}

func (p *migrationPrompt) input(ctx context.Context, message, fallback string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	label := message
	if fallback != "" {
		label += " [" + fallback + "]"
	}
	if _, err := fmt.Fprintf(p.writer, "%s: ", label); err != nil {
		return "", fmt.Errorf("Fprintf: %w", err)
	}
	line, err := p.reader.ReadSlice('\n')
	if err != nil {
		// Even an unterminated "yes" is not consent. EOF never accepts defaults.
		return "", fmt.Errorf("migration input cancelled or incomplete: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(line))
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("migration input must not contain control characters")
	}
	if value == "" {
		value = fallback
	}
	return value, nil
}

func (p *migrationPrompt) confirm(ctx context.Context, question string) (bool, error) {
	for {
		value, err := p.input(ctx, question+" [y/N]", "")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(value) {
		case "y", "yes":
			return true, nil
		case "", "n", "no":
			return false, nil
		default:
			if _, err := io.WriteString(p.writer, "Please answer yes or no; Enter means no.\n"); err != nil {
				return false, fmt.Errorf("WriteString: %w", err)
			}
		}
	}
}
