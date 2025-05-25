package storage

import "strings"

func sanitizePath(source string) string {
	return strings.ReplaceAll(source, "/", "-")
}
