package generation

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func runTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}
func writeTestLock(root string, lock v1.Lock) error {
	raw, err := yaml.Marshal(lock)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, v1.LockFile), raw, 0o600)
}
