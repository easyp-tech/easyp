package protosource

import (
	"bytes"
	"strings"
	"text/scanner"
)

// Package lexes only top-level package declarations. Scanner tokens keep
// strings and comments opaque, so an option containing "package" cannot select
// a file. It intentionally does not validate unrelated message definitions.
func Package(raw []byte) string {
	var lex scanner.Scanner
	lex.Init(bytes.NewReader(raw))
	lex.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanChars | scanner.ScanComments | scanner.SkipComments
	lex.Error = func(*scanner.Scanner, string) {}
	depth := 0
	start := true
	for token := lex.Scan(); token != scanner.EOF; token = lex.Scan() {
		if depth == 0 && start && token == scanner.Ident && lex.TokenText() == "package" {
			var parts []string
			for {
				if lex.Scan() != scanner.Ident {
					return ""
				}
				parts = append(parts, lex.TokenText())
				switch lex.Scan() {
				case ';':
					return strings.Join(parts, ".")
				case '.':
					continue
				default:
					return ""
				}
			}
		}
		switch token {
		case '{', '(', '[':
			depth++
			start = false
		case '}', ')', ']':
			if depth > 0 {
				depth--
			}
			start = depth == 0
		case ';':
			if depth == 0 {
				start = true
			}
		default:
			if depth == 0 {
				start = false
			}
		}
	}
	return ""
}
