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
// manifest at that suffix confirms the candidate at the requested revision.
func v1GitModuleCandidates(source string) ([]v1GitModuleCandidate, error) {
	major, err := v1.ModulePathMajor(source)
	if err != nil {
		return nil, fmt.Errorf("ModulePathMajor: %w", err)
	}
	// A slash major suffix identifies the module, not the repository or tag prefix.
	if strings.HasPrefix(major, "/") && !strings.Contains(source, "://") {
		source = strings.TrimSuffix(source, major)
	}
	if filepath.IsAbs(source) {
		return localV1GitModuleCandidates(source), nil
	}
	if strings.Contains(source, "://") {
		parsed, err := url.Parse(source)
		if err != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("invalid Git module URL %q", source)
		}
		if strings.HasPrefix(major, "/") {
			parsed.Path = strings.TrimSuffix(parsed.Path, major)
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
	candidates := []v1GitModuleCandidate{{remote: clean}}

	if localGitRoot(clean) {
		return candidates
	}
	// The requested revision may predate removal of the module directory or
	// replacement of any path component by a file. ReadGitDependencyAt verifies
	// identity in that revision; a repository candidate alone does not make an
	// unrelated directory a module.
	for remote := filepath.Dir(clean); remote != filepath.Dir(remote); remote = filepath.Dir(remote) {
		if localGitRoot(remote) {
			return appendLocalV1GitCandidate(candidates, remote, clean)
		}
	}
	return candidates
}

func localGitRoot(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func appendLocalV1GitCandidate(candidates []v1GitModuleCandidate, remote, clean string) []v1GitModuleCandidate {
	subdir, err := filepath.Rel(remote, clean)
	if err != nil || !filepath.IsLocal(subdir) {
		return candidates
	}
	return append(candidates, v1GitModuleCandidate{remote: remote, subdir: subdir})
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
		if slices.Contains(versions, strings.TrimSuffix(version, "+incompatible")) {
			return candidate, nil
		}
	}
	if !foundRepository && firstErr != nil {
		return v1GitModuleCandidate{}, fmt.Errorf("could not query Git tags for %s@%s; check repository access and credentials: %w", source, version, firstErr)
	}
	return v1GitModuleCandidate{}, fmt.Errorf("git tag %s@%s was not found", source, version)
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
		pathMajor, err := v1.ModulePathMajor(source)
		if err != nil {
			return nil, fmt.Errorf("ModulePathMajor: %w", err)
		}
		for i, version := range versions {
			if pathMajor == "" && semver.Major(version) != "v0" && semver.Major(version) != "v1" && semver.Build(version) == "" {
				// Enumeration is provisional. Fetch verifies pre-native metadata
				// before this compatibility marker can enter a published lock.
				versions[i] = version + "+incompatible"
			}
		}
		versions = slices.DeleteFunc(versions, func(version string) bool {
			return v1.ValidateModuleVersion(source, version) != nil
		})
		if len(versions) > 0 {
			return versions, nil
		}
	}
	if !foundRepository && firstErr != nil {
		return nil, fmt.Errorf("git tags for %s: %w", source, firstErr)
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
		if semver.IsValid(tag) && semver.Build(tag) != "+incompatible" {
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
