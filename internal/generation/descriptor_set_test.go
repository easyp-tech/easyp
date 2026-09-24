package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestRunDescriptorSetSkipsOptionsOnlyParentAcrossModules(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	files := map[string]string{
		"easyp.gen.yaml":            "version: v1\noptions:\n  go:\n    package_prefix: example.com/gen\n",
		"backend/easyp.gen.yaml":    "version: v1\ngenerate:\n  modules: [proto/orders, proto/users]\n",
		"proto/orders/protobuf.mod": "module example.com/orders\n",
		"proto/orders/orders.proto": "syntax = \"proto3\"; package orders.v1; message Order {}\n",
		"proto/users/protobuf.mod":  "module example.com/users\n",
		"proto/users/users.proto":   "syntax = \"proto3\"; package users.v1; message User {}\n",
	}
	for path, content := range files {
		writeV1GenerateFixture(t, root, path, content)
	}
	out := filepath.Join(root, "all.pb")

	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out}))

	set := readDescriptorSet(t, out)
	require.ElementsMatch(t, []string{"orders.proto", "users.proto"}, descriptorNames(set))
}

func TestRunDescriptorSetRejectsCombinedSymbolConflict(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	files := map[string]string{
		"easyp.gen.yaml":  "version: v1\ngenerate:\n  modules: [m1, m2]\n",
		"m1/protobuf.mod": "module example.com/m1\n",
		"m2/protobuf.mod": "module example.com/m2\n",
		"m1/first.proto":  "syntax = \"proto3\"; package acme.v1; message Item { string first = 1; }",
		"m2/second.proto": "syntax = \"proto3\"; package acme.v1; message Item { string second = 1; }",
	}
	for path, content := range files {
		writeV1GenerateFixture(t, root, path, content)
	}
	out := filepath.Join(root, "all.pb")
	previous := []byte("previous descriptor output")
	require.NoError(t, os.WriteFile(out, previous, 0o600))
	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: out, IncludeImports: true})

	require.ErrorContains(t, err, "name conflict")
	actual, readErr := os.ReadFile(out)
	require.NoError(t, readErr)
	require.Equal(t, previous, actual)
}

func TestRunDescriptorSetSharedImports(t *testing.T) {
	t.Parallel()
	for _, include := range []bool{false, true} {
		include := include
		t.Run(fmt.Sprintf("include_imports_%t", include), func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			files := map[string]string{
				"easyp.gen.yaml":      "version: v1\ngenerate:\n  modules: [m1, m2]\n",
				"shared/protobuf.mod": "module example.com/shared\n",
				"shared/common.proto": "syntax = \"proto3\"; package shared; import \"google/protobuf/timestamp.proto\"; message Common { google.protobuf.Timestamp created = 1; }\n",
			}
			for _, module := range []string{"m1", "m2"} {
				files[module+"/protobuf.mod"] = "module example.com/" + module + "\nrequire example.com/shared v1.0.0\nreplace example.com/shared => ../shared\n"
				files[module+"/"+module+".proto"] = "syntax = \"proto3\"; package " + module + "; import \"common.proto\"; message Item { shared.Common common = 1; }\n"
			}
			for path, content := range files {
				writeV1GenerateFixture(t, root, path, content)
			}
			out := filepath.Join(root, "all.pb")
			request := Request{WorkDir: root, DescriptorSetOut: out, IncludeImports: include}

			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))
			first, err := os.ReadFile(out)
			require.NoError(t, err)
			set := readDescriptorSet(t, out)
			want := []string{"m1.proto", "m2.proto"}
			if include {
				want = append(want, "common.proto", "google/protobuf/timestamp.proto")
				_, err := protodesc.NewFiles(set)
				require.NoError(t, err)
			}
			require.ElementsMatch(t, want, descriptorNames(set))

			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))
			second, err := os.ReadFile(out)
			require.NoError(t, err)
			require.Equal(t, first, second)
		})
	}
}

func readDescriptorSet(t *testing.T, path string) *descriptorpb.FileDescriptorSet {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	set := new(descriptorpb.FileDescriptorSet)
	require.NoError(t, proto.Unmarshal(raw, set))
	return set
}
func descriptorNames(set *descriptorpb.FileDescriptorSet) []string {
	names := make([]string, 0, len(set.File))
	for _, file := range set.File {
		names = append(names, file.GetName())
	}
	return names
}
