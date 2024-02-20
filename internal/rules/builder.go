package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

// Rules returns all rules.
func Rules() map[string]core.Rule {
	return map[string]core.Rule{
		// Minimal
		"DIRECTORY_SAME_PACKAGE":  &DirectorySamePackage{},
		"PACKAGE_DEFINED":         &PackageDefined{},
		"PACKAGE_DIRECTORY_MATCH": &PackageDirectoryMatch{},
		"PACKAGE_SAME_DIRECTORY":  &PackageSameDirectory{},
		// Basic
		"COMMENT_ENUM":                   &CommentEnum{},
		"COMMENT_ENUM_VALUE":             &CommentEnumValue{},
		"COMMENT_MESSAGE":                &CommentMessage{},
		"COMMENT_MESSAGE_FIELD":          &CommentMessageField{},
		"COMMENT_ONEOF":                  &CommentOneOf{},
		"COMMENT_RPC":                    &CommentRPC{},
		"COMMENT_SERVICE":                &CommentService{},
		"ENUM_FIRST_VALUE_ZERO":          &EnumFirstValueZero{},
		"ENUM_PASCAL_CASE":               &EnumPascalCase{},
		"ENUM_VALUE_UPPER_SNAKE_CASE":    &EnumValueUpperSnakeCase{},
		"MESSAGE_FIELD_LOWER_SNAKE_CASE": &MessageFieldLowerSnakeCase{},
		"MESSAGE_PASCAL_CASE":            &MessagePascalCase{},
		"ONEOF_LOWER_SNAKE_CASE":         &OneofLowerSnakeCase{},
		"PACKAGE_LOWER_SNAKE_CASE":       &PackageLowerSnakeCase{},
		"RPC_PASCAL_CASE":                &RpcPascalCase{},
		"SERVICE_PASCAL_CASE":            &ServicePascalCase{},
	}
}
