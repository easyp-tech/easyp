// Package syntax parses a single .proto file into easyp-protoc's syntax tree
// without resolving imports or linking. It keeps the token stream, so callers
// can recover comments and raw text that the tree does not store.
package syntax

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

type (
	// AST is the parsed file; see the parser package for node layout.
	AST = parser.ASTFile
	// TextSpan is a [Start, End) byte range into AST.Src.
	TextSpan = parser.TextSpan
	// Diagnostic is a parser diagnostic with a 1-based position.
	Diagnostic = parser.Diagnostic
	// TokenKind classifies a token.
	TokenKind = lexer.Kind
)

// Token kinds callers need to walk the stream.
const (
	KindIdent        = lexer.KindIdent
	KindString       = lexer.KindString
	KindPunct        = lexer.KindPunct
	KindLineComment  = lexer.KindLineComment
	KindBlockComment = lexer.KindBlockComment
)

// ErrSyntax wraps the syntax errors returned by Parse.
var ErrSyntax = errors.New("syntax error")

// File is one parsed file together with its tokens.
type File struct {
	AST    *AST
	tokens *lexer.TokenStream
}

// Parse lexes and parses src. It returns ErrSyntax (with the diagnostics in the
// message) when the file is not syntactically valid. Semantic diagnostics such
// as duplicate field numbers do not fail Parse; see File.Diagnostics.
func Parse(path string, src []byte) (*File, error) {
	// Cap the slice so the parser's appended synthetic bytes never write into
	// the caller's backing array.
	src = src[:len(src):len(src)]
	ts := lexer.NewTokenStream(len(src)/16+8, len(src)/64+8)
	if err := lexer.Lex(src, ts); err != nil {
		return nil, fmt.Errorf("lex %s: %w", path, err)
	}
	ast, _ := parser.NewParser(src, ts, path).ParseFile()
	f := &File{AST: ast, tokens: ts}
	var msgs []string
	for _, d := range ast.Errors {
		if !d.Semantic {
			msgs = append(msgs, d.String())
		}
	}
	if len(msgs) > 0 {
		return f, fmt.Errorf("%w: %s", ErrSyntax, strings.Join(msgs, "; "))
	}
	return f, nil
}

// FromAST wraps a file parsed elsewhere, such as by Link with syntax trees
// kept. The AST must still carry its tokens.
func FromAST(ast *AST) (*File, error) {
	if ast.Tokens == nil {
		return nil, fmt.Errorf("%s: no tokens kept", ast.Path)
	}
	return &File{AST: ast, tokens: ast.Tokens}, nil
}

// Diagnostics returns every parser diagnostic, syntax and semantic.
func (f *File) Diagnostics() []Diagnostic { return f.AST.Errors }

// Src returns the original source without synthetic bytes.
func (f *File) Src() []byte { return f.AST.Src[:f.AST.SrcLen] }

// Text returns the text of span, which may be synthetic.
func (f *File) Text(span TextSpan) string { return span.String(f.AST.Src) }

// IsSynthetic reports whether span points to bytes the parser added.
func (f *File) IsSynthetic(span TextSpan) bool { return span.Start >= f.AST.SrcLen }

// LineCol converts a byte offset to protoc's 1-based line and column: columns
// count bytes and a tab advances to the next multiple of 8.
func (f *File) LineCol(offset uint32) (int, int) { return f.tokens.LineCol(f.Src(), offset) }

// LineStarts returns the byte offset of every line start; LineStarts()[0] == 0.
func (f *File) LineStarts() []uint32 { return f.tokens.Lines }

// NumTokens returns the number of tokens, comments included.
func (f *File) NumTokens() int { return f.tokens.Len() }

// Token returns the kind and [start, end) byte range of the i-th token.
func (f *File) Token(i int) (TokenKind, uint32, uint32) {
	start := f.tokens.Starts[i]
	return f.tokens.Kinds[i], start, start + f.tokens.Lens[i]
}

// IsComment reports whether the i-th token is a comment.
func (f *File) IsComment(i int) bool {
	k := f.tokens.Kinds[i]
	return k == lexer.KindLineComment || k == lexer.KindBlockComment
}

// TokenAt returns the index of the first token that starts at or after offset.
func (f *File) TokenAt(offset uint32) int {
	starts := f.tokens.Starts
	return sort.Search(len(starts), func(i int) bool { return starts[i] >= offset })
}
