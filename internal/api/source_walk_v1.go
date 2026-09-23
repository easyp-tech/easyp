package api

import (
	"io/fs"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)

func walkV1ProtoFiles(root string, visit func(string) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path_helpers.ShouldSkipV1SourceDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".proto" {
			return nil
		}
		return visit(path)
	})
}
