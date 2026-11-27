package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

// issuePaths lists the file of every issue, in report order.
func issuePaths(t *testing.T, rule core.Rule, infos ...core.ProtoInfo) []string {
	t.Helper()
	moduleRule, ok := rule.(core.ModuleRule)
	require.True(t, ok)
	for _, info := range infos {
		require.NoError(t, moduleRule.Collect(info))
	}
	var paths []string
	for _, issue := range moduleRule.Finish() {
		paths = append(paths, issue.Path)
	}
	return paths
}

func TestPackageSameOptionReportsEveryFileOfThePackage(t *testing.T) {
	t.Parallel()

	a := parseProto(t, "a/v1/a.proto", "syntax = \"proto3\";\npackage a.v1;\noption go_package = \"x/a\";\n")
	b := parseProto(t, "a/v1/b.proto", "syntax = \"proto3\";\npackage a.v1;\noption go_package = \"x/b\";\n")
	missing := parseProto(t, "a/v1/c.proto", "syntax = \"proto3\";\npackage a.v1;\n")
	other := parseProto(t, "b/v1/d.proto", "syntax = \"proto3\";\npackage b.v1;\noption go_package = \"x/d\";\n")

	require.Equal(t, []string{"a/v1/a.proto", "a/v1/b.proto"}, issuePaths(t, &rules.PackageSameGoPackage{}, a, b, other))
	require.Equal(t, []string{"a/v1/a.proto", "a/v1/c.proto"}, issuePaths(t, &rules.PackageSameGoPackage{}, a, missing),
		"a file without the option counts as its own value")
	require.Empty(t, issuePaths(t, &rules.PackageSameGoPackage{}, missing, missing), "no file sets the option")

	rule := &rules.PackageSameGoPackage{}
	require.NoError(t, rule.Collect(missing))
	require.NoError(t, rule.Collect(a))
	issues := rule.Finish()
	require.Len(t, issues, 2)
	require.Equal(t, core.Position{Filename: "a/v1/c.proto", Line: 1, Column: 1}, issues[0].Position, "missing option is reported on the file")
	require.Equal(t, 3, issues[1].Position.Line, "set option is reported on its statement")
}

func TestPackageSameSwiftPrefixChecksSwiftPrefix(t *testing.T) {
	t.Parallel()

	a := parseProto(t, "a/v1/a.proto", "syntax = \"proto3\";\npackage a.v1;\noption swift_prefix = \"A\";\noption ruby_package = \"R\";\n")
	b := parseProto(t, "a/v1/b.proto", "syntax = \"proto3\";\npackage a.v1;\noption swift_prefix = \"B\";\noption ruby_package = \"R\";\n")

	require.Len(t, issuePaths(t, &rules.PackageSameSwiftPrefix{}, a, b), 2)
	require.Empty(t, issuePaths(t, &rules.PackageSameRubyPackage{}, a, b))
}

func TestDirectoryAndPackageGroupsReportEveryFile(t *testing.T) {
	t.Parallel()

	a := parseProto(t, "dir/a.proto", "syntax = \"proto3\";\npackage a.v1;\n")
	b := parseProto(t, "dir/b.proto", "syntax = \"proto3\";\npackage b.v1;\n")
	c := parseProto(t, "other/c.proto", "syntax = \"proto3\";\npackage a.v1;\n")

	require.Equal(t, []string{"dir/a.proto", "dir/b.proto"}, issuePaths(t, &rules.DirectorySamePackage{}, a, b, c))
	require.Equal(t, []string{"dir/a.proto", "other/c.proto"}, issuePaths(t, &rules.PackageSameDirectory{}, a, b, c))
}

func TestEnumZeroValueSuffixChecksEveryZeroValue(t *testing.T) {
	t.Parallel()

	info := parseProto(t, "e.proto", `syntax = "proto2";
package e.v1;
enum Kind {
  KIND_UNSPECIFIED = 0;
  KIND_NONE = 0;
  KIND_ONE = 1;
}
enum Other {
  OTHER_ONE = 1;
  OTHER_ZERO = 0;
}
`)
	issues, err := (&rules.EnumZeroValueSuffix{Suffix: "UNSPECIFIED"}).Validate(info)
	require.NoError(t, err)
	var names []string
	for _, issue := range issues {
		names = append(names, issue.SourceName)
	}
	require.Equal(t, []string{"KIND_NONE", "OTHER_ZERO"}, names)
}

func TestRPCRequestResponseUniqueReportsEveryRPCSharingAType(t *testing.T) {
	t.Parallel()

	info := parseProto(t, "s.proto", `syntax = "proto3";
package s.v1;
message A {}
message B {}
message C {}
service S {
  rpc One(A) returns (B);
  rpc Two(A) returns (C);
  rpc Same(C) returns (C);
}
`)
	rule := &rules.RPCRequestResponseUnique{}
	require.NoError(t, rule.Collect(info))
	byType := map[string]int{}
	for _, issue := range rule.Finish() {
		byType[issue.SourceName]++
	}
	require.Equal(t, map[string]int{"s.v1.A": 2, "s.v1.C": 3}, byType,
		"A is shared by One and Two; C by Two and Same, and Same also uses it for both")
}
