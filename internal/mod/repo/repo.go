package repo

import (
	"context"
)

type Repo interface {
	// GetFiles returns list of all files in repository
	GetFiles(ctx context.Context) ([]string, error)
}
