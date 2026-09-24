package v1

import (
	"errors"
	"strings"

	"golang.org/x/mod/semver"
)

// Plugin selects a local, bundled, or remote generator and its output options.
type Plugin struct {
	Name    string        `yaml:"name"`
	Path    string        `yaml:"path"`
	Command []string      `yaml:"command"`
	Remote  string        `yaml:"remote"`
	Version string        `yaml:"version"`
	Out     string        `yaml:"out"`
	Opts    PluginOptions `yaml:"opts"`
}

// Validate checks the plugin source and reproducibility requirements.
func (p Plugin) Validate() error {
	sources := 0
	if p.Name != "" {
		sources++
	}
	if p.Path != "" {
		sources++
	}
	if p.Command != nil {
		sources++
	}
	if p.Remote != "" {
		sources++
	}
	if sources != 1 {
		return errors.New("exactly one of name, path, command or remote is required")
	}
	if p.Command != nil && (len(p.Command) == 0 || strings.TrimSpace(p.Command[0]) == "") {
		return errors.New("command executable is required")
	}
	if p.Out == "" {
		return errors.New("out is required")
	}
	if strings.Contains(p.Version, ":latest") {
		return errors.New("latest version is not reproducible")
	}
	if p.Remote == "" {
		if p.Version != "" {
			return errors.New("local and bundled plugin versions are selected by the executable; version cannot be verified")
		}
		return nil
	}
	if !semver.IsValid(p.Version) {
		return errors.New("remote plugin requires a pinned semantic version")
	}
	lastSegment := p.Remote[strings.LastIndex(p.Remote, "/")+1:]
	if strings.Contains(lastSegment, ":") {
		return errors.New("specify the remote plugin version only in version")
	}
	return nil
}
