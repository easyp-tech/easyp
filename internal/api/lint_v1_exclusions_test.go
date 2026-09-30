package api

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestLintV1PathExclusions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, rootPolicy, childPolicy, file, proto, lintRoot string
		wantIssue                                            bool
	}{
		{name: "inherited_issue_source", rootPolicy: "issues:\n  exclude-rules:\n    - path: nested/api/**\n", childPolicy: "linters:\n  default: MINIMAL\n", file: "nested/api/bad.proto", proto: "invalid proto", lintRoot: "nested"},
		{name: "child_issue_source", childPolicy: "issues:\n  exclude-rules:\n    - path: api/**\n", file: "nested/api/bad.proto", proto: "invalid proto", lintRoot: "nested/api"},
		{name: "literal_directory", rootPolicy: "issues:\n  exclude-rules:\n    - path: nested/api\n", file: "nested/api/deep/bad.proto", proto: "invalid proto", lintRoot: "nested"},
		{name: "named_rule", rootPolicy: "issues:\n  exclude-rules:\n    - path: nested/api/**\n      linters: [PACKAGE_DEFINED]\n", file: "nested/api/file.proto", proto: "syntax = \"proto3\";", lintRoot: "nested/api"},
		{name: "named_group", rootPolicy: "issues:\n  exclude-rules:\n    - path: nested/api/**\n      linters: [MINIMAL]\n", file: "nested/api/file.proto", proto: "syntax = \"proto3\";", lintRoot: "nested/api"},
		{name: "other_rule_still_checked", rootPolicy: "issues:\n  exclude-rules:\n    - path: nested/api/**\n      linters: [FILE_LOWER_SNAKE_CASE]\n", file: "nested/api/file.proto", proto: "syntax = \"proto3\";", lintRoot: "nested/api", wantIssue: true},
		{name: "different_path_still_checked", rootPolicy: "issues:\n  exclude-rules:\n    - path: other/**\n", file: "nested/api/file.proto", proto: "syntax = \"proto3\";", lintRoot: "nested/api", wantIssue: true},
		{name: "child_clears_exclusion", rootPolicy: "issues:\n  exclude-rules:\n    - path: nested/api/**\n", childPolicy: "issues: {}\n", file: "nested/api/file.proto", proto: "syntax = \"proto3\";", lintRoot: "nested/api", wantIssue: true},
		{name: "pathless_all", rootPolicy: "issues:\n  exclude-rules:\n    - {}\n", file: "nested/api/file.proto", proto: "invalid proto", lintRoot: "nested"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "selected.yaml", "version: v1\nlinters:\n  default: MINIMAL\n"+tt.rootPolicy)
			if tt.childPolicy != "" {
				writeV1GenerateFixture(t, root, "nested/easyp.yaml", "version: v1\n"+tt.childPolicy)
			}
			writeV1GenerateFixture(t, root, tt.file, tt.proto)
			ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
			ctx.Context = t.Context()
			err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "selected.yaml"), root, filepath.Join(root, tt.lintRoot))
			if tt.wantIssue {
				require.ErrorIs(t, err, ErrHasLintIssue)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLintV1ExclusionBeforeDependencies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nlinters:\n  default: MINIMAL\nissues:\n  exclude-rules:\n    - path: module/proto/**\n")
	writeV1GenerateFixture(t, root, "module/protobuf.mod", "invalid manifest")
	writeV1GenerateFixture(t, root, "module/proto/bad.proto", "invalid proto")
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()
	err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
	require.NoError(t, err)
}

func TestLintV1ExcludedFileRemainsImportable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nlinters:\n  default: MINIMAL\nissues:\n  exclude-rules:\n    - path: dep.proto\n")
	writeV1GenerateFixture(t, root, "dep.proto", "syntax = \"proto3\"; message Dep {}")
	writeV1GenerateFixture(t, root, "api/use.proto", "syntax = \"proto3\"; package api; import \"dep.proto\"; message Use { Dep value = 1; }")
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()
	err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
	require.NoError(t, err)
}

func TestLintV1NamedExclusionPreservesOtherRuleState(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nlinters:\n  default: MINIMAL\n  disable: [PACKAGE_DIRECTORY_MATCH]\nissues:\n  exclude-rules:\n    - path: a.proto\n      linters: [FILE_LOWER_SNAKE_CASE]\n")
	writeV1GenerateFixture(t, root, "a.proto", "syntax = \"proto3\"; package first;")
	writeV1GenerateFixture(t, root, "b.proto", "syntax = \"proto3\"; package second;")
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()
	err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
	require.ErrorIs(t, err, ErrHasLintIssue)
}

func TestLintV1ExclusionsDoNotValidateUnsupportedPolicies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "easyp.yaml", "version: v1\nlinters:\n  extends: unavailable\nissues:\n  exclude-rules:\n    - {}\n")
	writeV1GenerateFixture(t, root, "item.proto", "invalid proto")
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()
	err := (Lint{}).actionV1(ctx, logger.NewNop(), filepath.Join(root, "easyp.yaml"), root, root)
	require.ErrorContains(t, err, "linters.extends")
}
