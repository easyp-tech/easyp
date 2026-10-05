package modules

import (
	"context"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// BSRResolver chooses a pinned Git source while preserving the original Buf request.
// Implementations may use local mappings or a service; the Git graph consumes the same requirement.
type BSRResolver interface {
	Resolve(context.Context, v1.BSRDependency) (v1.BSRResolution, error)
}
