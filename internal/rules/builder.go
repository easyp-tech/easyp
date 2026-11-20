package rules

import (
	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/core"
)

const (
	minGroup      = "MINIMAL"
	basicGroup    = "BASIC"
	defaultGroup  = "DEFAULT"
	commentsGroup = "COMMENTS"
	unaryRPCGroup = "UNARY_RPC"
)

// RuleGroup describes a named group of lint rules.
type RuleGroup struct {
	Name  string   // Human-readable name: "Minimal", "Basic", ...
	Key   string   // Config key: "MINIMAL", "BASIC", ...
	Rules []string // List of rule names in the group.
}

// AllGroups returns all rule groups in canonical order.
func AllGroups() []RuleGroup {
	return []RuleGroup{
		{Name: "Minimal", Key: minGroup, Rules: addMinimal(nil)},
		{Name: "Basic", Key: basicGroup, Rules: addBasic(nil)},
		{Name: "Default", Key: defaultGroup, Rules: addDefault(nil)},
		{Name: "Comments", Key: commentsGroup, Rules: addComments(nil)},
		{Name: "Unary RPC", Key: unaryRPCGroup, Rules: addUnary(nil)},
	}
}

// AllRuleNames returns a flat list of all rule names from all groups.
func AllRuleNames() []string {
	var res []string
	for _, g := range AllGroups() {
		res = append(res, g.Rules...)
	}
	return res
}

// AllLintUseValues returns all valid values for lint.use:
// group keys, grouped rule names, and uncategorized rule names.
func AllLintUseValues() []string {
	groups := AllGroups()
	values := make([]string, 0, len(groups)+len(AllRuleNames()))
	for _, group := range groups {
		values = append(values, group.Key)
	}
	values = append(values, AllRuleNames()...)

	return lo.Uniq(values)
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

	return lo.Uniq(res)
}

func removeExcept(except, use []string) []string {
	return lo.Filter(use, func(ruleName string, _ int) bool {
		return !lo.Contains(except, ruleName)
	})
}

// defaultIfEmpty returns val if non-empty, otherwise fallback.
func defaultIfEmpty(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func addMinimal(res []string) []string {
	res = append(res, core.GetRuleName(&DirectorySamePackageLinked{}))
	res = append(res, core.GetRuleName(&PackageDefinedLinked{}))
	res = append(res, core.GetRuleName(&PackageDirectoryMatchLinked{}))
	res = append(res, core.GetRuleName(&PackageSameDirectoryLinked{}))

	return res
}

func addBasic(res []string) []string {
	res = append(res, core.GetRuleName(&EnumFirstValueZeroLinked{}))
	res = append(res, core.GetRuleName(&EnumNoAllowAliasLinked{}))
	res = append(res, core.GetRuleName(&EnumPascalCaseLinked{}))
	res = append(res, core.GetRuleName(&EnumValueUpperSnakeCaseLinked{}))
	res = append(res, core.GetRuleName(&FieldLowerSnakeCaseLinked{}))
	res = append(res, core.GetRuleName(&ImportNoPublicLinked{}))
	res = append(res, core.GetRuleName(&ImportNoWeakLinked{}))
	res = append(res, core.GetRuleName(&ImportUsedLinked{}))
	res = append(res, core.GetRuleName(&MessagePascalCaseLinked{}))
	res = append(res, core.GetRuleName(&OneofLowerSnakeCaseLinked{}))
	res = append(res, core.GetRuleName(&PackageLowerSnakeCaseLinked{}))
	res = append(res, core.GetRuleName(&PackageSameCsharpNamespaceLinked{}))
	res = append(res, core.GetRuleName(&PackageSameGoPackageLinked{}))
	res = append(res, core.GetRuleName(&PackageSameJavaMultipleFilesLinked{}))
	res = append(res, core.GetRuleName(&PackageSameJavaPackageLinked{}))
	res = append(res, core.GetRuleName(&PackageSamePHPNamespaceLinked{}))
	res = append(res, core.GetRuleName(&PackageSameRubyPackageLinked{}))
	res = append(res, core.GetRuleName(&PackageSameSwiftPrefixLinked{}))
	res = append(res, core.GetRuleName(&RPCPascalCaseLinked{}))
	res = append(res, core.GetRuleName(&ServicePascalCaseLinked{}))
	return res
}

func addDefault(res []string) []string {
	res = append(res, core.GetRuleName(&EnumValuePrefixLinked{}))
	res = append(res, core.GetRuleName(&EnumZeroValueSuffixLinked{}))
	res = append(res, core.GetRuleName(&FileLowerSnakeCaseLinked{}))
	res = append(res, core.GetRuleName(&RPCRequestResponseUniqueLinked{}))
	res = append(res, core.GetRuleName(&RPCRequestStandardNameLinked{}))
	res = append(res, core.GetRuleName(&RPCResponseStandardNameLinked{}))
	res = append(res, core.GetRuleName(&PackageVersionSuffixLinked{}))
	res = append(res, core.GetRuleName(&ServiceSuffixLinked{}))
	return res
}

func addComments(res []string) []string {
	res = append(res, core.GetRuleName(&CommentEnumLinked{}))
	res = append(res, core.GetRuleName(&CommentEnumValueLinked{}))
	res = append(res, core.GetRuleName(&CommentFieldLinked{}))
	res = append(res, core.GetRuleName(&CommentMessageLinked{}))
	res = append(res, core.GetRuleName(&CommentOneofLinked{}))
	res = append(res, core.GetRuleName(&CommentRPCLinked{}))
	res = append(res, core.GetRuleName(&CommentServiceLinked{}))
	return res
}

func addUnary(res []string) []string {
	res = append(res, core.GetRuleName(&RPCNoClientStreamingLinked{}))
	res = append(res, core.GetRuleName(&RPCNoServerStreamingLinked{}))
	return res
}
