package config

// BreakingCheck is the configuration for `breaking` command
type BreakingCheck struct {
	Ignore []string `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	// git ref to compare with
	AgainstGitRef  string   `json:"against_git_ref,omitempty" yaml:"against_git_ref,omitempty"`
	Use            []string `json:"use,omitempty" yaml:"use,omitempty"`
	Categories     []string `json:"categories,omitempty" yaml:"categories,omitempty"`
	IgnoreUnstable bool     `json:"ignore_unstable,omitempty" yaml:"ignore_unstable,omitempty"`
}
