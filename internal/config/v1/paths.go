package v1

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Exclude portable path separators, glob characters, control characters and
// Unicode whitespace. Literal Unicode characters keep the schema pattern
// compatible with both Go and JSON Schema regular expressions.
const pathSelectorExcluded = `\\:*?\[\]"<>|\x00-\x20\x7f-\x9f` + "\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff"

const pathSelectorCharacter = `[^/` + pathSelectorExcluded + `]`
const pathSelectorNonDotCharacter = `[^/.` + pathSelectorExcluded + `]`
const pathSelectorForbiddenPattern = `[` + pathSelectorExcluded + `]`

// A segment can start with dots but cannot be exactly "." or "..".
const pathSelectorSegment = `(` + pathSelectorNonDotCharacter + `|\.` + pathSelectorNonDotCharacter + `|\.\.` + pathSelectorCharacter + `)` + pathSelectorCharacter + `*`

// PathSelectorPattern is the grammar of a canonical, portable import-relative
// file or subtree selector. A single dot selects the whole source namespace.
const PathSelectorPattern = `^(\.|` + pathSelectorSegment + `(/` + pathSelectorSegment + `)*)$`

var pathSelector = regexp.MustCompile(PathSelectorPattern)

// ValidatePathSelectors checks literal path selection without inspecting sources.
func ValidatePathSelectors(paths []string) error {
	for index, name := range paths {
		if !utf8.ValidString(name) || !pathSelector.MatchString(name) {
			return fmt.Errorf("generate.paths[%d]: %q is not a canonical portable import-relative file or directory path", index, name)
		}
	}
	return nil
}

// PathSelectorMatches reports whether a source import name equals a validated
// selector or lies in its subtree. Matching is bounded by path components.
func PathSelectorMatches(selector, name string) bool {
	return selector == "." || name == selector || strings.HasPrefix(name, selector+"/")
}
