package generate

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/generate/adapters"
)

const defaultCompiler = "protoc"

// Generate generates files.
func (g *Generator) Generate(ctx context.Context, root, directory string) error {
	q := Query{
		Compiler: defaultCompiler,
		Imports: []string{
			root,
		},
		Plugins: g.plugins,
	}

	for _, dep := range g.deps {
		modulePaths, err := g.moduleReflect.GetModulePath(ctx, dep)
		if err != nil {
			return fmt.Errorf("g.moduleReflect.GetModulePath: %w", err)
		}

		q.Imports = append(q.Imports, modulePaths)
	}

	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case d.IsDir():
			return nil
		case filepath.Ext(path) != ".proto":
			return nil
		case shouldIgnore(path, g.inputs.Dirs):
			return nil
		}

		q.Files = append(q.Files, path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("filepath.WalkDir: %w", err)
	}

	_, err = adapters.RunCmd(ctx, root, q.build())
	if err != nil {
		return fmt.Errorf("adapters.RunCmd: %w", err)
	}

	return nil
}

func shouldIgnore(path string, dirs []string) bool {
	if len(dirs) == 0 {
		return false
	}

	for _, dir := range dirs {
		if strings.HasPrefix(path, dir) {
			return false
		}
	}

	return true
}
