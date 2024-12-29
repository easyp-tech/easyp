package config

// LintConfig contains linter configuration.
type LintConfig struct {
	Use                 []string            `json:"use" yaml:"use" env:"USE"`                                                          // Use rules for linter.
	EnumZeroValueSuffix string              `json:"enum_zero_value_suffix" yaml:"enum_zero_value_suffix" env:"ENUM_ZERO_VALUE_SUFFIX"` // Enum zero value suffix.
	ServiceSuffix       string              `json:"service_suffix" yaml:"service_suffix" env:"SERVICE_SUFFIX"`                         // Service suffix.
	Ignore              []string            `json:"ignore" yaml:"ignore" env:"IGNORE"`                                                 // Ignore dirs with proto file.
	Except              []string            `json:"except" yaml:"except" env:"EXCEPT"`                                                 // Except linter rules.
	AllowCommentIgnores bool                `json:"allow_comment_ignores" yaml:"allow_comment_ignores" env:"ALLOW_COMMENT_IGNORES"`    // Allow comment ignore.
	IgnoreOnly          map[string][]string `json:"ignore_only" yaml:"ignore_only" env:"IGNORE_ONLY"`
}
