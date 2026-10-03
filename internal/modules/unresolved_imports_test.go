package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnresolvedImports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		source     string
		dependency string
		missing    []string
	}{
		{name: "local import", source: `syntax = "proto3"; import "local.proto";`, missing: []string{}},
		{name: "external import", source: `syntax = "proto3"; import "local.proto"; import "external.proto";`, missing: []string{"external.proto"}},
		{name: "dependency root", source: `syntax = "proto3"; import "external.proto";`, dependency: "external.proto", missing: []string{}},
		{name: "well-known import", source: `syntax = "proto3"; import "google/protobuf/timestamp.proto";`, missing: []string{}},
		{name: "sorted unique missing imports", source: `syntax = "proto3"; import "z.proto"; import "a.proto"; import "z.proto";`, missing: []string{"a.proto", "z.proto"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "main.proto", tt.source)
			writeV1GenerateFixture(t, root, "local.proto", `syntax = "proto3";`)
			var dependencies []string
			if tt.dependency != "" {
				dependency := t.TempDir()
				writeV1GenerateFixture(t, dependency, tt.dependency, `syntax = "proto3";`)
				dependencies = []string{dependency}
			}

			missing, err := unresolvedV1ImportsWithRoots(root, []string{"."}, dependencies)

			require.NoError(t, err)
			assert.Equal(t, tt.missing, missing)
		})
	}
}
