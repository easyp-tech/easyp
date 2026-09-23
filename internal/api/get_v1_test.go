package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestGetAddsDirectAndTransitiveRequirements(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "contracts")
	foo := filepath.Join(repository, "foo")
	bar := filepath.Join(repository, "bar")
	baz := filepath.Join(repository, "baz")
	for name, body := range map[string]string{
		"foo/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\nrequire %s\n", foo, bar),
		"foo/proto/foo.proto": "syntax = \"proto3\";\npackage foo.v1;\nimport \"bar.proto\";\nmessage Foo { bar.v1.Bar bar = 1; }\n",
		"bar/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\nrequire %s\n", bar, baz),
		"bar/proto/bar.proto": "syntax = \"proto3\";\npackage bar.v1;\nimport \"baz.proto\";\nmessage Bar { baz.v1.Baz baz = 1; }\n",
		"baz/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", baz),
		"baz/proto/baz.proto": "syntax = \"proto3\";\npackage baz.v1;\nmessage Baz {}\n",
	} {
		path := filepath.Join(repository, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "proto"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte("module example.com/app\nroots proto\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "proto", "app.proto"), []byte("syntax = \"proto3\";\nimport \"foo.proto\";\n"), 0o644))
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	t.Chdir(root)
	app := &cli.App{Commands: []*cli.Command{(Get{}).Command(), (Mod{}).Command()}, Metadata: map[string]any{}}
	args := []string{"easyp", "get", foo}
	require.NoError(t, app.Run(args))

	manifest, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Contains(t, string(manifest), "require "+foo+"\n")
	require.NotContains(t, string(manifest), "require "+foo+" // indirect")
	lock, err := readV1Lock(filepath.Join(root, "protobuf.lock"))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 3)
	for _, entry := range lock.Modules {
		require.True(t, entry.Source == foo || entry.Source == bar || entry.Source == baz)
		require.Equal(t, entry.Commit, entry.Version)
		require.True(t, strings.HasPrefix(entry.Hash, "h1:"))
		if entry.Source == bar || entry.Source == baz {
			require.Contains(t, string(manifest), "require "+entry.Source+" "+entry.Version+" // indirect\n")
		}
	}

	require.NoError(t, app.Run(args))
	again, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Equal(t, string(manifest), string(again))
	require.NoError(t, app.Run([]string{"easyp", "mod", "tidy"}))
	afterTidy, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Equal(t, string(manifest), string(afterTidy))
}

func TestAddDirectV1RequirementPromotesIndirectInBlock(t *testing.T) {
	t.Parallel()
	original := []byte("module example.com/app\nrequire ( // dependencies\n  example.com/dep v1.0.0 // indirect\n  example.com/other v1.0.0 // indirect\n)\n")
	updated, err := addDirectV1Requirement(original, v1.Requirement{Module: "example.com/dep", Version: "v1.1.0"})
	require.NoError(t, err)
	require.Equal(t, "module example.com/app\nrequire ( // dependencies\n  example.com/dep v1.1.0\n  example.com/other v1.0.0 // indirect\n)\n", string(updated))
}

func TestAddDirectV1RequirementPreservesFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original string
		version  string
		want     string
	}{
		{
			name:     "update preserves spacing and comment",
			original: "require\thttps://example.com/dep\t v1.0.0  // keep this\n",
			version:  "v1.1.0",
			want:     "require\thttps://example.com/dep\t v1.1.0  // keep this\n",
		},
		{
			name:     "versionless stays versionless",
			original: "require (\n\thttps://example.com/dep\t// keep this\n)\n",
			want:     "require (\n\thttps://example.com/dep\t// keep this\n)\n",
		},
		{
			name:     "pin versionless and retain comment",
			original: "require https://example.com/dep  // indirect needed by clients\n",
			version:  "v1.1.0",
			want:     "require https://example.com/dep v1.1.0  // needed by clients\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			updated, err := addDirectV1Requirement([]byte(tc.original), v1.Requirement{Module: "https://example.com/dep", Version: tc.version})
			require.NoError(t, err)
			require.Equal(t, tc.want, string(updated))
		})
	}
}

func TestParseV1GetRequirement(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		input   string
		module  string
		version string
		wantErr bool
	}{
		{name: "nested URL", input: "https://github.com/acme/repo.git/foo", module: "https://github.com/acme/repo.git/foo"},
		{name: "tag", input: "github.com/acme/repo/foo@v1.2.3", module: "github.com/acme/repo/foo", version: "v1.2.3"},
		{name: "commit", input: "github.com/acme/repo@" + strings.Repeat("a", 40), module: "github.com/acme/repo", version: strings.Repeat("a", 40)},
		{name: "invalid version", input: "github.com/acme/repo@latest", wantErr: true},
		{name: "invalid module", input: "github.com/acme/../repo", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseV1GetRequirement(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.module, got.Module)
			require.Equal(t, tc.version, got.Version)
		})
	}
}

func TestGetPromotesIndirectAndPinsExplicitCommit(t *testing.T) {
	dependency := filepath.Join(t.TempDir(), "dependency")
	require.NoError(t, os.MkdirAll(dependency, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dependency, "protobuf.mod"), []byte("module "+dependency+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dependency, "dep.proto"), []byte("syntax = \"proto3\";\nmessage Dep {}\n"), 0o644))
	runTestGit(t, dependency, "init", "-q")
	runTestGit(t, dependency, "add", ".")
	runTestGit(t, dependency, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	first := strings.TrimSpace(runTestGit(t, dependency, "rev-parse", "HEAD"))
	require.NoError(t, os.WriteFile(filepath.Join(dependency, "dep.proto"), []byte("syntax = \"proto3\";\nmessage Dep { string id = 1; }\n"), 0o644))
	runTestGit(t, dependency, "add", ".")
	runTestGit(t, dependency, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "update")
	latest := strings.TrimSpace(runTestGit(t, dependency, "rev-parse", "HEAD"))

	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/app\nrequire (\n  %s %s // indirect\n)\n", dependency, first)
	require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644))
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	t.Chdir(root)
	app := &cli.App{Commands: []*cli.Command{(Get{}).Command()}, Metadata: map[string]any{}}
	require.NoError(t, app.Run([]string{"easyp", "get", dependency + "@" + latest}))
	updated, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Contains(t, string(updated), dependency+" "+latest+"\n")
	require.NotContains(t, string(updated), "// indirect")
	lock, err := readV1Lock(filepath.Join(root, "protobuf.lock"))
	require.NoError(t, err)
	require.Len(t, lock.Modules, 1)
	require.Equal(t, latest, lock.Modules[0].Commit)
}

func TestGetRejectsBadArgumentWithoutChangingManifest(t *testing.T) {
	root := t.TempDir()
	manifest := []byte("module example.com/app\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), manifest, 0o644))
	t.Chdir(root)
	app := &cli.App{Commands: []*cli.Command{(Get{}).Command()}, Metadata: map[string]any{}}
	require.Error(t, app.Run([]string{"easyp", "get", "github.com/acme/repo@latest"}))
	current, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Equal(t, manifest, current)
	_, err = os.Stat(filepath.Join(root, "protobuf.lock"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestGetResolutionFailureLeavesManifestAndLockUnchanged(t *testing.T) {
	dependency := filepath.Join(t.TempDir(), "dependency")
	require.NoError(t, os.MkdirAll(dependency, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dependency, "protobuf.mod"), []byte("module "+dependency+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dependency, "dep.proto"), []byte("syntax = \"proto3\";\nmessage Dep {}\n"), 0o644))
	runTestGit(t, dependency, "init", "-q")
	runTestGit(t, dependency, "add", ".")
	runTestGit(t, dependency, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")

	root := t.TempDir()
	manifest := []byte("module example.com/app\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), manifest, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "app.proto"), []byte("syntax = \"proto3\";\nimport \"missing.proto\";\n"), 0o644))
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	t.Chdir(root)
	app := &cli.App{Commands: []*cli.Command{(Get{}).Command()}, Metadata: map[string]any{}}
	require.ErrorContains(t, app.Run([]string{"easyp", "get", dependency}), "cannot resolve imports")
	current, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, err)
	require.Equal(t, manifest, current)
	_, err = os.Stat(filepath.Join(root, "protobuf.lock"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
