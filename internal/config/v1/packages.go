package v1

import (
	"fmt"
	"regexp"
)

// PackageSelectorPattern is the grammar of an exact protobuf package name.
// Selectors are not file paths, glob patterns or package-name prefixes.
const PackageSelectorPattern = `^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`

var packageSelector = regexp.MustCompile(PackageSelectorPattern)

// ValidatePackageSelectors checks package selection without inspecting sources.
func ValidatePackageSelectors(packages []string) error {
	for index, name := range packages {
		if !packageSelector.MatchString(name) {
			return fmt.Errorf("generate.packages[%d]: %q is not an exact protobuf package name", index, name)
		}
	}
	return nil
}
