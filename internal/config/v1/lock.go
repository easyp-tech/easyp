package v1

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

// Lock is the v1 protobuf.lock document for one protobuf module.
type Lock struct {
	Version int            `yaml:"version"`
	Modules []LockedModule `yaml:"modules"`
}

// LockedModule identifies one exact Git dependency in protobuf.lock.
type LockedModule struct {
	Source  string          `yaml:"source"`
	Version string          `yaml:"version"`
	Commit  string          `yaml:"commit"`
	Hash    string          `yaml:"hash"`
	Roots   []string        `yaml:"roots,omitempty"`
	BSR     []BSRResolution `yaml:"bsr,omitempty"`
}

// ParseLock reads and validates a v1 protobuf.lock document.
func ParseLock(reader io.Reader) (Lock, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return Lock{}, fmt.Errorf("ReadAll: %w", err)
	}
	if err := validateLockRootsYAML(raw); err != nil {
		return Lock{}, fmt.Errorf("validateLockRootsYAML: %w", err)
	}
	var lock Lock
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&lock); err != nil {
		return Lock{}, fmt.Errorf("Decode: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return Lock{}, fmt.Errorf("Decode: %w", err)
		}
		return Lock{}, fmt.Errorf("protobuf.lock must contain one YAML document")
	}
	if err := lock.Validate(); err != nil {
		return Lock{}, fmt.Errorf("Validate: %w", err)
	}
	return lock, nil
}

// YAML's default string slice decoder coerces numeric scalars and drops null
// entries. Validate root nodes before decoding so the parsed lock matches its schema.
func validateLockRootsYAML(raw []byte) error {
	var document map[string]any
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("Unmarshal: %w", err)
	}
	entries, ok := document["modules"].([]any)
	if !ok {
		return nil
	}
	for _, rawEntry := range entries {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			continue
		}
		rawRoots, present := entry["roots"]
		if !present {
			continue
		}
		roots, ok := rawRoots.([]any)
		if !ok {
			return fmt.Errorf("roots must be a list of directory strings")
		}
		for index, root := range roots {
			if _, ok := root.(string); !ok {
				return fmt.Errorf("roots[%d] must be a directory string", index)
			}
		}
	}
	return nil
}

// Validate checks the version and integrity fields before a lock is used.
func (lock Lock) Validate() error {
	if lock.Version != 1 {
		return fmt.Errorf("unsupported protobuf.lock version %d", lock.Version)
	}
	seen := make(map[string]struct{}, len(lock.Modules))
	for _, entry := range lock.Modules {
		if entry.Source == "" || entry.Commit == "" || entry.Hash == "" {
			return fmt.Errorf("incomplete lock entry for %q", entry.Source)
		}
		if _, exists := seen[entry.Source]; exists {
			return fmt.Errorf("duplicate locked module %s", entry.Source)
		}
		seen[entry.Source] = struct{}{}
		if err := ValidateModuleRoots(entry.Roots); err != nil {
			return fmt.Errorf("ValidateModuleRoots: %s: %w", entry.Source, err)
		}
		if err := ValidateModuleVersion(entry.Source, entry.Version); err != nil {
			return fmt.Errorf("ValidateModuleVersion: %w", err)
		}
		if !semver.IsValid(entry.Version) && !IsCommitRef(entry.Version) {
			return fmt.Errorf("%s: invalid locked version %q", entry.Source, entry.Version)
		}
		if !IsCommitRef(entry.Commit) {
			return fmt.Errorf("%s: invalid commit %q", entry.Source, entry.Commit)
		}
		if IsCommitRef(entry.Version) && !strings.EqualFold(entry.Version, entry.Commit) {
			return fmt.Errorf("%s: pinned version %q differs from commit %q", entry.Source, entry.Version, entry.Commit)
		}
		if !strings.HasPrefix(entry.Hash, "h1:") {
			return fmt.Errorf("%s: invalid content hash %q", entry.Source, entry.Hash)
		}
		digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(entry.Hash, "h1:"))
		if err != nil || len(digest) != 32 {
			return fmt.Errorf("%s: invalid content hash %q", entry.Source, entry.Hash)
		}
		origins := make(map[string]bool, len(entry.BSR))
		for _, binding := range entry.BSR {
			if err := binding.Validate(); err != nil {
				return fmt.Errorf("Validate: %s: %w", entry.Source, err)
			}
			origin := binding.Dependency.Config + ":" + binding.Dependency.Module
			if origins[origin] {
				return fmt.Errorf("duplicate BSR binding for %s in %s", binding.Dependency.Module, binding.Dependency.Config)
			}
			origins[origin] = true
		}
	}
	return nil
}

// ValidateModuleRoots checks canonical portable module-relative directory names.
// Empty selections are omitted from the lock; duplicate directories are invalid.
func ValidateModuleRoots(roots []string) error {
	seen := make(map[string]bool, len(roots))
	for index, root := range roots {
		if !utf8.ValidString(root) || !pathSelector.MatchString(root) {
			return fmt.Errorf("roots[%d]: %q is not a canonical portable module-relative directory path", index, root)
		}
		if seen[root] {
			return fmt.Errorf("roots[%d]: duplicate directory %q", index, root)
		}
		seen[root] = true
	}
	return nil
}

func isHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil
}

// IsCommitRef reports whether value is an unambiguous full Git commit hash.
func IsCommitRef(value string) bool {
	return (len(value) == 40 || len(value) == 64) && isHex(value)
}
