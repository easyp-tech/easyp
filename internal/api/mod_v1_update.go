package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func (m Mod) updateV1IfPresent(ctx *cli.Context) (bool, error) {
	root, err := os.Getwd()
	if err != nil {
		return true, err
	}
	manifestPath := filepath.Join(root, "protobuf.mod")
	original, err := os.ReadFile(manifestPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	if !v1.IsModuleManifest(original) {
		return false, nil
	}
	module, err := v1.ParseModule(strings.NewReader(string(original)))
	if err != nil {
		return true, err
	}
	if len(module.Replaces) > 0 {
		return true, fmt.Errorf("module %s: remove local replacements before updating a reproducible lock", module.Name)
	}
	callCtx := ctx.Context
	if callCtx == nil {
		callCtx = context.Background()
	}
	updatedVersions := make(map[string]string, len(module.Requires))
	for _, requirement := range module.Requires {
		version, err := latestCompatibleV1Tag(callCtx, requirement.Module, requirement.Version)
		if err != nil {
			return true, err
		}
		updatedVersions[requirement.Module] = version
	}
	updated, err := rewriteV1RequiredVersions(original, updatedVersions)
	if err != nil {
		return true, err
	}
	updatedModule, err := v1.ParseModule(strings.NewReader(string(updated)))
	if err != nil {
		return true, err
	}
	lock, err := resolveV1Lock(ctx, root, updatedModule)
	if err != nil {
		return true, err
	}
	cacheRoot, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return true, err
	}
	updated, err = augmentV1ManifestRequirements(updated, root, updatedModule, lock, filepath.Join(cacheRoot, "v1", "git"))
	if err != nil {
		return true, err
	}
	return true, writeV1ResolvedFiles(root, original, updated, lock)
}

func latestCompatibleV1Tag(ctx context.Context, source, current string) (string, error) {
	if !semver.IsValid(current) {
		return "", fmt.Errorf("%s: invalid required version %q", source, current)
	}
	versions, err := listV1Tags(ctx, source)
	if err != nil {
		return "", err
	}
	best := current
	for _, version := range versions {
		if semver.Major(version) != semver.Major(current) {
			continue
		}
		if semver.Prerelease(current) == "" && semver.Prerelease(version) != "" {
			continue
		}
		if semver.Compare(version, best) > 0 {
			best = version
		}
	}
	return best, nil
}

func latestV1Tag(ctx context.Context, source string) (string, error) {
	versions, err := listV1Tags(ctx, source)
	if err != nil {
		return "", err
	}
	best := ""
	for _, version := range versions {
		if semver.Prerelease(version) != "" {
			continue
		}
		if best == "" || semver.Compare(version, best) > 0 {
			best = version
		}
	}
	if best == "" {
		return "", fmt.Errorf("legacy dependency %s has no stable semver Git tag; declare an explicit version in the dependent repository", source)
	}
	return best, nil
}

func listV1Tags(ctx context.Context, source string) ([]string, error) {
	raw, err := gitV1(ctx, "", "ls-remote", "--tags", "--", v1GitRemote(source))
	if err != nil {
		return nil, fmt.Errorf("list Git tags for %s: %w", source, err)
	}
	seen := make(map[string]struct{})
	for _, line := range strings.Split(raw, "\n") {
		_, ref, ok := strings.Cut(line, "\t")
		if !ok || !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}
		version := strings.TrimSuffix(strings.TrimPrefix(ref, "refs/tags/"), "^{}")
		if semver.IsValid(version) {
			seen[version] = struct{}{}
		}
	}
	versions := make([]string, 0, len(seen))
	for version := range seen {
		versions = append(versions, version)
	}
	slices.Sort(versions)
	return versions, nil
}

func rewriteV1RequiredVersions(original []byte, updates map[string]string) ([]byte, error) {
	lines := strings.Split(string(original), "\n")
	inRequire := false
	for i, line := range lines {
		body, comment, hasComment := splitV1ManifestComment(line)
		trimmed := strings.TrimSpace(body)
		if strings.HasSuffix(trimmed, "(") && strings.TrimSpace(strings.TrimSuffix(trimmed, "(")) == "require" {
			inRequire = true
			continue
		}
		if trimmed == ")" && inRequire {
			inRequire = false
			continue
		}
		fields := strings.Fields(body)
		var source, oldVersion string
		switch {
		case inRequire && len(fields) == 2:
			source, oldVersion = fields[0], fields[1]
		case !inRequire && len(fields) == 3 && fields[0] == "require":
			source, oldVersion = fields[1], fields[2]
		default:
			continue
		}
		version, ok := updates[source]
		if !ok || version == oldVersion {
			continue
		}
		at := strings.LastIndex(body, oldVersion)
		if at < 0 {
			return nil, fmt.Errorf("cannot rewrite require %s", source)
		}
		body = body[:at] + version + body[at+len(oldVersion):]
		if hasComment {
			body += "//" + comment
		}
		lines[i] = body
	}
	return []byte(strings.Join(lines, "\n")), nil
}

func splitV1ManifestComment(line string) (string, string, bool) {
	for i := 0; i+1 < len(line); i++ {
		if line[i] == '/' && line[i+1] == '/' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
			return line[:i], line[i+2:], true
		}
	}
	return line, "", false
}

func writeV1Manifest(root string, raw []byte) error {
	tmp, err := os.CreateTemp(root, ".protobuf.mod-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(root, "protobuf.mod"))
}
