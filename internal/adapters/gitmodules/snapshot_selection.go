package gitmodules

import (
	"io/fs"
	"path"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Snapshot bodies belong to protobuf contracts and their configuration. Other
// directory entries remain useful for discovering roots and bounded aliases.
func snapshotConfigFile(name string) bool {
	base := path.Base(name)
	return moduleconfig.IsGitDependencyConfigFile(base) || base == v1.GenerateFile || base == v1.LockFile
}

func snapshotSourceCandidate(name string, entry fs.DirEntry) bool {
	return entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 || path.Ext(name) == ".proto" || snapshotConfigFile(name)
}

func snapshotProtoCandidate(name string, entry fs.DirEntry) bool {
	return entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 || path.Ext(name) == ".proto"
}
