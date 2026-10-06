// Package gitindex parses staged Git index records.
package gitindex

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Mode is the file mode recorded in the Git index.
type Mode string

const (
	// RegularFile is the mode of a non-executable regular file.
	RegularFile Mode = "100644"
	// ExecutableFile is the mode of an executable regular file.
	ExecutableFile Mode = "100755"
	// Symlink is the mode of a symbolic link.
	Symlink Mode = "120000"
	// Gitlink is the mode of a submodule entry.
	Gitlink Mode = "160000"
)

// Entry is one staged Git index record.
type Entry struct {
	Mode Mode
	Path string
	// Raw preserves the original record without its NUL separator.
	Raw string
}

// IsRegular reports whether the entry is a regular or executable file.
func (e Entry) IsRegular() bool {
	return e.Mode == RegularFile || e.Mode == ExecutableFile
}

// Parse parses NUL-separated output from git ls-files --stage -z.
// It requires stage zero and local paths, leaving mode policy to callers.
func Parse(raw string) ([]Entry, error) {
	var entries []Entry
	for record := range strings.SplitSeq(raw, "\x00") {
		if record == "" {
			continue
		}
		metadata, path, ok := strings.Cut(record, "\t")
		fields := strings.Fields(metadata)
		if !ok || len(fields) != 3 || fields[2] != "0" {
			return nil, fmt.Errorf("invalid Git index entry %q", record)
		}
		if !filepath.IsLocal(filepath.FromSlash(path)) {
			return nil, fmt.Errorf("invalid tracked file path %q", path)
		}
		entries = append(entries, Entry{Mode: Mode(fields[0]), Path: path, Raw: record})
	}
	return entries, nil
}
