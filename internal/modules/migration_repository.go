package modules

import "context"

// MigrationRepository verifies a historical installed-tree hash and returns
// independently hashed v1 metadata for that same Git revision.
type MigrationRepository interface {
	FetchMigration(context.Context, string, string, string) (Fetched, error)
}
