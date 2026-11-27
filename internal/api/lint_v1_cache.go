package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
	"github.com/easyp-tech/easyp/internal/workspace"
)

// lintDirCache memoizes per-directory lookups for one lint run. The module
// directory and the effective policy depend only on the files above a
// directory, so every file in it shares them; errors are cached too, so each
// file still reports the first failure on its ancestor chain.
type lintDirCache struct {
	projectRoot, configPath string
	reader                  *sessionFileReader

	files    map[string]cachedPolicyFile
	modules  map[string]cachedModuleDir
	policies map[string]cachedLintPolicy
}

type cachedPolicyFile struct {
	file  v1PolicyFile
	found bool
	err   error
}

type cachedBoundary struct {
	dir string
	err error
}

type cachedModuleDir struct {
	dir string
	err error
}

type cachedLintPolicy struct {
	policy           v1.Policy
	sources          v1LintPolicySources
	lintersPresence  v1.PolicyPresence
	settingsPresence v1.PolicyPresence
	err              error
}

func newLintDirCache(projectRoot, configPath string, session *sourceview.Session) *lintDirCache {
	return &lintDirCache{
		projectRoot: projectRoot,
		configPath:  configPath,
		reader:      newSessionFileReader(session),
		files:       map[string]cachedPolicyFile{},
		modules:     map[string]cachedModuleDir{},
		policies:    map[string]cachedLintPolicy{},
	}
}

func (c *lintDirCache) moduleDir(dir string) (string, error) {
	return cachedFindV1PolicyModuleDir(c.modules, c.projectRoot, dir)
}

// cachedFindV1PolicyModuleDir is findV1PolicyModuleDir memoized in cache.
func cachedFindV1PolicyModuleDir(cache map[string]cachedModuleDir, projectRoot, dir string) (string, error) {
	if cached, ok := cache[dir]; ok {
		return cached.dir, cached.err
	}
	found, err := findV1PolicyModuleDir(projectRoot, dir)
	cache[dir] = cachedModuleDir{dir: found, err: err}
	return found, err
}

func (c *lintDirCache) lintPolicy(dir string) (v1.Policy, v1LintPolicySources, v1.PolicyPresence, v1.PolicyPresence, error) {
	if cached, ok := c.policies[dir]; ok {
		return cached.policy, cached.sources, cached.lintersPresence, cached.settingsPresence, cached.err
	}
	policy, sources, linters, settings, err := resolveV1LintPolicyWith(c.policyFile, dir, c.projectRoot, c.configPath)
	c.policies[dir] = cachedLintPolicy{policy, sources, linters, settings, err}
	return policy, sources, linters, settings, err
}

func (c *lintDirCache) policyFile(path string) (v1PolicyFile, bool, error) {
	if cached, ok := c.files[path]; ok {
		return cached.file, cached.found, cached.err
	}
	file, found, err := readV1PolicyFileWith(c.readOptionalFile, path)
	c.files[path] = cachedPolicyFile{file: file, found: found, err: err}
	return file, found, err
}

// readOptionalFile is readOptionalFile reading through the run's session.
func (c *lintDirCache) readOptionalFile(path string) ([]byte, bool, error) {
	raw, err := c.reader.readFileAt(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("ReadFile: %w", err)
	}
	return raw, true, nil
}

// sessionFileReader reads local metadata like workspace.ReadFileAt, through a
// run's session and with the workspace boundary of each directory looked up
// once. It is not safe for concurrent use.
type sessionFileReader struct {
	session    *sourceview.Session
	boundaries map[string]cachedBoundary
}

func newSessionFileReader(session *sourceview.Session) *sessionFileReader {
	return &sessionFileReader{session: session, boundaries: map[string]cachedBoundary{}}
}

func (r *sessionFileReader) readFileAt(path string) ([]byte, error) {
	dir := filepath.Dir(path)
	boundary, ok := r.boundaries[dir]
	if !ok {
		boundary.dir, boundary.err = workspace.Boundary(dir)
		r.boundaries[dir] = boundary
	}
	if boundary.err != nil {
		return nil, fmt.Errorf("Boundary: %w", boundary.err)
	}
	relative, err := filepath.Rel(boundary.dir, path)
	if err != nil {
		return nil, fmt.Errorf("Rel: %w", err)
	}
	return r.session.ReadLocal(context.Background(), boundary.dir, relative)
}

// backgroundChecks runs checks concurrently with the caller and reports the
// first failure in the order the checks were started.
type backgroundChecks struct {
	pending []chan error
	err     error
}

func (b *backgroundChecks) start(check func() error) {
	result := make(chan error, 1)
	b.pending = append(b.pending, result)
	go func() { result <- check() }()
}

// wait blocks until every started check has finished.
func (b *backgroundChecks) wait() error {
	for _, result := range b.pending {
		if err := <-result; err != nil && b.err == nil {
			b.err = err
		}
	}
	b.pending = nil
	return b.err
}
