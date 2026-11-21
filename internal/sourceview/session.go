package sourceview

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
)

// Session keeps bounded local roots open for one command, so opens below the
// same root share its directory handles instead of opening every ancestor
// directory again. Each open still resolves and checks its own path, rejects
// nested repositories and verifies the opened file, exactly as
// OpenLocalSelected does. A Session is safe for concurrent use; Close it when
// the command ends. A nil Session opens through the package functions.
type Session struct {
	mu    sync.Mutex
	roots map[string]*sessionRoot
}

type sessionRoot struct {
	once sync.Once
	tree *rootTree
	view *View
	err  error
}

// NewSession returns an empty session.
func NewSession() *Session {
	return &Session{roots: make(map[string]*sessionRoot)}
}

func (s *Session) root(rootPath string) (*rootTree, *View, error) {
	s.mu.Lock()
	entry, ok := s.roots[rootPath]
	if !ok {
		entry = &sessionRoot{}
		s.roots[rootPath] = entry
	}
	s.mu.Unlock()
	entry.once.Do(func() { entry.tree, entry.view, entry.err = localView(rootPath) })
	return entry.tree, entry.view, entry.err
}

// OpenLocalSelected is the package OpenLocalSelected through the session's roots.
func (s *Session) OpenLocalSelected(ctx context.Context, rootPath, logical string, selected func(Resolution) bool, nestedAllowed ...func(Resolution) bool) (fs.File, error) {
	if s == nil {
		return OpenLocalSelected(ctx, rootPath, logical, selected, nestedAllowed...)
	}
	root, view, err := s.root(filepath.Clean(rootPath))
	if err != nil {
		return nil, fmt.Errorf("localView: %w", err)
	}
	return openSelected(ctx, root, view, logical, selected, nestedAllowed...)
}

// OpenLocal is the package OpenLocal through the session's roots.
func (s *Session) OpenLocal(ctx context.Context, rootPath, logical string) (fs.File, error) {
	return s.OpenLocalSelected(ctx, rootPath, logical, nil)
}

// Close closes every root the session opened.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for _, entry := range s.roots {
		if entry.tree != nil {
			errs = append(errs, entry.tree.Close())
		}
	}
	s.roots = nil
	return errors.Join(errs...)
}

// ReadLocal is the package ReadLocal through the session's roots.
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
