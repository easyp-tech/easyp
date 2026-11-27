package generation

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

// The module index runs alongside compilation but reports as if it ran first.
func TestRunReportsImportCollisionBeforeCompileError(t *testing.T) {
	t.Parallel()

	for _, managed := range []bool{false, true} {
		root, dependency := t.TempDir(), t.TempDir()
		replacement, err := filepath.Rel(root, dependency)
		require.NoError(t, err)
		writeV1GenerateFixture(t, dependency, "protobuf.mod", "module example.com/dep\n")
		writeV1GenerateFixture(t, dependency, "dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1;\n")
		writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\nroots proto\nrequire example.com/dep v1.0.0\nreplace example.com/dep => "+replacement+"\n")
		writeV1GenerateFixture(t, root, "proto/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1;\n")
		writeV1GenerateFixture(t, root, "proto/app/v1/app.proto", "syntax = \"proto3\"; package app.v1; message {\n")
		config := "version: v1\n"
		if managed {
			config += "generate:\n  managed:\n    enabled: true\n"
		}
		writeV1GenerateFixture(t, root, "easyp.gen.yaml", config)

		err = Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: filepath.Join(root, "out.pb")})

		require.ErrorContains(t, err, "IndexFileModules: duplicate import path \"dep/v1/dep.proto\"", "managed=%v", managed)
	}
}
