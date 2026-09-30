package v1

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
	"unicode/utf8"
)

// issuePathMatcher is deliberately limited to issues.exclude-rules.path.
// Other config paths retain their existing semantics.
type issuePathMatcher struct {
	pattern  string
	segments []string
	literal  bool
}

func newIssuePathMatcher(pattern string) (issuePathMatcher, error) {
	if !utf8.ValidString(pattern) || strings.ContainsAny(pattern, "\\:\x00") || strings.HasPrefix(pattern, "/") {
		return issuePathMatcher{}, fmt.Errorf("expected a portable relative path pattern: %q", pattern)
	}
	if strings.Contains(pattern, "//") {
		return issuePathMatcher{}, fmt.Errorf("empty path segment in %q", pattern)
	}
	pattern = strings.TrimPrefix(pattern, "./")
	pattern = strings.TrimSuffix(pattern, "/")
	if pattern == "." || pattern == "" {
		pattern = "**"
	}
	segments := strings.Split(pattern, "/")
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return issuePathMatcher{}, fmt.Errorf("empty, dot, or traversing path segment in %q", pattern)
		}
		if segment == "**" {
			continue
		}
		if strings.Contains(segment, "**") {
			return issuePathMatcher{}, fmt.Errorf("** must be a complete path segment in %q", pattern)
		}
		_, err := path.Match(segment, "")
		if err != nil {
			return issuePathMatcher{}, fmt.Errorf("Match: %w", err)
		}
	}
	return issuePathMatcher{pattern: pattern, segments: segments, literal: !strings.ContainsAny(pattern, "*?[")}, nil
}

func (m issuePathMatcher) matches(relativePath string) bool {
	if !fs.ValidPath(relativePath) || strings.ContainsAny(relativePath, "\\:\x00") {
		return false
	}
	if m.literal {
		return relativePath == m.pattern || (!strings.HasSuffix(m.pattern, ".proto") && strings.HasPrefix(relativePath, m.pattern+"/"))
	}
	// Match complete segments. Dynamic programming bounds the work even when
	// several globstars can consume the same directories.
	segments := strings.Split(relativePath, "/")
	matched := make([]bool, len(segments)+1)
	matched[0] = true
	for _, pattern := range m.segments {
		next := make([]bool, len(matched))
		if pattern == "**" {
			next[0] = matched[0]
			for i := range segments {
				next[i+1] = matched[i+1] || next[i]
			}
		} else {
			for i, segment := range segments {
				// Syntax was validated at construction.
				ok, _ := path.Match(pattern, segment)
				next[i+1] = matched[i] && ok
			}
		}
		matched = next
	}
	return matched[len(segments)]
}

func (p IssuePolicy) validatePaths() error {
	for i, rule := range p.ExcludeRules {
		_, err := newIssuePathMatcher(rule.Path)
		if err != nil {
			return fmt.Errorf("issues.exclude-rules[%d].path: %w", i, err)
		}
	}
	return nil
}

// ExclusionsForPath selects exclusions for a slash-separated path relative to
// the policy file supplying the effective issues section. An empty linter list
// suppresses the whole file; pathless rules apply to every file in the policy.
func (p IssuePolicy) ExclusionsForPath(relativePath string) ([]string, bool, error) {
	var linters []string
	var all bool
	for i, rule := range p.ExcludeRules {
		matcher, err := newIssuePathMatcher(rule.Path)
		if err != nil {
			return nil, false, fmt.Errorf("issues.exclude-rules[%d].path: %w", i, err)
		}
		if rule.Path != "" && !matcher.matches(relativePath) {
			continue
		}
		if len(rule.Linters) == 0 {
			all = true
		}
		linters = append(linters, rule.Linters...)
	}
	return linters, all, nil
}
