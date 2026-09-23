package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnresolvedV1Imports(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{
		"main.proto":  "syntax = \"proto3\";\nimport \"local.proto\";\nimport \"external/v1/types.proto\";\n",
		"local.proto": "syntax = \"proto3\";\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	missing, err := unresolvedV1Imports(root, []string{"."})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 1 || missing[0] != "external/v1/types.proto" {
		t.Fatalf("missing = %v", missing)
	}
}
