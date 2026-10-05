package core

import (
	"regexp"
	"strings"
)

// Match the unstable forms documented by PACKAGE_VERSION_SUFFIX, with positive
// version numbers and optional positive alpha/beta sequence numbers.
var unstablePackageVersion = regexp.MustCompile(`^v[1-9][0-9]*(test.*|(alpha|beta)([1-9][0-9]*)?|p[1-9][0-9]*(alpha|beta)([1-9][0-9]*)?)$`)

func isUnstablePackage(name PackageName) bool {
	component := string(name)
	component = component[strings.LastIndexByte(component, '.')+1:]
	return unstablePackageVersion.MatchString(component)
}

func excludeUnstablePackages(data ProtoData) {
	for name := range data {
		if isUnstablePackage(name) {
			delete(data, name)
		}
	}
}
