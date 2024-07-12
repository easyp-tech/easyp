package rules

import (
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
