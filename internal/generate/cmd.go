package generate

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/easyp-tech/easyp/internal/generate/adapters"
)

const defaultCompiler = "protoc"

// Generate generates files.
func (g *Generator) Generate(ctx context.Context, directory string) error {
	q := Query{
		Compiler: defaultCompiler,
		Dir:      directory,
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
