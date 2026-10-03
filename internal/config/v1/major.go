package v1

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// ModulePathMajor extracts the Go major suffix from a logical module identity.
// Transport spelling is preserved: URLs and absolute fixture paths need not be
// valid Go import paths. Only their logical path participates in this check.
func ModulePathMajor(source string) (string, error) {
	logical := filepath.ToSlash(source)
	if strings.Contains(source, "://") {
		parsed, err := url.Parse(source)
		if err != nil {
			return "", fmt.Errorf("Parse: %w", err)
		}
		logical = parsed.Host + parsed.Path
	}
	_, major, ok := module.SplitPathVersion(logical)
	if !ok {
		return "", fmt.Errorf("module %q has an invalid major version suffix; use an unsuffixed path for v0/v1 or /vN for v2 and later (no leading zeros)", source)
	}
	return major, nil
}

// ValidateModuleVersion checks semantic import versioning without converting
// identities or versions. Empty versions and full commit pins still check paths.
func ValidateModuleVersion(source, version string) error {
	major, err := ModulePathMajor(source)
	if err != nil {
		return fmt.Errorf("ModulePathMajor: %w", err)
	}
	if semver.Build(version) == "+incompatible" {
		if major != "" || semver.Major(version) == "v0" || semver.Major(version) == "v1" {
			return fmt.Errorf("%s@%s: +incompatible requires an unsuffixed v2-or-later legacy module", source, version)
		}
		// Metadata provenance is verified on the exact fetched/cached revision.
		// Structural parsing cannot determine whether a native manifest exists.
		return nil
	}
	if !semver.IsValid(version) {
		return nil
	}
	if err := module.CheckPathMajor(version, major); err != nil {
		return fmt.Errorf("module %s@%s: %w; select a published /%s module for v2 or later, or an unsuffixed v0/v1 version; update protobuf.mod explicitly (migration never renames sources or versions)", source, version, err, semver.Major(version))
	}
	return nil
}
