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
	if name == "." || name == "" {
		return nil
	}
	current := directory
	for _, component := range strings.Split(filepath.FromSlash(name), string(filepath.Separator)) {
		requested := filepath.Join(current, component)
		_, err := os.Lstat(requested)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		entries, err := os.ReadDir(current)
		if err != nil {
			return fmt.Errorf("ReadDir: %w", err)
		}
		exact := false
		for _, entry := range entries {
			exact = exact || entry.Name() == component
		}
		if !exact {
			return fmt.Errorf("%w: %q resolves to a different spelling below %q", ErrPathCollision, name, current)
		}
		current = requested
	}
	return nil
}
