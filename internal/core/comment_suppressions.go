package core

import (
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
