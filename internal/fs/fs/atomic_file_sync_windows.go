//go:build windows

package fs

// Windows does not expose POSIX directory fsync through os.File.Sync.
// The atomic rename remains the durability boundary supported by this helper.
func syncDirectory(string) error { return nil }
