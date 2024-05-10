package config

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

// LintConfig contains linter configuration.
type LintConfig struct {
	Use                       []string `json:"use" yaml:"use" env:"USE"`                                                                       // Use rules for linter.
	EnumZeroValueSuffixPrefix string   `json:"enumZeroValueSuffixPrefix" yaml:"enumZeroValueSuffixPrefix" env:"ENUM_ZERO_VALUE_SUFFIX_PREFIX"` // Enum zero value suffix prefix.
	ServiceSuffixSuffix       string   `json:"serviceSuffixSuffix" yaml:"serviceSuffixSuffix" env:"SERVICE_SUFFIX_SUFFIX"`                     // Service suffix suffix.
	Ignore                    []string `json:"ignore" yaml:"ignore" env:"IGNORE"`                                                              // Ignore dirs with proto file.
}

func (cfg *Config) BuildLinterRules() ([]lint.Rule, error) {
	cfg.unwrapLintGroups()

	return cfg.buildFromUse()
}

func (cfg *Config) buildFromUse() ([]lint.Rule, error) {
	var useRule []lint.Rule

	for _, ruleName := range cfg.Lint.Use {
		rule, ok := rules.Rules(rules.Config{
			PackageDirectoryMatchRoot: ".",
			EnumZeroValueSuffixPrefix: cfg.Lint.EnumZeroValueSuffixPrefix,
			ServiceSuffixSuffix:       cfg.Lint.ServiceSuffixSuffix,
		})[ruleName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", lint.ErrInvalidRule, ruleName)
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
		}
	}

	cfg.Lint.Use = lo.FindUniques(res)
}

func (cfg *Config) addMinimal(res []string) []string {
	res = append(res, "DIRECTORY_SAME_PACKAGE")
	res = append(res, "PACKAGE_DEFINED")
	res = append(res, "PACKAGE_DIRECTORY_MATCH")
	res = append(res, "PACKAGE_SAME_DIRECTORY")
	return res
}

func (cfg *Config) addBasic(res []string) []string {
	res = append(res, "ENUM_FIRST_VALUE_ZERO")
	res = append(res, "ENUM_NO_ALLOW_ALIAS")
	res = append(res, "ENUM_PASCAL_CASE")
	res = append(res, "ENUM_VALUE_UPPER_SNAKE_CASE")
	res = append(res, "FIELD_LOWER_SNAKE_CASE")
	res = append(res, "IMPORT_NO_PUBLIC")
	res = append(res, "IMPORT_NO_WEAK")
	res = append(res, "IMPORT_USED")
	res = append(res, "MESSAGE_PASCAL_CASE")
	res = append(res, "ONEOF_LOWER_SNAKE_CASE")
	res = append(res, "PACKAGE_LOWER_SNAKE_CASE")
	res = append(res, "PACKAGE_SAME_CSHARP_NAMESPACE")
	res = append(res, "PACKAGE_SAME_GO_PACKAGE")
	res = append(res, "PACKAGE_SAME_JAVA_MULTIPLE_FILES")
	res = append(res, "PACKAGE_SAME_JAVA_PACKAGE")
	res = append(res, "PACKAGE_SAME_PHP_NAMESPACE")
	res = append(res, "PACKAGE_SAME_RUBY_PACKAGE")
	res = append(res, "PACKAGE_SAME_SWIFT_PREFIX")
	res = append(res, "RPC_PASCAL_CASE")
	res = append(res, "SERVICE_PASCAL_CASE")
	return res
}

func (cfg *Config) addDefault(res []string) []string {
	res = append(res, "ENUM_VALUE_PREFIX")
	res = append(res, "ENUM_ZERO_VALUE_SUFFIX")
	res = append(res, "FILE_LOWER_SNAKE_CASE")
	res = append(res, "RPC_REQUEST_RESPONSE_UNIQUE")
	res = append(res, "RPC_REQUEST_STANDARD_NAME")
	res = append(res, "RPC_RESPONSE_STANDARD_NAME")
	res = append(res, "PACKAGE_VERSION_SUFFIX")
	res = append(res, "SERVICE_SUFFIX")
	return res
}

func (cfg *Config) addComments(res []string) []string {
	res = append(res, "COMMENT_ENUM")
	res = append(res, "COMMENT_ENUM_VALUE")
	res = append(res, "COMMENT_FIELD")
	res = append(res, "COMMENT_MESSAGE")
	res = append(res, "COMMENT_ONEOF")
	res = append(res, "COMMENT_RPC")
	res = append(res, "COMMENT_SERVICE")
	return res
}

func (cfg *Config) addUnary(res []string) []string {
	res = append(res, "RPC_NO_CLIENT_STREAMING")
	res = append(res, "RPC_NO_SERVER_STREAMING")
	return res
}
