// Package sourceview resolves logical source names within a bounded filesystem.
package sourceview

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

var (
	// ErrOutsideRoot marks a pointer that leaves the filesystem boundary.
	ErrOutsideRoot = errors.New("source target is outside root")
	// ErrCycle marks an active link or directory expansion cycle.
	ErrCycle = errors.New("source link cycle")
	// ErrUnsupported marks an unsupported pointer spelling or file type.
	ErrUnsupported = errors.New("unsupported source target")
	// ErrChanged marks a target replaced between resolution and opening.
	ErrChanged = errors.New("source target changed during open")
)

// View resolves relative logical names against a borrowed, bounded filesystem.
// Its filesystem should implement fs.ReadLinkFS when it contains symbolic links.
// A local filesystem must enforce its own boundary, such as os.Root.FS.
// View does not close the filesystem and is safe for concurrent use when it is.
type View struct {
	fsys          fs.FS
	absoluteRoots []string
}

// Resolution describes a target and the physical links followed to reach it.
// Path is a relative physical path in the supplied filesystem. On failure it
// describes the last inspected target, and Info may be nil.
type Resolution struct {
	Path  string
	Info  fs.FileInfo
	Links []Link
}

// Link preserves a physical link's original target and its unfollowed metadata.
type Link struct {
	Path   string
	Target string
	Info   fs.FileInfo
}

// WalkFunc receives logical names, their resolution, and any resolution or
// directory-reading error. Returning nil ignores the error and skips that entry.
// fs.SkipDir and fs.SkipAll have the same meanings as in fs.WalkDir.
type WalkFunc func(logical string, resolved Resolution, err error) error

// New borrows fsys and rejects all absolute symbolic link targets.
func New(fsys fs.FS) *View {
	return &View{fsys: fsys}
}

// NewLocal borrows fsys and accepts absolute links beneath canonicalRoot.
// canonicalRoot must be the absolute, canonical OS path of fsys's boundary.
// The caller must canonicalize it before opening the bounded filesystem.
// rootAliases are alternative absolute spellings that the caller has verified
// identify the same physical root. Targets are never canonicalized on the host.
func NewLocal(fsys fs.FS, canonicalRoot string, rootAliases ...string) (*View, error) {
	roots := append([]string{canonicalRoot}, rootAliases...)
	for index, root := range roots {
		if !filepath.IsAbs(root) || strings.ContainsRune(root, '\x00') {
			return nil, &fs.PathError{Op: "NewLocal", Path: root, Err: fs.ErrInvalid}
		}
		roots[index] = filepath.ToSlash(filepath.Clean(root))
	}
	return &View{fsys: fsys, absoluteRoots: roots}, nil
}

// Resolve follows links component by component without cleaning away dots
// before preceding links have resolved. Relative parents cannot leave the root.
// Link traces are also returned for dangling or rejected targets.
func (v *View) Resolve(ctx context.Context, logical string) (Resolution, error) {
	resolved := Resolution{Path: "."}
	err := ctx.Err()
	if err != nil {
		return resolved, &fs.PathError{Op: "Resolve", Path: logical, Err: err}
	}
	if logical == "" || v.fsys == nil {
		return resolved, &fs.PathError{Op: "Resolve", Path: logical, Err: fs.ErrInvalid}
	}
	err = validatePointer(logical)
	if err != nil {
		return resolved, &fs.PathError{Op: "Resolve", Path: logical, Err: err}
	}
	if strings.HasPrefix(logical, "/") {
		return resolved, &fs.PathError{Op: "Resolve", Path: logical, Err: ErrOutsideRoot}
	}
	err = v.resolveComponents(ctx, ".", strings.Split(logical, "/"), make(map[string]bool), &resolved)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return resolved, &fs.PathError{Op: "Resolve", Path: logical, Err: err}
	}
	return resolved, nil
}

func (v *View) resolveComponents(ctx context.Context, base string, components []string, active map[string]bool, resolved *Resolution) error {
	err := ctx.Err()
	if err != nil {
		return err
	}
	info, err := fs.Lstat(v.fsys, base)
	resolved.Path, resolved.Info = base, info
	if err != nil {
		return fmt.Errorf("Lstat: %w", err)
	}
	if !info.IsDir() {
		return ErrChanged
	}
	current := base
	for _, component := range components {
		err = ctx.Err()
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fs.ErrInvalid
		}
		switch component {
		case "", ".":
			continue
		case "..":
			if current == "." {
				return ErrOutsideRoot
			}
			current = path.Dir(current)
		default:
			current = joinName(current, component)
		}
		info, err = fs.Lstat(v.fsys, current)
		resolved.Path, resolved.Info = current, info
		if err != nil {
			return fmt.Errorf("Lstat: %w", err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			linkPath := current
			target, readErr := fs.ReadLink(v.fsys, linkPath)
			resolved.Links = append(resolved.Links, Link{Path: linkPath, Target: target, Info: info})
			if readErr != nil {
				return fmt.Errorf("ReadLink: %w", readErr)
			}
			if active[linkPath] {
				return ErrCycle
			}
			err = validatePointer(target)
			if err != nil {
				return err
			}
			if target == "" {
				return ErrUnsupported
			}
			base = path.Dir(linkPath)
			if strings.HasPrefix(target, "/") {
				target, err = v.relativeAbsoluteTarget(target)
				if err != nil {
					return err
				}
				base = "."
			}
			active[linkPath] = true
			err = v.resolveComponents(ctx, base, strings.Split(target, "/"), active, resolved)
			delete(active, linkPath)
			if err != nil {
				return err
			}
			current, info = resolved.Path, resolved.Info
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return ErrUnsupported
		}
	}
	return nil
}

// relativeAbsoluteTarget deliberately retains dot components. Cleaning first
// would change the meaning of a parent component following a directory alias.
func (v *View) relativeAbsoluteTarget(target string) (string, error) {
	for _, root := range v.absoluteRoots {
		if target == root {
			return ".", nil
		}
		prefix := strings.TrimRight(root, "/") + "/"
		if strings.HasPrefix(target, prefix) {
			return strings.TrimPrefix(target, prefix), nil
		}
	}
	return "", ErrOutsideRoot
}

func validatePointer(target string) error {
	if strings.ContainsAny(target, "\\\x00") || strings.HasPrefix(target, "//") {
		return ErrUnsupported
	}
	if len(target) >= 2 && target[1] == ':' {
		return ErrUnsupported
	}
	return nil
}

// Open returns a resolved regular file. It checks the opened metadata against
// the resolution and closes the opened file if validation fails. Native OS
// metadata permits an identity check; synthetic metadata is checked by type,
// size, and modification time. The caller owns the returned file.
func (v *View) Open(ctx context.Context, logical string) (fs.File, error) {
	resolved, err := v.Resolve(ctx, logical)
	if err != nil {
		return nil, fmt.Errorf("Resolve: %w", err)
	}
	if !resolved.Info.Mode().IsRegular() {
		return nil, &fs.PathError{Op: "Open", Path: logical, Err: ErrUnsupported}
	}
	err = ctx.Err()
	if err != nil {
		return nil, &fs.PathError{Op: "Open", Path: logical, Err: err}
	}
	file, err := v.fsys.Open(resolved.Path)
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		return nil, closeWithError(file, fmt.Errorf("Stat: %w", err))
	}
	err = ctx.Err()
	if err != nil {
		return nil, closeWithError(file, &fs.PathError{Op: "Open", Path: logical, Err: err})
	}
	if !sameTarget(resolved.Info, info) {
		return nil, closeWithError(file, &fs.PathError{Op: "Open", Path: logical, Err: ErrChanged})
	}
	return file, nil
}

func sameTarget(expected, actual fs.FileInfo) bool {
	if actual == nil || !actual.Mode().IsRegular() || expected.Mode() != actual.Mode() {
		return false
	}
	if os.SameFile(expected, expected) && !os.SameFile(expected, actual) {
		return false
	}
	return expected.Size() == actual.Size() && expected.ModTime().Equal(actual.ModTime())
}

func closeWithError(file fs.File, err error) error {
	closeErr := file.Close()
	if closeErr != nil {
		return errors.Join(err, fmt.Errorf("Close: %w", closeErr))
	}
	return err
}

// Walk visits logicalRoot and sorted descendants, following directory aliases
// while preserving logical prefixes. An active physical directory is a cycle,
// but distinct aliases visited in separate branches are expanded independently.
// Resolution and directory-reading failures are handed to fn for classification.
func (v *View) Walk(ctx context.Context, logicalRoot string, fn WalkFunc) error {
	if fn == nil {
		return &fs.PathError{Op: "Walk", Path: logicalRoot, Err: fs.ErrInvalid}
	}
	err := v.walk(ctx, logicalRoot, fn, make(map[string]bool))
	if errors.Is(err, fs.SkipDir) || errors.Is(err, fs.SkipAll) {
		return nil
	}
	return err
}

func (v *View) walk(ctx context.Context, logical string, fn WalkFunc, active map[string]bool) error {
	err := ctx.Err()
	if err != nil {
		return &fs.PathError{Op: "Walk", Path: logical, Err: err}
	}
	resolved, resolveErr := v.Resolve(ctx, logical)
	err = ctx.Err()
	if err != nil {
		return &fs.PathError{Op: "Walk", Path: logical, Err: err}
	}
	if resolveErr == nil && resolved.Info.IsDir() && active[resolved.Path] {
		resolveErr = &fs.PathError{Op: "Walk", Path: logical, Err: ErrCycle}
	}
	err = fn(logical, resolved, resolveErr)
	if errors.Is(err, fs.SkipDir) && resolved.Info != nil && resolved.Info.IsDir() {
		return nil
	}
	if err != nil {
		return err
	}
	if resolveErr != nil || !resolved.Info.IsDir() {
		return nil
	}
	err = ctx.Err()
	if err != nil {
		return &fs.PathError{Op: "Walk", Path: logical, Err: err}
	}
	entries, err := fs.ReadDir(v.fsys, resolved.Path)
	contextErr := ctx.Err()
	if contextErr != nil {
		return &fs.PathError{Op: "Walk", Path: logical, Err: contextErr}
	}
	if err != nil {
		err = fn(logical, resolved, fmt.Errorf("ReadDir: %w", err))
		if errors.Is(err, fs.SkipDir) {
			return nil
		}
		return err
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	active[resolved.Path] = true
	defer delete(active, resolved.Path)
	for _, entry := range entries {
		name := entry.Name()
		if !fs.ValidPath(name) || strings.Contains(name, "/") || name == "." {
			err = fn(joinName(logical, name), Resolution{}, &fs.PathError{Op: "Walk", Path: name, Err: fs.ErrInvalid})
		} else {
			err = v.walk(ctx, joinName(logical, name), fn, active)
		}
		if errors.Is(err, fs.SkipDir) {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func joinName(parent, name string) string {
	if parent == "." {
		return name
	}
	return strings.TrimRight(parent, "/") + "/" + name
}
