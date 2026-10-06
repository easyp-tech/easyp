package migration

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationLocalSelectionUsesModuleRelativePaths(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		inputs string
		files  map[string]string
		paths  []string
	}{
		{name: "service_import_root", inputs: "[{directory: {root: api, path: easyp}}]", files: map[string]string{"api/easyp/generator.proto": "package easyp.generator.v1;"}, paths: []string{"api/easyp"}},
		{name: "mixed_roots", inputs: "[{directory: {root: proto, path: .}}, {directory: {root: schema, path: selected}}]", files: map[string]string{"proto/first.proto": "package first.v1;", "schema/selected/second.proto": "package second.v1;", "schema/other.proto": "package other.v1;"}, paths: []string{"proto", "schema/selected"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: "+tt.inputs+"\n")
			for name, content := range tt.files {
				writeFixture(t, root, name, content)
			}

			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/service"})
			require.NoError(t, err)
			gen, err := v1.ParseGenerate(bytes.NewReader(outputContent(t, plan, v1.GenerateFile)))
			require.NoError(t, err)
			assert.Empty(t, gen.Generate.Modules, "the generator must infer its sole local module")
			assert.Empty(t, gen.Generate.Packages, "literal physical paths must preserve mixed-root input scope")
			assert.Equal(t, tt.paths, gen.Generate.Paths)
			assert.Nil(t, gen.Options.Go.PackagePrefix)
		})
	}
}
