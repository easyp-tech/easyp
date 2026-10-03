package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// ValidateLegacyMajor verifies that +incompatible describes pre-native metadata.
// The suffix is a compatibility marker, never permission to bypass a native
// protobuf.mod identity. Call it on the exact verified revision, including caches.
func ValidateLegacyMajor(dir, source, version string) error {
	if semver.Build(version) != "+incompatible" {
		return nil
	}
	if err := v1.ValidateModuleVersion(source, version); err != nil {
		return fmt.Errorf("ValidateModuleVersion: %w", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, v1.ModuleFile))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ReadFile: %w", err)
	}
	if err == nil && v1.IsModuleManifest(raw) {
		return fmt.Errorf("%s@%s: +incompatible is only valid before a native protobuf.mod exists; use the published matching /vN module identity", source, version)
	}
	nested, err := readNestedGitDependencyModule(dir, source)
	if err != nil {
		return fmt.Errorf("readNestedGitDependencyModule: %w", err)
	}
	if nested.found {
		return fmt.Errorf("%s@%s: +incompatible cannot bypass a native nested protobuf.mod", source, version)
	}
	return nil
}
