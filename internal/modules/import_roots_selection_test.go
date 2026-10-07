package modules

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultIntrinsicRootsRequireDisconnectedMissingOwners(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		files     []RootProtoFile
		problems  []RootPathProblem
		wantError string
	}{
		{
			name: "unrelated_sources_in_same_directory",
			files: []RootProtoFile{
				intrinsicRootFile("api/service.proto", `import "api/types.proto";`),
				intrinsicRootFile("api/types.proto", ""),
				intrinsicRootFile("api/aux.proto", `import "aux_types.proto";`),
				intrinsicRootFile("api/aux_types.proto", ""),
			},
		},
		{
			name: "direct_contradiction",
			files: []RootProtoFile{
				intrinsicRootFile("api/request.proto", `import "api/service.proto"; import "types.proto";`),
				intrinsicRootFile("api/service.proto", ""),
				intrinsicRootFile("schema/types.proto", ""),
			},
			wantError: "previously resolved source",
		},
		{
			name: "chained_contradiction",
			files: []RootProtoFile{
				intrinsicRootFile("api/request.proto", `import "api/child.proto";`),
				intrinsicRootFile("api/child.proto", `import "types.proto";`),
				intrinsicRootFile("schema/types.proto", ""),
			},
			wantError: "previously resolved source",
		},
		{
			name: "later_missing_declaration_is_connected",
			files: []RootProtoFile{
				intrinsicRootFile("aaa/aux.proto", `import "aux_types.proto";`),
				intrinsicRootFile("aaa/aux_types.proto", ""),
				intrinsicRootFile("api/request.proto", `import "api/child.proto";`),
				intrinsicRootFile("api/child.proto", `import "types.proto";`),
				intrinsicRootFile("schema/types.proto", ""),
			},
			wantError: "previously resolved source",
		},
		{
			name: "missing_owner_aliases_default_edge_owner",
			files: []RootProtoFile{
				{Path: "public/service.proto", Identity: "physical/service.proto", Content: []byte(`import "public/types.proto"; import "aux_types.proto";`)},
				intrinsicRootFile("public/types.proto", ""),
				{Path: "tools/aux/service.proto", Identity: "physical/service.proto", Content: []byte(`import "public/types.proto"; import "aux_types.proto";`)},
				intrinsicRootFile("tools/aux/aux_types.proto", ""),
			},
			wantError: "previously resolved source",
		},
		{
			name: "missing_owner_aliases_default_edge_target",
			files: []RootProtoFile{
				intrinsicRootFile("public/service.proto", `import "public/types.proto";`),
				{Path: "public/types.proto", Identity: "physical/types.proto", Content: []byte(`import "aux_types.proto";`)},
				{Path: "tools/aux/types.proto", Identity: "physical/types.proto", Content: []byte(`import "aux_types.proto";`)},
				intrinsicRootFile("tools/aux/aux_types.proto", ""),
			},
			wantError: "previously resolved source",
		},
		{
			name: "no_default_intrinsic_bindings",
			files: []RootProtoFile{
				intrinsicRootFile("request.proto", `import "types.proto";`),
				intrinsicRootFile("schema/types.proto", ""),
			},
			wantError: "omit intrinsic import source",
		},
		{
			name:      "invalid_default_namespace",
			files:     disconnectedIntrinsicRootFiles(),
			problems:  []RootPathProblem{{Path: "bad.proto", Err: errors.New("outside root")}},
			wantError: "previously resolved source",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			module := importRootModule{name: "example.test/dep", roots: []string{"."}, inspection: &RootInspection{Files: tt.files, Problems: tt.problems}}

			roots, err := selectModuleImportRoots(t.Context(), module)

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Nil(t, roots)
		})
	}
}

func TestDefaultIntrinsicRootsDoNotHideAmbiguity(t *testing.T) {
	t.Parallel()
	files := []RootProtoFile{
		intrinsicRootFile("request.proto", `import "anchor.proto";`),
		intrinsicRootFile("anchor.proto", ""),
		{Path: "api/request.proto", Identity: "physical/request.proto", Content: []byte(`import "types.proto";`)},
		{Path: "sources/request.proto", Identity: "physical/request.proto", Content: []byte(`import "types.proto";`)},
		intrinsicRootFile("api/types.proto", ""),
		intrinsicRootFile("sources/types.proto", ""),
		{Path: "api/anchor.proto", Identity: "anchor.proto"},
		{Path: "sources/anchor.proto", Identity: "anchor.proto"},
	}

	_, err := selectModuleImportRoots(t.Context(), importRootModule{name: "example.test/dep", roots: []string{"."}, inspection: &RootInspection{Files: files}})

	require.ErrorContains(t, err, "ambiguous import roots")
	assert.ErrorContains(t, err, "api")
	assert.ErrorContains(t, err, "sources")
}

func TestDefaultIntrinsicRootsHonorCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := selectModuleImportRoots(ctx, importRootModule{name: "example.test/dep", roots: []string{"."}, inspection: &RootInspection{Files: disconnectedIntrinsicRootFiles()}})

	require.ErrorIs(t, err, context.Canceled)
}

func TestDefaultIntrinsicRootsHonorCandidateLimit(t *testing.T) {
	t.Parallel()
	files := disconnectedIntrinsicRootFiles()
	for number := range maxImportRootCandidates/4 + 1 {
		files = append(files, intrinsicRootFile(fmt.Sprintf("choices/%05d/inner/file_%05d.proto", number, number), ""))
	}

	_, err := selectModuleImportRoots(t.Context(), importRootModule{name: "example.test/dep", roots: []string{"."}, inspection: &RootInspection{Files: files}})

	require.ErrorContains(t, err, "candidate limit exceeded")
}

func intrinsicRootFile(name, imports string) RootProtoFile {
	return RootProtoFile{Path: name, Identity: name, Content: []byte(imports)}
}

func disconnectedIntrinsicRootFiles() []RootProtoFile {
	return []RootProtoFile{
		intrinsicRootFile("public/service.proto", `import "public/types.proto";`),
		intrinsicRootFile("public/types.proto", ""),
		intrinsicRootFile("tools/aux/service.proto", `import "aux_types.proto";`),
		intrinsicRootFile("tools/aux/aux_types.proto", ""),
	}
}
