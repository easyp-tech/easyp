package models

// CacheDownloadPaths collects cache download paths to:
// * archive
// * file with archive hash
// * info about downloaded module
type CacheDownloadPaths struct {
	// CacheDownload path to dir with downloaded cache
	CacheDownloadDir string

	// ArchiveFile full path to downloaded archive of module
	ArchiveFile string

	// ModuleInfoFile full path to file with info about downloaded module
	ModuleInfoFile string
}
