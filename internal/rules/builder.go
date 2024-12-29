package rules

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)

const (
	minGroup      = "MINIMAL"
	basicGroup    = "BASIC"
	defaultGroup  = "DEFAULT"
	commentsGroup = "COMMENTS"
	unaryRPCGroup = "UNARY_RPC"
)

// New returns a map of rules and a map of ignore only rules by configuration.
func New(cfg config.LintConfig) ([]core.Rule, map[string][]string, error) {
	allRules := []core.Rule{
		//	minGroup
		&DirectorySamePackage{},
		&PackageDefined{},
		&PackageDirectoryMatch{
			Root: ".", // TODO: fix me
		},
		&PackageSameDirectory{},

		//	basicGroup
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
		//	defaultGroup
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
		//	commentsGroup
		&CommentEnum{},
		&CommentEnumValue{},
		&CommentField{},
		&CommentMessage{},
		&CommentOneof{},
		&CommentRPC{},
		&CommentService{},
		//	unaryRPCGroup
		&RPCNoClientStreaming{},
		&RPCNoServerStreaming{},
		//	UNCATEGORIZED
		&PackageNoImportCycle{},
	}

	rules := make(map[string]core.Rule)
	for _, rule := range allRules {
		ruleName := core.GetRuleName(rule)
		rules[ruleName] = rule
	}

	use := unwrapLintGroups(cfg.Use)
	use = removeExcept(cfg.Except, use)

	res := make([]core.Rule, len(use))

	for i, ruleName := range use {
		rule, ok := rules[ruleName]
		if !ok {
			return nil, nil, fmt.Errorf("%w: %s", core.ErrInvalidRule, ruleName)
		}

		res[i] = rule
	}

	return res, unwrapIgnoreOnly(cfg.IgnoreOnly), nil
}

func unwrapIgnoreOnly(ignoreOnly map[string][]string) map[string][]string {
	res := make(map[string][]string)

	for ruleName, fileOrDirs := range ignoreOnly {
		switch ruleName {
		case minGroup:
			ruleNames := addMinimal(nil)
			for i := range ruleNames {
				res[ruleNames[i]] = fileOrDirs
			}
		case basicGroup:
			ruleNames := addBasic(nil)
			for i := range ruleNames {
				res[ruleNames[i]] = fileOrDirs
			}
		case defaultGroup:
			ruleNames := addDefault(nil)
			for i := range ruleNames {
				res[ruleNames[i]] = fileOrDirs
			}
		case commentsGroup:
			ruleNames := addComments(nil)
			for i := range ruleNames {
				res[ruleNames[i]] = fileOrDirs
			}
		case unaryRPCGroup:
			ruleNames := addUnary(nil)
			for i := range ruleNames {
				res[ruleNames[i]] = fileOrDirs
			}
		default:
			res[ruleName] = fileOrDirs
		}
	}

	return res
}

func unwrapLintGroups(use []string) []string {
	var res []string

	for _, ruleName := range use {
		switch ruleName {
		case minGroup:
			res = addMinimal(res)
		case basicGroup:
			res = addBasic(res)
		case defaultGroup:
			res = addDefault(res)
		case commentsGroup:
			res = addComments(res)
		case unaryRPCGroup:
			res = addUnary(res)
		default:
			res = append(res, ruleName)
		}
	}

	return lo.FindUniques(res)
}

func removeExcept(except, use []string) []string {
	return lo.Filter(use, func(ruleName string, _ int) bool {
		return !lo.Contains(except, ruleName)
	})
}

func addMinimal(res []string) []string {
	res = append(res, core.GetRuleName(&DirectorySamePackage{}))
	res = append(res, core.GetRuleName(&PackageDefined{}))
	res = append(res, core.GetRuleName(&PackageDirectoryMatch{}))
	res = append(res, core.GetRuleName(&PackageSameDirectory{}))

	return res
}

func addBasic(res []string) []string {
	res = append(res, core.GetRuleName(&EnumFirstValueZero{}))
	res = append(res, core.GetRuleName(&EnumNoAllowAlias{}))
	res = append(res, core.GetRuleName(&EnumPascalCase{}))
	res = append(res, core.GetRuleName(&EnumValueUpperSnakeCase{}))
	res = append(res, core.GetRuleName(&FieldLowerSnakeCase{}))
	res = append(res, core.GetRuleName(&ImportNoPublic{}))
	res = append(res, core.GetRuleName(&ImportNoWeak{}))
	res = append(res, core.GetRuleName(&ImportUsed{}))
	res = append(res, core.GetRuleName(&MessagePascalCase{}))
	res = append(res, core.GetRuleName(&OneofLowerSnakeCase{}))
	res = append(res, core.GetRuleName(&PackageLowerSnakeCase{}))
	res = append(res, core.GetRuleName(&PackageSameCsharpNamespace{}))
	res = append(res, core.GetRuleName(&PackageSameGoPackage{}))
	res = append(res, core.GetRuleName(&PackageSameJavaMultipleFiles{}))
	res = append(res, core.GetRuleName(&PackageSameJavaPackage{}))
	res = append(res, core.GetRuleName(&PackageSamePHPNamespace{}))
	res = append(res, core.GetRuleName(&PackageSameRubyPackage{}))
	res = append(res, core.GetRuleName(&PackageSameSwiftPrefix{}))
	res = append(res, core.GetRuleName(&RPCPascalCase{}))
	res = append(res, core.GetRuleName(&ServicePascalCase{}))
	return res
}

func addDefault(res []string) []string {
	res = append(res, core.GetRuleName(&EnumValuePrefix{}))
	res = append(res, core.GetRuleName(&EnumZeroValueSuffix{}))
	res = append(res, core.GetRuleName(&FileLowerSnakeCase{}))
	res = append(res, core.GetRuleName(&RPCRequestResponseUnique{}))
	res = append(res, core.GetRuleName(&RPCRequestStandardName{}))
	res = append(res, core.GetRuleName(&RPCResponseStandardName{}))
	res = append(res, core.GetRuleName(&PackageVersionSuffix{}))
	res = append(res, core.GetRuleName(&ServiceSuffix{}))
	return res
}

func addComments(res []string) []string {
	res = append(res, core.GetRuleName(&CommentEnum{}))
	res = append(res, core.GetRuleName(&CommentEnumValue{}))
	res = append(res, core.GetRuleName(&CommentField{}))
	res = append(res, core.GetRuleName(&CommentMessage{}))
	res = append(res, core.GetRuleName(&CommentOneof{}))
	res = append(res, core.GetRuleName(&CommentRPC{}))
	res = append(res, core.GetRuleName(&CommentService{}))
	return res
}

func addUnary(res []string) []string {
	res = append(res, "RPC_NO_CLIENT_STREAMING")
	res = append(res, "RPC_NO_SERVER_STREAMING")
	return res
}
