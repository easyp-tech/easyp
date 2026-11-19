package core

import (
	"context"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
)

// CompileRuleName reports compiler errors found while linking linted files.
const CompileRuleName = "COMPILE"

// linkSources is a temporary zero-value stub for the corpus-linker
// classification (plan 02-01, RED step). The GREEN step replaces this body
// with the protoc.Link-based classifier.
func (c *Core) linkSources(ctx context.Context, disk FS, paths []string, sources [][]byte) (*protoc.Linked, []IssueInfo) {
	return nil, nil
}
