package rules

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)

// New builds the lint engine's rule list (core.Rule) and the ignore_only map
// by configuration. The same single validateConfigRuleNames call (D-06:
// rules.New remains THE validator, same error wrapping), the same group
// order, and the same use/except/ignore_only plumbing as the old constructor
// deleted in the Phase-5 flip — but constructs the 41 port types under their
// canonical unsuffixed names (the Phase-5 consolidation renamed them; their
// registry names come from core.GetRuleName's pure reflection over the type
// names, with no per-type overrides).
//
// This constructor is THE registry — the flip deleted the old constructor in
// the same commit, and easyp lint binds here through the unchanged buildCore
// call site.
func New(cfg config.LintConfig) ([]core.Rule, map[string][]string, error) {
	if err := validateConfigRuleNames(cfg); err != nil {
		return nil, nil, fmt.Errorf("validateConfigRuleNames: %w", err)
	}
	allRules := []core.Rule{
		//	minGroup
		&DirectorySamePackage{},
		&PackageDefined{},
		&PackageDirectoryMatch{
			Root:   ".",
			Prefix: cfg.PackageDirectoryPrefix,
		},
		&PackageSameDirectory{},

		//	basicGroup — batch 2 (casing) and batch 3 (semantic)
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
		//	defaultGroup — batch 4
		&EnumValuePrefix{},
		&EnumZeroValueSuffix{
			Suffix: defaultIfEmpty(cfg.EnumZeroValueSuffix, "UNSPECIFIED"),
		},
		&FileLowerSnakeCase{},
		&RPCRequestResponseUnique{},
		&RPCRequestStandardName{},
		&RPCResponseStandardName{},
		&PackageVersionSuffix{},
		&ServiceSuffix{
			Suffix: defaultIfEmpty(cfg.ServiceSuffix, "Service"),
		},
		//	commentsGroup — batch 5
		&CommentEnum{},
		&CommentEnumValue{},
		&CommentField{},
		&CommentMessage{},
		&CommentOneof{},
		&CommentRPC{},
		&CommentService{},
		//	unaryRPCGroup — batch 6
		&RPCNoClientStreaming{},
		&RPCNoServerStreaming{},
	}

	rules := make(map[string]core.Rule)
	for _, rule := range allRules {
		ruleName := core.GetRuleName(rule)
		rules[ruleName] = rule
	}

	use := unwrapLintGroups(cfg.Use)
	use = removeExcept(unwrapLintGroups(cfg.Except), use)

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
