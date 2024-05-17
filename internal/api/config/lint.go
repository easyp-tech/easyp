package config

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

// LintConfig contains linter configuration.
type LintConfig struct {
	Use                 []string `json:"use" yaml:"use" env:"USE"`                                                          // Use rules for linter.
	EnumZeroValueSuffix string   `json:"enum_zero_value_suffix" yaml:"enum_zero_value_suffix" env:"ENUM_ZERO_VALUE_SUFFIX"` // Enum zero value suffix prefix.
	ServiceSuffix       string   `json:"service_suffix" yaml:"service_suffix" env:"SERVICE_SUFFIX"`                         // Service suffix suffix.
	Ignore              []string `json:"ignore" yaml:"ignore" env:"IGNORE"`                                                 // Ignore dirs with proto file.
	Except              []string `json:"except" yaml:"except" env:"EXCEPT"`                                                 // Except linter rules.
	AllowCommentIgnores bool     `json:"allow_comment_ignores" yaml:"allow_comment_ignores" env:"ALLOW_COMMENT_IGNORES"`    // Allow comment ignore.
}

func (cfg *Config) BuildLinterRules() ([]lint.Rule, error) {
	cfg.unwrapLintGroups()
	cfg.removeExcept()

	return cfg.buildFromUse()
}

func (cfg *Config) buildFromUse() ([]lint.Rule, error) {
	var useRule []lint.Rule

	for _, ruleName := range cfg.Lint.Use {
		rule, ok := rules.Rules(rules.Config{
			PackageDirectoryMatchRoot: ".",
			EnumZeroValueSuffix:       cfg.Lint.EnumZeroValueSuffix,
			ServiceSuffix:             cfg.Lint.ServiceSuffix,
		})[ruleName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", rules.ErrInvalidRule, ruleName)
		}

		useRule = append(useRule, rule)
	}

	return useRule, nil
}

const (
	minGroup      = "MINIMAL"
	basicGroup    = "BASIC"
	defaultGroup  = "DEFAULT"
	commentsGroup = "COMMENTS"
	unaryRPCGroup = "UNARY_RPC"
)

func (cfg *Config) unwrapLintGroups() {
	var res []string

	for _, ruleName := range cfg.Lint.Use {
		switch ruleName {
		case minGroup:
			res = cfg.addMinimal(res)
		case basicGroup:
			res = cfg.addBasic(res)
		case defaultGroup:
			res = cfg.addDefault(res)
		case commentsGroup:
			res = cfg.addComments(res)
		case unaryRPCGroup:
			res = cfg.addUnary(res)
		default:
			res = append(res, ruleName)
		}
	}

	cfg.Lint.Use = lo.FindUniques(res)
}

func (cfg *Config) removeExcept() {
	cfg.Lint.Use = lo.Filter(cfg.Lint.Use, func(ruleName string, _ int) bool {
		return !lo.Contains(cfg.Lint.Except, ruleName)
	})
}

func (cfg *Config) addMinimal(res []string) []string {
	res = append(res, rules.DIRECTORY_SAME_PACKAGE)
	res = append(res, rules.PACKAGE_DEFINED)
	res = append(res, rules.PACKAGE_DIRECTORY_MATCH)
	res = append(res, rules.PACKAGE_SAME_DIRECTORY)
	return res
}

func (cfg *Config) addBasic(res []string) []string {
	res = append(res, rules.ENUM_FIRST_VALUE_ZERO)
	res = append(res, rules.ENUM_NO_ALLOW_ALIAS)
	res = append(res, rules.ENUM_PASCAL_CASE)
	res = append(res, rules.ENUM_VALUE_UPPER_SNAKE_CASE)
	res = append(res, rules.FIELD_LOWER_SNAKE_CASE)
	res = append(res, rules.IMPORT_NO_PUBLIC)
	res = append(res, rules.IMPORT_NO_WEAK)
	res = append(res, rules.IMPORT_USED)
	res = append(res, rules.MESSAGE_PASCAL_CASE)
	res = append(res, rules.ONEOF_LOWER_SNAKE_CASE)
	res = append(res, rules.PACKAGE_LOWER_SNAKE_CASE)
	res = append(res, rules.PACKAGE_SAME_CSHARP_NAMESPACE)
	res = append(res, rules.PACKAGE_SAME_GO_PACKAGE)
	res = append(res, rules.PACKAGE_SAME_JAVA_MULTIPLE_FILES)
	res = append(res, rules.PACKAGE_SAME_JAVA_PACKAGE)
	res = append(res, rules.PACKAGE_SAME_PHP_NAMESPACE)
	res = append(res, rules.PACKAGE_SAME_RUBY_PACKAGE)
	res = append(res, rules.PACKAGE_SAME_SWIFT_PREFIX)
	res = append(res, rules.RPC_PASCAL_CASE)
	res = append(res, rules.SERVICE_PASCAL_CASE)
	return res
}

func (cfg *Config) addDefault(res []string) []string {
	res = append(res, rules.ENUM_VALUE_PREFIX)
	res = append(res, rules.ENUM_ZERO_VALUE_SUFFIX)
	res = append(res, rules.FILE_LOWER_SNAKE_CASE)
	res = append(res, rules.RPC_REQUEST_RESPONSE_UNIQUE)
	res = append(res, rules.RPC_REQUEST_STANDARD_NAME)
	res = append(res, rules.RPC_RESPONSE_STANDARD_NAME)
	res = append(res, rules.PACKAGE_VERSION_SUFFIX)
	res = append(res, rules.SERVICE_SUFFIX)
	return res
}

func (cfg *Config) addComments(res []string) []string {
	res = append(res, rules.COMMENT_ENUM)
	res = append(res, rules.COMMENT_ENUM_VALUE)
	res = append(res, rules.COMMENT_FIELD)
	res = append(res, rules.COMMENT_MESSAGE)
	res = append(res, rules.COMMENT_ONEOF)
	res = append(res, rules.COMMENT_RPC)
	res = append(res, rules.COMMENT_SERVICE)
	return res
}

func (cfg *Config) addUnary(res []string) []string {
	res = append(res, rules.RPC_NO_CLIENT_STREAMING)
	res = append(res, rules.RPC_NO_SERVER_STREAMING)
	return res
}
