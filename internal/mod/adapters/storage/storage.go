package storage

const (
	// root cache dir
	cacheDir = "cache"
	// dir for downloaded (check sum, archive)
	cacheDownloadDir = "download"
	// dir for installed packages
	installedDir = "mod"
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
	dirPerm = 0755
)
