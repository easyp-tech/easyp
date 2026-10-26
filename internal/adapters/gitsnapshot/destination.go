package gitsnapshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrPathCollision means distinct logical Git paths share one host destination.
var ErrPathCollision = errors.New("snapshot path collision")

// ValidateDestination rejects distinct Git paths that the host filesystem aliases.
// Git paths stay case-sensitive even on a case-insensitive staging filesystem.
func ValidateDestination(directory, name string) error {
	return validateDestination(directory, name, nil, nil)
}

// DestinationValidator reuses validated path prefixes while one writer builds a
// private snapshot. It must be reset after removing or renaming staged paths.
type DestinationValidator struct {
	directory string
	checked   map[string]bool
	names     map[string]map[string]bool
}

// NewDestinationValidator creates a validator for an exclusively owned stage.
func NewDestinationValidator(directory string) *DestinationValidator {
	return &DestinationValidator{directory: directory, checked: make(map[string]bool), names: make(map[string]map[string]bool)}
}

// Validate rejects host aliases without rereading unchanged parent directories.
func (v *DestinationValidator) Validate(name string) error {
	return validateDestination(v.directory, name, v.checked, v.names)
}

// Record remembers a successfully created path after Validate has accepted it.
// The single snapshot writer must record every creation before the next check.
func (v *DestinationValidator) Record(name string) {
	if name == "." || name == "" {
		return
	}
	current := v.directory
	for _, component := range strings.Split(filepath.FromSlash(name), string(filepath.Separator)) {
		if names := v.names[current]; names != nil {
			names[component] = true
		}
		current = filepath.Join(current, component)
		v.checked[current] = true
	}
}

// Reset discards prefixes whose directories may have been removed or renamed.
func (v *DestinationValidator) Reset() {
	clear(v.checked)
	clear(v.names)
}

func validateDestination(directory, name string, checked map[string]bool, directories map[string]map[string]bool) error {
	if name == "." || name == "" {
		return nil
	}
	current := directory
	for _, component := range strings.Split(filepath.FromSlash(name), string(filepath.Separator)) {
		requested := filepath.Join(current, component)
		if checked[requested] {
			current = requested
			continue
		}
		_, err := os.Lstat(requested)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		names := directories[current]
		exact := false
		if names == nil {
			entries, err := os.ReadDir(current)
			if err != nil {
				return fmt.Errorf("ReadDir: %w", err)
			}
			if directories != nil {
				names = make(map[string]bool, len(entries))
			}
			for _, entry := range entries {
				exact = exact || entry.Name() == component
				if names != nil {
					names[entry.Name()] = true
				}
			}
			if directories != nil {
				directories[current] = names
			}
		} else {
			exact = names[component]
		}
		if !exact {
			return fmt.Errorf("%w: %q resolves to a different spelling below %q", ErrPathCollision, name, current)
		}
		if checked != nil {
			checked[requested] = true
		}
		current = requested
	}
	return nil
}
