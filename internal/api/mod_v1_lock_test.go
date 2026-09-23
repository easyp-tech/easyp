package api

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/sumdb/dirhash"
	"gopkg.in/yaml.v3"
)

func TestBuildV1LockPinsGitCommitAndGoStyleH1(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(filepath.Join(remote, "dep", "v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf("module %s\n", remote)
	proto := "syntax = \"proto3\";\npackage dep.v1;\nmessage Item { string id = 1; }\n"
	for name, content := range map[string]string{
		"protobuf.mod":     manifest,
		"dep/v1/dep.proto": proto,
	} {
		if err := os.WriteFile(filepath.Join(remote, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))

	wantHash, err := dirhash.Hash1([]string{"dep/v1/dep.proto", "protobuf.mod"}, func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(remote, filepath.FromSlash(name)))
	})
	if err != nil {
		t.Fatal(err)
	}
	module := v1.Module{
		Name:     "example.com/root",
		Roots:    []string{"."},
		Requires: []v1.Requirement{{Module: remote, Version: "v1.0.0"}},
	}
	lock, err := buildV1Lock(context.Background(), module, filepath.Join(t.TempDir(), "cache"))
	if err != nil {
		t.Fatal(err)
	}
	if lock.Version != 1 || len(lock.Modules) != 1 {
		t.Fatalf("unexpected lock: %#v", lock)
	}
	entry := lock.Modules[0]
	if entry.Source != remote || entry.Version != "v1.0.0" || entry.Commit != commit || entry.Hash != wantHash {
		t.Fatalf("lock entry = %#v; want source=%q version=v1.0.0 commit=%q hash=%q", entry, remote, commit, wantHash)
	}
}

func TestBuildV1LockSelectsModulesFromOneGitRepository(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "monorepo")
	foo := filepath.Join(repository, "foo")
	bar := filepath.Join(repository, "bar")
	for name, content := range map[string]string{
		"foo/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", foo),
		"foo/proto/foo.proto": "syntax = \"proto3\";\nmessage Foo {}\n",
		"bar/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", bar),
		"bar/proto/bar.proto": "syntax = \"proto3\";\nmessage Bar {}\n",
	} {
		path := filepath.Join(repository, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	for _, tag := range []string{"foo/v1.0.0", "foo/v1.1.0", "bar/v1.0.0"} {
		runTestGit(t, repository, "tag", tag)
	}
	module := v1.Module{Requires: []v1.Requirement{
		{Module: foo, Version: "v1.0.0"},
		{Module: bar, Version: "v1.0.0"},
	}}
	cache := t.TempDir()
	lock, err := buildV1Lock(context.Background(), module, cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Modules) != 2 || lock.Modules[0].Source != bar || lock.Modules[1].Source != foo {
		t.Fatalf("unexpected multi-module lock: %#v", lock)
	}
	if err := downloadV1Lock(context.Background(), lock, cache); err != nil {
		t.Fatal(err)
	}
	roots, err := sourcesFromV1Lock(lock, cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 2 || roots[0].module != bar || roots[1].module != foo {
		t.Fatalf("unexpected multi-module roots: %#v", roots)
	}
	for _, root := range roots {
		name := filepath.Base(filepath.Dir(root.path))
		if _, err := os.Stat(filepath.Join(root.path, name+".proto")); err != nil {
			t.Fatal(err)
		}
	}
	latest, err := latestCompatibleV1Tag(context.Background(), foo, "v1.0.0")
	if err != nil || latest != "v1.1.0" {
		t.Fatalf("latest foo tag = %s, %v", latest, err)
	}
}

func TestBuildV1LockSelectsUntaggedModulesByCommit(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "monorepo")
	foo := filepath.Join(repository, "foo")
	bar := filepath.Join(repository, "bar")
	for name, content := range map[string]string{
		"foo/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", foo),
		"foo/proto/foo.proto": "syntax = \"proto3\";\nmessage Foo {}\n",
		"bar/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", bar),
		"bar/proto/bar.proto": "syntax = \"proto3\";\nmessage Bar {}\n",
	} {
		path := filepath.Join(repository, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	commit := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	module := v1.Module{Requires: []v1.Requirement{
		{Module: foo, Version: commit},
		{Module: bar, Version: commit},
	}}
	cache := t.TempDir()
	lock, err := buildV1Lock(context.Background(), module, cache)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(lock.Modules) != 2 {
		t.Fatalf("lock has %d modules, want two", len(lock.Modules))
	}
	for _, entry := range lock.Modules {
		if entry.Commit != commit || entry.Version != commit {
			t.Fatalf("unexpected pinned module: %#v", entry)
		}
		updated, err := latestCompatibleV1Tag(context.Background(), entry.Source, entry.Version)
		if err != nil || updated != commit {
			t.Fatalf("update changed pinned commit: %s, %v", updated, err)
		}
	}
	if err := validateV1Requirements(module.Requires, lock); err != nil {
		t.Fatal(err)
	}
	if err := downloadV1Lock(context.Background(), lock, cache); err != nil {
		t.Fatal(err)
	}
	roots, err := sourcesFromV1Lock(lock, cache)
	if err != nil || len(roots) != 2 {
		t.Fatalf("roots: %#v, %v", roots, err)
	}
}

func TestTidyV1ResolvesUntaggedModulesWithoutVersionsAndKeepsLock(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "monorepo")
	foo := filepath.Join(repository, "foo")
	bar := filepath.Join(repository, "bar")
	for name, content := range map[string]string{
		"foo/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", foo),
		"foo/proto/foo.proto": "syntax = \"proto3\";\npackage foo.v1;\nmessage Foo {}\n",
		"bar/protobuf.mod":    fmt.Sprintf("module %s\nroots proto\n", bar),
		"bar/proto/bar.proto": "syntax = \"proto3\";\npackage bar.v1;\nmessage Bar {}\n",
	} {
		path := filepath.Join(repository, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestGit(t, repository, "init", "-q")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	initial := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/app\nrequire (\n  %s\n  %s\n)\n", foo, bar)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	appProto := "syntax = \"proto3\";\npackage app.v1;\nimport \"foo.proto\";\nimport \"bar.proto\";\nmessage App { foo.v1.Foo foo = 1; bar.v1.Bar bar = 2; }\n"
	if err := os.WriteFile(filepath.Join(root, "app.proto"), []byte(appProto), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASYPPATH", t.TempDir())
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	cliCtx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	readLock := func() v1.Lock {
		t.Helper()
		lock, err := readV1Lock(filepath.Join(root, "protobuf.lock"))
		if err != nil {
			t.Fatal(err)
		}
		return lock
	}
	if err := (Mod{}).Tidy(cliCtx); err != nil {
		t.Fatal(err)
	}
	lock := readLock()
	if len(lock.Modules) != 2 {
		t.Fatalf("want two modules, got %#v", lock)
	}
	for _, entry := range lock.Modules {
		if entry.Version != initial || entry.Commit != initial {
			t.Fatalf("unexpected auto-resolved module: %#v", entry)
		}
	}
	if err := os.WriteFile(filepath.Join(foo, "proto", "foo.proto"), []byte("syntax = \"proto3\";\npackage foo.v1;\nmessage Foo { string id = 1; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "foo update")
	updated := strings.TrimSpace(runTestGit(t, repository, "rev-parse", "HEAD"))
	if err := (Mod{}).Tidy(cliCtx); err != nil {
		t.Fatal(err)
	}
	for _, entry := range readLock().Modules {
		if entry.Commit != initial {
			t.Fatalf("tidy changed a pinned dependency: %#v", entry)
		}
	}
	if err := (Mod{}).Update(cliCtx); err != nil {
		t.Fatal(err)
	}
	for _, entry := range readLock().Modules {
		if entry.Commit != updated || entry.Version != updated {
			t.Fatalf("update did not refresh an unversioned dependency: %#v", entry)
		}
	}
}

func TestDownloadV1LockKeepsCommitAfterTagMoves(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte(fmt.Sprintf("module %s\n", remote)), 0o644); err != nil {
		t.Fatal(err)
	}
	proto := filepath.Join(remote, "dep.proto")
	oldContent := []byte("syntax = \"proto3\";\nmessage Old {}\n")
	if err := os.WriteFile(proto, oldContent, 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "old")
	runTestGit(t, remote, "tag", "v1.0.0")
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	module := v1.Module{Name: "example.com/root", Requires: []v1.Requirement{{Module: remote, Version: "v1.0.0"}}}
	lock, err := buildV1Lock(context.Background(), module, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proto, []byte("syntax = \"proto3\";\nmessage New {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "new")
	runTestGit(t, remote, "tag", "-f", "v1.0.0")
	if err := downloadV1Lock(context.Background(), lock, cacheRoot); err != nil {
		t.Fatal(err)
	}
	installed, err := os.ReadFile(filepath.Join(v1ModuleCachePath(cacheRoot, lock.Modules[0]), "dep.proto"))
	if err != nil {
		t.Fatal(err)
	}
	if string(installed) != string(oldContent) {
		t.Fatalf("download followed moved tag instead of locked commit: %s", installed)
	}
}

func TestTidyV1WritesLockedGitDependency(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte(fmt.Sprintf("module %s\n", remote)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "dep.proto"), []byte("syntax = \"proto3\";\npackage dep.v1;\nmessage Item { string id = 1; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
	wantHash, err := dirhash.Hash1([]string{"dep.proto", "protobuf.mod"}, func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(remote, name))
	})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", remote)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	cacheBase := filepath.Join(t.TempDir(), "cache")
	t.Setenv("EASYPPATH", cacheBase)
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	cliCtx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	if err := (Mod{}).Tidy(cliCtx); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "protobuf.lock"))
	if err != nil {
		t.Fatal(err)
	}
	var lock v1.Lock
	if err := yaml.Unmarshal(raw, &lock); err != nil {
		t.Fatal(err)
	}
	if lock.Version != 1 || len(lock.Modules) != 1 || lock.Modules[0].Commit != commit || lock.Modules[0].Hash != wantHash {
		t.Fatalf("unexpected lock: %#v", lock)
	}
	if err := (Mod{}).Download(cliCtx); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(v1ModuleCachePath(filepath.Join(cacheBase, "v1", "git"), lock.Modules[0]), "dep.proto")); err != nil {
		t.Fatal(err)
	}
	module, err := v1.ParseModule(strings.NewReader(manifest))
	if err != nil {
		t.Fatal(err)
	}
	roots, err := lockedV1DependencyRoots(context.Background(), root, module, filepath.Join(cacheBase, "v1", "git"))
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("dependency roots = %v", roots)
	}
	if _, err := os.Stat(filepath.Join(roots[0], "dep.proto")); err != nil {
		t.Fatal(err)
	}
}

func TestDownloadV1LockPinsCommitAndChecksCachedH1(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"protobuf.mod": fmt.Sprintf("module %s\n", remote),
		"dep.proto":    "syntax = \"proto3\";\npackage dep.v1;\nmessage Item { string id = 1; }\n",
	} {
		if err := os.WriteFile(filepath.Join(remote, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	lock, err := buildV1Lock(context.Background(), v1.Module{Requires: []v1.Requirement{{Module: remote, Version: "v1.0.0"}}}, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := downloadV1Lock(context.Background(), lock, cacheRoot); err != nil {
		t.Fatal(err)
	}
	installed := v1ModuleCachePath(cacheRoot, lock.Modules[0])
	if _, err := os.Stat(filepath.Join(installed, "dep.proto")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installed, "dep.proto"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := downloadV1Lock(context.Background(), lock, cacheRoot); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("expected cached hash mismatch, got %v", err)
	}
}

func TestDownloadV1LockRejectsUnsafeCommit(t *testing.T) {
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{{
		Source: "example.com/dependency", Version: "v1.0.0",
		Commit: "../../outside", Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}}}
	if err := downloadV1Lock(context.Background(), lock, t.TempDir()); err == nil || !strings.Contains(err.Error(), "invalid commit") {
		t.Fatalf("expected invalid commit error, got %v", err)
	}
}

func TestGenerateV1UsesLockedGitDependency(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"protobuf.mod": fmt.Sprintf("module %s\n", remote),
		"dep.proto":    "syntax = \"proto3\";\npackage dep.v1;\nmessage Item { string id = 1; }\n",
	} {
		if err := os.WriteFile(filepath.Join(remote, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", remote)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "user.proto"), []byte("syntax = \"proto3\";\npackage user.v1;\nimport \"dep.proto\";\nmessage User { dep.v1.Item item = 1; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	module, err := v1.ParseModule(strings.NewReader(manifest))
	if err != nil {
		t.Fatal(err)
	}
	cacheBase := filepath.Join(t.TempDir(), "cache")
	t.Setenv("EASYPPATH", cacheBase)
	lock, err := buildV1Lock(context.Background(), module, filepath.Join(cacheBase, "v1", "git"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := yaml.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "protobuf.lock"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	gen, err := v1.ParseGenerate(strings.NewReader("version: v1\nplugins:\n  - name: python\n    out: ./gen/python\n"))
	if err != nil {
		t.Fatal(err)
	}
	cliCtx := cli.NewContext(&cli.App{Metadata: map[string]any{}}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	cliCtx.Context = context.Background()
	if err := generateV1Module(cliCtx, logger.NewNop(), filepath.Join(root, "easyp.gen.yaml"), root, gen); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "gen", "python", "user_pb2.py")); err != nil {
		t.Fatal(err)
	}
}

func TestTidyV1RejectsUnresolvedImportWithRequirement(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte(fmt.Sprintf("module %s\n", remote)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "dep.proto"), []byte("syntax = \"proto3\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", remote)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "user.proto"), []byte("syntax = \"proto3\";\nimport \"missing.proto\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	cliCtx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	if err := (Mod{}).Tidy(cliCtx); err == nil || !strings.Contains(err.Error(), "missing.proto") {
		t.Fatalf("expected missing import error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "protobuf.lock")); !os.IsNotExist(err) {
		t.Fatalf("tidy wrote an invalid lockfile: %v", err)
	}
}

func TestUpdateV1SelectsLatestCompatibleTag(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte(fmt.Sprintf("module %s\n", remote)), 0o644); err != nil {
		t.Fatal(err)
	}
	protoPath := filepath.Join(remote, "dep.proto")
	if err := os.WriteFile(protoPath, []byte("syntax = \"proto3\";\npackage dep.v1;\nmessage Item { string id = 1; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "v1.0")
	runTestGit(t, remote, "tag", "v1.0.0")
	if err := os.WriteFile(protoPath, []byte("syntax = \"proto3\";\npackage dep.v1;\nmessage Item { string id = 1; string name = 2; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "v1.1")
	runTestGit(t, remote, "tag", "v1.1.0")
	wantCommit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
	if err := os.WriteFile(protoPath, []byte("syntax = \"proto3\";\npackage dep.v2;\nmessage Item {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "v2")
	runTestGit(t, remote, "tag", "v2.0.0")

	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", remote)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "user.proto"), []byte("syntax = \"proto3\";\nimport \"dep.proto\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	cliCtx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	if err := (Mod{}).Update(cliCtx); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), remote+" v1.1.0") || strings.Contains(string(updated), "v2.0.0") {
		t.Fatalf("unexpected updated manifest: %s", updated)
	}
	lockRaw, err := os.ReadFile(filepath.Join(root, "protobuf.lock"))
	if err != nil {
		t.Fatal(err)
	}
	var lock v1.Lock
	if err := yaml.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatal(err)
	}
	if len(lock.Modules) != 1 || lock.Modules[0].Version != "v1.1.0" || lock.Modules[0].Commit != wantCommit {
		t.Fatalf("unexpected updated lock: %#v", lock)
	}
}

func TestBuildV1LockSelectsTransitiveMinimum(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, "common")
	weather := filepath.Join(root, "weather")
	for _, dir := range []string{common, weather} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		runTestGit(t, dir, "init", "-q")
	}
	if err := os.WriteFile(filepath.Join(common, "protobuf.mod"), []byte(fmt.Sprintf("module %s\n", common)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(common, "common.proto"), []byte("syntax = \"proto3\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, common, "add", ".")
	runTestGit(t, common, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "common 1.0")
	runTestGit(t, common, "tag", "v1.0.0")
	if err := os.WriteFile(filepath.Join(common, "common.proto"), []byte("syntax = \"proto3\";\npackage common.v1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, common, "add", ".")
	runTestGit(t, common, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "common 1.1")
	runTestGit(t, common, "tag", "v1.1.0")
	if err := os.WriteFile(filepath.Join(weather, "protobuf.mod"), []byte(fmt.Sprintf("module %s\nrequire %s v1.1.0\n", weather, common)), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, weather, "add", ".")
	runTestGit(t, weather, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "weather")
	runTestGit(t, weather, "tag", "v1.0.0")
	module := v1.Module{Name: "example.com/root", Roots: []string{"."}, Requires: []v1.Requirement{
		{Module: common, Version: "v1.0.0"}, {Module: weather, Version: "v1.0.0"},
	}}
	lock, err := buildV1Lock(context.Background(), module, filepath.Join(t.TempDir(), "cache"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Modules) != 2 {
		t.Fatalf("lock modules = %#v", lock.Modules)
	}
	for _, entry := range lock.Modules {
		if entry.Source == common && entry.Version != "v1.1.0" {
			t.Fatalf("common version = %s", entry.Version)
		}
	}
}

func TestDownloadV1RejectsStaleDirectRequirement(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte(fmt.Sprintf("module %s\n", remote)), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	runTestGit(t, remote, "tag", "v1.0.0")
	root := t.TempDir()
	module := v1.Module{Name: "example.com/root", Roots: []string{"."}, Requires: []v1.Requirement{{Module: remote, Version: "v1.0.0"}}}
	cacheBase := filepath.Join(t.TempDir(), "cache")
	lock, err := buildV1Lock(context.Background(), module, filepath.Join(cacheBase, "v1", "git"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := yaml.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "protobuf.lock"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.1.0\n", remote)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASYPPATH", cacheBase)
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	cliCtx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	if err := (Mod{}).Download(cliCtx); err == nil || !strings.Contains(err.Error(), "does not satisfy") {
		t.Fatalf("expected stale lock error, got %v", err)
	}
}

func TestTidyV1RecordsTransitiveRequirementAsIndirect(t *testing.T) {
	deps := t.TempDir()
	common := filepath.Join(deps, "common")
	weather := filepath.Join(deps, "weather")
	for _, dir := range []string{common, weather} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		runTestGit(t, dir, "init", "-q")
	}
	for path, body := range map[string]string{
		filepath.Join(common, "protobuf.mod"):   fmt.Sprintf("module %s\n", common),
		filepath.Join(common, "common.proto"):   "syntax = \"proto3\";\npackage common.v1;\nmessage Common {}\n",
		filepath.Join(weather, "protobuf.mod"):  fmt.Sprintf("module %s\nrequire %s v1.0.0\n", weather, common),
		filepath.Join(weather, "weather.proto"): "syntax = \"proto3\";\npackage weather.v1;\nimport \"common.proto\";\nmessage Weather { common.v1.Common common = 1; }\n",
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{common, weather} {
		runTestGit(t, dir, "add", ".")
		runTestGit(t, dir, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
		runTestGit(t, dir, "tag", "v1.0.0")
	}
	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", weather)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "user.proto"), []byte("syntax = \"proto3\";\nimport \"weather.proto\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	ctx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	if err := (Mod{}).Tidy(ctx); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	if err != nil {
		t.Fatal(err)
	}
	want := common + " v1.0.0 // indirect"
	if !strings.Contains(string(raw), want) {
		t.Fatalf("missing transitive requirement %q in:\n%s", want, raw)
	}
	if err := (Mod{}).Tidy(ctx); err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(again) {
		t.Fatalf("tidy changed an already tidy manifest:\n%s\n---\n%s", raw, again)
	}
}

func TestTidyV1MixedBufAndLegacyEasyPDependencies(t *testing.T) {
	base := t.TempDir()
	common := filepath.Join(base, "common")
	legacy := filepath.Join(base, "legacy")
	buf := filepath.Join(base, "buf")
	for _, dir := range []string{common, legacy, buf} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		runTestGit(t, dir, "init", "-q")
	}
	files := map[string]string{
		filepath.Join(common, "protobuf.mod"):           fmt.Sprintf("module %s\n", common),
		filepath.Join(common, "common.proto"):           "syntax = \"proto3\";\npackage common.v1;\nmessage Common {}\n",
		filepath.Join(legacy, "protobuf.mod"):           fmt.Sprintf("direct (\n  %s@v1.0.0\n)\n", common),
		filepath.Join(legacy, "easyp.yaml"):             "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n",
		filepath.Join(legacy, "proto", "weather.proto"): "syntax = \"proto3\";\nimport \"common.proto\";\nmessage Weather {}\n",
		filepath.Join(buf, "buf.yaml"):                  "version: v2\nmodules:\n  - path: schemas\n",
		filepath.Join(buf, "schemas", "buf.proto"):      "syntax = \"proto3\";\nmessage BufItem {}\n",
	}
	for path, body := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{common, legacy, buf} {
		runTestGit(t, dir, "add", ".")
		runTestGit(t, dir, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
		runTestGit(t, dir, "tag", "v1.0.0")
	}
	root := t.TempDir()
	manifest := fmt.Sprintf("module example.com/root\nrequire (\n  %s v1.0.0\n  %s v1.0.0\n)\n", legacy, buf)
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.proto"), []byte("syntax = \"proto3\";\nimport \"weather.proto\";\nimport \"buf.proto\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cacheBase := filepath.Join(t.TempDir(), "cache")
	t.Setenv("EASYPPATH", cacheBase)
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	ctx := &cli.Context{Context: context.Background(), App: &cli.App{Metadata: map[string]any{}}}
	if err := (Mod{}).Tidy(ctx); err != nil {
		t.Fatal(err)
	}
	lock, err := readV1Lock(filepath.Join(root, "protobuf.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Modules) != 3 {
		t.Fatalf("lock should contain Buf, legacy EasyP, and transitive module: %#v", lock.Modules)
	}
	updated, err := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), common+" v1.0.0 // indirect") {
		t.Fatalf("transitive old EasyP dependency missing from manifest: %s", updated)
	}
	roots, err := rootsFromV1Lock(lock, filepath.Join(cacheBase, "v1", "git"))
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 3 {
		t.Fatalf("dependency roots = %v", roots)
	}
}

func TestRewriteV1RequiredVersionsHTTPSIdentity(t *testing.T) {
	original := []byte("module https://example.com/user.git\nrequire https://example.com/common.git v1.0.0 // indirect\n")
	updated, err := rewriteV1RequiredVersions(original, map[string]string{"https://example.com/common.git": "v1.1.0"})
	if err != nil {
		t.Fatal(err)
	}
	want := "require https://example.com/common.git v1.1.0 // indirect"
	if !strings.Contains(string(updated), want) {
		t.Fatalf("missing %q in %s", want, updated)
	}
}

func TestAugmentV1ManifestMarksDirectTransitiveImport(t *testing.T) {
	root := t.TempDir()
	cache := t.TempDir()
	entry := v1.LockedModule{Source: "example.com/common", Version: "v1.0.0", Commit: strings.Repeat("a", 40)}
	install := v1ModuleCachePath(cache, entry)
	if err := os.MkdirAll(install, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(install, "protobuf.mod"), []byte("module example.com/common\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(install, "common.proto"), []byte("syntax = \"proto3\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "root.proto"), []byte("syntax = \"proto3\";\nimport \"common.proto\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	original := []byte("module example.com/root\n")
	updated, err := augmentV1ManifestRequirements(original, root, v1.Module{Name: "example.com/root", Roots: []string{"."}}, v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "require example.com/common v1.0.0\n") || strings.Contains(string(updated), "// indirect") {
		t.Fatalf("root-imported module should be direct: %s", updated)
	}
}

func runTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}
