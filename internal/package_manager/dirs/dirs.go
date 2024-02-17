package dirs

// Dirs implements workflows with directories
type Dirs struct {
	cacheRootDir string
}

func New(cacheRootDir string) *Dirs {
	return &Dirs{
		cacheRootDir: cacheRootDir,
	}
}

const (
	cacheDirPerm = 0755
)
