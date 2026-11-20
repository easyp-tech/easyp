package sourceview

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const defaultSessionMaxDirs = 128

// Session keeps bounded local roots open for one command, sharing a capped LRU
// of directory handles across opens below those roots. Each open still resolves
// and checks its own path, rejects nested repositories, and verifies the opened
// file, matching OpenLocalSelected. A Session is safe for concurrent use; Close
// it when the command ends. A nil Session delegates to the package functions.
type Session struct {
	mu     sync.Mutex
	closed atomic.Bool
	cache  *dirCache
	roots  map[string]*sessionRoot
}

type sessionRoot struct {
	mu   sync.Mutex
	init bool
	tree *rootTree
	view *View
	err  error
}

// NewSession returns an empty session with the default directory handle limit.
func NewSession() *Session {
	return newSessionWithMaxDirs(defaultSessionMaxDirs)
}

func newSessionWithMaxDirs(maxDirs int) *Session {
	return &Session{
		cache: newDirCache(maxDirs),
		roots: make(map[string]*sessionRoot),
	}
}

func (s *Session) root(rootPath string) (*rootTree, *View, error) {
	s.mu.Lock()
	if s.closed.Load() {
		s.mu.Unlock()
		return nil, nil, fs.ErrClosed
	}
	entry, ok := s.roots[rootPath]
	if !ok {
		entry = &sessionRoot{}
		s.roots[rootPath] = entry
	}
	s.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()
	if s.closed.Load() {
		return nil, nil, fs.ErrClosed
	}
	if !entry.init {
		entry.tree, entry.view, entry.err = localTreeView(rootPath, s.cache)
		entry.init = true
	}
	return entry.tree, entry.view, entry.err
}

// OpenLocalSelected opens a logical source through the session's shared roots.
func (s *Session) OpenLocalSelected(ctx context.Context, rootPath, logical string, selected func(Resolution) bool, nestedAllowed ...func(Resolution) bool) (fs.File, error) {
	if s == nil {
		return OpenLocalSelected(ctx, rootPath, logical, selected, nestedAllowed...)
	}
	tree, view, err := s.root(filepath.Clean(rootPath))
	if err != nil {
		return nil, fmt.Errorf("localTreeView: %w", err)
	}
	return openSelected(ctx, tree.Open, tree.Lstat, view, logical, selected, nestedAllowed...)
}

// OpenLocal opens a logical source through the session's shared roots.
func (s *Session) OpenLocal(ctx context.Context, rootPath, logical string) (fs.File, error) {
	return s.OpenLocalSelected(ctx, rootPath, logical, nil)
}

// ReadLocal reads a logical regular source through the session's shared roots.
func (s *Session) ReadLocal(ctx context.Context, rootPath, logical string) (_ []byte, resultErr error) {
	file, err := s.OpenLocal(ctx, rootPath, logical)
	if err != nil {
		return nil, fmt.Errorf("OpenLocal: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source %q is not a regular file", logical)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return data, nil
}

// Close closes every root and cached directory handle opened by the session.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.closed.Swap(true) {
		s.mu.Unlock()
		return nil
	}
	entries := make([]*sessionRoot, 0, len(s.roots))
	for _, entry := range s.roots {
		entries = append(entries, entry)
	}
	s.roots = nil
	s.mu.Unlock()

	var errs []error
	for _, entry := range entries {
		entry.mu.Lock()
		if entry.tree != nil {
			errs = append(errs, entry.tree.Close())
		}
		entry.mu.Unlock()
	}
	errs = append(errs, s.cache.drainCloseErrs()...)
	return errors.Join(errs...)
}
