package repo

import (
	"context"
)

const (
	CacheArchiveName = "cache.zip"
)

type Repo interface {
	// GetFiles returns list of all files in repository
	GetFiles(ctx context.Context, dirs ...string) ([]string, error)

	// Archive passed dirs to archive and return full path to archive
	Archive(ctx context.Context, dirs ...string) (string, error)
}
