package models

// Revision collects references to module's commit
// Revision is actual module information from repository
type Revision struct {
	CommitHash string // commit's hash
	Version    string // commit's tag or generated version
}
