package api

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

func TestGenerateExplicitConfigPreservesSourceAndOutputPaths(t *testing.T) {
	tests := []struct {
		name             string
		workDir          string
		configDir        string
		absolute         bool
		implicitModule   bool
		frozen           bool
		malformedSibling bool
		alias            bool
	}{
		{name: "relative file"},
		{name: "absolute file", absolute: true},
		{name: "file above invocation directory", workDir: "invocation"},
		{name: "file in another directory", configDir: "profiles", implicitModule: true},
		{name: "frozen file", frozen: true},
		{name: "malformed standard sibling is ignored", malformedSibling: true},
		{name: "internal file alias keeps logical output directory", alias: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.alias && runtime.GOOS == "windows" {
				t.Skip("creating symlinks requires Windows privileges")
			}
			root := explicitGenerateFixture(t)
			workDir := filepath.Join(root, tt.workDir)
			require.NoError(t, os.MkdirAll(workDir, 0o755))
			t.Chdir(workDir)
			t.Setenv(envEasypPath, t.TempDir())

			configDir := filepath.Join(root, tt.configDir)
			profile := filepath.Join(configDir, "private.easyp.gen.yaml")
			content := "version: v1\ngenerate:\n  modules:\n    - module: example.test/app\n      paths: [proto/item.proto]\nplugins:\n  - name: python\n    out: sdk/private\n"
			if tt.implicitModule {
				content = "version: v1\nplugins:\n  - name: python\n    out: sdk/private\n"
			}
			if tt.alias {
				target := filepath.Join(root, "profiles", "source.yaml")
				writeV1GenerateFixture(t, root, "profiles/source.yaml", content)
				require.NoError(t, os.Symlink(target, profile))
			} else {
				writeV1GenerateFixture(t, configDir, filepath.Base(profile), content)
			}
			standard := "version: v1\nplugins:\n  - name: python\n    out: sdk/default\n"
			if tt.malformedSibling {
				standard = "version: v1\nplugins: [\n"
			}
			writeV1GenerateFixture(t, configDir, v1.GenerateFile, standard)
			argument := profile
			if !tt.absolute {
				var err error
				argument, err = filepath.Rel(workDir, profile)
				require.NoError(t, err)
			}
			args := []string{"easyp", "generate", "--gen-config", argument}
			if tt.frozen {
				args = append(args, "--frozen")
			}

			err := explicitGenerateApp().RunContext(t.Context(), args)

			require.NoError(t, err)
			generated, err := os.ReadFile(filepath.Join(configDir, "sdk", "private", "item_pb2.py"))
			require.NoError(t, err)
			require.Contains(t, string(generated), "Item")
			require.Contains(t, string(generated), "item.proto")
			require.NoDirExists(t, filepath.Join(configDir, "sdk", "default"))
			if tt.workDir != "" {
				require.NoDirExists(t, filepath.Join(workDir, "sdk"))
			}
			if tt.alias {
				require.NoDirExists(t, filepath.Join(root, "profiles", "sdk"))
			}
			lock, err := os.ReadFile(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			require.Equal(t, "version: 1\nmodules: []\n", string(lock))
		})
	}
}

func TestGenerateExplicitConfigPreservesReadErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		missing bool
		folder  bool
		wantErr error
		message string
	}{
		{name: "missing file", missing: true, wantErr: os.ErrNotExist},
		{name: "directory", folder: true, wantErr: sourceview.ErrUnsupported},
		{name: "malformed YAML", content: "version: v1\nplugins: [\n", message: "yaml"},
		{name: "legacy config", content: "generate:\n  inputs: []\n", message: "field inputs not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := explicitGenerateFixture(t)
			t.Chdir(root)
			t.Setenv(envEasypPath, t.TempDir())
			writeV1GenerateFixture(t, root, v1.GenerateFile, "version: v1\nplugins:\n  - name: python\n    out: sdk/default\n")
			if tt.folder {
				require.NoError(t, os.Mkdir(filepath.Join(root, "private.easyp.gen.yaml"), 0o755))
			} else if !tt.missing {
				writeV1GenerateFixture(t, root, "private.easyp.gen.yaml", tt.content)
			}

			err := explicitGenerateApp().RunContext(t.Context(), []string{"easyp", "generate", "--gen-config", "private.easyp.gen.yaml"})

			require.Error(t, err)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.ErrorContains(t, err, tt.message)
			}
			if tt.name == "legacy config" {
				var typeError *yaml.TypeError
				require.ErrorAs(t, err, &typeError)
			}
			require.NoDirExists(t, filepath.Join(root, "sdk"))
		})
	}
}

func TestGenerateExplicitConfigRejectsSelectorConflicts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "empty file", args: []string{"--gen-config", ""}, want: "--gen-config must not be empty"},
		{name: "with all", args: []string{"--gen-config", "private.yaml", "--all"}, want: "--gen-config"},
		{name: "with project", args: []string{"--gen-config", "private.yaml", "--project", "."}, want: "--gen-config"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			args := append([]string{"easyp", "generate"}, tt.args...)
			err := explicitGenerateApp().RunContext(t.Context(), args)
			require.ErrorContains(t, err, tt.want)
			require.NotContains(t, err.Error(), "flag provided but not defined")
		})
	}
}

func TestGenerateExplicitConfigRejectsWorkspaceEscape(t *testing.T) {
	tests := []struct {
		name  string
		alias bool
	}{
		{name: "parent path"},
		{name: "metadata alias", alias: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.alias && runtime.GOOS == "windows" {
				t.Skip("creating symlinks requires Windows privileges")
			}
			root := explicitGenerateFixture(t)
			workspace := filepath.Join(root, "limited")
			require.NoError(t, os.Mkdir(workspace, 0o755))
			t.Chdir(workspace)
			t.Setenv(envEasypPath, t.TempDir())
			outside := filepath.Join(root, "private.easyp.gen.yaml")
			writeV1GenerateFixture(t, root, filepath.Base(outside), "version: v1\nplugins:\n  - name: python\n    out: sdk/private\n")
			argument := "../private.easyp.gen.yaml"
			if tt.alias {
				argument = "alias.easyp.gen.yaml"
				require.NoError(t, os.Symlink(outside, filepath.Join(workspace, argument)))
			}

			err := explicitGenerateApp().RunContext(t.Context(), []string{"easyp", "generate", "--workspace", ".", "--gen-config", argument})

			require.ErrorIs(t, err, sourceview.ErrOutsideRoot)
			require.NoDirExists(t, filepath.Join(root, "sdk"))
		})
	}
}

func TestGenerateExplicitConfigDoesNotDiscoverUnrelatedChildren(t *testing.T) {
	root := explicitGenerateFixture(t)
	t.Chdir(root)
	t.Setenv(envEasypPath, t.TempDir())
	writeV1GenerateFixture(t, root, "private.easyp.gen.yaml", "version: v1\nplugins: []\n")
	writeV1GenerateFixture(t, root, "child/easyp.gen.yaml", "version: v1\nplugins:\n  - name: python\n    out: sdk/child\n")

	err := explicitGenerateApp().RunContext(t.Context(), []string{"easyp", "generate", "--gen-config", "private.easyp.gen.yaml"})

	require.NoError(t, err)
	require.NoDirExists(t, filepath.Join(root, "child", "sdk"))
}

func TestGenerateExplicitConfigInheritsOnlyAncestorOptions(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    string
	}{
		{name: "standard sibling is independent", profile: "private.easyp.gen.yaml", want: "example.test/original;privatev1"},
		{name: "standard ancestor is inherited", profile: "profiles/private.easyp.gen.yaml", want: "example.test/shared/private/v1;privatev1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := explicitGenerateFixture(t)
			t.Chdir(root)
			t.Setenv(envEasypPath, t.TempDir())
			writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package private.v1; option go_package = \"example.test/original;privatev1\"; message Item {}\n")
			writeV1GenerateFixture(t, root, v1.GenerateFile, "version: v1\nplugins: []\noptions:\n  go:\n    package_prefix: example.test/shared\n")
			writeV1GenerateFixture(t, root, tt.profile, "version: v1\nplugins: []\n")

			err := explicitGenerateApp().RunContext(t.Context(), []string{"easyp", "generate", "--gen-config", tt.profile, "--descriptor_set_out", "private.pb"})

			require.NoError(t, err)
			raw, err := os.ReadFile(filepath.Join(root, "private.pb"))
			require.NoError(t, err)
			var descriptors descriptorpb.FileDescriptorSet
			require.NoError(t, proto.Unmarshal(raw, &descriptors))
			require.Len(t, descriptors.File, 1)
			require.Equal(t, "item.proto", descriptors.File[0].GetName())
			require.Equal(t, tt.want, descriptors.File[0].GetOptions().GetGoPackage())
		})
	}
}

func explicitGenerateFixture(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o755))
	writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/app\nroots proto\n")
	writeV1GenerateFixture(t, root, v1.LockFile, "version: 1\nmodules: []\n")
	writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package private.v1; message Item { string id = 1; }\n")
	return root
}

func explicitGenerateApp() *cli.App {
	command := (Generate{}).Command()
	// urfave's automatic HelpFlag is global mutable state across apps.
	command.HideHelp = true
	return &cli.App{Commands: []*cli.Command{command}, HideHelp: true, Writer: io.Discard, ErrWriter: io.Discard}
}
