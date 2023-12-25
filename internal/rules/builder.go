package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

// Rules returns all rules.
func Rules() map[string]core.Rule {
	return map[string]core.Rule{
		"COMMENT_ENUM":                   &CommentEnum{},
		"COMMENT_ONEOF":                  &CommentOneOf{},
		"COMMENT_RPC":                    &CommentRPC{},
		"COMMENT_SERVICE":                &CommentService{},
		"COMMENT_MESSAGE":                &CommentMessage{},
		"COMMENT_MESSAGE_FIELD":          &CommentMessageField{},
		"COMMENT_ENUM_VALUE":             &CommentEnumValue{},
		"ENUM_PASCAL_CASE":               &EnumPascalCase{},
		"ENUM_VALUE_UPPER_SNAKE_CASE":    &EnumValueUpperSnakeCase{},
		"MESSAGE_PASCAL_CASE":            &MessagePascalCase{},
		"MESSAGE_FIELD_LOWER_SNAKE_CASE": &MessageFieldLowerSnakeCase{},
		"ONEOF_LOWER_SNAKE_CASE":         &OneofLowerSnakeCase{},
		"PACKAGE_LOWER_SNAKE_CASE":       &PackageLowerSnakeCase{},
	}
}
