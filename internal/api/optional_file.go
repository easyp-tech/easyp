package api

import (
	"errors"
	"fmt"
	"os"

	"github.com/easyp-tech/easyp/internal/workspace"
)

func readOptionalFile(path string) ([]byte, bool, error) {
	raw, err := workspace.ReadFileAt(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("ReadFile: %w", err)
	}
	return raw, true, nil
}
