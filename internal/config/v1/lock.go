package v1

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

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
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
	Commit  string `yaml:"commit"`
	Hash    string `yaml:"hash"`
}

// ParseLock reads and validates a v1 protobuf.lock document.
func ParseLock(reader io.Reader) (Lock, error) {
	var lock Lock
	decoder := yaml.NewDecoder(reader)
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
		if !semver.IsValid(entry.Version) {
			return fmt.Errorf("%s: invalid locked version %q", entry.Source, entry.Version)
		}
		if (len(entry.Commit) != 40 && len(entry.Commit) != 64) || !isHex(entry.Commit) {
			return fmt.Errorf("%s: invalid commit %q", entry.Source, entry.Commit)
		}
		if !strings.HasPrefix(entry.Hash, "h1:") {
			return fmt.Errorf("%s: invalid content hash %q", entry.Source, entry.Hash)
		}
		digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(entry.Hash, "h1:"))
		if err != nil || len(digest) != 32 {
			return fmt.Errorf("%s: invalid content hash %q", entry.Source, entry.Hash)
		}
	}
	return nil
}

func isHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil
}
