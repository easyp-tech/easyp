package sourceview

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestSessionSharesRootsAndKeepsChecks(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"one.proto", "two.proto"} {
		if err := os.WriteFile(filepath.Join(root, "a", "b", name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.proto"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.proto"), filepath.Join(root, "a", "b", "escape.proto")); err != nil {
		t.Fatal(err)
	}

	session := NewSession()
	defer session.Close()
	for _, name := range []string{"one.proto", "two.proto"} {
		file, err := session.OpenLocal(t.Context(), root, "a/b/"+name)
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		data, err := io.ReadAll(file)
		if err != nil || string(data) != name {
			t.Fatalf("read %s: %q %v", name, data, err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if len(session.roots) != 1 {
		t.Fatalf("roots = %d, want one shared root", len(session.roots))
	}
	if _, err := session.OpenLocal(t.Context(), root, "a/b/escape.proto"); !errors.Is(err, ErrOutsideRoot) {
		t.Fatalf("escape through the session = %v, want ErrOutsideRoot", err)
	}
	if _, err := session.OpenLocal(t.Context(), root, "a/b/missing.proto"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file = %v, want not exist", err)
	}
}
