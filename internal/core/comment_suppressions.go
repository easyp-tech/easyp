package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/protoast"
)

type suppressionRange struct {
	rule        string
	first, last int
}
type suppressionRanges []suppressionRange

func (ranges suppressionRanges) contains(rule string, line int) bool {
	for _, r := range ranges {
		if r.rule == rule && line >= r.first && line <= r.last {
			return true
		}
	}
	return false
}

// directiveComment is a comment with the line range of the declaration it
// annotates; first is 0 when it annotates nothing.
type directiveComment struct {
	comment     *protoast.Comment
	first, last int
}

func directiveComments(file *protoast.File) []directiveComment {
	if file == nil {
		return nil
	}
	comments := make([]directiveComment, 0, len(file.Comments))
	for _, comment := range file.Comments {
		d := directiveComment{comment: comment}
		if owner := file.Owner(comment); owner != nil {
			d.first, d.last = owner.Pos.Line, max(owner.Pos.Line, owner.End.Line)
		}
		comments = append(comments, d)
	}
	return comments
}

type parsedDirective struct {
	action  string
	rules   []string
	comment directiveComment
}

// commentSuppressions uses parser comments, not textual substring matching, so
// marker-like text inside strings is never interpreted as a directive.
func (c *Core) commentSuppressions(info ProtoInfo) (suppressionRanges, error) {
	comments := directiveComments(info.File)
	known := make(map[string]bool)
	for _, name := range c.knownLintRules {
		known[name] = true
	}
	if len(known) == 0 {
		for _, rule := range c.rules {
			known[GetRuleName(rule)] = true
		}
	}
	var parsed []parsedDirective
	for _, comment := range comments {
		for _, text := range comment.comment.Lines() {
			text = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "*"))
			action, rest, recognized := parseDirectivePrefix(text)
			if !recognized {
				continue
			}
			names := strings.FieldsFunc(rest, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
			if action == "" || len(names) == 0 {
				return nil, fmt.Errorf("%s:%d:%d: malformed lint directive %q", info.Path, comment.comment.Pos.Line, comment.comment.Pos.Column, text)
			}
			names = compactDirectiveNames(names)
			for _, name := range names {
				if !known[name] {
					return nil, fmt.Errorf("%s:%d:%d: unknown or unsupported lint rule %q in comment directive", info.Path, comment.comment.Pos.Line, comment.comment.Pos.Column, name)
				}
			}
			parsed = append(parsed, parsedDirective{action: action, rules: names, comment: comment})
		}
	}
	var ranges suppressionRanges
	pending := make(map[string][]int)
	paired := make(map[string]bool)
	for index, directive := range parsed {
		if directive.action == "legacy" {
			continue
		}
		for _, name := range directive.rules {
			if directive.action == "disable" {
				pending[name] = append(pending[name], index)
				continue
			}
			stack := pending[name]
			if len(stack) == 0 {
				return nil, fmt.Errorf("%s:%d: easyp:enable %s has no preceding disable", info.Path, directive.comment.comment.Pos.Line, name)
			}
			begin := stack[len(stack)-1]
			pending[name] = stack[:len(stack)-1]
			start := parsed[begin].comment.comment.Pos.Line
			ranges = append(ranges, suppressionRange{name, start, directive.comment.comment.Pos.Line - 1})
			paired[fmt.Sprintf("%d:%s", begin, name)] = true
		}
	}
	for index, directive := range parsed {
		if directive.action == "enable" {
			continue
		}
		for _, name := range directive.rules {
			if paired[fmt.Sprintf("%d:%s", index, name)] {
				continue
			}
			if directive.comment.first <= 0 {
				return nil, fmt.Errorf("%s:%d: suppression %s must annotate a declaration or have a matching easyp:enable", info.Path, directive.comment.comment.Pos.Line, name)
			}
			ranges = append(ranges, suppressionRange{name, directive.comment.first, directive.comment.last})
		}
	}
	return ranges, nil
}

func compactDirectiveNames(names []string) []string {
	var result []string
	for _, name := range names {
		if !slices.Contains(result, name) {
			result = append(result, name)
		}
	}
	return result
}

func parseDirectivePrefix(text string) (string, string, bool) {
	for _, prefix := range []struct{ text, action string }{{"easyp:disable", "disable"}, {"easyp:enable", "enable"}, {"nolint:", "legacy"}, {"buf:lint:ignore", "legacy"}} {
		if !strings.HasPrefix(text, prefix.text) {
			continue
		}
		rest := strings.TrimPrefix(text, prefix.text)
		if !strings.HasSuffix(prefix.text, ":") && rest != "" && rest[0] != ' ' && rest[0] != '\t' {
			return "", "", true
		}
		rest, _, _ = strings.Cut(rest, " -- ")
		rest, _, _ = strings.Cut(rest, " // ")
		return prefix.action, strings.TrimSpace(rest), true
	}
	if strings.HasPrefix(text, "easyp:") {
		return "", "", true
	}
	return "", "", false
}
