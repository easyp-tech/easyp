package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// EnvProxy is the environment variable that selects the Git host used to fetch dependencies.
const EnvProxy = "EASYP_PROXY"

// LoadEnvironment reads process settings into cfg.
// EnvProxy, when set, replaces the host used to fetch Git dependencies.
func (cfg *Config) LoadEnvironment() error {
	proxy, err := parseProxyHost(os.Getenv(EnvProxy))
	if err != nil {
		return fmt.Errorf("parseProxyHost: %w", err)
	}
	cfg.Proxy = proxy
	return nil
}

// parseProxyHost accepts an empty value, a host, or host:port.
// A scheme, path, user, or query is rejected.
func parseProxyHost(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if strings.Contains(value, "://") || strings.ContainsAny(value, "/?#@\\") {
		return "", fmt.Errorf("EASYP_PROXY must be a host, got %q", value)
	}
	parsed, err := url.Parse("https://" + value)
	if err != nil || parsed.Host != value || parsed.Hostname() == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("EASYP_PROXY must be a host, got %q", value)
	}
	return parsed.Host, nil
}
