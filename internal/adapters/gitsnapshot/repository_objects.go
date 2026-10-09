package gitsnapshot

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strconv"
	"strings"
)

func (f *repositoryFS) object(kind, id string) ([]byte, error) {
	data, err := f.command(nil, "cat-file", kind, id)
	if err != nil {
		return nil, fmt.Errorf("command: %w", err)
	}
	digest := sha1.New()
	if len(id) == 64 {
		digest = sha256.New()
	}
	_, _ = fmt.Fprintf(digest, "%s %d\x00", kind, len(data))
	_, _ = digest.Write(data)
	if hex.EncodeToString(digest.Sum(nil)) != id {
		return nil, fmt.Errorf("object hash mismatch for Git %s %s", kind, id)
	}
	return data, nil
}

func (f *repositoryFS) objectSize(id string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if size, ok := f.sizes[id]; ok {
		return size, nil
	}
	raw, err := f.command(nil, "cat-file", "-s", id)
	if err != nil {
		return 0, fmt.Errorf("command: %w", err)
	}
	size, err := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("ParseInt: %w", err)
	}
	f.sizes[id] = size
	return size, nil
}

func (f *repositoryFS) verifiedTree(id string) (map[string]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if tree, ok := f.trees[id]; ok {
		return tree, nil
	}
	data, err := f.object("tree", id)
	if err != nil {
		return nil, fmt.Errorf("object: %w", err)
	}
	tree := make(map[string]string)
	width := len(id) / 2
	for len(data) > 0 {
		split := bytes.IndexByte(data, 0)
		if split < 0 || len(data) < split+1+width {
			return nil, fmt.Errorf("invalid Git tree")
		}
		_, name, ok := strings.Cut(string(data[:split]), " ")
		if !ok {
			return nil, fmt.Errorf("invalid Git tree entry")
		}
		tree[name] = hex.EncodeToString(data[split+1 : split+1+width])
		data = data[split+1+width:]
	}
	f.trees[id] = tree
	return tree, nil
}

func (f *repositoryFS) verifyPath(name string) error {
	id := f.entries["."].object
	if name == "." {
		_, err := f.verifiedTree(id)
		return err
	}
	prefix := "."
	for _, part := range strings.Split(name, "/") {
		tree, err := f.verifiedTree(id)
		if err != nil {
			return fmt.Errorf("verifiedTree: %w", err)
		}
		prefix = path.Join(prefix, part)
		entry := f.entries[prefix]
		child, ok := tree[part]
		if !ok || child != entry.object {
			return fmt.Errorf("object hash mismatch in Git path %q", prefix)
		}
		id = child
	}
	return nil
}
