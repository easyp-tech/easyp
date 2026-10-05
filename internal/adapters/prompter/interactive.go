package prompter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// InteractivePrompter provides deterministic line-oriented terminal prompts.
// It deliberately avoids terminal capability/background queries so invoking an
// EasyP command never writes OSC/CSI probes before the command itself runs.
type InteractivePrompter struct {
	In  io.Reader
	Out io.Writer
}

var _ Prompter = InteractivePrompter{}

func (p InteractivePrompter) input() io.Reader {
	if p.In != nil {
		return p.In
	}
	return os.Stdin
}

func (p InteractivePrompter) output() io.Writer {
	if p.Out != nil {
		return p.Out
	}
	return os.Stdout
}

func (p InteractivePrompter) Confirm(ctx context.Context, message string, defaultValue bool) (bool, error) {
	hint := "[y/N]"
	if defaultValue {
		hint = "[Y/n]"
	}
	for {
		if err := ctx.Err(); err != nil {
			return defaultValue, err
		}
		if _, err := fmt.Fprintf(p.output(), "%s %s: ", message, hint); err != nil {
			return defaultValue, fmt.Errorf("write prompt: %w", err)
		}
		line, err := readPromptLine(p.input())
		if err != nil {
			return defaultValue, fmt.Errorf("read prompt: %w", err)
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "":
			return defaultValue, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "q":
			return defaultValue, context.Canceled
		default:
			if _, err := fmt.Fprintln(p.output(), "Please answer yes or no."); err != nil {
				return defaultValue, fmt.Errorf("write prompt: %w", err)
			}
		}
	}
}

func (p InteractivePrompter) Select(ctx context.Context, message string, options []string, defaultIndex int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("select requires at least one option")
	}
	if defaultIndex < 0 || defaultIndex >= len(options) {
		return 0, fmt.Errorf("default selection %d is outside 0..%d", defaultIndex, len(options)-1)
	}
	for {
		if err := ctx.Err(); err != nil {
			return defaultIndex, err
		}
		if _, err := fmt.Fprintln(p.output(), message); err != nil {
			return defaultIndex, fmt.Errorf("write prompt: %w", err)
		}
		for i, option := range options {
			if _, err := fmt.Fprintf(p.output(), "  %d) %s\n", i+1, option); err != nil {
				return defaultIndex, fmt.Errorf("write prompt: %w", err)
			}
		}
		if _, err := fmt.Fprintf(p.output(), "Selection [%d]: ", defaultIndex+1); err != nil {
			return defaultIndex, fmt.Errorf("write prompt: %w", err)
		}
		line, err := readPromptLine(p.input())
		if err != nil {
			return defaultIndex, fmt.Errorf("read prompt: %w", err)
		}
		value := strings.TrimSpace(line)
		if value == "" {
			return defaultIndex, nil
		}
		if strings.EqualFold(value, "q") {
			return defaultIndex, context.Canceled
		}
		index, err := strconv.Atoi(value)
		if err == nil && index >= 1 && index <= len(options) {
			return index - 1, nil
		}
		if _, err := fmt.Fprintln(p.output(), "Enter one of the listed numbers."); err != nil {
			return defaultIndex, fmt.Errorf("write prompt: %w", err)
		}
	}
}

func (p InteractivePrompter) MultiSelect(ctx context.Context, message string, options []string, defaults []bool) ([]int, error) {
	if len(defaults) > len(options) {
		return nil, fmt.Errorf("default selections exceed available options")
	}
	if err := ctx.Err(); err != nil {
		return selectedDefaults(defaults), err
	}
	if _, err := fmt.Fprintln(p.output(), message); err != nil {
		return nil, fmt.Errorf("write prompt: %w", err)
	}
	for i, option := range options {
		selected := " "
		if i < len(defaults) && defaults[i] {
			selected = "x"
		}
		if _, err := fmt.Fprintf(p.output(), "  %d) [%s] %s\n", i+1, selected, option); err != nil {
			return nil, fmt.Errorf("write prompt: %w", err)
		}
	}
	if _, err := fmt.Fprint(p.output(), "Selections (comma separated; Enter keeps defaults): "); err != nil {
		return nil, fmt.Errorf("write prompt: %w", err)
	}
	line, err := readPromptLine(p.input())
	if err != nil {
		return selectedDefaults(defaults), fmt.Errorf("read prompt: %w", err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return selectedDefaults(defaults), nil
	}
	if strings.EqualFold(value, "q") {
		return selectedDefaults(defaults), context.Canceled
	}
	seen := make(map[int]bool)
	var result []int
	for _, item := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }) {
		index, err := strconv.Atoi(item)
		if err != nil || index < 1 || index > len(options) {
			return selectedDefaults(defaults), fmt.Errorf("selection %q is not one of 1..%d", item, len(options))
		}
		index--
		if !seen[index] {
			seen[index] = true
			result = append(result, index)
		}
	}
	return result, nil
}

func (p InteractivePrompter) Input(ctx context.Context, message string, defaultValue string) (string, error) {
	if err := ctx.Err(); err != nil {
		return defaultValue, err
	}
	suffix := ": "
	if defaultValue != "" {
		suffix = fmt.Sprintf(" [%s]: ", defaultValue)
	}
	if _, err := fmt.Fprint(p.output(), message+suffix); err != nil {
		return defaultValue, fmt.Errorf("write prompt: %w", err)
	}
	line, err := readPromptLine(p.input())
	if err != nil {
		return defaultValue, fmt.Errorf("read prompt: %w", err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func selectedDefaults(defaults []bool) []int {
	var result []int
	for i, selected := range defaults {
		if selected {
			result = append(result, i)
		}
	}
	return result
}

// readPromptLine reads exactly through one newline without buffering bytes from
// the next answer. That matters when several prompts share a pipe or TTY.
func readPromptLine(reader io.Reader) (string, error) {
	var result strings.Builder
	var one [1]byte
	for {
		n, err := reader.Read(one[:])
		if n == 1 {
			switch one[0] {
			case '\n':
				return strings.TrimSuffix(result.String(), "\r"), nil
			default:
				result.WriteByte(one[0])
			}
		}
		if err != nil {
			if err == io.EOF && result.Len() > 0 {
				return strings.TrimSuffix(result.String(), "\r"), nil
			}
			return "", err
		}
	}
}
