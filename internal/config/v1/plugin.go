package v1

import (
	"errors"
	"strings"

	"golang.org/x/mod/semver"
)

// Plugin selects a local, bundled, or remote generator and its output options.
type Plugin struct {
	Name    string        `yaml:"name"`
	Remote  string        `yaml:"remote"`
	Version string        `yaml:"version"`
	Out     string        `yaml:"out"`
	Opts    PluginOptions `yaml:"opts"`
}

// Validate checks the plugin source and reproducibility requirements.
func (p Plugin) Validate() error {
	if (p.Name == "") == (p.Remote == "") {
		return errors.New("exactly one of name or remote is required")
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
