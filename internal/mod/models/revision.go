package models

type Revision struct {
	CommitHash string // commit's hash
	Version    string // tag or HEAD if version was omitted
}
