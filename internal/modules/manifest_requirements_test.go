package modules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestAugmentManifestPromotesExistingIndirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		original string
		want     string
	}{
		{
			name:     "standalone",
			original: "module example.com/app // keep\n\nrequire\texample.com/dep\t v1.0.0 // indirect\n",
			want:     "module example.com/app // keep\n\nrequire\texample.com/dep\t v1.0.0\n",
		},
		{
			name:     "block_with_other_comments",
			original: "module example.com/app\nrequire ( // dependencies\n\t// keep above\n\texample.com/dep\t v1.0.0  // indirect needed by clients\n) // keep below\n",
			want:     "module example.com/app\nrequire ( // dependencies\n\t// keep above\n\texample.com/dep\t v1.0.0  // needed by clients\n) // keep below\n",
		},
		{
			name:     "versionless",
			original: "module example.com/app\nrequire example.com/dep // indirect keep HEAD\n",
			want:     "module example.com/app\nrequire example.com/dep // keep HEAD\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, dependency := t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, root, "root.proto", `syntax = "proto3"; import "dep.proto";`)
			writeV1GenerateFixture(t, dependency, "proto/dep.proto", `syntax = "proto3";`)
			module, err := v1.ParseModule(strings.NewReader(tt.original))
			require.NoError(t, err)
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{{Source: "example.com/dep", Version: "v1.2.0", Commit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}}
			before := v1.Lock{Version: lock.Version, Modules: append([]v1.LockedModule(nil), lock.Modules...)}
			repository := &fakeRepository{directory: dependency, module: v1.Module{Name: "example.com/dep", Roots: []string{"proto"}}}
			original := []byte(tt.original)

			updated, err := augmentV1ManifestRequirements(original, root, module, lock, repository)

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(updated))
			assert.Equal(t, tt.original, string(original))
			assert.Equal(t, before, lock)
			assert.False(t, repository.installed)
		})
	}
}
