package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Update refreshes tagged requirements within their major versions and
// versionless requirements to Git HEAD; explicit commit pins remain fixed.
func (m Mod) Update(ctx *cli.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	original, module, err := readV1Manifest(root)
	if err != nil {
		return fmt.Errorf("readV1Manifest: %w", err)
	}
	if len(module.Replaces) > 0 {
		return fmt.Errorf("module %s: remove local replacements before updating a reproducible lock", module.Name)
	}
	callCtx := ctx.Context
	if callCtx == nil {
		callCtx = context.Background()
	}
	updatedVersions := make(map[string]string, len(module.Requires))
	for _, requirement := range module.Requires {
		version, err := latestCompatibleV1Tag(callCtx, requirement.Module, requirement.Version)
		if err != nil {
			return fmt.Errorf("latestCompatibleV1Tag: %w", err)
		}
		updatedVersions[requirement.Module] = version
	}
	updated, err := rewriteV1RequiredVersions(original, updatedVersions)
	if err != nil {
		return fmt.Errorf("rewriteV1RequiredVersions: %w", err)
	}
	updatedModule, err := v1.ParseModule(strings.NewReader(string(updated)))
	if err != nil {
		return fmt.Errorf("ParseModule: %w", err)
	}
	lock, err := resolveV1Lock(ctx, root, updatedModule)
	if err != nil {
		return fmt.Errorf("resolveV1Lock: %w", err)
	}
	cacheRoot, err := getEasypPath(getLogger(ctx))
	if err != nil {
		return fmt.Errorf("getEasypPath: %w", err)
	}
	updated, err = augmentV1ManifestRequirements(updated, root, updatedModule, lock, filepath.Join(cacheRoot, "v1", "git"))
	if err != nil {
		return fmt.Errorf("augmentV1ManifestRequirements: %w", err)
	}
	return writeV1ResolvedFiles(root, original, updated, lock)
}

func latestCompatibleV1Tag(ctx context.Context, source, current string) (string, error) {
	if current == "" {
		return "", nil // versionless requirements resolve the current Git HEAD
	}
	if v1.IsCommitRef(current) {
		return current, nil
	}
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

func listV1Tags(ctx context.Context, source string) ([]string, error) {
	return listV1ModuleTags(ctx, source)
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
