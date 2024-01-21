package repo

type Repo interface {
}

// WorkDir returns the name of the cached work directory to use for the
// given repository type and name.
func WorkDir() error {
	return nil
}
