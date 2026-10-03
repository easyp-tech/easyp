package migration

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

var legacyPseudo = regexp.MustCompile(`^v0\.0\.0-([0-9]{14})-([0-9a-fA-F]+)$`)
var fullSemver = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+].*)?$`)

type requirements struct {
	items    []v1.Requirement
	indirect map[string]bool
}

func (r *requirements) add(raw string, indirect bool) error {
	name, version, hasVersion := strings.Cut(raw, "@")
	if hasVersion && version == "" {
		return fmt.Errorf("empty dependency version in %q", raw)
	}
	if err := validIdentity(name); err != nil {
		return fmt.Errorf("validIdentity: %w", err)
	}
	if err := gitmodules.ValidateSource(name); err != nil {
		return fmt.Errorf("ValidateSource: %w", err)
	}
	version, err := migrationVersion(version)
	if err != nil {
		return fmt.Errorf("migrationVersion: %w", err)
	}
	if r.indirect == nil {
		r.indirect = make(map[string]bool)
	}
	for i, previous := range r.items {
		if previous.Module != name {
			continue
		}
		if previous.Version != "" && version != "" && previous.Version != version {
			return fmt.Errorf("conflicting legacy versions for %s (%s and %s); choose one explicit pin before migration", name, previous.Version, version)
		}
		if previous.Version == "" {
			r.items[i].Version = version
		}
		r.indirect[name] = r.indirect[name] && indirect
		return nil
	}
	r.items = append(r.items, v1.Requirement{Module: name, Version: version})
	r.indirect[name] = indirect
	return nil
}

func migrationVersion(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if strings.Contains(value, "$") {
		return "", fmt.Errorf("dependency version placeholder prevents safe analysis; migrate manually without persisting expanded secrets")
	}
	if match := legacyPseudo.FindStringSubmatch(value); match != nil {
		if _, err := time.Parse("20060102150405", match[1]); err != nil {
			return "", fmt.Errorf("Parse: %w", err)
		}
		if !v1.IsCommitRef(match[2]) {
			return "", fmt.Errorf("legacy pseudo-version %q requires a full Git commit; recover the original pin manually", value)
		}
		return strings.ToLower(match[2]), nil
	}
	if v1.IsCommitRef(value) {
		return strings.ToLower(value), nil
	}
	if semver.IsValid(value) && fullSemver.MatchString(value) {
		return value, nil
	}
	return "", fmt.Errorf("unsupported ref %q: use a semantic version or full Git commit; resolve branches, short hashes and non-semver tags manually", value)
}

func validIdentity(value string) error {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, " \t\r\n@#$()\\") {
		return fmt.Errorf("invalid module identity %q; supply a literal canonical identity with --module", value)
	}
	return nil
}

type legacyReplacement struct{ module, version, target string }

func parseManifest(raw []byte, deps *requirements) ([]v1.Replacement, error) {
	var replacements []legacyReplacement
	block := ""
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(manifestComment(scanner.Text()))
		if text == "" {
			continue
		}
		if text == ")" {
			if block == "" {
				return nil, fmt.Errorf("protobuf.mod:%d: unmatched closing block", line)
			}
			block = ""
			continue
		}
		if strings.HasSuffix(text, "(") {
			if block != "" {
				return nil, fmt.Errorf("protobuf.mod:%d: nested block", line)
			}
			block = strings.TrimSpace(strings.TrimSuffix(text, "("))
			if block != "direct" && block != "indirect" && block != "replace" {
				return nil, fmt.Errorf("protobuf.mod:%d: unknown block %q", line, block)
			}
			continue
		}
		directive, value := block, text
		if block == "" {
			fields := strings.Fields(text)
			if len(fields) > 1 {
				directive, value = fields[0], strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
			} else {
				directive = "direct"
			}
		}
		switch directive {
		case "direct", "indirect":
			if len(strings.Fields(value)) != 1 {
				return nil, fmt.Errorf("protobuf.mod:%d: expected module@version", line)
			}
			if err := deps.add(value, directive == "indirect"); err != nil {
				return nil, fmt.Errorf("add: %w", err)
			}
		case "replace":
			fields := strings.Fields(value)
			if len(fields) != 3 || fields[1] != "=>" {
				return nil, fmt.Errorf("protobuf.mod:%d: expected replace module[@version] => local-path", line)
			}
			name, version, hasVersion := strings.Cut(fields[0], "@")
			if hasVersion && version == "" {
				return nil, fmt.Errorf("empty replacement version for %s", name)
			}
			version, err := migrationVersion(version)
			if err != nil {
				return nil, fmt.Errorf("migrationVersion: %w", err)
			}
			if err := validIdentity(name); err != nil {
				return nil, fmt.Errorf("validIdentity: %w", err)
			}
			if strings.ContainsAny(fields[2], "$@()#") || strings.Contains(fields[2], "://") {
				return nil, fmt.Errorf("replacement for %s must be a literal local path; manual migration required", name)
			}
			replacements = append(replacements, legacyReplacement{module: name, version: version, target: fields[2]})
		default:
			return nil, fmt.Errorf("protobuf.mod:%d: unknown directive %q", line, directive)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Err: %w", err)
	}
	if block != "" {
		return nil, fmt.Errorf("protobuf.mod: unclosed %s block", block)
	}
	var result []v1.Replacement
	seen := make(map[string]bool)
	for _, replacement := range replacements {
		if seen[replacement.module] {
			return nil, fmt.Errorf("ambiguous version-specific replacements for %s; v1 replacements apply to all versions", replacement.module)
		}
		seen[replacement.module] = true
		if replacement.version != "" {
			matched := false
			for _, dep := range deps.items {
				matched = matched || (dep.Module == replacement.module && dep.Version == replacement.version)
			}
			if !matched {
				return nil, fmt.Errorf("version-specific replacement %s@%s does not match a unique required pin; manual migration required", replacement.module, replacement.version)
			}
		}
		result = append(result, v1.Replacement{Module: replacement.module, Target: replacement.target})
	}
	return result, nil
}

func manifestComment(line string) string {
	for i := range line {
		if line[i] == '#' || (strings.HasPrefix(line[i:], "//") && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t')) {
			return line[:i]
		}
	}
	return line
}

func formatManifest(module v1.Module, indirect map[string]bool) []byte {
	var out strings.Builder
	fmt.Fprintf(&out, "module %s\n\nroots (\n", module.Name)
	for _, root := range module.Roots {
		fmt.Fprintf(&out, "\t%s\n", root)
	}
	out.WriteString(")\n")
	for _, req := range module.Requires {
		fmt.Fprintf(&out, "\nrequire %s", req.Module)
		if req.Version != "" {
			fmt.Fprintf(&out, " %s", req.Version)
		}
		if indirect[req.Module] {
			out.WriteString(" // indirect")
		}
		out.WriteByte('\n')
	}
	for _, replacement := range module.Replaces {
		fmt.Fprintf(&out, "\nreplace %s => %s\n", replacement.Module, replacement.Target)
	}
	return []byte(out.String())
}
