package rules

import (
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// CheckIsIgnored check if passed ruleName has to be ignored due to ignore command in comments
func CheckIsIgnored(comments []parser.Comment, ruleName string) bool {
	return false
}
