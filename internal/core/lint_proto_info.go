package core

import (
	"context"
	"fmt"
)

// lintProtoInfoRead is the lint-owned read-model constructor (D-06): the lint
// path's own twin of breakingProtoInfoRead. Behaviorally identical in Phase 1;
// only the read-model type differs.
func (c *Core) lintProtoInfoRead(ctx context.Context, fs FS, path string) (ProtoInfo, error) {
	f, err := c.openDiskSource(fs, path)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("Open: %w", err)
	}
	defer c.close(ctx, f, path)

	protoFile, comments, err := readProtoFileWithDirectives(f)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("readProtoFileWithDirectives: %w", err)
	}

	protoFilesFromImport, err := c.readFilesFromImport(ctx, fs, protoFile)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("readFilesFromImport: %w", err)
	}

	protoInfo := ProtoInfo{
		directiveComments:    comments,
		Path:                 path,
		Info:                 protoFile,
		ProtoFilesFromImport: protoFilesFromImport,
	}
	protoInfo.ImportPath = c.sourceImportPath(fs, path)
	return protoInfo, nil
}
