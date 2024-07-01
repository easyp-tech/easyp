package rules

import (
	"unicode"

	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/lint"
)

// Config is the configuration for the rules.
type Config struct {
	PackageDirectoryMatchRoot string
	EnumZeroValueSuffix       string
	ServiceSuffix             string
}

// Rules returns all rules.
func Rules(cfg Config) map[string]lint.Rule {
	// List rules

	rules := []lint.Rule{
		//	MINIMAL
		&DirectorySamePackage{},
		&PackageDefined{},
		&PackageDirectoryMatch{
			Root: cfg.PackageDirectoryMatchRoot,
		},
		&PackageSameDirectory{},

		//	BASIC
		&EnumFirstValueZero{},
		&EnumNoAllowAlias{},
		&EnumPascalCase{},
		&EnumValueUpperSnakeCase{},
		&FieldLowerSnakeCase{},
		&ImportNoPublic{},
		&ImportNoWeak{},
		&ImportUsed{},
		&MessagePascalCase{},
		&OneofLowerSnakeCase{},
		&PackageLowerSnakeCase{},
		&PackageSameCsharpNamespace{},
		&PackageSameGoPackage{},
		&PackageSameJavaMultipleFiles{},
		&PackageSameJavaPackage{},
		&PackageSamePHPNamespace{},
		&PackageSameRubyPackage{},
		&PackageSameSwiftPrefix{},
		&RPCPascalCase{},
		&ServicePascalCase{},
		//	DEFAULT
		&EnumValuePrefix{},
		&EnumZeroValueSuffix{
			Suffix: cfg.EnumZeroValueSuffix,
		},
		&FileLowerSnakeCase{},
		&RPCRequestResponseUnique{},
		&RPCRequestStandardName{},
		&RPCResponseStandardName{},
		&PackageVersionSuffix{},
		&ServiceSuffix{
			Suffix: cfg.ServiceSuffix,
		},
		//	COMMENTS
		&CommentEnum{},
		&CommentEnumValue{},
		&CommentField{},
		&CommentMessage{},
		&CommentOneof{},
		&CommentRPC{},
		&CommentService{},
		//	UNARY_RPC
		&RPCNoClientStreaming{},
		&RPCNoServerStreaming{},
		//	UNCATEGORIZED
		&PackageNoImportCycle{},
	}

	return lo.SliceToMap(rules, func(item lint.Rule) (string, lint.Rule) {
		return item.Name(), item
	})

	//return map[string]lint.Rule{
	//	// Minimal
	//	DirectorySamePackageName: &DirectorySamePackage{},
	//	PackageDefinedName:       &PackageDefined{},
	//	PackageDirectoryMatchName: &PackageDirectoryMatch{
	//		Root: cfg.PackageDirectoryMatchRoot,
	//	},
	//	PackageSameDirectoryName: &PackageSameDirectory{},
	//
	//	// Basic
	//	"ENUM_FIRST_VALUE_ZERO":            &EnumFirstValueZero{},
	//	"ENUM_NO_ALLOW_ALIAS":              &EnumNoAllowAlias{},
	//	"ENUM_PASCAL_CASE":                 &EnumPascalCase{},
	//	"ENUM_VALUE_UPPER_SNAKE_CASE":      &EnumValueUpperSnakeCase{},
	//	"FIELD_LOWER_SNAKE_CASE":           &FieldLowerSnakeCase{},
	//	"IMPORT_NO_PUBLIC":                 &ImportNoPublic{},
	//	"IMPORT_NO_WEAK":                   &ImportNoWeak{},
	//	"IMPORT_USED":                      &ImportUsed{},
	//	"MESSAGE_PASCAL_CASE":              &MessagePascalCase{},
	//	"ONEOF_LOWER_SNAKE_CASE":           &OneofLowerSnakeCase{},
	//	"PACKAGE_LOWER_SNAKE_CASE":         &PackageLowerSnakeCase{},
	//	"PACKAGE_SAME_CSHARP_NAMESPACE":    &PackageSameCsharpNamespace{},
	//	"PACKAGE_SAME_GO_PACKAGE":          &PackageSameGoPackage{},
	//	"PACKAGE_SAME_JAVA_MULTIPLE_FILES": &PackageSameJavaMultipleFiles{},
	//	"PACKAGE_SAME_JAVA_PACKAGE":        &PackageSameJavaPackage{},
	//	"PACKAGE_SAME_PHP_NAMESPACE":       &PackageSamePHPNamespace{},
	//	"PACKAGE_SAME_RUBY_PACKAGE":        &PackageSameRubyPackage{},
	//	"PACKAGE_SAME_SWIFT_PREFIX":        &PackageSameSwiftPrefix{},
	//	"RPC_PASCAL_CASE":                  &RPCPascalCase{},
	//	"SERVICE_PASCAL_CASE":              &ServicePascalCase{},
	//
	//	// Default
	//	"ENUM_VALUE_PREFIX": &EnumValuePrefix{},
	//	"ENUM_ZERO_VALUE_SUFFIX": &EnumZeroValueSuffix{
	//		Suffix: cfg.EnumZeroValueSuffix,
	//	},
	//	"FILE_LOWER_SNAKE_CASE":       &FileLowerSnakeCase{},
	//	"RPC_REQUEST_RESPONSE_UNIQUE": &RPCRequestResponseUnique{},
	//	"RPC_REQUEST_STANDARD_NAME":   &RPCRequestStandardName{},
	//	"RPC_RESPONSE_STANDARD_NAME":  &RPCResponseStandardName{},
	//	"PACKAGE_VERSION_SUFFIX":      &PackageVersionSuffix{},
	//	"PROTOVALIDATE":               &ProtoValidate{}, // TODO: This, rule, is not implemented yet
	//	"SERVICE_SUFFIX": &ServiceSuffix{
	//		Suffix: cfg.ServiceSuffix,
	//	},
	//
	//	// Comments
	//	"COMMENT_ENUM":       &CommentEnum{},
	//	"COMMENT_ENUM_VALUE": &CommentEnumValue{},
	//	"COMMENT_FIELD":      &CommentField{},
	//	"COMMENT_MESSAGE":    &CommentMessage{},
	//	"COMMENT_ONEOF":      &CommentOneOf{},
	//	"COMMENT_RPC":        &CommentRPC{},
	//	"COMMENT_SERVICE":    &CommentService{},
	//
	//	// UNARY_RPC
	//	"RPC_NO_CLIENT_STREAMING": &RPCNoClientStreaming{},
	//	"RPC_NO_SERVER_STREAMING": &RPCNoServerStreaming{},
	//
	//	// Uncategorized
	//	"PACKAGE_NO_IMPORT_CYCLE": &PackageNoImportCycle{}, // TODO: This, rule, is not implemented yet
	//}
}

// toUpperSnakeCase converts a string from PascalCase or camelCase to UPPER_SNEAK_CASE.
func toUpperSnakeCase(s string) string {
	var result []rune

	for i, r := range s {
		if unicode.IsUpper(r) {
			// Добавляем подчеркивание, когда:
			// 1. Не первый символ.
			// 2. Предыдущий символ не был заглавной буквой, либо следующий является прописной буквой.
			if i > 0 && (unicode.IsLower(rune(s[i-1])) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1])))) {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToUpper(r))
	}

	return string(result)

}
