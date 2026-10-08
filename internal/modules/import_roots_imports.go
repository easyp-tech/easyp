package modules

import (
	"bytes"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/lexer/scanner"
)

// Lex top-level import declarations independently of message bodies, then use
// the compiler parser for modifiers and escaped/concatenated string literals.
// An unrelated unterminated body cannot discard an earlier valid import.
func intrinsicProtoImports(name string, raw []byte) []string {
	lex := scanner.NewScanner(bytes.NewReader(raw), scanner.WithFilename(name))
	lex.Mode = scanner.ScanKeyword | scanner.ScanStrLit
	depth := 0
	var imports []string
	token, _, _, _ := lex.Scan()
	for token != scanner.TEOF {
		if depth == 0 && token == scanner.TIMPORT {
			var declaration strings.Builder
			declaration.WriteString("import ")
			var text string
			var err error
			token, text, _, err = lex.Scan()
			if token == scanner.TWEAK || token == scanner.TPUBLIC {
				declaration.WriteString(text)
				declaration.WriteByte(' ')
				token, text, _, err = lex.Scan()
			}
			for token == scanner.TSTRLIT && err == nil {
				declaration.WriteString(text)
				declaration.WriteByte(' ')
				token, text, _, err = lex.Scan()
			}
			if token == scanner.TSEMICOLON && err == nil {
				declaration.WriteByte(';')
				parsed, err := ParseProtoImports(name, []byte(declaration.String()))
				if err == nil {
					imports = append(imports, parsed...)
				}
			}
			// Keep the current token for the outer scope tracker when an
			// incomplete declaration ends at a brace or another keyword.
			continue
		}
		switch token {
		case scanner.TLEFTCURLY, scanner.TLEFTPAREN, scanner.TLEFTSQUARE:
			depth++
		case scanner.TRIGHTCURLY, scanner.TRIGHTPAREN, scanner.TRIGHTSQUARE:
			if depth > 0 {
				depth--
			}
		}
		token, _, _, _ = lex.Scan()
	}
	return imports
}
