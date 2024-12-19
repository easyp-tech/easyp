package lockfile

// IsEmpty check if lock file doesn't have any deps
func (l *LockFile) IsEmpty() bool {
	return len(l.cache) == 0
}
