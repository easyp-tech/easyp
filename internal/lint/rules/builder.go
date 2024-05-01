package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

// Config is the configuration for the rules.
type Config struct {
	PackageDirectoryMatchRoot string
	EnumZeroValueSuffixPrefix string
	ServiceSuffixSuffix       string
}

// Rules returns all rules.
func Rules(cfg Config) map[string]lint.Rule {
	return map[string]lint.Rule{
		// Minimal
		"DIRECTORY_SAME_PACKAGE": &DirectorySamePackage{},
		"PACKAGE_DEFINED":        &PackageDefined{},
		"PACKAGE_DIRECTORY_MATCH": &PackageDirectoryMatch{
			Root: cfg.PackageDirectoryMatchRoot,
		},
		"PACKAGE_SAME_DIRECTORY": &PackageSameDirectory{},

		// Basic
		"ENUM_FIRST_VALUE_ZERO":            &EnumFirstValueZero{},
		"ENUM_NO_ALLOW_ALIAS":              &EnumNoAllowAlias{},
		"ENUM_PASCAL_CASE":                 &EnumPascalCase{},
		"ENUM_VALUE_UPPER_SNAKE_CASE":      &EnumValueUpperSnakeCase{},
		"FIELD_LOWER_SNAKE_CASE":           &FieldLowerSnakeCase{},
		"IMPORT_NO_PUBLIC":                 &ImportNoPublic{},
		"IMPORT_NO_WEAK":                   &ImportNoWeak{},
		"IMPORT_USED":                      &ImportUsed{},
		"MESSAGE_PASCAL_CASE":              &MessagePascalCase{},
		"ONEOF_LOWER_SNAKE_CASE":           &OneofLowerSnakeCase{},
		"PACKAGE_LOWER_SNAKE_CASE":         &PackageLowerSnakeCase{},
		"PACKAGE_SAME_CSHARP_NAMESPACE":    &PackageSameCSharpNamespace{},
		"PACKAGE_SAME_GO_PACKAGE":          &PackageSameGoPackage{},
		"PACKAGE_SAME_JAVA_MULTIPLE_FILES": &PackageSameJavaMultipleFiles{},
		"PACKAGE_SAME_JAVA_PACKAGE":        &PackageSameJavaPackage{},
		"PACKAGE_SAME_PHP_NAMESPACE":       &PackageSamePHPNamespace{},
		"PACKAGE_SAME_RUBY_PACKAGE":        &PackageSameRubyPackage{},
		"PACKAGE_SAME_SWIFT_PREFIX":        &PackageSameSwiftPrefix{},
		"RPC_PASCAL_CASE":                  &RpcPascalCase{},
		"SERVICE_PASCAL_CASE":              &ServicePascalCase{},

		// Default
		"ENUM_VALUE_PREFIX": &EnumValuePrefix{},
		"ENUM_ZERO_VALUE_SUFFIX": &EnumZeroValueSuffix{
			Suffix: cfg.EnumZeroValueSuffixPrefix,
		},
		"FILE_LOWER_SNAKE_CASE":       &FileLowerSnakeCase{},
		"RPC_REQUEST_RESPONSE_UNIQUE": &RPCRequestResponseUnique{},
		"RPC_REQUEST_STANDARD_NAME":   &RPCRequestStandardName{},
		"RPC_RESPONSE_STANDARD_NAME":  &RPCResponseStandardName{},
		"PACKAGE_VERSION_SUFFIX":      &PackageVersionSuffix{},
		"PROTOVALIDATE":               &ProtoValidate{}, // TODO: This, rule, is not implemented yet
		"SERVICE_SUFFIX": &ServiceSuffix{
			Suffix: cfg.ServiceSuffixSuffix,
		},

		// Comments
		"COMMENT_ENUM":       &CommentEnum{},
		"COMMENT_ENUM_VALUE": &CommentEnumValue{},
		"COMMENT_FIELD":      &CommentField{},
		"COMMENT_MESSAGE":    &CommentMessage{},
		"COMMENT_ONEOF":      &CommentOneOf{},
		"COMMENT_RPC":        &CommentRPC{},
		"COMMENT_SERVICE":    &CommentService{},

		// UNARY_RPC
		"RPC_NO_CLIENT_STREAMING": &RPCNoClientStreaming{},
		"RPC_NO_SERVER_STREAMING": &RPCNoServerStreaming{},

		// Uncategorized
		"PACKAGE_NO_IMPORT_CYCLE": &PackageNoImportCycle{}, // TODO: This, rule, is not implemented yet
	}
}
