package config

// BreakingCheck is the configuration for `breaking` command
type BreakingCheck struct {
	Ignore []string `yaml:"ignore" yaml:"ignore"`
	// git ref to compare with
	AgainstGitRef string `json:"against_git_ref" yaml:"against_git_ref"`
}
