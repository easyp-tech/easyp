package v1

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// Plugin selects a local, bundled, or remote generator and its output options.
type Plugin struct {
	Name        string        `yaml:"name"`
	Path        string        `yaml:"path"`
	Command     []string      `yaml:"command"`
	Remote      string        `yaml:"remote"`
	Version     string        `yaml:"version"`
	Out         string        `yaml:"out"`
	Opts        PluginOptions `yaml:"opts"`
	WithImports bool          `yaml:"with_imports"`
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
	lastSegment := p.Remote[strings.LastIndex(p.Remote, "/")+1:]
	if colon := strings.LastIndexByte(lastSegment, ':'); colon >= 0 {
		name, embedded := p.Remote[:strings.LastIndexByte(p.Remote, ':')], lastSegment[colon+1:]
		return fmt.Errorf("specify the remote plugin version only in version: split remote into remote: %q and version: %q", name, embedded)
	}
	if !semver.IsValid(p.Version) {
		return errors.New("remote plugin requires a pinned semantic version")
	}
	return nil
}
