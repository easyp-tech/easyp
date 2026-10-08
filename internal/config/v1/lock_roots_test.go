package v1

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLockRootsParsing(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		roots string
		valid bool
	}{
		{name: "omitted", valid: true},
		{name: "empty", roots: "[]", valid: true},
		{name: "repository default", roots: "[.]", valid: true},
		{name: "nested", roots: "[api, third_party/google]", valid: true},
		{name: "duplicate", roots: "[api, api]"},
		{name: "duplicate default", roots: "[., .]"},
		{name: "null list", roots: "null"},
		{name: "null path", roots: "[null]"},
		{name: "numeric path", roots: "[123]"},
		{name: "mapping", roots: "{api: true}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := lockedRootsYAML(tt.roots)
			lock, err := ParseLock(strings.NewReader(raw))
			if !tt.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			serialized, err := yaml.Marshal(lock)
			require.NoError(t, err)
			if tt.roots == "" || tt.roots == "[]" {
				assert.NotContains(t, string(serialized), "roots:")
			} else {
				assert.Contains(t, string(serialized), "roots:")
			}
		})
	}
}

func TestLockRootsRejectNonPortableDirectories(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"", "..", "../api", "api/../other", "./api", "/api", "C:/api", `api\v1`, "api/*", "api?", "api[ab]", "api//v1", "api/", "api/.", "api path", "api\npath"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseLock(strings.NewReader(lockedRootsYAML("[" + strconv.Quote(name) + "]")))
			require.Error(t, err)
		})
	}
}

func TestLockRootsValidateMergedYAMLTypes(t *testing.T) {
	t.Parallel()
	raw := strings.Replace(lockedRootsYAML("[]"), "roots: []", "<<: {roots: [123]}", 1)
	_, err := ParseLock(strings.NewReader(raw))
	require.Error(t, err)
}

func TestLockRootsSchema(t *testing.T) {
	t.Parallel()
	raw, err := SchemaJSON("protobuf.lock")
	require.NoError(t, err)
	var document map[string]any
	require.NoError(t, json.Unmarshal(raw, &document))
	properties := document["properties"].(map[string]any)
	entry := properties["modules"].(map[string]any)["items"].(map[string]any)
	roots, ok := entry["properties"].(map[string]any)["roots"].(map[string]any)
	require.True(t, ok, "the lock schema must describe optional roots")
	assert.Equal(t, "array", roots["type"])
	assert.Equal(t, true, roots["uniqueItems"])
	assert.Equal(t, PathSelectorPattern, roots["items"].(map[string]any)["pattern"])
	assert.NotContains(t, entry["required"], "roots")
}

func lockedRootsYAML(roots string) string {
	raw := "version: 1\nmodules:\n  - source: example.test/module\n    version: v1.0.0\n    commit: " + strings.Repeat("a", 40) + "\n    hash: h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n"
	if roots != "" {
		raw += "    roots: " + roots + "\n"
	}
	return raw
}
