package api

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/lint/rules"
)

// Handler is an interface for a handling command.
type Handler interface {
	// Command returns a command.
	Command() *cli.Command
}

// Config is the configuration of easyp.
type Config struct {
	// LintConfig is the lint configuration.
	Lint LintConfig `json:"lint" yaml:"lint" env:"EASYP_LINT"`

	// Deps is the dependencies repositories
	Deps []string `json:"deps" yaml:"deps" env:"EASYP_DEPS"`
}

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

func readConfig(ctx *cli.Context) (*Config, error) {
	cfgFile, err := os.Open(ctx.String(flagCfg.Name))
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	cfg := &Config{}
	err = yaml.NewDecoder(cfgFile).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml.NewDecoder.Decode: %w", err)
	}

	return cfg, nil
}
