package storage

const (
	cacheDir      = "cache"
	cacheDownload = "download"
)

// Storage implements workflows with directories
type Storage struct {
	rootDir string
}

func New(rootDir string) *Storage {
	return &Storage{
		rootDir: rootDir,
	}
}

const (
	cacheDirPerm = 0755
)
