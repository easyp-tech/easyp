package core

import "context"

// lintProtoInfoRead is the lint-owned read-model constructor (D-06): the lint
// path's own twin of breakingProtoInfoRead. Behaviorally identical in Phase 1;
// only the read-model type differs.
func (c *Core) lintProtoInfoRead(ctx context.Context, fs FS, path string) (ProtoInfo, error) {
	return ProtoInfo{}, nil
}
