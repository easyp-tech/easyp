package compiler

import (
	"io"
	"io/fs"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileOpensOnlyReachableDependencyFiles(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		imports string
		value   string
		opened  []string
		fails   bool
	}{
		{
			name:    "primary through a transitive import",
			imports: `import "service.proto";`,
			value:   "service.Response",
			opened:  []string{"request.proto", "service.proto", "google/type/money.proto"},
		},
		{
			name:    "preview independently",
			imports: `import "preview/google/type/money.proto";`,
			value:   "google.type.Money",
			opened:  []string{"request.proto", "preview/google/type/money.proto"},
		},
		{
			name:    "both declarations in one graph",
			imports: `import "service.proto"; import "preview/google/type/money.proto";`,
			value:   "service.Response",
			opened:  []string{"request.proto", "service.proto", "google/type/money.proto", "preview/google/type/money.proto"},
			fails:   true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sources := map[string]string{
				"request.proto":                   `syntax = "proto3"; package app; ` + tt.imports + ` message Request { ` + tt.value + ` value = 1; }`,
				"service.proto":                   `syntax = "proto3"; package service; import "google/type/money.proto"; message Response { google.type.Money value = 1; }`,
				"google/type/money.proto":         `syntax = "proto3"; package google.type; message Money { int64 units = 1; }`,
				"preview/google/type/money.proto": `syntax = "proto3"; package google.type; message Money { string units = 1; }`,
				"preview/broken.proto":            `this unimported file must never be parsed`,
			}
			var mu sync.Mutex
			opened := make(map[string]bool)
			files, err := Compile(t.Context(), []string{"request.proto"}, Options{
				Resolve: func(name string) (io.ReadCloser, error) {
					mu.Lock()
					opened[name] = true
					mu.Unlock()
					source, ok := sources[name]
					if !ok {
						return nil, fs.ErrNotExist
					}
					return io.NopCloser(strings.NewReader(source)), nil
				},
			})
			if tt.fails {
				require.ErrorContains(t, err, `"Money" is already defined`)
				assert.ErrorContains(t, err, "money.proto")
				assert.Nil(t, files)
			} else {
				require.NoError(t, err)
				require.Len(t, files, 1)
				assert.Equal(t, "request.proto", files[0].Path())
			}
			mu.Lock()
			var names []string
			for name := range opened {
				names = append(names, name)
			}
			mu.Unlock()
			assert.ElementsMatch(t, tt.opened, names)
		})
	}
}
