package repository

import (
	"context"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
)

const (
	CacheArchiveName = "cache.zip"
)

type Repo interface {
	// GetFiles returns list of all files in repository
	GetFiles(ctx context.Context, revision models.Revision, dirs ...string) ([]string, error)

	// Archive passed storage to archive and return full path to archive
	Archive(ctx context.Context, revision models.Revision, archivePath string, dirs ...string) error

	// ReadRevision reads commit's revision by passed version
	// or return the latest commit if version is empty
	ReadRevision(ctx context.Context, version string) (models.Revision, error)

	// Fetch from remote repository specified version
	Fetch(ctx context.Context, revision models.Revision) error
}
