package gitmodules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

// Shared checkouts resolve both objects and their sizes through explicitly
// rooted stores. Go-git's alternate filesystem can lose Windows drive names or
// reinterpret relative paths, and EncodedObjectSize never follows alternates.
type snapshotObjectStorage struct {
	storage.Storer
	alternates []storage.Storer
}

func (s snapshotObjectStorage) EncodedObject(kind plumbing.ObjectType, hash plumbing.Hash) (plumbing.EncodedObject, error) {
	object, err := s.Storer.EncodedObject(kind, hash)
	if !errors.Is(err, plumbing.ErrObjectNotFound) {
		return object, err
	}
	for _, alternate := range s.alternates {
		object, err = alternate.EncodedObject(kind, hash)
		if !errors.Is(err, plumbing.ErrObjectNotFound) {
			return object, err
		}
	}
	return nil, err
}

func (s snapshotObjectStorage) EncodedObjectSize(hash plumbing.Hash) (int64, error) {
	size, err := s.Storer.EncodedObjectSize(hash)
	if !errors.Is(err, plumbing.ErrObjectNotFound) {
		return size, err
	}
	for _, alternate := range s.alternates {
		size, err = alternate.EncodedObjectSize(hash)
		if !errors.Is(err, plumbing.ErrObjectNotFound) {
			return size, err
		}
	}
	return 0, err
}

func snapshotStorage(checkout string) (storage.Storer, error) {
	gitDir := filepath.Join(checkout, ".git")
	primary := filesystem.NewStorageWithOptions(osfs.New(gitDir), cache.NewObjectLRUDefault(), filesystem.Options{AlternatesFS: osfs.New(string(filepath.Separator))})
	data, err := os.ReadFile(filepath.Join(gitDir, "objects", "info", "alternates"))
	if errors.Is(err, os.ErrNotExist) {
		return primary, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}
	store := snapshotObjectStorage{Storer: primary}
	for _, name := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if !filepath.IsAbs(name) {
			name = filepath.Join(gitDir, "objects", name)
		}
		store.alternates = append(store.alternates, filesystem.NewStorage(osfs.New(filepath.Dir(name)), cache.NewObjectLRUDefault()))
	}
	return store, nil
}
