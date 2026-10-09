package gitmodules

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
)

func gitV1(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := gitcommand.Run(ctx, dir, nil, args...)
	if err != nil {
		return "", fmt.Errorf("Run: %w", err)
	}
	return string(out), nil
}

func v1CacheSourceKey(source string) string {
	hash := sha256.Sum256([]byte(source))
	return fmt.Sprintf("%x", hash)
}
