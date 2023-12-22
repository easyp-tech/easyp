package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var (
	Rules = map[string]core.Rule{
		"COMMENT_ENUM":                &CommentEnum{},
		"COMMENT_ONEOF":               &CommentOneOf{},
		"COMMENT_RPC":                 &CommentRPC{},
		"COMMENT_SERVICE":             &CommentService{},
		"COMMENT_MESSAGE":             &CommentMessage{},
		"COMMENT_MESSAGE_FIELD":       &CommentMessageField{},
		"ENUM_PASCAL_CASE":            &EnumPascalCase{},
		"ENUM_VALUE_UPPER_SNAKE_CASE": &EnumValueUpperSnakeCase{},
	}
)
