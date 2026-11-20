package compilerbackend

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestCompileImportPathsWKTAndSourceInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sourceInfo     bool
		wantSourceInfo bool
	}{
		{name: "with_source_info", sourceInfo: true, wantSourceInfo: true},
		{name: "without_source_info", sourceInfo: false, wantSourceInfo: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			files := map[string]string{
				"a/common/v1/common.proto": `syntax = "proto3";
package common.v1;
import "google/protobuf/timestamp.proto";
// Common message comment.
message Common {
  google.protobuf.Timestamp at = 1;
}
`,
				"b/api/v1/api.proto": `syntax = "proto3";
package api.v1;
import "common/v1/common.proto";
// Request message comment.
message Req {
  // Common field comment.
  common.v1.Common common = 1;
}
service Api {
  rpc Get(Req) returns (Req);
}
`,
			}
			for rel, src := range files {
				full := filepath.Join(dir, rel)
				require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
				require.NoError(t, os.WriteFile(full, []byte(src), 0o600))
			}

			result, err := Compile(t.Context(), Request{
				Files:           []string{"api/v1/api.proto", "common/v1/common.proto"},
				ImportPaths:     []string{filepath.Join(dir, "b"), filepath.Join(dir, "a")},
				Accessor:        func(path string) (io.ReadCloser, error) { return os.Open(path) },
				StandardImports: true,
				SourceInfo:      tc.sourceInfo,
			})
			require.NoError(t, err)
			require.Len(t, result.Files, 2)
			assert.Equal(t, "api/v1/api.proto", result.Files[0].Path())
			assert.Equal(t, "common/v1/common.proto", result.Files[1].Path())

			require.Contains(t, result.Protos, "api/v1/api.proto")
			require.Contains(t, result.Protos, "common/v1/common.proto")
			require.Contains(t, result.Protos, "google/protobuf/timestamp.proto")

			apiProto := result.Protos["api/v1/api.proto"]
			if tc.wantSourceInfo {
				require.NotNil(t, apiProto.GetSourceCodeInfo())
				assert.NotEmpty(t, apiProto.GetSourceCodeInfo().GetLocation())
				msg := result.Files[0].Messages().ByName("Req")
				require.NotNil(t, msg)
				loc := result.Files[0].SourceLocations().ByDescriptor(msg)
				assert.Contains(t, loc.LeadingComments, "Request message comment.")
			} else {
				assert.Nil(t, apiProto.GetSourceCodeInfo())
			}
		})
	}
}

func TestCompileFieldShadowsPackage(t *testing.T) {
	t.Parallel()

	sources := map[string]string{
		"svc.proto":  `syntax = "proto3"; package service; message Service {}`,
		"main.proto": `syntax = "proto3"; import "svc.proto"; message C { service.Service service = 1; }`,
	}

	result, err := Compile(t.Context(), Request{
		Files:    []string{"main.proto"},
		Accessor: mapAccessor(sources),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	field := result.Files[0].Messages().ByName("C").Fields().ByName("service")
	require.NotNil(t, field)
	assert.Equal(t, protoreflect.FullName("service.Service"), field.Message().FullName())
}

// TestCompileAccessorIsolationNoCWDFallback verifies that compilation never
// falls back to reading files from the process working directory when the
// configured Accessor reports them absent. Sequential because it uses os.Chdir.
func TestCompileAccessorIsolationNoCWDFallback(t *testing.T) {
	cwdDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(cwdDir, "leaked.proto"),
		[]byte(`syntax = "proto3"; package leaked.v1; message Secret { string x = 1; }`),
		0o600,
	))

	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(cwdDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origWD))
	})

	rootDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(rootDir, "app.proto"),
		[]byte(`syntax = "proto3"; package app.v1; import "leaked.proto"; message App { leaked.v1.Secret s = 1; }`),
		0o600,
	))

	_, err = Compile(t.Context(), Request{
		Files:       []string{"app.proto"},
		ImportPaths: []string{rootDir},
		Accessor: func(filename string) (io.ReadCloser, error) {
			if strings.HasSuffix(filename, "leaked.proto") {
				return nil, &os.PathError{Op: "open", Path: filename, Err: os.ErrNotExist}
			}
			return os.Open(filename)
		},
		StandardImports: true,
	})
	require.Error(t, err)
}

func TestCompilePreservesAccessorErrors(t *testing.T) {
	t.Parallel()

	errBlocked := errors.New("source target is outside root")
	sources := map[string]string{
		"app.proto": `syntax = "proto3"; package app.v1; import "escape.proto"; message App {}`,
	}

	_, err := Compile(t.Context(), Request{
		Files: []string{"app.proto"},
		Accessor: func(name string) (io.ReadCloser, error) {
			if name == "escape.proto" {
				return nil, errBlocked
			}
			return mapAccessor(sources)(name)
		},
		StandardImports: true,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errBlocked)
}

func TestCompileSemanticErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		source          string
		wantErrContains string
	}{
		{
			name:            "duplicate_enum_without_allow_alias",
			source:          `syntax = "proto3"; package p; enum E { A = 0; B = 0; }`,
			wantErrContains: "allow_alias",
		},
		{
			name:            "proto3_nonzero_first_enum",
			source:          `syntax = "proto3"; package p; enum E { A = 1; }`,
			wantErrContains: "zero",
		},
		{
			name:            "unknown_field_type",
			source:          `syntax = "proto3"; package p; message M { Missing x = 1; }`,
			wantErrContains: "Missing",
		},
		{
			name:            "duplicate_field_number",
			source:          `syntax = "proto3"; package p; message M { string a = 1; string b = 1; }`,
			wantErrContains: "1",
		},
		{
			name:            "missing_import",
			source:          `syntax = "proto3"; package p; import "missing.proto"; message M {}`,
			wantErrContains: "missing.proto",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Compile(t.Context(), Request{
				Files:           []string{"schema.proto"},
				Accessor:        mapAccessor(map[string]string{"schema.proto": tc.source}),
				StandardImports: true,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains)
		})
	}
}

func TestCompileEditions2023Features(t *testing.T) {
	t.Parallel()

	src := `edition = "2023";
package ed.v1;
import "google/protobuf/cpp_features.proto";

message M {
  option features.json_format = LEGACY_BEST_EFFORT;
  string implicit_str = 1 [features.field_presence = IMPLICIT];
  string no_utf8 = 2 [features.utf8_validation = NONE];
  string cpp_view = 3 [features.(pb.cpp).string_type = VIEW];
}
`
	result, err := Compile(t.Context(), Request{
		Files:           []string{"ed.proto"},
		Accessor:        mapAccessor(map[string]string{"ed.proto": src}),
		StandardImports: true,
		SourceInfo:      true,
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	fd := result.Files[0]
	assert.Equal(t, protoreflect.Editions, fd.Syntax())
	fdp := result.Protos["ed.proto"]
	require.NotNil(t, fdp)
	assert.Equal(t, descriptorpb.Edition_EDITION_2023, fdp.GetEdition())

	msg := fd.Messages().ByName("M")
	require.NotNil(t, msg)
	msgOpts := msg.Options().(*descriptorpb.MessageOptions)
	assert.Equal(t, descriptorpb.FeatureSet_LEGACY_BEST_EFFORT, msgOpts.GetFeatures().GetJsonFormat())

	implicitField := msg.Fields().ByName("implicit_str")
	require.NotNil(t, implicitField)
	assert.False(t, implicitField.HasPresence())

	noUTF8Field := msg.Fields().ByName("no_utf8")
	require.NotNil(t, noUTF8Field)
	noUTF8Opts := noUTF8Field.Options().(*descriptorpb.FieldOptions)
	assert.Equal(t, descriptorpb.FeatureSet_NONE, noUTF8Opts.GetFeatures().GetUtf8Validation())

	cppField := msg.Fields().ByName("cpp_view")
	require.NotNil(t, cppField)
	cppFeatures := cppField.Options().(*descriptorpb.FieldOptions).GetFeatures()
	require.NotNil(t, cppFeatures)
	cppStringType := ""
	cppFeatures.ProtoReflect().Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if string(f.FullName()) != "pb.cpp" || f.Kind() != protoreflect.MessageKind {
			return true
		}
		member := v.Message().Descriptor().Fields().ByName("string_type")
		if member != nil && v.Message().Has(member) {
			if ev := member.Enum().Values().ByNumber(v.Message().Get(member).Enum()); ev != nil {
				cppStringType = string(ev.Name())
			}
		}
		return false
	})
	assert.Equal(t, "VIEW", cppStringType)
}

func TestCompileValidationAndHelpers(t *testing.T) {
	t.Parallel()

	t.Run("empty_files_rejected", func(t *testing.T) {
		t.Parallel()

		_, err := Compile(t.Context(), Request{})
		require.ErrorIs(t, err, ErrNoInputFiles)
	})

	t.Run("canceled_context_rejected", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := Compile(ctx, Request{
			Files:    []string{"a.proto"},
			Accessor: mapAccessor(map[string]string{"a.proto": `syntax = "proto3";`}),
		})
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("reader_and_bytes_accessors", func(t *testing.T) {
		t.Parallel()

		src := `syntax = "proto3"; package p; message M { string id = 1; }`
		resReader, err := Compile(t.Context(), Request{
			Files:          []string{"a.proto"},
			MaxParallelism: 1,
			Accessor: ReaderAccessor(func(name string) (io.Reader, error) {
				if name != "a.proto" {
					return nil, os.ErrNotExist
				}
				return strings.NewReader(src), nil
			}),
		})
		require.NoError(t, err)
		require.Len(t, resReader.Files, 1)

		resBytes, err := Compile(t.Context(), Request{
			Files:          []string{"a.proto"},
			MaxParallelism: 1,
			Accessor: BytesAccessor(func(name string) ([]byte, error) {
				if name != "a.proto" {
					return nil, os.ErrNotExist
				}
				return []byte(src), nil
			}),
		})
		require.NoError(t, err)
		require.Len(t, resBytes.Files, 1)
	})
}

func mapAccessor(files map[string]string) func(string) (io.ReadCloser, error) {
	return BytesAccessor(func(name string) ([]byte, error) {
		src, ok := files[name]
		if !ok {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		return []byte(src), nil
	})
}
