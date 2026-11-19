package core

import "context"

// LintLinked lints the proto files with the two-pass linked engine over the
// easypc linked syntax trees. RED stub — the engine body lands with the GREEN
// commit.
func (c *Core) LintLinked(ctx context.Context, fsWalker DirWalker) ([]IssueInfo, error) {
	return nil, nil
}
