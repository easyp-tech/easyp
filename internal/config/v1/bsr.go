package v1

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	bsrModulePattern = `^[^/:\s]+/[^/:\s]+/[^/:\s]+$`
	bsrDigestPattern = `^[^:\s]+:[0-9a-fA-F]+$`

	// BSRCompatibilitySnapshot identifies a Git snapshot without verified BSR revision equivalence.
	BSRCompatibilitySnapshot = "compatibility_snapshot"
	// BSRExactRevision identifies a Git source proven to match the requested BSR revision.
	BSRExactRevision = "exact"
)

var (
	bsrModuleRE = regexp.MustCompile(bsrModulePattern)
	bsrDigestRE = regexp.MustCompile(bsrDigestPattern)
)

// BSRDependency preserves a Buf request and its optional lock pin, relative to the Git checkout.
type BSRDependency struct {
	Module    string `yaml:"module"`
	Reference string `yaml:"reference,omitempty"`
	Commit    string `yaml:"commit,omitempty"`
	Digest    string `yaml:"digest,omitempty"`
	Config    string `yaml:"config"`
}

// BSRResolution records the Git requirement chosen for a BSR dependency.
type BSRResolution struct {
	Dependency BSRDependency `yaml:"dependency"`
	Git        Requirement   `yaml:"git"`
	Resolution string        `yaml:"resolution"`
}

// Validate checks the preserved metadata, without verifying BSR content or contacting a registry.
func (dependency BSRDependency) Validate() error {
	if !bsrModuleRE.MatchString(dependency.Module) {
		return fmt.Errorf("invalid BSR module %q", dependency.Module)
	}
	if strings.ContainsAny(dependency.Reference, ": \t\r\n") {
		return fmt.Errorf("invalid BSR reference %q", dependency.Reference)
	}
	if dependency.Commit != "" && (len(dependency.Commit) != 32 || !isHex(dependency.Commit)) {
		return fmt.Errorf("invalid BSR commit %q", dependency.Commit)
	}
	if len(dependency.Reference) == 32 && isHex(dependency.Reference) && dependency.Commit != "" && !strings.EqualFold(dependency.Reference, dependency.Commit) {
		return fmt.Errorf("conflicting BSR commit for %s: requested %s, locked %s", dependency.Module, dependency.Reference, dependency.Commit)
	}
	if dependency.Digest != "" && !bsrDigestRE.MatchString(dependency.Digest) {
		return fmt.Errorf("invalid BSR digest %q", dependency.Digest)
	}
	if !filepath.IsLocal(filepath.FromSlash(dependency.Config)) || path.Clean(dependency.Config) != dependency.Config || strings.Contains(dependency.Config, "\\") {
		return fmt.Errorf("invalid BSR config path %q", dependency.Config)
	}
	return nil
}

// Validate requires an explicit, pinned Git target and a declared resolution policy.
func (binding BSRResolution) Validate() error {
	if err := binding.Dependency.Validate(); err != nil {
		return fmt.Errorf("Validate: %w", err)
	}
	if binding.Git.Module == "" || binding.Git.Version == "" {
		return fmt.Errorf("unpinned BSR Git target for %s", binding.Dependency.Module)
	}
	if err := ValidateModuleVersion(binding.Git.Module, binding.Git.Version); err != nil {
		return fmt.Errorf("ValidateModuleVersion: %w", err)
	}
	if !semver.IsValid(binding.Git.Version) && !IsCommitRef(binding.Git.Version) {
		return fmt.Errorf("invalid version %q for BSR Git target %s", binding.Git.Version, binding.Git.Module)
	}
	if binding.Resolution != BSRCompatibilitySnapshot && binding.Resolution != BSRExactRevision {
		return fmt.Errorf("invalid BSR resolution %q for %s", binding.Resolution, binding.Dependency.Module)
	}
	return nil
}
