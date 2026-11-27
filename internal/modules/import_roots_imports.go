package modules

import (
	"strings"

	"github.com/easyp-tech/easyprotoc/lexer"
)

// Lex top-level import declarations independently of message bodies, then use
// the compiler parser for modifiers and escaped/concatenated string literals.
// An unrelated unterminated body cannot discard an earlier valid import.
func intrinsicProtoImports(name string, raw []byte) []string {
	tokens := lexer.NewTokenStream(len(raw)/16+8, len(raw)/64+8)
	if err := lexer.Lex(raw, tokens); err != nil {
		return nil
	}
	// next returns the index of the first non-comment token at or after i.
	next := func(i int) int {
		for i < tokens.Len() {
			if kind := tokens.Kinds[i]; kind != lexer.KindLineComment && kind != lexer.KindBlockComment {
				return i
			}
			i++
		}
		return i
	}
	text := func(i int) string {
		token := tokens.Token(i)
		return string(raw[token.Start : token.Start+token.Len])
	}
	is := func(i int, kind lexer.Kind, value string) bool {
		return i < tokens.Len() && tokens.Kinds[i] == kind && text(i) == value
	}
	depth := 0
	var imports []string
	for i := next(0); i < tokens.Len(); {
		if depth == 0 && is(i, lexer.KindIdent, "import") {
			var declaration strings.Builder
			declaration.WriteString("import ")
			i = next(i + 1)
			if is(i, lexer.KindIdent, "weak") || is(i, lexer.KindIdent, "public") {
				declaration.WriteString(text(i))
				declaration.WriteByte(' ')
				i = next(i + 1)
			}
			for i < tokens.Len() && tokens.Kinds[i] == lexer.KindString {
				declaration.WriteString(text(i))
				declaration.WriteByte(' ')
				i = next(i + 1)
			}
			if is(i, lexer.KindPunct, ";") {
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
		if i < tokens.Len() && tokens.Kinds[i] == lexer.KindPunct {
			switch text(i) {
			case "{", "(", "[":
				depth++
			case "}", ")", "]":
				if depth > 0 {
					depth--
				}
			}
		}
		i = next(i + 1)
	}
	return imports
}
