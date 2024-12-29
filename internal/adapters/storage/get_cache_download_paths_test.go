package storage

import (
	"path/filepath"
)

func (s *storageSuite) Test_GetCacheDownloadPaths() {
	module := getFakeModule()
	revision := getFakeRevision()

	// ref values
	expectedCacheDownloadDir := filepath.Join(s.rootDir, cacheDir, cacheDownloadDir, module.Name)
	expectedArchiveFile := filepath.Join(expectedCacheDownloadDir, revision.Version) + ".zip"
	expectedArchiveHashFile := filepath.Join(expectedCacheDownloadDir, revision.Version) + ".ziphash"
	expectedModuleInfoFile := filepath.Join(expectedCacheDownloadDir, revision.Version) + ".info"

	res := s.storage.GetCacheDownloadPaths(module, revision)

	s.Equal(expectedCacheDownloadDir, res.CacheDownloadDir)
	s.Equal(expectedArchiveFile, res.ArchiveFile)
	s.Equal(expectedArchiveHashFile, res.ArchiveHashFile)
	s.Equal(expectedModuleInfoFile, res.ModuleInfoFile)
}
