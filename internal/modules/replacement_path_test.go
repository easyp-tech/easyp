package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestResolveReplacementPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	owner := filepath.Join(root, "project")
	dep := filepath.Join(root, "dependency")
	tests := []struct{ name, target, want string }{
		{name: "absolute", target: dep, want: dep},
		{name: "parent relative", target: "../dependency", want: dep},
		{name: "child relative", target: "local/../local", want: filepath.Join(owner, "local")},
		{name: "same directory", target: ".", want: owner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { t.Parallel(); assert.Equal(t, tt.want, ResolveReplacementPath(owner, tt.target)) })
	}
}

func TestLocalDependencySourcesUseAbsoluteReplacement(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	owner := filepath.Join(root, "app")
	dep := filepath.Join(root, "dependency")
	require.NoError(t, os.MkdirAll(dep, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dep, "protobuf.mod"), []byte("module example.com/dep\n"), 0o644))
	manifest := "module example.com/app\nrequire example.com/dep\nreplace example.com/dep => " + dep + "\n"
	module, err := v1.ParseModule(strings.NewReader(manifest))
	require.NoError(t, err)
	roots, err := localDependencySources(owner, module)
	require.NoError(t, err)
	require.Len(t, roots, 1)
	assert.Equal(t, dep, roots[0].Path)
}
