package sourceview

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
)

var errCacheExhausted = errors.New("directory handle cache exhausted")

// dirCache bounds the total number of cached child *os.Root directory handles
// across all rootTrees in a Session. Entries are reference-counted so evicting
// a directory never closes a handle while another goroutine is using it.
type dirCache struct {
	mu        sync.Mutex
	maxDirs   int
	head      *cachedDir
	tail      *cachedDir
	count     int
	closeErrs []error
}

type cachedDir struct {
	tree    *rootTree
	name    string
	handle  *os.Root
	refs    int
	evicted bool
	inLRU   bool
	prev    *cachedDir
	next    *cachedDir
}

func newDirCache(maxDirs int) *dirCache {
	return &dirCache{maxDirs: maxDirs}
}

func (c *dirCache) pushFront(entry *cachedDir) {
	entry.prev = nil
	entry.next = c.head
	if c.head != nil {
		c.head.prev = entry
	} else {
		c.tail = entry
	}
	c.head = entry
	entry.inLRU = true
	c.count++
}

func (c *dirCache) removeLRU(entry *cachedDir) {
	if !entry.inLRU {
		return
	}
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		c.head = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		c.tail = entry.prev
	}
	entry.prev = nil
	entry.next = nil
	entry.inLRU = false
	c.count--
}

func (c *dirCache) moveToFront(entry *cachedDir) {
	if !entry.inLRU || c.head == entry {
		return
	}
	c.removeLRU(entry)
	c.pushFront(entry)
}

func (c *dirCache) evictOneLocked() *os.Root {
	// Prefer an idle directory handle so eviction closes an FD immediately.
	victim := c.tail
	for candidate := c.tail; candidate != nil; candidate = candidate.prev {
		if candidate.refs == 0 {
			victim = candidate
			break
		}
	}
	if victim == nil {
		return nil
	}
	c.removeLRU(victim)
	delete(victim.tree.dirs, victim.name)
	victim.evicted = true
	if victim.refs == 0 {
		return victim.handle
	}
	return nil
}

func (c *dirCache) evictAllIdle() {
	var toClose []*os.Root
	c.mu.Lock()
	for entry := c.tail; entry != nil; {
		prev := entry.prev
		if entry.refs == 0 {
			c.removeLRU(entry)
			delete(entry.tree.dirs, entry.name)
			entry.evicted = true
			toClose = append(toClose, entry.handle)
		}
		entry = prev
	}
	c.mu.Unlock()
	c.closeHandles(toClose)
}

func (c *dirCache) release(entry *cachedDir) {
	if entry == nil {
		return
	}
	c.mu.Lock()
	entry.refs--
	shouldClose := entry.refs == 0 && entry.evicted
	c.mu.Unlock()
	if shouldClose {
		c.recordCloseErr(entry.handle.Close())
	}
}

func (c *dirCache) closeHandles(handles []*os.Root) {
	for _, handle := range handles {
		c.recordCloseErr(handle.Close())
	}
}

func (c *dirCache) recordCloseErr(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	c.closeErrs = append(c.closeErrs, err)
	c.mu.Unlock()
}

func (c *dirCache) drainCloseErrs() []error {
	c.mu.Lock()
	errs := c.closeErrs
	c.closeErrs = nil
	c.mu.Unlock()
	return errs
}

func (c *dirCache) cachedCount() int {
	c.mu.Lock()
	count := c.count
	c.mu.Unlock()
	return count
}

func isTooManyOpenFiles(err error) bool {
	return errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE)
}

// rootTree serves an os.Root by slash-separated name through a bounded LRU of
// directory handles, so repeated lookups below the same root cost one system
// call instead of one per path component without risking FD exhaustion.
type rootTree struct {
	root     *os.Root
	cache    *dirCache
	closed   bool
	rootRefs int
	dirs     map[string]*cachedDir
}

func newRootTree(root *os.Root, cache *dirCache) *rootTree {
	if cache == nil {
		cache = newDirCache(defaultSessionMaxDirs)
	}
	return &rootTree{
		root:  root,
		cache: cache,
		dirs:  make(map[string]*cachedDir),
	}
}

func (t *rootTree) acquireRoot() (*os.Root, func(), error) {
	t.cache.mu.Lock()
	if t.closed {
		t.cache.mu.Unlock()
		return nil, nil, fs.ErrClosed
	}
	t.rootRefs++
	t.cache.mu.Unlock()
	return t.root, t.releaseRoot, nil
}

func (t *rootTree) releaseRoot() {
	t.cache.mu.Lock()
	t.rootRefs--
	shouldClose := t.closed && t.rootRefs == 0
	t.cache.mu.Unlock()
	if shouldClose {
		t.cache.recordCloseErr(t.root.Close())
	}
}

func (t *rootTree) acquireDirHandle(name string) (*os.Root, func(), error) {
	if name == "." {
		return t.acquireRoot()
	}
	entry, err := t.acquireDir(name)
	if err != nil {
		return nil, nil, err
	}
	return entry.handle, func() { t.cache.release(entry) }, nil
}

func (t *rootTree) acquireDir(name string) (*cachedDir, error) {
	t.cache.mu.Lock()
	if t.closed {
		t.cache.mu.Unlock()
		return nil, fs.ErrClosed
	}
	if entry, ok := t.dirs[name]; ok {
		entry.refs++
		t.cache.moveToFront(entry)
		t.cache.mu.Unlock()
		return entry, nil
	}
	t.cache.mu.Unlock()

	parent, releaseParent, err := t.acquireDirHandle(path.Dir(name))
	if err != nil {
		return nil, err
	}
	base := path.Base(name)
	handle, err := parent.OpenRoot(base)
	if err != nil && isTooManyOpenFiles(err) {
		t.cache.evictAllIdle()
		handle, err = parent.OpenRoot(base)
	}
	releaseParent()
	if err != nil {
		if isTooManyOpenFiles(err) {
			return nil, errCacheExhausted
		}
		return nil, err
	}

	var toClose []*os.Root
	t.cache.mu.Lock()
	if t.closed {
		t.cache.mu.Unlock()
		t.cache.recordCloseErr(handle.Close())
		return nil, fs.ErrClosed
	}
	if existing, ok := t.dirs[name]; ok {
		existing.refs++
		t.cache.moveToFront(existing)
		t.cache.mu.Unlock()
		t.cache.recordCloseErr(handle.Close())
		return existing, nil
	}
	if t.cache.maxDirs <= 0 {
		t.cache.mu.Unlock()
		return &cachedDir{tree: t, name: name, handle: handle, refs: 1, evicted: true}, nil
	}
	entry := &cachedDir{tree: t, name: name, handle: handle, refs: 1}
	t.dirs[name] = entry
	t.cache.pushFront(entry)
	for t.cache.count > t.cache.maxDirs {
		if evicted := t.cache.evictOneLocked(); evicted != nil {
			toClose = append(toClose, evicted)
		} else {
			break
		}
	}
	t.cache.mu.Unlock()
	t.cache.closeHandles(toClose)
	return entry, nil
}

func (t *rootTree) acquireTarget(op, name string) (*os.Root, string, func(), error) {
	if !fs.ValidPath(name) {
		return nil, "", nil, &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		root, release, err := t.acquireRoot()
		if err != nil {
			return nil, "", nil, pathError(op, name, err)
		}
		return root, ".", release, nil
	}
	dirName := path.Dir(name)
	if dirName == "." {
		root, release, err := t.acquireRoot()
		if err != nil {
			return nil, "", nil, pathError(op, name, err)
		}
		return root, path.Base(name), release, nil
	}
	entry, err := t.acquireDir(dirName)
	if errors.Is(err, errCacheExhausted) {
		root, release, rootErr := t.acquireRoot()
		if rootErr != nil {
			return nil, "", nil, pathError(op, name, rootErr)
		}
		return root, filepath.FromSlash(name), release, nil
	}
	if err != nil {
		return nil, "", nil, pathError(op, name, err)
	}
	return entry.handle, path.Base(name), func() { t.cache.release(entry) }, nil
}

func (t *rootTree) Open(name string) (fs.File, error) {
	handle, target, release, err := t.acquireTarget("open", name)
	if err != nil {
		return nil, err
	}
	defer release()
	file, err := handle.Open(target)
	if err != nil {
		return nil, pathError("open", name, err)
	}
	return file, nil
}

func (t *rootTree) Lstat(name string) (fs.FileInfo, error) {
	handle, target, release, err := t.acquireTarget("lstat", name)
	if err != nil {
		return nil, err
	}
	defer release()
	info, err := handle.Lstat(target)
	if err != nil {
		return nil, pathError("lstat", name, err)
	}
	return info, nil
}

func (t *rootTree) ReadLink(name string) (string, error) {
	handle, target, release, err := t.acquireTarget("readlink", name)
	if err != nil {
		return "", err
	}
	defer release()
	linkTarget, err := handle.Readlink(target)
	if err != nil {
		return "", pathError("readlink", name, err)
	}
	return linkTarget, nil
}

func (t *rootTree) ReadDir(name string) ([]fs.DirEntry, error) {
	file, err := t.Open(name)
	if err != nil {
		return nil, err
	}
	dir, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, errors.Join(&fs.PathError{Op: "readdir", Path: name, Err: errors.New("not implemented")}, file.Close())
	}
	entries, err := dir.ReadDir(-1)
	slices.SortFunc(entries, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, errors.Join(err, file.Close())
}

// Close closes every idle cached directory handle and the root itself once
// in-flight operations release their references.
func (t *rootTree) Close() error {
	var toClose []*os.Root
	t.cache.mu.Lock()
	if t.closed {
		t.cache.mu.Unlock()
		return nil
	}
	t.closed = true
	for name, entry := range t.dirs {
		t.cache.removeLRU(entry)
		delete(t.dirs, name)
		entry.evicted = true
		if entry.refs == 0 {
			toClose = append(toClose, entry.handle)
		}
	}
	closeRoot := t.rootRefs == 0
	t.cache.mu.Unlock()

	var errs []error
	for _, handle := range toClose {
		errs = append(errs, handle.Close())
	}
	if closeRoot {
		errs = append(errs, t.root.Close())
	}
	return errors.Join(errs...)
}

// pathError reports the full name instead of the handle-relative base.
func pathError(op, name string, err error) error {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return &fs.PathError{Op: op, Path: name, Err: pathErr.Err}
	}
	return &fs.PathError{Op: op, Path: name, Err: err}
}
