package core

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
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

type directiveComment struct {
	comment     *parser.Comment
	first, last int
}

type parsedDirective struct {
	action  string
	rules   []string
	comment directiveComment
}

// commentSuppressions uses parser comments, not textual substring matching, so
// marker-like text inside strings is never interpreted as a directive.
func (c *Core) commentSuppressions(info ProtoInfo) (suppressionRanges, error) {
	comments := info.directiveComments
	if comments == nil {
		comments = collectDirectiveComments(info.Info)
	}
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
				return nil, fmt.Errorf("%s:%d:%d: malformed lint directive %q", info.Path, comment.comment.Meta.Pos.Line, comment.comment.Meta.Pos.Column, text)
			}
			names = compactDirectiveNames(names)
			for _, name := range names {
				if !known[name] {
					return nil, fmt.Errorf("%s:%d:%d: unknown or unsupported lint rule %q in comment directive", info.Path, comment.comment.Meta.Pos.Line, comment.comment.Meta.Pos.Column, name)
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
				return nil, fmt.Errorf("%s:%d: easyp:enable %s has no preceding disable", info.Path, directive.comment.comment.Meta.Pos.Line, name)
			}
			begin := stack[len(stack)-1]
			pending[name] = stack[:len(stack)-1]
			start := parsed[begin].comment.comment.Meta.Pos.Line
			ranges = append(ranges, suppressionRange{name, start, directive.comment.comment.Meta.Pos.Line - 1})
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
				return nil, fmt.Errorf("%s:%d: suppression %s must annotate a declaration or have a matching easyp:enable", info.Path, directive.comment.comment.Meta.Pos.Line, name)
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

// collectDirectiveComments reads declaration metadata generically across the
// parser's ordered and unordered node types, including inline comments.
func collectDirectiveComments(root any) []directiveComment {
	found := make(map[*parser.Comment]directiveComment)
	seen := make(map[uintptr]bool)
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Interface {
			if !value.IsNil() {
				visit(value.Elem())
			}
			return
		}
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return
			}
			if comment, ok := value.Interface().(*parser.Comment); ok {
				if _, exists := found[comment]; !exists {
					found[comment] = directiveComment{comment: comment}
				}
				return
			}
			pointer := value.Pointer()
			if seen[pointer] {
				return
			}
			seen[pointer] = true
			visit(value.Elem())
			return
		}
		switch value.Kind() {
		case reflect.Struct:
			var location meta.Meta
			field := value.FieldByName("Meta")
			if field.IsValid() && field.CanInterface() {
				location, _ = field.Interface().(meta.Meta)
			}
			if location.Pos.Line > 0 {
				last := max(location.Pos.Line, location.LastPos.Line)
				for _, name := range []string{"Comments", "InlineComment", "InlineCommentBehindLeftCurly"} {
					field := value.FieldByName(name)
					if !field.IsValid() || !field.CanInterface() {
						continue
					}
					var attached []*parser.Comment
					switch comments := field.Interface().(type) {
					case []*parser.Comment:
						attached = comments
					case *parser.Comment:
						if comments != nil {
							attached = []*parser.Comment{comments}
						}
					}
					for _, comment := range attached {
						found[comment] = directiveComment{comment, location.Pos.Line, last}
					}
				}
			}
			for i := 0; i < value.NumField(); i++ {
				field := value.Field(i)
				if field.CanInterface() {
					visit(field)
				}
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < value.Len(); i++ {
				visit(value.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(root))
	result := make([]directiveComment, 0, len(found))
	for _, comment := range found {
		result = append(result, comment)
	}
	slices.SortFunc(result, func(a, b directiveComment) int {
		if a.comment.Meta.Pos.Line != b.comment.Meta.Pos.Line {
			return a.comment.Meta.Pos.Line - b.comment.Meta.Pos.Line
		}
		return a.comment.Meta.Pos.Column - b.comment.Meta.Pos.Column
	})
	return result
}
