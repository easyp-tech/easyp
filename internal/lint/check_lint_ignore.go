package lint

import (
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"
)

const (
	// for backward compatibility with buf
	bufLintIgnorePrefix = "buf:lint:ignore "
	lintIgnorePrefix    = "nolint:"
)

// CheckIsIgnored check if passed ruleName has to be ignored due to ignore command in comments
func CheckIsIgnored(comments []*parser.Comment, ruleName string) bool {
	if len(comments) == 0 {
		return false
	}

	bufIgnore := bufLintIgnorePrefix + ruleName
	easypIgnore := lintIgnorePrefix + ruleName

	for _, comment := range comments {
		if strings.Contains(comment.Raw, bufIgnore) {
			return true
		}
		if strings.Contains(comment.Raw, easypIgnore) {
			return true
		}
	}

	return false
}
