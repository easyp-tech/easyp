package parser

import (
	"strings"
)

// JSONName converts a protobuf field name to its canonical default json_name,
// exactly matching protoc 36.2's algorithm:
// - Naive splitting on underscores ('_')
// - The first character of each word after the first is capitalized
// - Existing uppercase letters are preserved (no lowercasing)
//
// Examples:
//   foo_bar   -> fooBar
//   _foo      -> Foo
//   foo__bar  -> fooBar
//   FOO_BAR   -> FOOBAR
//   foo_      -> foo
func JSONName(name string) string {
	var buf strings.Builder
	buf.Grow(len(name))
	parts := strings.Split(name, "_")
	firstWord := true
	for _, word := range parts {
		if word == "" {
			firstWord = false
			continue
		}
		firstRune := true
		for _, r := range word {
			if (!firstWord) && firstRune {
				buf.WriteRune(toUpper(r))
			} else {
				buf.WriteRune(r)
			}
			firstRune = false
		}
		firstWord = false
	}
	return buf.String()
}

// MapEntryTypeName computes the synthetic PascalCase MapEntry message name,
// e.g. "my_map" -> "MyMapEntry".
func MapEntryTypeName(fieldName string) string {
	var buf strings.Builder
	buf.Grow(len(fieldName) + 5)
	parts := strings.Split(fieldName, "_")
	for _, word := range parts {
		if word == "" {
			continue
		}
		firstRune := true
		for _, r := range word {
			if firstRune {
				buf.WriteRune(toUpper(r))
			} else {
				buf.WriteRune(r)
			}
			firstRune = false
		}
	}
	buf.WriteString("Entry")
	return buf.String()
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - ('a' - 'A')
	}
	return r
}
