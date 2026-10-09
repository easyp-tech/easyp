package gitmodules

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type pinnedSourceBinding struct {
	Repository string `json:"repository"`
}

func sourceBindingPath(root string, entry v1.LockedModule) string {
	return filepath.Join(root, "sources", v1CacheSourceKey(entry.Source), strings.ToLower(entry.Commit)+".json")
}

func readSourceBinding(root string, entry v1.LockedModule) (string, error) {
	data, err := os.ReadFile(sourceBindingPath(root, entry))
	if err != nil {
		return "", fmt.Errorf("ReadFile: %w", err)
	}
	var binding pinnedSourceBinding
	if err := json.Unmarshal(data, &binding); err != nil {
		return "", fmt.Errorf("Unmarshal: %w", err)
	}
	decoded, err := hex.DecodeString(binding.Repository)
	if err != nil || len(decoded) != 32 {
		return "", fmt.Errorf("invalid pinned repository binding")
	}
	return filepath.Join(root, "objects", binding.Repository), nil
}

func writeSourceBinding(ctx context.Context, root, checkout string, entry v1.LockedModule) (err error) {
	// Git may repack a shallow --shared clone instead of writing alternates.
	// Its declared origin still identifies our local, persistent object store.
	data, err := gitV1(ctx, checkout, "config", "--get", "remote.origin.url")
	if err != nil {
		return fmt.Errorf("gitV1: %w", err)
	}
	repository := strings.TrimSpace(data)
	objectsRoot, err := filepath.Abs(filepath.Join(root, "objects"))
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	key, err := filepath.Rel(objectsRoot, repository)
	if err != nil || !filepath.IsLocal(key) || strings.ContainsAny(key, "/\\") {
		return fmt.Errorf("invalid cached object repository %q", repository)
	}
	encoded, err := json.Marshal(pinnedSourceBinding{Repository: key})
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}
	filename := sourceBindingPath(root, entry)
	old, readErr := os.ReadFile(filename)
	if readErr == nil && bytes.Equal(old, encoded) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(filename), ".source-*")
	if err != nil {
		return fmt.Errorf("CreateTemp: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("Remove: %w", removeErr))
		}
	}()
	_, writeErr := file.Write(encoded)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.Join(writeErr, closeErr)
	}
	return os.Rename(file.Name(), filename)
}

// Native snapshots do not contain arbitrary policy fragments. Explicit download
// retains their immutable object source too; read-only validation never repairs it.
func (c *Cache) ensurePinnedSource(ctx context.Context, entry v1.LockedModule, quarantineIncomplete bool) (err error) {
	var refetchRepository string
	if repository, readErr := readSourceBinding(c.root, entry); readErr == nil {
		incomplete := false
		_, err := gitV1(ctx, repository, "cat-file", "-e", entry.Commit+"^{commit}")
		if err == nil {
			// Connectivity checks read Git tree/index metadata and check object
			// presence; they do not hash unrelated blob contents.
			_, err = gitV1(ctx, repository, "fsck", "--connectivity-only", "--no-dangling", "--no-reflogs", entry.Commit)
			if err == nil {
				return nil
			}
			incomplete = true
		}
		if gitcommand.Interrupted(err) {
			return fmt.Errorf("gitV1: %w", err)
		}
		if incomplete && quarantineIncomplete {
			unlock, err := lockObjectRepository(ctx, repository+".lock")
			if err != nil {
				return fmt.Errorf("lockObjectRepository: %w", err)
			}
			if _, checkErr := gitV1(ctx, repository, "fsck", "--connectivity-only", "--no-dangling", "--no-reflogs", entry.Commit); checkErr == nil {
				unlock()
				return nil
			} else if gitcommand.Interrupted(checkErr) {
				unlock()
				return fmt.Errorf("gitV1: %w", checkErr)
			}
			quarantine, err := os.MkdirTemp(c.root, "incomplete-objects-*")
			if err == nil {
				err = os.Rename(repository, filepath.Join(quarantine, "repository"))
			}
			unlock()
			if err != nil {
				return fmt.Errorf("quarantine pinned objects: %w", err)
			}
			gitcommand.Debug(ctx, "Quarantined incomplete pinned objects before explicit download", slog.String("directory", quarantine))
		} else if incomplete {
			refetchRepository = repository
		}
	}
	candidates, err := v1GitModuleCandidates(entry.Source)
	if err != nil {
		return fmt.Errorf("v1GitModuleCandidates: %w", err)
	}
	var failures []error
	for _, candidate := range candidates {
		directory, err := os.MkdirTemp(c.root, "source-*")
		if err != nil {
			return fmt.Errorf("MkdirTemp: %w", err)
		}
		var fetchErr error
		if refetchRepository == filepath.Join(c.root, "objects", v1CacheSourceKey(candidate.remote)) {
			fetchErr = refetchPinnedSource(ctx, refetchRepository, candidate.remote, entry.Commit)
		}
		if fetchErr == nil {
			_, fetchErr = checkoutCachedCommit(ctx, directory, entry, candidate)
		}
		if fetchErr == nil {
			fetchErr = writeSourceBinding(ctx, c.root, directory, entry)
		}
		removeErr := os.RemoveAll(directory)
		if removeErr != nil {
			return errors.Join(fetchErr, fmt.Errorf("RemoveAll: %w", removeErr))
		}
		if fetchErr == nil {
			return nil
		}
		failures = append(failures, fetchErr)
		if gitcommand.Interrupted(fetchErr) {
			break
		}
	}
	return errors.Join(failures...)
}

func refetchPinnedSource(ctx context.Context, repository, remote, commit string) error {
	unlock, err := lockObjectRepository(ctx, repository+".lock")
	if err != nil {
		return fmt.Errorf("lockObjectRepository: %w", err)
	}
	defer unlock()
	if _, err := gitV1(ctx, repository, "fsck", "--connectivity-only", "--no-dangling", "--no-reflogs", commit); err == nil {
		return nil
	} else if gitcommand.Interrupted(err) {
		return fmt.Errorf("gitV1: %w", err)
	}
	_, err = gitV1(ctx, repository, "-c", "fetch.fsckObjects=true", "fetch", "--quiet", "--refetch", "--depth=1", "--no-tags", "--no-write-fetch-head", "--", remote, commit)
	if err != nil {
		return fmt.Errorf("gitV1: %w", err)
	}
	return nil
}
