package core

// Standing constraint (a) — ROADMAP backlog 999.1: the disk-instrumentation
// cell. readSource is the linked engine's single read pass; this test proves
// the no-disk-reread guarantee behaviorally: on a compiling corpus every
// corpus file is opened from disk exactly once, total. Corpus files imported
// by other corpus files are served from the link pass's in-memory sources
// (lintResolver), so an intra-corpus import costs no second Open.

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

// countingWalker counts every Open by name. It embeds *fs.FSWalker and
// overrides only Open; the wrapper value itself satisfies DirWalker via the
// embedding plus this override. Opens happen in parallel (read pass), so the
// counter is mutex-guarded.
type countingWalker struct {
	*fs.FSWalker

	mu    sync.Mutex
	opens map[string]int
}

func (w *countingWalker) Open(name string) (io.ReadCloser, error) {
	w.mu.Lock()
	w.opens[name]++
	w.mu.Unlock()
	return w.FSWalker.Open(name)
}

func (w *countingWalker) count(name string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.opens[name]
}

func (w *countingWalker) total() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	sum := 0
	for _, n := range w.opens {
		sum += n
	}
	return sum
}

func TestReadSourceOnce(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Compiling two-file corpus: b.proto imports a.proto, both linted. A
	// package-scope message reference keeps the corpus genuinely linked.
	files := map[string]string{
		"a.proto": "syntax = \"proto3\";\n\npackage once;\n\nmessage A {}\n",
		"b.proto": "syntax = \"proto3\";\n\npackage once;\n\nimport \"a.proto\";\n\nmessage B {\n\tA a = 1;\n}\n",
	}
	for name, src := range files {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(src), 0o644))
	}

	walker := &countingWalker{FSWalker: fs.NewFSWalker(root, "."), opens: map[string]int{}}

	// Literal rule-name list: package core cannot import internal/rules (the
	// rules package imports core — an import cycle; 03-01 precedent).
	app := New(Options{
		Logger:              logger.NewNop(),
		AllowCommentIgnores: true,
		KnownLintRules:      []string{"PACKAGE_DIRECTORY_MATCH", "MESSAGE_PASCAL_CASE"},
	})

	issues, err := app.Lint(t.Context(), walker)
	require.NoError(t, err)
	require.Empty(t, issues, "compiling corpus, no rules wired: no findings expected")

	for name := range files {
		require.Equal(t, 1, walker.count(name), "corpus file %s must be opened from disk exactly once", name)
	}
	require.Equal(t, 2, walker.total(), "total disk opens == corpus size; an intra-corpus import must not re-read a linted file")
}
