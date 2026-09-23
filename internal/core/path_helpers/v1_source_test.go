package path_helpers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldSkipV1SourceDir(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{".cache", "buf-module", "legacy-module", "regular"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range map[string]string{
		filepath.Join(root, "buf-module", "buf.yaml"):        "version: v2\n",
		filepath.Join(root, "legacy-module", "protobuf.mod"): "direct (\n)\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name string
		skip bool
	}{
		{".", false},
		{".cache", true},
		{"buf-module", true},
		{"legacy-module", true},
		{"regular", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldSkipV1SourceDir(root, filepath.Join(root, tc.name)); got != tc.skip {
				t.Fatalf("skip = %t, want %t", got, tc.skip)
			}
		})
	}
}
