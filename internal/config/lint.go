package config

// LintConfig contains linter configuration.
type LintConfig struct {
	Use                 []string            `json:"use,omitempty" yaml:"use,omitempty" env:"USE"`                                                          // Use rules for linter.
	EnumZeroValueSuffix string              `json:"enum_zero_value_suffix,omitempty" yaml:"enum_zero_value_suffix,omitempty" env:"ENUM_ZERO_VALUE_SUFFIX"` // Enum zero value suffix.
	ServiceSuffix       string              `json:"service_suffix,omitempty" yaml:"service_suffix,omitempty" env:"SERVICE_SUFFIX"`                         // Service suffix.
	Ignore              []string            `json:"ignore,omitempty" yaml:"ignore,omitempty" env:"IGNORE"`                                                 // Ignore dirs with proto file.
	Except              []string            `json:"except,omitempty" yaml:"except,omitempty" env:"EXCEPT"`                                                 // Except linter rules.
	AllowCommentIgnores bool                `json:"allow_comment_ignores,omitempty" yaml:"allow_comment_ignores,omitempty" env:"ALLOW_COMMENT_IGNORES"`    // Allow comment ignore.
	IgnoreOnly          map[string][]string `json:"ignore_only,omitempty" yaml:"ignore_only,omitempty" env:"IGNORE_ONLY"`
}
