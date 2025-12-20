package core

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/wellknownimports"

	"github.com/easyp-tech/easyp/internal/core/models"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

// FileSource represents the origin of a proto file or search root
type FileSource string

const (
	FileSourceWorkspace  FileSource = "workspace"
	FileSourceDependency FileSource = "dependency"
	FileSourceGitRepo    FileSource = "git_repo"
	FileSourceWellKnown  FileSource = "wellknown"

	wellKnownBase = "/wellknownimports"
)

// ListFilesOptions controls behaviour of ListFiles
type ListFilesOptions struct {
	// IncludeImports when true returns transitive imports for discovered .proto files
	IncludeImports bool
}

// FileInfo represents a discovered proto file
type FileInfo struct {
	AbsPath    string     `json:"abs_path"`
	ImportPath string     `json:"import_path"`
	Source     FileSource `json:"source"`
	Root       string     `json:"root"`
}

// RootInfo represents a search root
type RootInfo struct {
	Path   string     `json:"path"`
	Source FileSource `json:"source"`
}

// ErrorInfo contains error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// LsFilesResult is the output for ls-files command
type LsFilesResult struct {
	Files  []FileInfo  `json:"files"`
	Roots  []RootInfo  `json:"roots"`
	Errors []ErrorInfo `json:"errors,omitempty"`
}

type lsContext struct {
	searchRoots []RootInfo
	rootSource  map[string]FileSource
	seenFiles   map[string]struct{}
	fileRoot    map[string]string
	files       []FileInfo
	errors      []ErrorInfo
}

func newLsContext() *lsContext {
	return &lsContext{
		searchRoots: make([]RootInfo, 0),
		rootSource:  make(map[string]FileSource),
		seenFiles:   make(map[string]struct{}),
		fileRoot:    make(map[string]string),
		files:       make([]FileInfo, 0),
		errors:      make([]ErrorInfo, 0),
	}
}

// ListFiles returns .proto files information
func (c *Core) ListFiles(ctx context.Context, workingRoot string, opts ListFilesOptions) (LsFilesResult, error) {
	state := newLsContext()

	c.collectRoots(workingRoot, state)
	c.scanWorkspaceFiles(ctx, workingRoot, state)
	if opts.IncludeImports {
		c.collectImports(ctx, state)
	}

	return LsFilesResult{
		Files:  state.sortedFiles(),
		Roots:  state.sortedRoots(),
		Errors: state.errors,
	}, nil
}

// Helpers reused from previous implementation but cleaned up

func normalizeImportPath(importPath string) string {
	normalized := strings.Trim(importPath, " \t\r\n\"")
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimSuffix(normalized, "/")
	normalized = filepath.ToSlash(filepath.Clean(normalized))
	return normalized
}

func normalizeAbsPath(path string) string {
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		abs, err := filepath.Abs(cleaned)
		if err == nil {
			cleaned = abs
		}
	}
	return filepath.ToSlash(cleaned)
}

// Helpers to keep ListFiles readable

func (c *Core) collectRoots(workingRoot string, state *lsContext) {
	// Well-known imports virtual root
	state.addRoot(wellKnownBase, FileSourceWellKnown)

	for _, inputFilesDir := range c.inputs.InputFilesDir {
		root := inputFilesDir.Root
		if !filepath.IsAbs(root) {
			root = filepath.Join(workingRoot, root)
		}
		state.addRoot(root, FileSourceWorkspace)
	}

	for lockInfo := range c.lockFile.DepsIter() {
		modulePath := c.storage.GetInstallDir(lockInfo.Name, lockInfo.Version)
		state.addRoot(modulePath, FileSourceDependency)
	}

	for _, repo := range c.inputs.InputGitRepos {
		module := models.NewModule(repo.URL)
		modulePath, err := c.modulePath(module)
		if err != nil {
			state.errors = append(state.errors, ErrorInfo{
				Code:    "git_repo_path_error",
				Message: "failed to get path for git repo " + repo.URL + ": " + err.Error(),
			})
			continue
		}

		root := modulePath
		if repo.Root != "" {
			root = filepath.Join(modulePath, repo.Root)
		}
		state.addRoot(root, FileSourceGitRepo)
	}
}

func (c *Core) scanWorkspaceFiles(ctx context.Context, workingRoot string, state *lsContext) {
	for _, inputFilesDir := range c.inputs.InputFilesDir {
		root := inputFilesDir.Root
		if !filepath.IsAbs(root) {
			root = filepath.Join(workingRoot, root)
		}
		root = normalizeAbsPath(root)

		walker := fs.NewFSWalker(root, inputFilesDir.Path)
		walker.WalkDir(func(path string, err error) error {
			if err != nil {
				state.errors = append(state.errors, ErrorInfo{
					Code:    "workspace_walk_error",
					Message: err.Error(),
				})
				return nil
			}
			if ctx.Err() != nil || filepath.Ext(path) != ".proto" {
				return nil
			}
			absPath := filepath.Join(root, path)
			state.addFile(absPath, path, FileSourceWorkspace, root)
			return nil
		})
	}
}

func (c *Core) collectImports(ctx context.Context, state *lsContext) {
	queue := make([]string, 0, len(state.files))
	for _, f := range state.files {
		queue = append(queue, f.AbsPath)
	}

	visited := make(map[string]struct{}, len(queue))
	for _, q := range queue {
		visited[q] = struct{}{}
	}

	head := 0
	for head < len(queue) {
		currentAbs := queue[head]
		head++

		if ctx.Err() != nil {
			return
		}

		proto, ok := c.readProtoForPath(ctx, currentAbs, state)
		if !ok {
			continue
		}

		c.processImports(proto, currentAbs, &queue, state, visited)
	}
}

func (c *Core) processImports(
	proto *unordered.Proto,
	currentAbs string,
	queue *[]string,
	state *lsContext,
	visited map[string]struct{},
) {
	for _, imp := range proto.ProtoBody.Imports {
		impPath := strings.Trim(imp.Location, "\"")

		if abs, src, root, ok := state.resolveImport(currentAbs, impPath); ok {
			if _, seen := visited[abs]; !seen {
				visited[abs] = struct{}{}
				*queue = append(*queue, abs)
				state.addFile(abs, impPath, src, root)
			}
		} else {
			state.errors = append(state.errors, ErrorInfo{
				Code:    "import_not_found",
				Message: "cannot resolve import " + impPath + " from " + currentAbs,
			})
		}
	}
}

func (c *Core) readProtoForPath(ctx context.Context, absPath string, state *lsContext) (*unordered.Proto, bool) {
	source := state.sourceForAbs(absPath)

	if source == FileSourceWellKnown {
		rel, err := filepath.Rel(state.fileRoot[absPath], absPath)
		if err != nil {
			state.errors = append(state.errors, ErrorInfo{
				Code:    "open_error",
				Message: err.Error(),
			})
			return nil, false
		}

		r, err := wellknownimports.Content.Open(rel)
		if err != nil {
			state.errors = append(state.errors, ErrorInfo{
				Code:    "open_error",
				Message: err.Error(),
			})
			return nil, false
		}
		defer r.Close()

		proto, err := readProtoFile(r)
		if err != nil {
			state.errors = append(state.errors, ErrorInfo{
				Code:    "parse_error",
				Message: err.Error(),
			})
			return nil, false
		}
		return proto, true
	}

	f, err := os.Open(absPath)
	if err != nil {
		state.errors = append(state.errors, ErrorInfo{
			Code:    "open_error",
			Message: err.Error(),
		})
		return nil, false
	}
	defer c.close(ctx, f, absPath)

	proto, err := readProtoFile(f)
	if err != nil {
		state.errors = append(state.errors, ErrorInfo{
			Code:    "parse_error",
			Message: err.Error(),
		})
		return nil, false
	}
	return proto, true
}

func (s *lsContext) resolveImport(currentAbs, importPath string) (string, FileSource, string, bool) {
	currentRoot := s.fileRoot[currentAbs]
	if currentRoot != "" {
		candidate := filepath.Join(currentRoot, importPath)
		if _, err := os.Stat(candidate); err == nil {
			candidate = normalizeAbsPath(candidate)
			if source, ok := s.rootSource[currentRoot]; ok {
				return candidate, source, currentRoot, true
			}
			return candidate, "", currentRoot, true
		}
	}

	// Direct well-known lookup before FS roots
	if f, err := wellknownimports.Content.Open(importPath); err == nil {
		_ = f.Close()
		candidate := filepath.Join(wellKnownBase, importPath)
		return normalizeAbsPath(candidate), FileSourceWellKnown, wellKnownBase, true
	}

	for _, r := range s.searchRoots {
		candidate := filepath.Join(r.Path, importPath)
		if _, err := os.Stat(candidate); err == nil {
			candidate = normalizeAbsPath(candidate)
			return candidate, r.Source, r.Path, true
		}
	}

	return "", "", "", false
}

func (s *lsContext) addRoot(path string, source FileSource) {
	normalized := normalizeAbsPath(path)
	if _, ok := s.rootSource[normalized]; ok {
		return
	}
	s.rootSource[normalized] = source
	s.searchRoots = append(s.searchRoots, RootInfo{Path: normalized, Source: source})
}

func (s *lsContext) addFile(absPath, importPath string, source FileSource, root string) {
	absPath = normalizeAbsPath(absPath)
	if _, ok := s.seenFiles[absPath]; ok {
		return
	}
	s.seenFiles[absPath] = struct{}{}
	root = normalizeAbsPath(root)
	s.fileRoot[absPath] = root
	s.files = append(s.files, FileInfo{
		AbsPath:    absPath,
		ImportPath: normalizeImportPath(importPath),
		Source:     source,
		Root:       root,
	})
}

func (s *lsContext) sortedRoots() []RootInfo {
	roots := make([]RootInfo, len(s.searchRoots))
	copy(roots, s.searchRoots)
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Path < roots[j].Path
	})
	return roots
}

func (s *lsContext) sortedFiles() []FileInfo {
	files := make([]FileInfo, len(s.files))
	copy(files, s.files)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ImportPath < files[j].ImportPath
	})
	return files
}

func (s *lsContext) sourceForAbs(absPath string) FileSource {
	absPath = normalizeAbsPath(absPath)
	root := s.fileRoot[absPath]
	if src, ok := s.rootSource[root]; ok {
		return src
	}
	return ""
}
