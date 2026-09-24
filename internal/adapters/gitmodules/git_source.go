package gitmodules

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/semver"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type v1GitModuleCandidate struct {
	remote string
	subdir string
}

func (candidate v1GitModuleCandidate) tag(version string) string {
	if candidate.subdir == "" {
		return version
	}
	return filepath.ToSlash(candidate.subdir) + "/" + version
}

// v1GitModuleCandidates tries the complete module identity first, then each
// parent Git URL with the removed suffix as its module directory. The module
// manifest at that suffix confirms the candidate for untagged requirements.
func v1GitModuleCandidates(source string) ([]v1GitModuleCandidate, error) {
	if filepath.IsAbs(source) {
		return localV1GitModuleCandidates(source), nil
	}
	if strings.Contains(source, "://") {
		parsed, err := url.Parse(source)
		if err != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("invalid Git module URL %q", source)
		}
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if !validV1GitModuleParts(parts) {
			return nil, fmt.Errorf("invalid Git module URL %q", source)
		}
		return pathV1GitModuleCandidates(parts, func(prefix []string) string {
			candidate := *parsed
			candidate.Path = "/" + strings.Join(prefix, "/")
			candidate.RawPath = ""
			return candidate.String()
		}, 1), nil
	}
	parts := strings.Split(strings.Trim(source, "/"), "/")
	if len(parts) < 2 || !validV1GitModuleParts(parts) {
		return nil, fmt.Errorf("invalid Git module identity %q", source)
	}
	return pathV1GitModuleCandidates(parts, func(prefix []string) string {
		return "https://" + strings.Join(prefix, "/")
	}, 2), nil
}

func validV1GitModuleParts(parts []string) bool {
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return len(parts) > 0
}

func localV1GitModuleCandidates(source string) []v1GitModuleCandidate {
	clean := filepath.Clean(source)
	var candidates []v1GitModuleCandidate
	for remote := clean; remote != filepath.Dir(remote); remote = filepath.Dir(remote) {
		subdir, err := filepath.Rel(remote, clean)
		if err != nil {
			continue
		}
		if subdir == "." {
			subdir = ""
		}
		candidates = append(candidates, v1GitModuleCandidate{remote: remote, subdir: subdir})
	}
	return candidates
}

func pathV1GitModuleCandidates(parts []string, remote func([]string) string, minParts int) []v1GitModuleCandidate {
	var candidates []v1GitModuleCandidate
	for length := len(parts); length >= minParts; length-- {
		candidates = append(candidates, v1GitModuleCandidate{
			remote: remote(parts[:length]),
			subdir: strings.Join(parts[length:], "/"),
		})
	}
	return candidates
}

func findV1GitModuleTag(ctx context.Context, source, version string) (v1GitModuleCandidate, error) {
	candidates, err := v1GitModuleCandidates(source)
	if err != nil {
		return v1GitModuleCandidate{}, fmt.Errorf("v1GitModuleCandidates: %w", err)
	}
	var firstErr error
	foundRepository := false
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return v1GitModuleCandidate{}, fmt.Errorf("Err: %w", err)
		}
		versions, err := listV1CandidateTags(ctx, candidate)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		foundRepository = true
		if slices.Contains(versions, version) {
			return candidate, nil
		}
	}
	if !foundRepository && firstErr != nil {
		return v1GitModuleCandidate{}, fmt.Errorf("Git tag %s@%s was not found: %w", source, version, firstErr)
	}
	return v1GitModuleCandidate{}, fmt.Errorf("Git tag %s@%s was not found", source, version)
}

func listV1ModuleTags(ctx context.Context, source string) ([]string, error) {
	candidates, err := v1GitModuleCandidates(source)
	if err != nil {
		return nil, fmt.Errorf("v1GitModuleCandidates: %w", err)
	}
	var firstErr error
	foundRepository := false
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("Err: %w", err)
		}
		versions, err := listV1CandidateTags(ctx, candidate)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		foundRepository = true
		if len(versions) > 0 {
			return versions, nil
		}
	}
	if !foundRepository && firstErr != nil {
		return nil, fmt.Errorf("Git tags for %s: %w", source, firstErr)
	}
	return nil, nil
}

func listV1CandidateTags(ctx context.Context, candidate v1GitModuleCandidate) ([]string, error) {
	raw, err := gitV1(ctx, "", "ls-remote", "--tags", "--", candidate.remote)
	if err != nil {
		return nil, fmt.Errorf("ls-remote %s: %w", candidate.remote, err)
	}
	seen := make(map[string]struct{})
	for _, line := range strings.Split(raw, "\n") {
		_, ref, ok := strings.Cut(line, "\t")
		if !ok || !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}
		tag := strings.TrimSuffix(strings.TrimPrefix(ref, "refs/tags/"), "^{}")
		if candidate.subdir != "" {
			prefix := filepath.ToSlash(candidate.subdir) + "/"
			if !strings.HasPrefix(tag, prefix) {
				continue
			}
			tag = strings.TrimPrefix(tag, prefix)
		}
		if semver.IsValid(tag) {
			seen[tag] = struct{}{}
		}
	}
	versions := make([]string, 0, len(seen))
	for version := range seen {
		versions = append(versions, version)
	}
	slices.Sort(versions)
	return versions, nil
}

// cloneHeadV1GitModule finds the repository default branch containing source.
// The caller owns the returned checkout and must remove it.
func cloneHeadV1GitModule(ctx context.Context, source, cacheRoot string) (string, v1.Module, string, error) {
	candidates, err := v1GitModuleCandidates(source)
	if err != nil {
		return "", v1.Module{}, "", err
	}
	var firstErr, moduleErr error
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return "", v1.Module{}, "", err
		}
		checkout, err := os.MkdirTemp(cacheRoot, "git-*")
		if err != nil {
			return "", v1.Module{}, "", err
		}
		module, commit, cloned, err := checkoutHeadV1GitModuleCandidate(ctx, checkout, source, candidate)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			if cloned {
				moduleErr = err
			}
			if removeErr := os.RemoveAll(checkout); removeErr != nil {
				return "", v1.Module{}, "", removeErr
			}
			continue
		}
		return checkout, module, commit, nil
	}
	if moduleErr != nil {
		return "", v1.Module{}, "", fmt.Errorf("module %s was not found at a Git repository HEAD: %w", source, moduleErr)
	}
	if firstErr != nil {
		return "", v1.Module{}, "", fmt.Errorf("Git repository for %s was not found: %w", source, firstErr)
	}
	return "", v1.Module{}, "", fmt.Errorf("no Git repository candidate for %s", source)
}

func checkoutHeadV1GitModuleCandidate(ctx context.Context, checkout, source string, candidate v1GitModuleCandidate) (v1.Module, string, bool, error) {
	if _, err := gitV1(ctx, "", "clone", "--quiet", "--depth=1", "--no-checkout", "--", candidate.remote, checkout); err != nil {
		return v1.Module{}, "", false, err
	}
	commit, err := gitV1(ctx, checkout, "rev-parse", "HEAD")
	if err != nil {
		return v1.Module{}, "", true, err
	}
	commit = strings.TrimSpace(commit)
	if _, err := gitV1(ctx, checkout, "checkout", "--quiet", "--detach", commit); err != nil {
		return v1.Module{}, "", true, err
	}
	module, err := moduleconfig.ReadGitDependencyAt(checkout, source, candidate.subdir)
	if err != nil {
		return v1.Module{}, "", true, err
	}
	return module, commit, true, nil
}

// clonePinnedV1GitModule finds a repository containing the locked commit and
// the requested module. It does not depend on the tag still pointing there.
func clonePinnedV1GitModule(ctx context.Context, entry v1.LockedModule, cacheRoot string) (string, error) {
	candidates, err := v1GitModuleCandidates(entry.Source)
	if err != nil {
		return "", fmt.Errorf("v1GitModuleCandidates: %w", err)
	}
	var firstErr error
	var moduleErr error
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		checkout, err := os.MkdirTemp(cacheRoot, "git-*")
		if err != nil {
			return "", fmt.Errorf("MkdirTemp: %w", err)
		}
		cloned, err := checkoutPinnedV1GitModuleCandidate(ctx, checkout, entry, candidate)
		if err != nil {
			if cloned {
				moduleErr = err
			}
			if firstErr == nil {
				firstErr = err
			}
			if removeErr := os.RemoveAll(checkout); removeErr != nil {
				return "", fmt.Errorf("RemoveAll: %w", removeErr)
			}
			continue
		}
		return checkout, nil
	}
	if moduleErr != nil {
		return "", fmt.Errorf("%s: %w", entry.Source, moduleErr)
	}
	if firstErr == nil {
		return "", fmt.Errorf("%s: no Git repository candidate", entry.Source)
	}
	return "", fmt.Errorf("%s: locked commit %s is unavailable: %w", entry.Source, entry.Commit, firstErr)
}

func checkoutPinnedV1GitModuleCandidate(ctx context.Context, checkout string, entry v1.LockedModule, candidate v1GitModuleCandidate) (bool, error) {
	if _, err := gitV1(ctx, "", "clone", "--quiet", "--no-checkout", "--", candidate.remote, checkout); err != nil {
		return false, err
	}
	if _, err := gitV1(ctx, checkout, "rev-parse", "--verify", entry.Commit+"^{commit}"); err != nil {
		return true, err
	}
	if _, err := gitV1(ctx, checkout, "checkout", "--quiet", "--detach", entry.Commit); err != nil {
		return true, err
	}
	if _, err := moduleconfig.ReadGitDependencyAt(checkout, entry.Source, candidate.subdir); err != nil {
		return true, err
	}
	return true, nil
}
