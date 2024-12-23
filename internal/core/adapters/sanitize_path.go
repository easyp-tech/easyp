package adapters

import (
	"strings"
)

// SanitizePath replace `/` with `-` to prevent troubles with creating files/dirs for modules that have '/' in version
func SanitizePath(source string) string {
	return strings.Replace(source, "/", "-", -1)
}
