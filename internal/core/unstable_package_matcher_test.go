package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestUnstablePackageMatcher(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, pkg string
		want      bool
	}{
		{name: "test", pkg: "foo.v1test", want: true},
		{name: "test_suffix", pkg: "foo.v12testFeature", want: true},
		{name: "alpha", pkg: "foo.v2alpha", want: true},
		{name: "alpha_number", pkg: "foo.v2alpha3", want: true},
		{name: "beta", pkg: "foo.v3beta", want: true},
		{name: "beta_number", pkg: "foo.v3beta12", want: true},
		{name: "point_alpha", pkg: "foo.v2p3alpha4", want: true},
		{name: "point_beta", pkg: "v2p3beta", want: true},
		{name: "stable", pkg: "foo.v1"},
		{name: "not_last", pkg: "foo.v1alpha1.service"},
		{name: "test_not_last", pkg: "v1test.foo"},
		{name: "component_prefix", pkg: "foo.prefixv1alpha1"},
		{name: "component_suffix", pkg: "foo.v1alpha1suffix"},
		{name: "zero_version", pkg: "foo.v0alpha1"},
		{name: "zero_point", pkg: "foo.v1p0beta1"},
		{name: "zero_prerelease", pkg: "foo.v1beta0"},
		{name: "missing_version", pkg: "foo.valpha1"},
		{name: "stable_point", pkg: "foo.v1p2"},
		{name: "empty", pkg: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isUnstablePackage(PackageName(tt.pkg)))
		})
	}
}

func TestCompareBreakingIgnoreUnstable(t *testing.T) {
	t.Parallel()
	const declarations = `message Item { string id = 1; oneof value { string text = 2; } }
enum State { STATE_UNSPECIFIED = 0; }
service ItemService { rpc Get(Item) returns (Item); }`
	tests := []struct {
		name, baselinePackage, currentPackage, currentFile, currentBody string
		ignore, files                                                   bool
		wantCount                                                       int
	}{
		{name: "unstable_enabled", baselinePackage: "example.v1beta1", currentPackage: "example.v1beta1", ignore: true},
		{name: "unstable_disabled", baselinePackage: "example.v1beta1", currentPackage: "example.v1beta1", wantCount: 4},
		{name: "stable_enabled", baselinePackage: "example.v1", currentPackage: "example.v1", ignore: true, wantCount: 4},
		{name: "baseline_unstable_current_stable", baselinePackage: "example.v1beta1", currentPackage: "example.v1", ignore: true},
		{name: "baseline_stable_current_unstable", baselinePackage: "example.v1", currentPackage: "example.v1beta1", currentBody: declarations, ignore: true, wantCount: 4},
		{name: "stable_file_moves", baselinePackage: "example.v1", currentPackage: "example.v1", currentFile: "new.proto", currentBody: declarations, ignore: true, files: true, wantCount: 4},
		{name: "unstable_file_moves", baselinePackage: "example.v1alpha", currentPackage: "example.v1alpha", currentFile: "new.proto", currentBody: declarations, ignore: true, files: true},
		{name: "unstable_not_last", baselinePackage: "example.v1beta1.service", currentPackage: "example.v1beta1.service", ignore: true, wantCount: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseline, current := t.TempDir(), t.TempDir()
			writeUnstableProto(t, baseline, "old.proto", "syntax = \"proto3\"; package "+tt.baselinePackage+"; "+declarations)
			currentFile := tt.currentFile
			if currentFile == "" {
				currentFile = "old.proto"
			}
			writeUnstableProto(t, current, currentFile, "syntax = \"proto3\"; package "+tt.currentPackage+"; "+tt.currentBody)
			app := New(Options{Logger: logger.NewNop(), BreakingCheckConfig: BreakingCheckConfig{IgnoreUnstable: tt.ignore, FilesCheck: tt.files}})
			issues, err := app.CompareBreaking(t.Context(), fs.NewFSWalker(current, "."), fs.NewFSWalker(baseline, "."), nil)
			require.NoError(t, err)
			assert.Len(t, issues, tt.wantCount, "%+v", issues)
		})
	}
}

func TestIgnoreUnstablePreservesStableReferencesAndDependencyBaselines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, dependencyPackage string
		wantCount               int
	}{
		{name: "unstable_dependency", dependencyPackage: "dep.v1beta1", wantCount: 4},
		{name: "stable_dependency", dependencyPackage: "dep.v1", wantCount: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseline, current, baselineDeps, currentDeps := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
			writeUnstableProto(t, baselineDeps, "dep.proto", "syntax = \"proto3\"; package "+tt.dependencyPackage+"; message Item { string id = 1; }")
			writeUnstableProto(t, currentDeps, "dep.proto", "syntax = \"proto3\"; package "+tt.dependencyPackage+"; message Item {} message Other {}")
			writeUnstableProto(t, baselineDeps, "removed.proto", "syntax = \"proto3\"; package old.v1beta1;")
			writeUnstableProto(t, baseline, "use.proto", "syntax = \"proto3\"; package stable.v1; import \"dep.proto\"; import \"removed.proto\"; message Use { "+tt.dependencyPackage+".Item item = 1; } service API { rpc Get("+tt.dependencyPackage+".Item) returns ("+tt.dependencyPackage+".Item); }")
			writeUnstableProto(t, current, "use.proto", "syntax = \"proto3\"; package stable.v1; import \"dep.proto\"; message Use { "+tt.dependencyPackage+".Other item = 1; } service API { rpc Get("+tt.dependencyPackage+".Other) returns ("+tt.dependencyPackage+".Other); }")
			app := New(Options{Logger: logger.NewNop(), ImportRoots: []string{currentDeps}, BreakingCheckConfig: BreakingCheckConfig{IgnoreUnstable: true}})
			issues, err := app.CompareBreaking(t.Context(), fs.NewFSWalker(current, "."), fs.NewFSWalker(baseline, "."), []string{baselineDeps})
			require.NoError(t, err)
			assert.Len(t, issues, tt.wantCount, "%+v", issues)
			stableCount := 0
			for _, issue := range issues {
				if issue.Path == "use.proto" {
					stableCount++
				}
			}
			assert.Equal(t, 4, stableCount)
		})
	}
}

func writeUnstableProto(t *testing.T, root, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), 0o600))
}

func TestLegacyBreakingCheckIgnoreUnstable(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		ignore    bool
		wantCount int
	}{
		{name: "enabled", ignore: true},
		{name: "disabled", wantCount: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseline, current := t.TempDir(), t.TempDir()
			writeUnstableProto(t, baseline, "item.proto", "syntax = \"proto3\"; package example.v1test; message Item {}")
			writeUnstableProto(t, current, "item.proto", "syntax = \"proto3\"; package example.v1test;")
			app := New(Options{Logger: logger.NewNop(), BreakingCheckConfig: BreakingCheckConfig{IgnoreUnstable: tt.ignore}, CurrentProjectGitWalker: unstableBaselineWalker{DirWalker: fs.NewFSWalker(baseline, ".")}})
			issues, err := app.BreakingCheck(t.Context(), current, current, ".")
			require.NoError(t, err)
			assert.Len(t, issues, tt.wantCount)
		})
	}
}

type unstableBaselineWalker struct{ DirWalker }

func (w unstableBaselineWalker) GetDirWalker(_, _, _ string) (DirWalker, error) {
	return w.DirWalker, nil
}
