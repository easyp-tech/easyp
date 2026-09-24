package gitmodules

import (
	"context"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorkspaceIdentity returns the origin module identity for a repository root, or an empty string.
func WorkspaceIdentity(ctx context.Context, root string) string {
	repoRoot, err := gitV1(ctx, root, "rev-parse", "--show-toplevel")
	if err != nil || filepath.Clean(strings.TrimSpace(repoRoot)) != filepath.Clean(root) {
		return ""
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "remote", "get-url", "origin")
	raw, err := cmd.Output()
	if err != nil {
		return ""
	}
	remote := strings.TrimSpace(string(raw))
	if strings.Contains(remote, "://") {
		parsed, err := url.Parse(remote)
		if err != nil || parsed.Host == "" {
			return ""
		}
		return strings.TrimSuffix(parsed.Host+"/"+strings.TrimPrefix(parsed.Path, "/"), ".git")
	}
	if _, rest, ok := strings.Cut(remote, "@"); ok {
		host, path, ok := strings.Cut(rest, ":")
		if ok && host != "" && path != "" {
			return strings.TrimSuffix(host+"/"+path, ".git")
		}
	}
	return ""
}
