package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/wellknownimports"
)

const wellKnownV1Root = "/wellknownimports"

// LsFiles lists module sources and reachable transitive imports.
type LsFiles struct{}

var _ Handler = (*LsFiles)(nil)

var flagLsFilesIncludeImports = &cli.BoolFlag{
	Name:    "include-imports",
	Aliases: []string{"I"},
	Usage:   "include reachable imports from dependencies and well-known protos",
	Value:   true,
}

type v1ListedFile struct {
	AbsPath    string `json:"abs_path"`
	ImportPath string `json:"import_path"`
	Source     string `json:"source"`
	Root       string `json:"root"`
}

type v1ListedRoot struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}

type v1ListError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type v1ListResult struct {
	Files  []v1ListedFile `json:"files"`
	Roots  []v1ListedRoot `json:"roots"`
	Errors []v1ListError  `json:"errors,omitempty"`
}

// Command implements Handler.
func (l LsFiles) Command() *cli.Command {
	return &cli.Command{
		Name:    "ls-files",
		Aliases: []string{"ls"},
		Usage:   "list v1 module proto files and reachable imports",
		Action:  l.Action,
		Flags:   []cli.Flag{flagLsFilesIncludeImports, flags.Format},
	}
}

// Action implements Handler.
func (l LsFiles) Action(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	_, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
	}
	listed, err := listV1Files(ctx, root, module)
	if err != nil {
		return fmt.Errorf("listV1Files: %w", err)
	}
	format := flags.GetFormat(ctx, flags.JSONFormat)
	switch format {
	case flags.JSONFormat:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(listed)
	case flags.TextFormat:
		return printV1ListedFiles(listed)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func listV1Files(ctx *cli.Context, moduleDir string, module v1.Module) (v1ListResult, error) {
	result := v1ListResult{Files: []v1ListedFile{}, Roots: []v1ListedRoot{}}
	index := make(map[string]v1ListedFile)
	for _, root := range module.Roots {
		path := filepath.Join(moduleDir, root)
		result.Roots = append(result.Roots, v1ListedRoot{Path: filepath.ToSlash(path), Source: "workspace"})
		if err := indexV1ProtoRoot(path, "workspace", index, &result.Files); err != nil {
			return v1ListResult{}, fmt.Errorf("indexV1ProtoRoot: %w", err)
		}
	}
	if ctx.Bool(flagLsFilesIncludeImports.Name) {
		dependencies, err := resolveV1DependencySources(ctx.Context, getLogger(ctx), moduleDir, module)
		if err != nil {
			return v1ListResult{}, fmt.Errorf("resolveV1DependencySources: %w", err)
		}
		if err := checkV1ImportPathCollisions(moduleDir, module.Roots, dependencies.paths()); err != nil {
			return v1ListResult{}, fmt.Errorf("checkV1ImportPathCollisions: %w", err)
		}
		for _, dependency := range dependencies {
			result.Roots = append(result.Roots, v1ListedRoot{Path: filepath.ToSlash(dependency.path), Source: "dependency"})
			if err := indexV1ProtoRoot(dependency.path, "dependency", index, nil); err != nil {
				return v1ListResult{}, fmt.Errorf("indexV1ProtoRoot: %w", err)
			}
		}
		collectV1ListedImports(index, &result)
	}
	slices.SortFunc(result.Roots, func(a, b v1ListedRoot) int { return strings.Compare(a.Path, b.Path) })
	slices.SortFunc(result.Files, func(a, b v1ListedFile) int { return strings.Compare(a.ImportPath, b.ImportPath) })
	return result, nil
}

func indexV1ProtoRoot(root, source string, index map[string]v1ListedFile, selected *[]v1ListedFile) error {
	return walkV1ProtoFiles(root, func(path string) error {
		importPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		importPath = filepath.ToSlash(importPath)
		file := v1ListedFile{AbsPath: filepath.ToSlash(path), ImportPath: importPath, Source: source, Root: filepath.ToSlash(root)}
		if previous, exists := index[importPath]; exists && previous.AbsPath != file.AbsPath {
			return fmt.Errorf("duplicate import path %q: %s and %s", importPath, previous.AbsPath, file.AbsPath)
		}
		index[importPath] = file
		if selected != nil {
			*selected = append(*selected, file)
		}
		return nil
	})
}

func collectV1ListedImports(index map[string]v1ListedFile, result *v1ListResult) {
	queue := append([]v1ListedFile(nil), result.Files...)
	seen := make(map[string]bool, len(queue))
	for _, file := range queue {
		seen[file.ImportPath] = true
	}
	for len(queue) > 0 {
		file := queue[0]
		queue = queue[1:]
		imports, err := readV1ListedImports(file)
		if err != nil {
			result.Errors = append(result.Errors, v1ListError{Code: "parse_error", Message: fmt.Sprintf("%s: %v", file.ImportPath, err)})
			continue
		}
		for _, importPath := range imports {
			if seen[importPath] {
				continue
			}
			resolved, ok := index[importPath]
			if !ok {
				if !filepath.IsLocal(importPath) {
					result.Errors = append(result.Errors, v1ListError{Code: "invalid_import", Message: fmt.Sprintf("%s imports %q", file.ImportPath, importPath)})
					continue
				}
				_, err := wellknownimports.Content.ReadFile(importPath)
				if errors.Is(err, os.ErrNotExist) {
					result.Errors = append(result.Errors, v1ListError{Code: "import_not_found", Message: fmt.Sprintf("%s imports %q", file.ImportPath, importPath)})
					continue
				}
				if err != nil {
					result.Errors = append(result.Errors, v1ListError{Code: "open_error", Message: fmt.Sprintf("%s: %v", importPath, err)})
					continue
				}
				resolved = v1ListedFile{AbsPath: wellKnownV1Root + "/" + importPath, ImportPath: importPath, Source: "wellknown", Root: wellKnownV1Root}
				index[importPath] = resolved
			}
			seen[importPath] = true
			result.Files = append(result.Files, resolved)
			queue = append(queue, resolved)
		}
	}
}

func readV1ListedImports(file v1ListedFile) ([]string, error) {
	if file.Source != "wellknown" {
		return readV1ProtoImports(filepath.FromSlash(file.AbsPath))
	}
	raw, err := wellknownimports.Content.ReadFile(file.ImportPath)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}
	return parseV1ProtoImports(file.ImportPath, raw)
}

func printV1ListedFiles(result v1ListResult) error {
	for _, file := range result.Files {
		if _, err := fmt.Fprintf(os.Stdout, "%s\t%s\t%s\n", file.ImportPath, file.Source, file.AbsPath); err != nil {
			return fmt.Errorf("Fprintf: %w", err)
		}
	}
	for _, issue := range result.Errors {
		if _, err := fmt.Fprintf(os.Stderr, "%s: %s\n", issue.Code, issue.Message); err != nil {
			return fmt.Errorf("Fprintf: %w", err)
		}
	}
	return nil
}
