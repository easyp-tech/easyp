package gitmodules

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/easyp-tech/easyp/internal/adapters/gitsnapshot"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// PolicyFiles supplies a lazy, read-only source for a reached and verified pin.
// Prefix is the module's logical manifest directory, not its import root.
func (c *Cache) PolicyFiles(entry v1.LockedModule, prefix string) modules.PolicyFiles {
	return pinnedPolicyFiles{root: c.root, entry: entry, prefix: prefix}
}

type pinnedPolicyFiles struct {
	root   string
	entry  v1.LockedModule
	prefix string
}

func (p pinnedPolicyFiles) Read(ctx context.Context, relative string) (modules.PolicyFile, error) {
	if !fs.ValidPath(relative) || !fs.ValidPath(p.prefix) {
		return modules.PolicyFile{}, fmt.Errorf("invalid module policy path %q", relative)
	}
	repository, err := readSourceBinding(p.root, p.entry)
	if err != nil {
		return modules.PolicyFile{}, fmt.Errorf("cached policy objects are unavailable; run easyp mod download: %w", err)
	}
	var tree *gitsnapshot.FS
	if len(p.entry.Commit) == 64 {
		tree, err = gitsnapshot.NewRepository(ctx, repository, p.entry.Commit)
	} else {
		repo, openErr := git.PlainOpen(repository)
		if openErr != nil {
			return modules.PolicyFile{}, fmt.Errorf("cached policy objects are unavailable; run easyp mod download: %w", openErr)
		}
		tree, err = gitsnapshot.New(repo, plumbing.NewHash(p.entry.Commit))
	}
	if err != nil {
		return modules.PolicyFile{}, fmt.Errorf("open pinned policy source; run easyp mod download: %w", err)
	}
	root, err := sourceview.New(tree).Resolve(ctx, p.prefix)
	if err != nil {
		return modules.PolicyFile{}, fmt.Errorf("Resolve: module root: %w", err)
	}
	if !root.Info.IsDir() {
		return modules.PolicyFile{}, fmt.Errorf("module policy root is not a directory")
	}
	bounded, err := fs.Sub(tree, root.Path)
	if err != nil {
		return modules.PolicyFile{}, fmt.Errorf("Sub: %w", err)
	}
	view := sourceview.New(bounded)
	logical := relative
	resolved, err := view.Resolve(ctx, logical)
	if err != nil {
		return modules.PolicyFile{}, fmt.Errorf("Resolve: %w", err)
	}
	if resolved.Info.IsDir() {
		relative = path.Join(relative, v1.PolicyFile)
		logical = relative
		resolved, err = view.Resolve(ctx, logical)
		if err != nil {
			return modules.PolicyFile{}, fmt.Errorf("Resolve: %w", err)
		}
	}
	file, err := view.Open(ctx, logical)
	if err != nil {
		return modules.PolicyFile{}, fmt.Errorf("Open: %w", err)
	}
	data, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil {
		return modules.PolicyFile{}, fmt.Errorf("ReadAll: %w", readErr)
	}
	if closeErr != nil {
		return modules.PolicyFile{}, fmt.Errorf("Close: %w", closeErr)
	}
	return modules.PolicyFile{Path: filepath.ToSlash(relative), Canonical: p.entry.Source + "@" + strings.ToLower(p.entry.Commit) + ":" + path.Join(root.Path, resolved.Path), Content: data}, nil
}
