package config

// BreakingCheck is the configuration for `breaking` command
type BreakingCheck struct {
	Ignore []string `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	// git ref to compare with
	AgainstGitRef string `json:"against_git_ref,omitempty" yaml:"against_git_ref,omitempty"`
}
