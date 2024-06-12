package rules

import (
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"
)

const (
	bufLintIgnorePrefix = "buf:lint:ignore "
)

// CheckIsIgnored check if passed ruleName has to be ignored due to ignore command in comments
func CheckIsIgnored(comments []parser.Comment, ruleName string) bool {
	if len(comments) == 0 {
		return false
	}

	s := bufLintIgnorePrefix + ruleName

	for _, comment := range comments {
		if strings.Contains(comment.Raw, s) {
			return true
		}
	}

	return false
}
