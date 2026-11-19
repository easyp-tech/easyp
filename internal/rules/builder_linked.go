package rules

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)

// NewLinked builds the linked lint engine's rule list (core.LintRule) and the
// ignore_only map by configuration. It mirrors New structurally — the same
// single validateConfigRuleNames call (D-06: rules.New remains THE validator,
// same error wrapping), the same group order, the same use/except/ignore_only
// plumbing — but constructs the Phase-4 port types, whose names come from
// core.LinkedRuleName (each port type reports its canonical registry name via
// LintRuleName).
//
// Dormancy: nothing populates core.Options.LinkedRules from production code
// until the Phase-5 flip; this constructor is reachable from tests only, and
// easyp lint keeps reading the old registry.
//
// Rule-type rename map (old type -> linked port type), kept for the Phase-5
// consolidation pass. Each batch plan replaces its placeholders; the finishing
// batch asserts the list is complete:
//
//	PackageDirectoryMatch        -> PackageDirectoryMatchLinked        (Wave 0, this file batch)
//	DirectorySamePackage         -> DirectorySamePackageLinked         (batch 1)
//	PackageDefined               -> PackageDefinedLinked               (batch 1)
//	PackageSameDirectory         -> PackageSameDirectoryLinked         (batch 1)
//	EnumFirstValueZero           -> (placeholder, batch 3)
//	EnumNoAllowAlias             -> (placeholder, batch 3)
//	EnumPascalCase               -> EnumPascalCaseLinked               (batch 2)
//	EnumValueUpperSnakeCase      -> EnumValueUpperSnakeCaseLinked      (batch 2)
//	FieldLowerSnakeCase          -> FieldLowerSnakeCaseLinked          (batch 2)
//	ImportNoPublic               -> ImportNoPublicLinked               (batch 2)
//	ImportNoWeak                 -> ImportNoWeakLinked                 (batch 2)
//	ImportUsed                   -> (placeholder, batch 3)
//	MessagePascalCase            -> MessagePascalCaseLinked            (batch 2)
//	OneofLowerSnakeCase          -> OneofLowerSnakeCaseLinked          (batch 2)
//	PackageLowerSnakeCase        -> PackageLowerSnakeCaseLinked        (batch 2)
//	PackageSameCsharpNamespace   -> (placeholder, batch 3)
//	PackageSameGoPackage         -> (placeholder, batch 3)
//	PackageSameJavaMultipleFiles -> (placeholder, batch 3)
//	PackageSameJavaPackage       -> (placeholder, batch 3)
//	PackageSamePHPNamespace      -> (placeholder, batch 3)
//	PackageSameRubyPackage       -> (placeholder, batch 3)
//	PackageSameSwiftPrefix       -> (placeholder, batch 3)
//	RPCPascalCase                -> RPCPascalCaseLinked                (batch 2)
//	ServicePascalCase            -> ServicePascalCaseLinked            (batch 2)
//	EnumValuePrefix              -> (placeholder, batch 4)
//	EnumZeroValueSuffix          -> (placeholder, batch 4)
//	FileLowerSnakeCase           -> (placeholder, batch 4)
//	RPCRequestResponseUnique     -> (placeholder, batch 4)
//	RPCRequestStandardName       -> (placeholder, batch 4)
//	RPCResponseStandardName      -> (placeholder, batch 4)
//	PackageVersionSuffix         -> (placeholder, batch 4)
//	ServiceSuffix                -> (placeholder, batch 4)
//	CommentEnum                  -> (placeholder, batch 5)
//	CommentEnumValue             -> (placeholder, batch 5)
//	CommentField                 -> (placeholder, batch 5)
//	CommentMessage               -> (placeholder, batch 5)
//	CommentOneof                 -> (placeholder, batch 5)
//	CommentRPC                   -> (placeholder, batch 5)
//	CommentService               -> (placeholder, batch 5)
//	RPCNoClientStreaming         -> (placeholder, batch 6)
//	RPCNoServerStreaming         -> (placeholder, batch 6)
func NewLinked(cfg config.LintConfig) ([]core.LintRule, map[string][]string, error) {
	if err := validateConfigRuleNames(cfg); err != nil {
		return nil, nil, fmt.Errorf("validateConfigRuleNames: %w", err)
	}
	allRules := []core.LintRule{
		//	minGroup
		&DirectorySamePackageLinked{},
		&PackageDefinedLinked{},
		&PackageDirectoryMatchLinked{
			Root:   ".",
			Prefix: cfg.PackageDirectoryPrefix,
		},
		&PackageSameDirectoryLinked{},

		//	basicGroup — batch 2 (casing) and batch 3 (semantic)
		// &EnumFirstValueZeroLinked{},
		// &EnumNoAllowAliasLinked{},
		&EnumPascalCaseLinked{},
		&EnumValueUpperSnakeCaseLinked{},
		&FieldLowerSnakeCaseLinked{},
		&ImportNoPublicLinked{},
		&ImportNoWeakLinked{},
		// &ImportUsedLinked{},
		&MessagePascalCaseLinked{},
		&OneofLowerSnakeCaseLinked{},
		&PackageLowerSnakeCaseLinked{},
		// &PackageSameCsharpNamespaceLinked{},
		// &PackageSameGoPackageLinked{},
		// &PackageSameJavaMultipleFilesLinked{},
		// &PackageSameJavaPackageLinked{},
		// &PackageSamePHPNamespaceLinked{},
		// &PackageSameRubyPackageLinked{},
		// &PackageSameSwiftPrefixLinked{},
		&RPCPascalCaseLinked{},
		&ServicePascalCaseLinked{},
		//	defaultGroup — batch 4
		// &EnumValuePrefixLinked{},
		// &EnumZeroValueSuffixLinked{
		// 	Suffix: defaultIfEmpty(cfg.EnumZeroValueSuffix, "UNSPECIFIED"),
		// },
		// &FileLowerSnakeCaseLinked{},
		// &RPCRequestResponseUniqueLinked{},
		// &RPCRequestStandardNameLinked{},
		// &RPCResponseStandardNameLinked{},
		// &PackageVersionSuffixLinked{},
		// &ServiceSuffixLinked{
		// 	Suffix: defaultIfEmpty(cfg.ServiceSuffix, "Service"),
		// },
		//	commentsGroup — batch 5
		// &CommentEnumLinked{},
		// &CommentEnumValueLinked{},
		// &CommentFieldLinked{},
		// &CommentMessageLinked{},
		// &CommentOneofLinked{},
		// &CommentRPCLinked{},
		// &CommentServiceLinked{},
		//	unaryRPCGroup — batch 6
		// &RPCNoClientStreamingLinked{},
		// &RPCNoServerStreamingLinked{},
	}

	rules := make(map[string]core.LintRule)
	for _, rule := range allRules {
		ruleName := core.LinkedRuleName(rule)
		rules[ruleName] = rule
	}

	use := unwrapLintGroups(cfg.Use)
	use = removeExcept(unwrapLintGroups(cfg.Except), use)

	res := make([]core.LintRule, len(use))

	for i, ruleName := range use {
		rule, ok := rules[ruleName]
		if !ok {
			return nil, nil, fmt.Errorf("%w: %s", core.ErrInvalidRule, ruleName)
		}

		res[i] = rule
	}

	return res, unwrapIgnoreOnly(cfg.IgnoreOnly), nil
}
