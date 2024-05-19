package generate

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"text/template"

	"github.com/easyp-tech/easyp/internal/generate/adapters"
)

const defaultCompiler = "protoc"

// Generate generates files.
func (g *Generator) Generate(ctx context.Context, directory string, storage fs.FS) error {
	q := Query{
		Compiler: defaultCompiler,
		Dir:      ".",
		Imports: []string{
			".",
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

	err := fs.WalkDir(storage, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case d.IsDir():
			return nil
		case filepath.Ext(path) != ".proto":
			return nil
		}

		q.Files = append(q.Files, path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("fs.WalkDir: %w", err)
	}

	tmpl, err := template.New("query").Parse(queryTmpl)
	if err != nil {
		return fmt.Errorf("template.New: %w", err)
	}

	buffer := &bytes.Buffer{}
	err = tmpl.Execute(buffer, q)
	if err != nil {
		return fmt.Errorf("tmpl.Execute: %w", err)
	}

	cmd := buffer.String()

	_, err = adapters.RunCmd(ctx, directory, cmd)
	if err != nil {
		return fmt.Errorf("adapters.RunCmd: %w", err)
	}

	return nil
}
