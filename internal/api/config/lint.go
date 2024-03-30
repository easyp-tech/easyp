package config

import (
	"encoding/json"
	"fmt"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

type rule[T any] struct {
	Value     T
	Activated bool
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *rule[T]) UnmarshalJSON(data []byte) error {
	fmt.Println("UnmarshalJSON", string(data))

	r.Activated = true

	// create not pointer type for empty
	if string(data) == "null" {
		var empty T
		r.Value = empty
		return nil
	}

	return json.Unmarshal(data, &r.Value)
}

// LintConfig contains linter configuration.
type LintConfig struct {
	Use []string `json:"use" yaml:"use" env:"USE"` // For supporting buf format.

	// Minimal
	DirectorySamePackage  rule[rules.DirectorySamePackage]  `json:"directory_same_package" yaml:"directory_same_package" env:"DIRECTORY_SAME_PACKAGE"`
	PackageDefined        rule[rules.PackageDefined]        `json:"package_defined" yaml:"package_defined" env:"PACKAGE_DEFINED"`
	PackageDirectoryMatch rule[rules.PackageDirectoryMatch] `json:"package_directory_match" yaml:"package_directory_match" env:"PACKAGE_DIRECTORY_MATCH"`
	PackageSameDirectory  rule[rules.PackageSameDirectory]  `json:"package_same_directory" yaml:"package_same_directory" env:"PACKAGE_SAME_DIRECTORY"`

	// Basic
	EnumFirstValueZero      rule[rules.EnumFirstValueZero]      `json:"enum_first_value_zero" yaml:"enum_first_value_zero" env:"ENUM_FIRST_VALUE_ZERO"`
	EnumNoAllowAlias        rule[rules.EnumNoAllowAlias]        `json:"enum_no_allow_alias" yaml:"enum_no_allow_alias" env:"ENUM_NO_ALLOW_ALIAS"`
	EnumPascalCase          rule[rules.EnumPascalCase]          `json:"enum_pascal_case" yaml:"enum_pascal_case" env:"ENUM_PASCAL_CASE"`
	EnumValueUpperSnakeCase rule[rules.EnumValueUpperSnakeCase] `json:"enum_value_upper_snake_case" yaml:"enum_value_upper_snake_case" env:"ENUM_VALUE_UPPER_SNAKE_CASE"`
	FieldLowerSnakeCase     rule[rules.FieldLowerSnakeCase]     `json:"field_lower_snake_case" yaml:"field_lower_snake_case" env:"FIELD_LOWER_SNAKE_CASE"`
	ImportNoPublic          rule[rules.ImportNoPublic]          `json:"import_no_public" yaml:"import_no_public" env:"IMPORT_NO_PUBLIC"`
	ImportNoWeak            rule[rules.ImportNoWeak]            `json:"import_no_weak" yaml:"import_no_weak" env:"IMPORT_NO_WEAK"`
	ImportUsed              rule[rules.ImportUsed]              `json:"import_used" yaml:"import_used" env:"IMPORT_USED"`
	MessagePascalCase       rule[rules.MessagePascalCase]       `json:"message_pascal_case" yaml:"message_pascal_case" env:"MESSAGE_PASCAL_CASE"`
	OneofLowerSnakeCase     rule[rules.OneofLowerSnakeCase]     `json:"oneof_lower_snake_case" yaml:"oneof_lower_snake_case" env:"ONEOF_LOWER_SNAKE_CASE"`
	PackageLowerSnakeCase   rule[rules.PackageLowerSnakeCase]   `json:"package_lower_snake_case" yaml:"package_lower_snake_case" env:"PACKAGE_LOWER_SNAKE_CASE"`

	// Default
	EnumValuePrefix          rule[rules.EnumValuePrefix]          `json:"enum_value_prefix" yaml:"enum_value_prefix" env:"ENUM_VALUE_PREFIX"`
	EnumZeroValueSuffix      rule[rules.EnumZeroValueSuffix]      `json:"enum_zero_value_suffix" yaml:"enum_zero_value_suffix" env:"ENUM_ZERO_VALUE_SUFFIX"`
	FileLowerSnakeCase       rule[rules.FileLowerSnakeCase]       `json:"file_lower_snake_case" yaml:"file_lower_snake_case" env:"FILE_LOWER_SNAKE_CASE"`
	RPCRequestResponseUnique rule[rules.RPCRequestResponseUnique] `json:"rpc_request_response_unique" yaml:"rpc_request_response_unique" env:"RPC_REQUEST_RESPONSE_UNIQUE"`
	RPCRequestStandardName   rule[rules.RPCRequestStandardName]   `json:"rpc_request_standard_name" yaml:"rpc_request_standard_name" env:"RPC_REQUEST_STANDARD_NAME"`
	RPCResponseStandardName  rule[rules.RPCResponseStandardName]  `json:"rpc_response_standard_name" yaml:"rpc_response_standard_name" env:"RPC_RESPONSE_STANDARD_NAME"`
	PackageVersionSuffix     rule[rules.PackageVersionSuffix]     `json:"package_version_suffix" yaml:"package_version_suffix" env:"PACKAGE_VERSION_SUFFIX"`
	ServiceSuffix            rule[rules.ServiceSuffix]            `json:"service_suffix" yaml:"service_suffix" env:"SERVICE_SUFFIX"`

	// Comments
	CommentEnum      rule[rules.CommentEnum]      `json:"comment_enum" yaml:"comment_enum" env:"COMMENT_ENUM"`
	CommentEnumValue rule[rules.CommentEnumValue] `json:"comment_enum_value" yaml:"comment_enum_value" env:"COMMENT_ENUM_VALUE"`
	CommentField     rule[rules.CommentField]     `json:"comment_field" yaml:"comment_field" env:"COMMENT_FIELD"`
	CommentMessage   rule[rules.CommentMessage]   `json:"comment_message" yaml:"comment_message" env:"COMMENT_MESSAGE"`
	CommentOneof     rule[rules.CommentOneOf]     `json:"comment_oneof" yaml:"comment_oneof" env:"COMMENT_ONEOF"`
	CommentRPC       rule[rules.CommentRPC]       `json:"comment_rpc" yaml:"comment_rpc" env:"COMMENT_RPC"`
	CommentService   rule[rules.CommentService]   `json:"comment_service" yaml:"comment_service" env:"COMMENT_SERVICE"`

	// Unary rpc
	RPCNoClientStreaming rule[rules.RPCNoClientStreaming] `json:"rpc_no_client_streaming" yaml:"rpc_no_client_streaming" env:"RPC_NO_CLIENT_STREAMING"`
	RPCNoServerStreaming rule[rules.RPCNoServerStreaming] `json:"rpc_no_server_streaming" yaml:"rpc_no_server_streaming" env:"RPC_NO_SERVER_STREAMING"`
}

func (cfg Config) BuildLinterRules() ([]lint.Rule, error) {
	if len(cfg.Lint.Use) > 0 {
		return cfg.buildFromUse()
	}

	return cfg.buildStdRules()
}

func (cfg Config) buildFromUse() ([]lint.Rule, error) {
	var useRule []lint.Rule

	for _, ruleName := range cfg.Lint.Use {
		rule, ok := rules.Rules(rules.Config{
			PackageDirectoryMatchRoot: ".",           // TODO: Move to config
			EnumZeroValueSuffixPrefix: "UNSPECIFIED", // TODO: Move to config
			ServiceSuffixSuffix:       "Service",     // TODO: Move to config
		})[ruleName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", lint.ErrInvalidRule, ruleName)
		}

		useRule = append(useRule, rule)
	}

	return useRule, nil
}

// todo: reflect
func (cfg Config) buildStdRules() ([]lint.Rule, error) {
	var useRule []lint.Rule

	// Minimal
	if cfg.Lint.DirectorySamePackage.Activated {
		useRule = append(useRule, &cfg.Lint.DirectorySamePackage.Value)
	}

	if cfg.Lint.PackageDefined.Activated {
		useRule = append(useRule, &cfg.Lint.PackageDefined.Value)
	}

	if cfg.Lint.PackageDirectoryMatch.Activated {
		useRule = append(useRule, &cfg.Lint.PackageDirectoryMatch.Value)
	}

	if cfg.Lint.PackageSameDirectory.Activated {
		useRule = append(useRule, &cfg.Lint.PackageSameDirectory.Value)
	}

	// Basic

	if cfg.Lint.EnumFirstValueZero.Activated {
		useRule = append(useRule, &cfg.Lint.EnumFirstValueZero.Value)
	}

	if cfg.Lint.EnumNoAllowAlias.Activated {
		useRule = append(useRule, &cfg.Lint.EnumNoAllowAlias.Value)
	}

	if cfg.Lint.EnumPascalCase.Activated {
		useRule = append(useRule, &cfg.Lint.EnumPascalCase.Value)
	}

	if cfg.Lint.EnumValueUpperSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.EnumValueUpperSnakeCase.Value)
	}

	if cfg.Lint.FieldLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.FieldLowerSnakeCase.Value)
	}

	if cfg.Lint.ImportNoPublic.Activated {
		useRule = append(useRule, &cfg.Lint.ImportNoPublic.Value)
	}

	if cfg.Lint.ImportNoWeak.Activated {
		useRule = append(useRule, &cfg.Lint.ImportNoWeak.Value)
	}

	if cfg.Lint.ImportUsed.Activated {
		useRule = append(useRule, &cfg.Lint.ImportUsed.Value)
	}

	if cfg.Lint.MessagePascalCase.Activated {
		useRule = append(useRule, &cfg.Lint.MessagePascalCase.Value)
	}

	if cfg.Lint.OneofLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.OneofLowerSnakeCase.Value)
	}

	if cfg.Lint.PackageLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.PackageLowerSnakeCase.Value)
	}

	// Default
	if cfg.Lint.EnumValuePrefix.Activated {
		useRule = append(useRule, &cfg.Lint.EnumValuePrefix.Value)
	}

	if cfg.Lint.EnumZeroValueSuffix.Activated {
		useRule = append(useRule, &cfg.Lint.EnumZeroValueSuffix.Value)
	}

	if cfg.Lint.FileLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.FileLowerSnakeCase.Value)
	}

	if cfg.Lint.RPCRequestResponseUnique.Activated {
		useRule = append(useRule, &cfg.Lint.RPCRequestResponseUnique.Value)
	}

	if cfg.Lint.RPCRequestStandardName.Activated {
		useRule = append(useRule, &cfg.Lint.RPCRequestStandardName.Value)
	}

	if cfg.Lint.RPCResponseStandardName.Activated {
		useRule = append(useRule, &cfg.Lint.RPCResponseStandardName.Value)
	}

	if cfg.Lint.PackageVersionSuffix.Activated {
		useRule = append(useRule, &cfg.Lint.PackageVersionSuffix.Value)
	}

	if cfg.Lint.ServiceSuffix.Activated {
		useRule = append(useRule, &cfg.Lint.ServiceSuffix.Value)
	}

	// Comments
	if cfg.Lint.RPCNoClientStreaming.Activated {
		useRule = append(useRule, &cfg.Lint.RPCNoClientStreaming.Value)
	}

	if cfg.Lint.RPCNoServerStreaming.Activated {
		useRule = append(useRule, &cfg.Lint.RPCNoServerStreaming.Value)
	}

	return useRule, nil
}
