package moduleconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestLegacyDepsAndGitInputsAreCombined(t *testing.T) {
	t.Parallel()
	sha := strings.Repeat("a", 40)
	tests := []struct {
		name, yaml string
		want       []v1.Requirement
	}{
		{name: "deps only", yaml: "deps: [example.com/common@v1.2.0]\n", want: []v1.Requirement{{Module: "example.com/common", Version: "v1.2.0"}}},
		{name: "unversioned", yaml: "deps: [example.com/common]\n", want: []v1.Requirement{{Module: "example.com/common"}}},
		{name: "commit pin", yaml: "deps: [example.com/common@" + sha + "]\n", want: []v1.Requirement{{Module: "example.com/common", Version: sha}}},
		{name: "duplicates and distinct minima", yaml: "deps: [example.com/common@v1.2.0, example.com/common@v1.2.0]\ngenerate:\n  inputs:\n    - git_repo: {url: example.com/common@v1.2.0}\n    - git_repo: {url: example.com/common@v1.3.0}\n    - git_repo: {url: example.com/other@v1.0.0}\n", want: []v1.Requirement{{Module: "example.com/common", Version: "v1.2.0"}, {Module: "example.com/common", Version: "v1.3.0"}, {Module: "example.com/other", Version: "v1.0.0"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "easyp.yaml"), []byte(tt.yaml), 0o644))
			module, err := ReadGitDependency(dir, "example.com/dependency")
			require.NoError(t, err)
			assert.Equal(t, tt.want, module.Requires)
			assert.Equal(t, []string{"."}, module.Roots)
		})
	}
}

func TestInvalidLegacyDepsAreNotSilentlyDropped(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, yaml, want string }{
		{name: "empty", yaml: "deps: ['']\n", want: "deps[0]"},
		{name: "null", yaml: "deps: [null]\n", want: "deps[0]"},
		{name: "unsupported branch", yaml: "deps: [example.com/common@main]\n", want: "SemVer tag or full Git commit"},
		{name: "invalid YAML shape", yaml: "deps: {url: example.com/common}\n", want: "Unmarshal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "easyp.yaml"), []byte(tt.yaml), 0o644))
			_, err := ReadGitDependency(dir, "example.com/dependency")
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestNativeManifestTakesPrecedenceOverLegacyDeps(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "protobuf.mod"), []byte("module example.com/dependency\nrequire example.com/current v1.0.0\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "easyp.yaml"), []byte("deps: {invalid: legacy}\n"), 0o644))
	module, err := ReadGitDependency(dir, "example.com/dependency")
	require.NoError(t, err)
	assert.Equal(t, []v1.Requirement{{Module: "example.com/current", Version: "v1.0.0"}}, module.Requires)
}
