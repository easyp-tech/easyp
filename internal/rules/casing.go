package rules

import (
	"strings"
	"unicode"
)

// The casing helpers follow buf's lint rules so that easyp and buf agree on
// names. They are ported from buf.build/go/standard xstrings (Apache-2.0):
// digits do not start a new word, and '.', '-', '_' and whitespace delimit words.

// toPascalCase drops delimiters and capitalizes the first letter of each word.
// Letters that are already uppercase stay uppercase.
func toPascalCase(s string) string {
	var b strings.Builder
	var previous rune
	for i, c := range strings.TrimSpace(s) {
		if !isCaseDelimiter(c) {
			if i == 0 || isCaseDelimiter(previous) || unicode.IsUpper(c) {
				b.WriteRune(unicode.ToUpper(c))
			} else {
				b.WriteRune(unicode.ToLower(c))
			}
		}
		previous = c
	}
	return b.String()
}

func toLowerSnakeCase(s string) string { return strings.ToLower(toSnakeCase(s)) }

func toUpperSnakeCase(s string) string { return strings.ToUpper(toSnakeCase(s)) }

func toSnakeCase(s string) string {
	output := ""
	s = strings.TrimFunc(s, isCaseDelimiter)
	for i, c := range s {
		if isCaseDelimiter(c) {
			c = '_'
		}
		switch {
		case i == 0:
			output += string(c)
		case unicode.IsUpper(c) &&
			output[len(output)-1] != '_' &&
			((i < len(s)-1 && !isSnakeCaseNewWord(rune(s[i+1])) && !isCaseDelimiter(rune(s[i+1]))) ||
				unicode.IsLower(rune(s[i-1]))):
			output += "_" + string(c)
		case !isCaseDelimiter(c) || output[len(output)-1] != '_':
			output += string(c)
		}
	}
	return output
}

// isSnakeCaseNewWord matches buf's check for the character after an uppercase
// letter, where digits count as part of a new word.
func isSnakeCaseNewWord(r rune) bool { return unicode.IsUpper(r) || unicode.IsDigit(r) }

func isCaseDelimiter(r rune) bool {
	return r == '.' || r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
