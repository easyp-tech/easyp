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
		ruleName := lint.GetRuleName(item)
		return ruleName, item
	})
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
