package parser

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
)

type Parser struct {
	src    []byte
	tokens *lexer.TokenStream
	tokIdx int
	file   *ASTFile
	// prevEnd is the end offset of the last consumed non-comment token.
	prevEnd uint32

	internMap map[string]TextSpan
}

func NewParser(src []byte, ts *lexer.TokenStream, path string) *Parser {
	f := &ASTFile{
		Path:            path,
		DisplayPath:     path,
		Src:             src,
		SrcLen:          uint32(len(src)),
		Tokens:          ts,
		Syntax:          SyntaxProto2, // Default
		ChildIndices:    make([]uint32, 0, 32),
		Imports:         make([]ImportNode, 0, 4),
		Messages:        make([]MessageNode, 0, 4),
		Fields:          make([]FieldNode, 0, 16),
		Enums:           make([]EnumNode, 0, 2),
		EnumValues:      make([]EnumValueNode, 0, 8),
		Options:         make([]OptionNode, 0, 8),
		OptionNameParts: make([]OptionNamePartNode, 0, 8),
		Decls:           make([]DeclRef, 0, 8),
	}
	return &Parser{
		src:       src,
		tokens:    ts,
		tokIdx:    0,
		file:      f,
		internMap: make(map[string]TextSpan),
	}
}

func (p *Parser) SetDisplayPath(dp string) {
	if dp != "" {
		p.file.DisplayPath = dp
	}
}

func (p *Parser) displayPath() string {
	if p.file.DisplayPath != "" {
		return p.file.DisplayPath
	}
	return p.file.Path
}

func (p *Parser) storeIndices(indices []uint32) Span32 {
	if len(indices) == 0 {
		return Span32{Start: 0, Count: 0}
	}
	start := uint32(len(p.file.ChildIndices))
	p.file.ChildIndices = append(p.file.ChildIndices, indices...)
	return Span32{Start: start, Count: uint32(len(indices))}
}

// internBytes appends bytes to p.src and returns a TextSpan referencing it.
func (p *Parser) internBytes(b []byte) TextSpan {
	s := string(b)
	if span, ok := p.internMap[s]; ok {
		return span
	}
	start := uint32(len(p.src))
	p.src = append(p.src, b...)
	end := uint32(len(p.src))
	p.file.Src = p.src
	span := TextSpan{Start: start, End: end}
	p.internMap[s] = span
	return span
}

func (p *Parser) internString(s string) TextSpan {
	return p.internBytes([]byte(s))
}

func (p *Parser) atEnd() bool {
	return p.tokIdx >= p.tokens.Len()
}

func (p *Parser) skipIgnored() {
	for p.tokIdx < p.tokens.Len() {
		k := p.tokens.Kinds[p.tokIdx]
		if k == lexer.KindLineComment || k == lexer.KindBlockComment {
			p.tokIdx++
			continue
		}
		break
	}
}

func (p *Parser) peek() lexer.Token {
	p.skipIgnored()
	if p.atEnd() {
		return lexer.Token{Kind: 0, Start: uint32(len(p.src)), Len: 0}
	}
	return p.tokens.Token(p.tokIdx)
}

func (p *Parser) next() lexer.Token {
	t := p.peek()
	if !p.atEnd() {
		p.tokIdx++
		p.prevEnd = t.Start + t.Len
	}
	return t
}

func (p *Parser) lastTokenEnd() uint32 {
	if p.tokIdx == 0 {
		return 0
	}
	return p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]
}

func (p *Parser) matchPunct(b byte) bool {
	t := p.peek()
	if t.Kind == lexer.KindPunct && t.Len == 1 && p.src[t.Start] == b {
		p.next()
		return true
	}
	return false
}

func (p *Parser) matchIdent(keyword string) bool {
	t := p.peek()
	if t.Kind == lexer.KindIdent && int(t.Len) == len(keyword) && string(p.src[t.Start:t.Start+t.Len]) == keyword {
		p.next()
		return true
	}
	return false
}

func (p *Parser) expectPunct(b byte) bool {
	if p.matchPunct(b) {
		return true
	}
	p.errorAt(p.peek(), fmt.Sprintf("expected '%c'", b))
	return false
}

func (p *Parser) expectIdent(desc string) (TextSpan, bool) {
	t := p.peek()
	if t.Kind == lexer.KindIdent {
		p.next()
		return TextSpan{Start: t.Start, End: t.Start + t.Len}, true
	}
	p.errorAt(t, fmt.Sprintf("expected %s", desc))
	return TextSpan{}, false
}

func (p *Parser) errorAt(t lexer.Token, msg string) {
	line, col := p.tokens.LineCol(p.src, t.Start)
	p.file.Errors = append(p.file.Errors, Diagnostic{
		Path:    p.displayPath(),
		Line:    line,
		Col:     col,
		Message: msg,
	})
}

// semanticErrorAt reports a check on syntactically valid input.
func (p *Parser) semanticErrorAt(t lexer.Token, msg string) {
	p.errorAt(t, msg)
	p.file.Errors[len(p.file.Errors)-1].Semantic = true
}

func (p *Parser) handleLexError(tok lexer.Token) {
	if tok.Len == 0 {
		return
	}
	b := p.src[tok.Start]
	if b == 0x00 {
		p.errorAt(tok, "Invalid control characters encountered in text.")
		return
	}
	if tok.Len >= 2 && p.src[tok.Start] == '/' && p.src[tok.Start+1] == '*' {
		line, col := p.tokens.LineCol(p.src, tok.Start)
		eofLine, eofCol := p.tokens.LineCol(p.src, uint32(len(p.src)))
		p.file.Errors = append(p.file.Errors, Diagnostic{
			Path:    p.displayPath(),
			Line:    eofLine,
			Col:     eofCol,
			Message: "End-of-file inside block comment.",
		})
		p.file.Errors = append(p.file.Errors, Diagnostic{
			Path:    p.displayPath(),
			Line:    line,
			Col:     col,
			Message: "  Comment started here.",
		})
		return
	}
	if b == '"' || b == '\'' {
		line, col := p.tokens.LineCol(p.src, tok.Start+tok.Len)
		p.file.Errors = append(p.file.Errors, Diagnostic{
			Path:    p.displayPath(),
			Line:    line,
			Col:     col,
			Message: fmt.Sprintf("Multiline strings are not allowed. Did you miss a %c?.", b),
		})
		return
	}
	if tok.Len == 2 && p.src[tok.Start] == '0' && (p.src[tok.Start+1] == 'x' || p.src[tok.Start+1] == 'X') {
		p.errorAt(tok, "\"0x\" must be followed by hex digits.")
		return
	}
	if b == '0' && tok.Len > 1 {
		p.errorAt(tok, "Numbers starting with leading zero must be in octal.")
		return
	}
	p.errorAt(tok, fmt.Sprintf("unexpected character %q", p.src[tok.Start:tok.Start+tok.Len]))
}

func (p *Parser) recoverStatement() {
	for !p.atEnd() {
		t := p.peek()
		if t.Kind == lexer.KindPunct && t.Len == 1 {
			if p.src[t.Start] == '}' {
				return
			}
			if p.src[t.Start] == ';' {
				p.next()
				return
			}
		}
		p.next()
	}
}

func (p *Parser) skipBlock() {
	depth := 1
	for !p.atEnd() && depth > 0 {
		t := p.next()
		if t.Kind == lexer.KindPunct && t.Len == 1 {
			if p.src[t.Start] == '{' {
				depth++
			} else if p.src[t.Start] == '}' {
				depth--
			}
		}
	}
}

// ParseFile parses the entire proto file.
func (p *Parser) ParseFile() (*ASTFile, error) {
	p.skipIgnored()

	// Parse optional syntax or edition statement at start of file
	syntaxTok := p.peek()
	if syntaxTok.Kind == lexer.KindError {
		p.handleLexError(syntaxTok)
		p.next()
	} else if p.matchIdent("syntax") {
		p.parseSyntax(syntaxTok)
		p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclSyntax, Index: 0})
	} else if p.matchIdent("edition") {
		p.parseEdition(syntaxTok)
		p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclSyntax, Index: 0})
	}

	topMsgs := make([]uint32, 0, 32)
	topEnums := make([]uint32, 0, 16)
	topOptions := make([]uint32, 0, 16)
	seenTypes := make(map[string]bool)

	for !p.atEnd() {
		p.skipIgnored()
		if p.atEnd() {
			break
		}
		if p.matchPunct(';') {
			continue
		}

		tok := p.peek()
		if tok.Kind == lexer.KindError {
			p.handleLexError(tok)
			p.next()
			continue
		}
		if tok.Kind != lexer.KindIdent {
			p.errorAt(tok, fmt.Sprintf("unexpected token at top level (%s)", tok.Kind.String()))
			before := p.tokIdx
			p.recoverStatement()
			// recoverStatement stops before '}', which has no enclosing block here.
			if p.tokIdx == before {
				p.next()
			}
			continue
		}

		ident := string(p.src[tok.Start : tok.Start+tok.Len])
		switch ident {
		case "package":
			p.next()
			p.parsePackage(tok)
			p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclPackage, Index: 0})
		case "import":
			p.next()
			impIdx := p.parseImport(tok)
			p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclImport, Index: impIdx})
		case "option":
			p.next()
			if optIdx, ok := p.parseOption(tok); ok {
				topOptions = append(topOptions, optIdx)
				p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclOption, Index: optIdx})
			}
		case "message":
			p.next()
			mIdx, ok := p.parseMessage(nil, tok)
			if !ok {
				continue
			}
			mName := p.file.Messages[mIdx].Name.String(p.src)
			if seenTypes[mName] {
				p.semanticErrorAt(lexer.Token{Start: p.file.Messages[mIdx].Name.Start, Len: p.file.Messages[mIdx].Name.Len()},
					fmt.Sprintf("%q is already defined.", mName))
			} else {
				seenTypes[mName] = true
			}
			topMsgs = append(topMsgs, mIdx)
			p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclMessage, Index: mIdx})
		case "enum":
			p.next()
			eIdx, ok := p.parseEnum(nil, tok)
			if !ok {
				continue
			}
			eName := p.file.Enums[eIdx].Name.String(p.src)
			if seenTypes[eName] {
				p.semanticErrorAt(lexer.Token{Start: p.file.Enums[eIdx].Name.Start, Len: p.file.Enums[eIdx].Name.Len()},
					fmt.Sprintf("%q is already defined.", eName))
			} else {
				seenTypes[eName] = true
			}
			topEnums = append(topEnums, eIdx)
			p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclEnum, Index: eIdx})
		case "service":
			p.next()
			svcIdx, ok := p.parseService(tok)
			if !ok {
				continue
			}
			p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclService, Index: svcIdx})
		case "extend":
			p.next()
			extBlockIdx, _ := p.parseExtend(nil, tok)
			p.file.Decls = append(p.file.Decls, DeclRef{Kind: DeclExtend, Index: extBlockIdx})
		default:
			p.errorAt(tok, fmt.Sprintf("unknown top-level declaration '%s'", ident))
			p.recoverStatement()
		}
	}

	p.file.TopLevelMessages = p.storeIndices(topMsgs)
	p.file.TopLevelEnums = p.storeIndices(topEnums)
	p.file.FileOptions = p.storeIndices(topOptions)

	// Gate post-processing: do not process malformed AST
	if len(p.file.Errors) > 0 {
		return p.file, fmt.Errorf("%s: %d errors encountered", p.file.Path, len(p.file.Errors))
	}

	p.postProcess()

	return p.file, nil
}

func (p *Parser) parseSyntax(startTok lexer.Token) {
	if !p.expectPunct('=') {
		p.recoverStatement()
		return
	}
	tok := p.peek()
	if tok.Kind != lexer.KindString {
		p.errorAt(tok, "expected string literal for syntax")
		p.recoverStatement()
		return
	}
	rawVal, _ := p.consumeStringLit()
	val := string(rawVal)
	switch val {
	case "proto2":
		p.file.Syntax = SyntaxProto2
	case "proto3":
		p.file.Syntax = SyntaxProto3
	default:
		p.errorAt(tok, fmt.Sprintf("Unrecognized syntax identifier %q.  This parser only recognizes \"proto2\" and \"proto3\".", val))
	}
	p.expectPunct(';')
	p.file.SyntaxSpan = TextSpan{Start: startTok.Start, End: p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]}
}

func (p *Parser) parseEdition(startTok lexer.Token) {
	if !p.expectPunct('=') {
		p.recoverStatement()
		return
	}
	tok := p.peek()
	if tok.Kind != lexer.KindString {
		p.errorAt(tok, "expected string literal for edition")
		p.recoverStatement()
		return
	}
	rawVal, _ := p.consumeStringLit()
	val := string(rawVal)
	p.file.Syntax = SyntaxEditions
	switch val {
	case "2023":
		p.file.Edition = Edition2023
	case "2024":
		p.file.Edition = Edition2024
	case "2026":
		p.file.Edition = Edition2026
	default:
		p.file.Edition = EditionUnknown
	}
	p.expectPunct(';')
	p.file.SyntaxSpan = TextSpan{Start: startTok.Start, End: p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]}
}

func (p *Parser) parsePackage(startTok lexer.Token) {
	pkgName := p.parseFullIdent()
	if pkgName.IsEmpty() {
		p.errorAt(p.peek(), "expected package name")
		p.recoverStatement()
		return
	}
	p.file.Package = pkgName
	p.expectPunct(';')
	p.file.PackageSpan = TextSpan{Start: startTok.Start, End: p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]}
}

// consumeStringLit consumes one or more adjacent string literals and returns
// their decoded concatenation and the end offset of the last literal, matching
// protoc's ConsumeString. The caller must have peeked a string token.
func (p *Parser) consumeStringLit() ([]byte, uint32) {
	tok := p.peek()
	p.next()
	val := p.unescapeString(tok.Start, tok.Len)
	end := tok.Start + tok.Len
	if p.peek().Kind != lexer.KindString {
		return val, end
	}
	combined := append([]byte(nil), val...)
	for p.peek().Kind == lexer.KindString {
		t := p.peek()
		p.next()
		combined = append(combined, p.unescapeString(t.Start, t.Len)...)
		end = t.Start + t.Len
	}
	return combined, end
}

func (p *Parser) parseImport(startTok lexer.Token) uint32 {
	kind := ImportDefault
	var modSpan TextSpan
	peekTok := p.peek()
	if p.matchIdent("public") {
		kind = ImportPublic
		modSpan = TextSpan{Start: peekTok.Start, End: peekTok.Start + peekTok.Len}
	} else if p.matchIdent("weak") {
		kind = ImportWeak
		modSpan = TextSpan{Start: peekTok.Start, End: peekTok.Start + peekTok.Len}
	}

	tok := p.peek()
	if tok.Kind != lexer.KindString {
		p.errorAt(tok, "expected string literal for import path")
		p.recoverStatement()
		return 0
	}
	unquoted, _ := p.consumeStringLit()
	pathSpan := p.internBytes(unquoted)

	p.expectPunct(';')
	end := p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]

	impIdx := uint32(len(p.file.Imports))
	p.file.Imports = append(p.file.Imports, ImportNode{
		Path:         pathSpan,
		Kind:         kind,
		Span:         TextSpan{Start: startTok.Start, End: end},
		ModifierSpan: modSpan,
	})
	return impIdx
}

func (p *Parser) parseFullIdentWithSpan() (TextSpan, TextSpan) {
	t := p.peek()
	leadingDot := false
	sourceStart := t.Start
	if t.Kind == lexer.KindPunct && t.Len == 1 && p.src[t.Start] == '.' {
		leadingDot = true
		p.next()
	}

	startTok := p.peek()
	if startTok.Kind != lexer.KindIdent {
		return TextSpan{}, TextSpan{}
	}
	if !leadingDot {
		sourceStart = startTok.Start
	}
	p.next()

	var parts []string
	if leadingDot {
		parts = append(parts, ".")
	}
	parts = append(parts, string(p.src[startTok.Start:startTok.Start+startTok.Len]))

	hasWhitespace := false
	prevEnd := startTok.Start + startTok.Len

	for {
		nextTok := p.peek()
		if nextTok.Kind == lexer.KindPunct && nextTok.Len == 1 && p.src[nextTok.Start] == '.' {
			if nextTok.Start != prevEnd {
				hasWhitespace = true
			}
			p.next() // consume '.'

			idTok := p.peek()
			if idTok.Kind != lexer.KindIdent {
				p.errorAt(idTok, "expected identifier after '.'")
				break
			}
			if idTok.Start != nextTok.Start+1 {
				hasWhitespace = true
			}
			p.next() // consume ident

			parts = append(parts, "."+string(p.src[idTok.Start:idTok.Start+idTok.Len]))
			prevEnd = idTok.Start + idTok.Len
		} else {
			break
		}
	}

	srcSpan := TextSpan{Start: sourceStart, End: prevEnd}

	if len(parts) == 1 && !leadingDot {
		return srcSpan, srcSpan
	}
	if !hasWhitespace && !leadingDot {
		return srcSpan, srcSpan
	}
	if !hasWhitespace && leadingDot && startTok.Start == t.Start+1 {
		return srcSpan, srcSpan
	}

	// Contains whitespace or formatted components: assemble clean string
	clean := strings.Join(parts, "")
	return p.internString(clean), srcSpan
}

func (p *Parser) parseFullIdent() TextSpan {
	name, _ := p.parseFullIdentWithSpan()
	return name
}

func (p *Parser) parseOption(startTok lexer.Token) (uint32, bool) {
	opt := p.parseSingleOption(startTok, true)
	if opt != nil {
		idx := uint32(len(p.file.Options))
		p.file.Options = append(p.file.Options, *opt)
		return idx, true
	}
	return 0, false
}

func (p *Parser) parseSingleOption(startTok lexer.Token, requireSemicolon bool) *OptionNode {
	namePartsStart := uint32(len(p.file.OptionNameParts))
	nameStart := p.peek().Start

	for {
		isExt := false
		var partName TextSpan
		if p.matchPunct('(') {
			isExt = true
			partName = p.parseFullIdent()
			if !p.expectPunct(')') {
				p.recoverStatement()
				return nil
			}
		} else {
			var ok bool
			partName, ok = p.expectIdent("option name")
			if !ok {
				p.recoverStatement()
				return nil
			}
		}

		p.file.OptionNameParts = append(p.file.OptionNameParts, OptionNamePartNode{
			Name:        partName,
			IsExtension: isExt,
		})

		if p.matchPunct('.') {
			continue
		}
		break
	}

	namePartsCount := uint32(len(p.file.OptionNameParts)) - namePartsStart
	nameSpan := TextSpan{Start: nameStart, End: p.prevEnd}

	if !p.expectPunct('=') {
		p.recoverStatement()
		return nil
	}

	valNode := p.parseOptionValue()
	if valNode == nil {
		p.recoverStatement()
		return nil
	}

	valNode.NameParts = Span32{Start: namePartsStart, Count: namePartsCount}
	valNode.NameSpan = nameSpan
	if namePartsCount == 1 {
		part := p.file.OptionNameParts[namePartsStart]
		if !part.IsExtension {
			optName := part.Name.String(p.src)
			if optName == "deprecated" {
				valStr := valNode.ValSpan.String(p.src)
				if valNode.ValKind != OptionValIdent || (valStr != "true" && valStr != "false") {
					p.semanticErrorAt(lexer.Token{Start: valNode.Span.Start, Len: valNode.Span.Len()},
						"Value must be identifier for boolean option \"google.protobuf.MessageOptions.deprecated\".")
				}
			}
		}
	}
	if requireSemicolon {
		p.expectPunct(';')
		valNode.Span = TextSpan{Start: startTok.Start, End: p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]}
	} else {
		valNode.Span = TextSpan{Start: startTok.Start, End: p.peek().Start}
	}

	return valNode
}

func (p *Parser) parseOptionValue() *OptionNode {
	start := p.peek().Start
	val := p.parseOptionValueLiteral()
	if val != nil {
		val.ValRawSpan = TextSpan{Start: start, End: p.prevEnd}
	}
	return val
}

func (p *Parser) parseOptionValueLiteral() *OptionNode {
	p.skipIgnored()
	tok := p.peek()

	if tok.Kind == lexer.KindError {
		p.handleLexError(tok)
		p.next()
		return nil
	}

	// 1. String literal (possibly concatenated)
	if tok.Kind == lexer.KindString {
		var combined []byte
		start := tok.Start
		end := tok.Start + tok.Len
		for {
			t := p.peek()
			if t.Kind != lexer.KindString {
				break
			}
			p.next()
			combined = append(combined, p.unescapeString(t.Start, t.Len)...)
			end = t.Start + t.Len
		}
		span := p.internBytes(combined)
		return &OptionNode{
			ValKind: OptionValString,
			ValSpan: span,
			Span:    TextSpan{Start: start, End: end},
		}
	}

	// 2. Aggregate message literal: { field: val, ... } or < field: val, ... >
	if tok.Kind == lexer.KindPunct && tok.Len == 1 && (p.src[tok.Start] == '{' || p.src[tok.Start] == '<') {
		closeChar := byte('}')
		if p.src[tok.Start] == '<' {
			closeChar = '>'
		}
		p.next()
		var childIdxs []uint32
		for !p.atEnd() {
			p.skipIgnored()
			if p.matchPunct(closeChar) {
				break
			}
			if p.matchPunct(';') || p.matchPunct(',') {
				continue
			}

			subOpt := p.parseAggregateField()
			if subOpt != nil {
				idx := uint32(len(p.file.Options))
				p.file.Options = append(p.file.Options, *subOpt)
				childIdxs = append(childIdxs, idx)
			} else {
				p.skipBlock()
				break
			}
			p.matchPunct(',')
			p.matchPunct(';')
		}
		subStart := uint32(len(p.file.ChildIndices))
		p.file.ChildIndices = append(p.file.ChildIndices, childIdxs...)
		subCount := uint32(len(childIdxs))
		return &OptionNode{
			ValKind:    OptionValAggregate,
			SubOptions: Span32{Start: subStart, Count: subCount},
			Span:       TextSpan{Start: tok.Start, End: p.peek().Start},
		}
	}

	// 2b. List literal: [ elem1, elem2, ... ]
	if tok.Kind == lexer.KindPunct && tok.Len == 1 && p.src[tok.Start] == '[' {
		p.next()
		var childIdxs []uint32
		for !p.atEnd() {
			p.skipIgnored()
			if p.matchPunct(']') {
				break
			}
			if p.matchPunct(';') || p.matchPunct(',') {
				continue
			}

			subOpt := p.parseOptionValue()
			if subOpt != nil {
				idx := uint32(len(p.file.Options))
				p.file.Options = append(p.file.Options, *subOpt)
				childIdxs = append(childIdxs, idx)
			} else {
				break
			}
			p.matchPunct(',')
			p.matchPunct(';')
		}
		subStart := uint32(len(p.file.ChildIndices))
		p.file.ChildIndices = append(p.file.ChildIndices, childIdxs...)
		subCount := uint32(len(childIdxs))
		return &OptionNode{
			ValKind:    OptionValList,
			SubOptions: Span32{Start: subStart, Count: subCount},
			Span:       TextSpan{Start: tok.Start, End: p.peek().Start},
		}
	}

	// 3. Number (optional sign + int/float)
	isNeg := false
	if tok.Kind == lexer.KindPunct && tok.Len == 1 && (p.src[tok.Start] == '-' || p.src[tok.Start] == '+') {
		if p.src[tok.Start] == '-' {
			isNeg = true
		}
		p.next()
		tok = p.peek()
	}

	if tok.Kind == lexer.KindInt {
		p.next()
		raw := string(p.src[tok.Start : tok.Start+tok.Len])
		uval, _ := strconv.ParseUint(raw, 0, 64)
		var valSpan TextSpan
		if isNeg {
			valSpan = p.internString("-" + raw)
		} else {
			valSpan = TextSpan{Start: tok.Start, End: tok.Start + tok.Len}
		}
		return &OptionNode{
			ValKind:    OptionValInt,
			ValSpan:    valSpan,
			IntVal:     uval,
			IsNegative: isNeg,
		}
	}

	if tok.Kind == lexer.KindFloat {
		p.next()
		raw := string(p.src[tok.Start : tok.Start+tok.Len])
		fval, _ := strconv.ParseFloat(raw, 64)
		if isNeg {
			fval = -fval
		}
		var valSpan TextSpan
		if isNeg {
			valSpan = p.internString("-" + raw)
		} else {
			valSpan = TextSpan{Start: tok.Start, End: tok.Start + tok.Len}
		}
		return &OptionNode{
			ValKind:    OptionValFloat,
			ValSpan:    valSpan,
			FloatVal:   fval,
			IsNegative: isNeg,
		}
	}

	// 4. Identifier (or bool, or enum constant)
	if tok.Kind == lexer.KindIdent {
		p.next()
		var valSpan TextSpan
		if isNeg {
			valSpan = p.internString("-" + string(p.src[tok.Start:tok.Start+tok.Len]))
		} else {
			valSpan = TextSpan{Start: tok.Start, End: tok.Start + tok.Len}
		}
		return &OptionNode{
			ValKind:    OptionValIdent,
			ValSpan:    valSpan,
			IsNegative: isNeg,
		}
	}

	p.errorAt(tok, fmt.Sprintf("unexpected token in option value: %s", tok.Kind.String()))
	return nil
}

func (p *Parser) parseAggregateField() *OptionNode {
	startTok := p.peek()
	namePartsStart := uint32(len(p.file.OptionNameParts))

	isExt := false
	var partName TextSpan
	if p.matchPunct('[') {
		isExt = true
		partName = p.parseFullIdent()
		if !p.expectPunct(']') {
			return nil
		}
	} else if p.matchPunct('(') {
		isExt = true
		partName = p.parseFullIdent()
		if !p.expectPunct(')') {
			return nil
		}
	} else {
		var ok bool
		partName, ok = p.expectIdent("aggregate field name")
		if !ok {
			return nil
		}
	}

	p.file.OptionNameParts = append(p.file.OptionNameParts, OptionNamePartNode{
		Name:        partName,
		IsExtension: isExt,
	})

	for p.matchPunct('.') {
		subName, ok := p.expectIdent("aggregate subfield name")
		if !ok {
			return nil
		}
		p.file.OptionNameParts = append(p.file.OptionNameParts, OptionNamePartNode{
			Name:        subName,
			IsExtension: false,
		})
	}

	namePartsCount := uint32(len(p.file.OptionNameParts)) - namePartsStart
	nameSpan := TextSpan{Start: startTok.Start, End: p.prevEnd}

	// Separator in aggregate can be ':', '=', or omitted before '{' or '<'
	if !p.matchPunct(':') && !p.matchPunct('=') {
		tok := p.peek()
		if tok.Kind != lexer.KindPunct || (p.src[tok.Start] != '{' && p.src[tok.Start] != '<') {
			p.errorAt(tok, "expected ':' or '='")
			return nil
		}
	}

	valNode := p.parseOptionValue()
	if valNode == nil {
		return nil
	}

	valNode.NameParts = Span32{Start: namePartsStart, Count: namePartsCount}
	valNode.NameSpan = nameSpan
	valNode.Span = TextSpan{Start: startTok.Start, End: p.peek().Start}
	return valNode
}

func (p *Parser) parseMessage(parent *MessageNode, msgTok lexer.Token) (uint32, bool) {
	name, ok := p.expectIdent("message name")
	if !ok {
		p.recoverStatement()
		return 0, false
	}

	msgIdx := uint32(len(p.file.Messages))
	p.file.Messages = append(p.file.Messages, MessageNode{
		Name: name,
		Span: TextSpan{Start: msgTok.Start, End: name.End},
	})

	openBraceTok := p.peek()
	if !p.expectPunct('{') {
		p.recoverStatement()
		return msgIdx, true
	}
	openBraceSpan := TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}

	fieldIndices := make([]uint32, 0, 16)
	nestedMsgIndices := make([]uint32, 0, 8)
	nestedEnumIndices := make([]uint32, 0, 4)
	oneofIndices := make([]uint32, 0, 4)
	extRangeIndices := make([]uint32, 0, 4)
	extIndices := make([]uint32, 0, 4)
	resRangeIndices := make([]uint32, 0, 4)
	resNameIndices := make([]uint32, 0, 4)
	mOptionIndices := make([]uint32, 0, 4)
	decls := make([]DeclRef, 0, 16)

	msg := &p.file.Messages[msgIdx]
	msgNameStr := name.String(p.src)
	seenFields := make(map[string]bool)
	usedTags := make(map[int32]string)

	checkField := func(fIdx uint32) {
		fld := &p.file.Fields[fIdx]
		fldName := fld.Name.String(p.src)
		if seenFields[fldName] {
			p.semanticErrorAt(lexer.Token{Start: fld.Name.Start, Len: fld.Name.Len()},
				fmt.Sprintf("%q is already defined in %q.", fldName, msgNameStr))
		} else {
			seenFields[fldName] = true
		}
		if fld.Number > 0 && fld.Number <= 536870911 && (fld.Number < 19000 || fld.Number > 19999) {
			if firstField, used := usedTags[fld.Number]; used {
				nextAvail := int32(1)
				for usedTags[nextAvail] != "" {
					nextAvail++
				}
				p.semanticErrorAt(lexer.Token{Start: fld.NumberSpan.Start, Len: fld.NumberSpan.Len()},
					fmt.Sprintf("Field number %d has already been used in %q by field %q. Next available field number is %d.", fld.Number, msgNameStr, firstField, nextAvail))
			} else {
				usedTags[fld.Number] = fldName
			}
		}
	}

	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct('}') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			continue
		}

		tok := p.peek()
		isLeadingDot := tok.Kind == lexer.KindPunct && tok.Len == 1 && p.src[tok.Start] == '.'
		if tok.Kind != lexer.KindIdent && !isLeadingDot {
			p.errorAt(tok, "unexpected token in message body")
			p.recoverStatement()
			continue
		}

		kw := string(p.src[tok.Start : tok.Start+tok.Len])
		switch kw {
		case "message":
			p.next()
			subIdx, ok := p.parseMessage(msg, tok)
			if !ok {
				continue
			}
			nestedMsgIndices = append(nestedMsgIndices, subIdx)
			decls = append(decls, DeclRef{Kind: DeclMessage, Index: subIdx})
		case "enum":
			p.next()
			subIdx, ok := p.parseEnum(msg, tok)
			if !ok {
				continue
			}
			nestedEnumIndices = append(nestedEnumIndices, subIdx)
			decls = append(decls, DeclRef{Kind: DeclEnum, Index: subIdx})
		case "extend":
			p.next()
			extBlockIdx, newExts := p.parseExtend(msg, tok)
			extIndices = append(extIndices, newExts...)
			decls = append(decls, DeclRef{Kind: DeclExtend, Index: extBlockIdx})
		case "oneof":
			p.next()
			ooIdx, ooFields := p.parseOneof(msg, int32(len(oneofIndices)), tok)
			oneofIndices = append(oneofIndices, ooIdx)
			fieldIndices = append(fieldIndices, ooFields...)
			decls = append(decls, DeclRef{Kind: DeclOneof, Index: ooIdx})
			for _, fIdx := range ooFields {
				decls = append(decls, DeclRef{Kind: DeclField, Index: fIdx})
				checkField(fIdx)
			}
		case "extensions":
			p.next()
			erDeclIdx := p.parseExtensions(tok, &extRangeIndices)
			decls = append(decls, DeclRef{Kind: DeclExtensionRange, Index: erDeclIdx})
		case "reserved":
			p.next()
			resDeclIdx := p.parseReserved(tok, false, &resRangeIndices, &resNameIndices)
			decls = append(decls, DeclRef{Kind: DeclReserved, Index: resDeclIdx})
		case "option":
			p.next()
			if optIdx, ok := p.parseOption(tok); ok {
				mOptionIndices = append(mOptionIndices, optIdx)
				decls = append(decls, DeclRef{Kind: DeclOption, Index: optIdx})
			}
		case "map":
			p.next()
			fldIdx, entryMsgIdx := p.parseMapField(msg, tok)
			nestedMsgIndices = append(nestedMsgIndices, entryMsgIdx)
			fieldIndices = append(fieldIndices, fldIdx)
			decls = append(decls, DeclRef{Kind: DeclField, Index: fldIdx})
			checkField(fldIdx)
		default:
			// Normal field or group
			fldIdx, groupMsgIdx, isGroup, ok := p.parseFieldOrGroup(msg)
			if ok {
				if isGroup {
					nestedMsgIndices = append(nestedMsgIndices, groupMsgIdx)
				}
				fieldIndices = append(fieldIndices, fldIdx)
				decls = append(decls, DeclRef{Kind: DeclField, Index: fldIdx})
				checkField(fldIdx)
			}
		}
	}

	if !closed {
		p.errorAt(p.peek(), "expected '}'")
	}

	end := p.tokens.Starts[p.tokIdx-1] + p.tokens.Lens[p.tokIdx-1]
	m := &p.file.Messages[msgIdx]
	m.Span = TextSpan{Start: msgTok.Start, End: end}
	m.OpenBraceSpan = openBraceSpan
	m.Fields = p.storeIndices(fieldIndices)
	m.NestedMessages = p.storeIndices(nestedMsgIndices)
	m.NestedEnums = p.storeIndices(nestedEnumIndices)
	m.Oneofs = p.storeIndices(oneofIndices)
	m.ExtensionRanges = p.storeIndices(extRangeIndices)
	m.Extensions = p.storeIndices(extIndices)
	m.ReservedRanges = p.storeIndices(resRangeIndices)
	m.ReservedNames = p.storeIndices(resNameIndices)
	m.Options = p.storeIndices(mOptionIndices)
	m.Decls = decls

	return msgIdx, true
}

func (p *Parser) parseFieldOrGroup(msg *MessageNode) (fldIdx uint32, groupMsgIdx uint32, isGroup bool, ok bool) {
	startTok := p.peek()
	label := LabelUnspecified
	var labelSpan TextSpan
	isProto3Opt := false

	// Check optional/required/repeated label
	if p.matchIdent("optional") {
		label = LabelOptional
		labelSpan = TextSpan{Start: startTok.Start, End: startTok.Start + startTok.Len}
		if p.file.Syntax == SyntaxProto3 {
			isProto3Opt = true
		}
	} else if p.matchIdent("required") {
		label = LabelRequired
		labelSpan = TextSpan{Start: startTok.Start, End: startTok.Start + startTok.Len}
	} else if p.matchIdent("repeated") {
		label = LabelRepeated
		labelSpan = TextSpan{Start: startTok.Start, End: startTok.Start + startTok.Len}
	} else {
		// Label omitted
		if p.file.Syntax == SyntaxProto3 || p.file.Syntax == SyntaxEditions {
			label = LabelOptional
		}
	}

	// Check for group
	groupKwTok := p.peek()
	if p.matchIdent("group") {
		var grpOk bool
		fldIdx, groupMsgIdx, grpOk = p.parseGroup(msg, label, labelSpan, startTok, groupKwTok)
		return fldIdx, groupMsgIdx, true, grpOk
	}

	// Field type
	typeSpan, typeSrcSpan := p.parseFullIdentWithSpan()
	if typeSpan.IsEmpty() {
		p.errorAt(p.peek(), "expected field type")
		p.recoverStatement()
		return 0, 0, false, false
	}

	// Field name
	nameSpan, okIdent := p.expectIdent("field name")
	if !okIdent {
		p.recoverStatement()
		return 0, 0, false, false
	}

	// '='
	if !p.expectPunct('=') {
		p.recoverStatement()
		return 0, 0, false, false
	}

	// Field number
	numTok := p.peek()
	if numTok.Kind != lexer.KindInt {
		p.errorAt(numTok, "expected field number")
		p.recoverStatement()
		return 0, 0, false, false
	}
	p.next()
	numVal, _ := strconv.ParseInt(string(p.src[numTok.Start:numTok.Start+numTok.Len]), 0, 32)
	numSpan := TextSpan{Start: numTok.Start, End: numTok.Start + numTok.Len}

	msgName := ""
	if msg != nil {
		msgName = msg.Name.String(p.src)
	}
	if numVal <= 0 {
		p.semanticErrorAt(numTok, "Field numbers must be positive integers.")
		if msgName != "" {
			p.semanticErrorAt(numTok, fmt.Sprintf("Suggested field numbers for %s: 1", msgName))
		}
	} else if numVal >= 19000 && numVal <= 19999 {
		p.file.Errors = append(p.file.Errors, Diagnostic{
			Path:     p.displayPath(),
			Line:     0,
			Col:      0,
			Message:  "Field numbers 19000 through 19999 are reserved for the protocol buffer library implementation.",
			Semantic: true,
		})
	} else if numVal > 536870911 {
		p.semanticErrorAt(numTok, "Field numbers cannot be greater than 536870911.")
		if msgName != "" {
			p.semanticErrorAt(numTok, fmt.Sprintf("Suggested field numbers for %s: 1", msgName))
		}
	}

	// Field options [ ... ]
	var customJSONName TextSpan
	var customJSONOpt TextSpan
	var customJSONVal TextSpan
	var defaultVal TextSpan
	var defaultValSpan TextSpan
	var optionsSpan TextSpan
	var defaultOpt TextSpan
	fldOptionIndices := p.parseFieldOptions(&customJSONName, &customJSONOpt, &customJSONVal, &defaultVal, &defaultValSpan, &defaultOpt, &optionsSpan)

	if !p.matchPunct(';') {
		if len(p.file.Errors) > 0 {
			p.recoverStatement()
			return 0, 0, false, false
		}
		p.expectPunct(';')
	}

	typeKind := resolvePrimitiveType(typeSpan.String(p.src))

	fldIdx = uint32(len(p.file.Fields))
	p.file.Fields = append(p.file.Fields, FieldNode{
		Name:              nameSpan,
		TypeName:          typeSpan,
		TypeSpan:          typeSrcSpan,
		Number:            int32(numVal),
		Label:             label,
		TypeKind:          typeKind,
		OneofIndex:        -1,
		Proto3Optional:    isProto3Opt,
		Span:              TextSpan{Start: startTok.Start, End: p.lastTokenEnd()},
		Options:           p.storeIndices(fldOptionIndices),
		CustomJSONName:    customJSONName,
		DefaultVal:        defaultVal,
		LabelSpan:         labelSpan,
		NumberSpan:        numSpan,
		OptionsSpan:       optionsSpan,
		CustomJSONNameOpt: customJSONOpt,
		CustomJSONNameVal: customJSONVal,
		DefaultValSpan:    defaultValSpan,
		DefaultOpt:        defaultOpt,
	})
	return fldIdx, 0, false, true
}

// parseFieldOptions parses an option bracket list: [opt1 = val1, opt2 = val2]
func (p *Parser) parseFieldOptions(customJSON *TextSpan, customJSONOpt *TextSpan, customJSONVal *TextSpan, defaultVal *TextSpan, defaultValSpan *TextSpan, defaultOpt *TextSpan, optionsSpan *TextSpan) []uint32 {
	startTok := p.peek()
	if !p.matchPunct('[') {
		return nil
	}
	openBracketTok := startTok
	optIndices := make([]uint32, 0, 4)
	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct(']') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			break
		}

		optIdx, isUninterpreted, ok := p.parseFieldOptionItem(customJSON, customJSONOpt, customJSONVal, defaultVal, defaultValSpan, defaultOpt)
		if ok && isUninterpreted {
			optIndices = append(optIndices, optIdx)
		}

		p.skipIgnored()
		if p.matchPunct(']') {
			closed = true
			break
		}
		if p.matchPunct(',') {
			continue
		}

		// Malformed or unexpected token between options: recover to next option or bracket
		if !ok || (!p.atEnd() && p.peek().Kind == lexer.KindPunct && (p.src[p.peek().Start] == ';' || p.src[p.peek().Start] == '}')) {
			p.recoverOptionList()
			if p.matchPunct(']') {
				closed = true
			}
			break
		}

		p.errorAt(p.peek(), "expected ',' or ']'")
		p.recoverOptionList()
		if p.matchPunct(']') {
			closed = true
		}
		break
	}
	if !closed {
		p.errorAt(p.peek(), "Expected \"]\".")
	}
	if optionsSpan != nil {
		*optionsSpan = TextSpan{Start: openBracketTok.Start, End: p.lastTokenEnd()}
	}
	return optIndices
}

func formatFloatProto(f float64) string {
	if math.IsNaN(f) {
		return "nan"
	}
	if math.IsInf(f, 1) {
		return "inf"
	}
	if math.IsInf(f, -1) {
		return "-inf"
	}
	abs := math.Abs(f)
	if abs == 0 {
		return "0"
	}
	if abs >= 1e-4 && abs < 1e14 {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.TrimRight(s, ".")
		}
		return s
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func (p *Parser) normalizeDefaultValue(valNode *OptionNode) TextSpan {
	switch valNode.ValKind {
	case OptionValInt:
		if valNode.IsNegative {
			return p.internString("-" + strconv.FormatUint(valNode.IntVal, 10))
		}
		return p.internString(strconv.FormatUint(valNode.IntVal, 10))
	case OptionValFloat:
		return p.internString(formatFloatProto(valNode.FloatVal))
	default:
		return valNode.ValSpan
	}
}

func (p *Parser) recoverOptionList() {
	for !p.atEnd() {
		t := p.peek()
		if t.Kind == lexer.KindPunct && t.Len == 1 {
			b := p.src[t.Start]
			if b == ']' || b == ';' || b == '}' {
				return
			}
			if b == ',' {
				p.next()
				return
			}
		}
		p.next()
	}
}

func (p *Parser) parseFieldOptionItem(customJSON *TextSpan, customJSONOpt *TextSpan, customJSONVal *TextSpan, defaultVal *TextSpan, defaultValSpan *TextSpan, defaultOpt *TextSpan) (uint32, bool, bool) {
	startTok := p.peek()
	namePartsStart := uint32(len(p.file.OptionNameParts))

	isExt := false
	var partName TextSpan
	if p.matchPunct('(') {
		isExt = true
		partName = p.parseFullIdent()
		if !p.expectPunct(')') {
			return 0, false, false
		}
	} else {
		var ok bool
		partName, ok = p.expectIdent("option name")
		if !ok {
			return 0, false, false
		}
	}

	p.file.OptionNameParts = append(p.file.OptionNameParts, OptionNamePartNode{
		Name:        partName,
		IsExtension: isExt,
	})

	for p.matchPunct('.') {
		subName, ok := p.expectIdent("option subfield name")
		if !ok {
			return 0, false, false
		}
		p.file.OptionNameParts = append(p.file.OptionNameParts, OptionNamePartNode{
			Name:        subName,
			IsExtension: false,
		})
	}

	namePartsCount := uint32(len(p.file.OptionNameParts)) - namePartsStart
	nameSpan := TextSpan{Start: startTok.Start, End: p.prevEnd}

	if !p.expectPunct('=') {
		return 0, false, false
	}

	valNode := p.parseOptionValue()
	if valNode == nil {
		return 0, false, false
	}
	valNode.NameSpan = nameSpan

	// Check if this option is 'default' or 'json_name'
	if !isExt && namePartsCount == 1 {
		optNameStr := partName.String(p.src)
		if optNameStr == "default" {
			*defaultVal = p.normalizeDefaultValue(valNode)
			if defaultValSpan != nil {
				*defaultValSpan = valNode.ValSpan
			}
			if defaultOpt != nil {
				*defaultOpt = TextSpan{Start: startTok.Start, End: p.prevEnd}
			}
			return 0, false, true
		}
		if optNameStr == "json_name" {
			*customJSON = valNode.ValSpan
			if customJSONOpt != nil {
				*customJSONOpt = TextSpan{Start: startTok.Start, End: valNode.Span.End}
			}
			if customJSONVal != nil {
				*customJSONVal = valNode.Span
			}
			return 0, false, true
		}
	}

	valNode.NameParts = Span32{Start: namePartsStart, Count: namePartsCount}
	valNode.Span = TextSpan{Start: startTok.Start, End: p.lastTokenEnd()}
	optIdx := uint32(len(p.file.Options))
	p.file.Options = append(p.file.Options, *valNode)
	return optIdx, true, true
}

func (p *Parser) parseGroup(parent *MessageNode, label FieldLabel, labelSpan TextSpan, startTok lexer.Token, groupKwTok lexer.Token) (fldIdx uint32, groupMsgIdx uint32, ok bool) {
	groupNameSpan, okIdent := p.expectIdent("group name")
	if !okIdent {
		p.recoverStatement()
		return 0, 0, false
	}

	if !p.expectPunct('=') {
		p.recoverStatement()
		return 0, 0, false
	}

	numTok := p.peek()
	if numTok.Kind != lexer.KindInt {
		p.errorAt(numTok, "expected field number for group")
		p.recoverStatement()
		return 0, 0, false
	}
	p.next()
	numVal, _ := strconv.ParseInt(string(p.src[numTok.Start:numTok.Start+numTok.Len]), 0, 32)
	numSpan := TextSpan{Start: numTok.Start, End: numTok.Start + numTok.Len}

	openBraceTok := p.peek()
	if !p.expectPunct('{') {
		p.recoverStatement()
		return 0, 0, false
	}
	openBraceSpan := TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}

	// Create synthetic nested message for group
	groupMsgIdx = uint32(len(p.file.Messages))
	p.file.Messages = append(p.file.Messages, MessageNode{
		Name:          groupNameSpan,
		OpenBraceSpan: openBraceSpan,
		IsGroup:       true,
	})

	fieldIndices := make([]uint32, 0, 8)
	nestedMsgIndices := make([]uint32, 0, 4)
	decls := make([]DeclRef, 0, 8)

	groupMsg := &p.file.Messages[groupMsgIdx]

	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct('}') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			continue
		}
		subFldIdx, subGroupIdx, isSubGroup, subOk := p.parseFieldOrGroup(groupMsg)
		if subOk {
			if isSubGroup {
				nestedMsgIndices = append(nestedMsgIndices, subGroupIdx)
			}
			fieldIndices = append(fieldIndices, subFldIdx)
			decls = append(decls, DeclRef{Kind: DeclField, Index: subFldIdx})
		}
	}
	if !closed {
		p.errorAt(p.peek(), "expected '}'")
	}

	groupEnd := p.lastTokenEnd()
	gm := &p.file.Messages[groupMsgIdx]
	gm.Span = TextSpan{Start: startTok.Start, End: groupEnd}
	gm.Fields = p.storeIndices(fieldIndices)
	gm.NestedMessages = p.storeIndices(nestedMsgIndices)
	gm.Decls = decls

	// Add group field to parent
	groupNameStr := groupNameSpan.String(p.src)
	fieldLowerName := p.internString(strings.ToLower(groupNameStr))

	groupKwSpan := TextSpan{Start: groupKwTok.Start, End: groupKwTok.Start + groupKwTok.Len}

	fldIdx = uint32(len(p.file.Fields))
	p.file.Fields = append(p.file.Fields, FieldNode{
		Name:             fieldLowerName,
		TypeName:         groupNameSpan,
		Number:           int32(numVal),
		Label:            label,
		TypeKind:         TypeGroup,
		OneofIndex:       -1,
		Span:             TextSpan{Start: startTok.Start, End: groupEnd},
		LabelSpan:        labelSpan,
		GroupKeywordSpan: groupKwSpan,
		NumberSpan:       numSpan,
	})
	return fldIdx, groupMsgIdx, true
}

func (p *Parser) parseMapField(parent *MessageNode, mapTok lexer.Token) (fldIdx uint32, entryMsgIdx uint32) {
	if !p.expectPunct('<') {
		p.recoverStatement()
		return 0, 0
	}

	keyTypeSpan, keySrcSpan := p.parseFullIdentWithSpan()
	if keyTypeSpan.IsEmpty() {
		p.errorAt(p.peek(), "expected map key type")
		p.recoverStatement()
		return 0, 0
	}

	if !p.expectPunct(',') {
		p.recoverStatement()
		return 0, 0
	}

	valTypeSpan, valSrcSpan := p.parseFullIdentWithSpan()
	if valTypeSpan.IsEmpty() {
		p.errorAt(p.peek(), "expected map value type")
		p.recoverStatement()
		return 0, 0
	}

	if !p.expectPunct('>') {
		p.recoverStatement()
		return 0, 0
	}

	mapTypeSpan := TextSpan{Start: mapTok.Start, End: p.lastTokenEnd()}

	nameSpan, ok := p.expectIdent("map field name")
	if !ok {
		p.recoverStatement()
		return 0, 0
	}

	if !p.expectPunct('=') {
		p.recoverStatement()
		return 0, 0
	}

	numTok := p.peek()
	if numTok.Kind != lexer.KindInt {
		p.errorAt(numTok, "expected field number for map")
		p.recoverStatement()
		return 0, 0
	}
	p.next()
	numVal, _ := strconv.ParseInt(string(p.src[numTok.Start:numTok.Start+numTok.Len]), 0, 32)
	numSpan := TextSpan{Start: numTok.Start, End: numTok.Start + numTok.Len}

	// Options
	var customJSON TextSpan
	var customJSONOpt TextSpan
	var customJSONVal TextSpan
	var defaultVal TextSpan
	var defaultValSpan TextSpan
	var optionsSpan TextSpan
	var defaultOpt TextSpan
	fldOptionIndices := p.parseFieldOptions(&customJSON, &customJSONOpt, &customJSONVal, &defaultVal, &defaultValSpan, &defaultOpt, &optionsSpan)
	p.expectPunct(';')

	mapEnd := p.lastTokenEnd()

	// Desugar into synthetic <PascalName>Entry nested message
	fieldNameStr := nameSpan.String(p.src)
	entryTypeNameStr := MapEntryTypeName(fieldNameStr)
	entryTypeNameSpan := p.internString(entryTypeNameStr)

	entryMsgIdx = uint32(len(p.file.Messages))
	p.file.Messages = append(p.file.Messages, MessageNode{
		Name:       entryTypeNameSpan,
		Span:       nameSpan,
		IsMapEntry: true,
	})

	// Add key field (tag 1) and value field (tag 2) to entry message
	keyNameSpan := p.internString("key")
	valNameSpan := p.internString("value")

	keyTypeKind := resolvePrimitiveType(keyTypeSpan.String(p.src))
	valTypeKind := resolvePrimitiveType(valTypeSpan.String(p.src))

	keyFldIdx := uint32(len(p.file.Fields))
	p.file.Fields = append(p.file.Fields, FieldNode{
		Name:       keyNameSpan,
		TypeName:   keyTypeSpan,
		TypeSpan:   keySrcSpan,
		Number:     1,
		Label:      LabelOptional,
		TypeKind:   keyTypeKind,
		OneofIndex: -1,
	})
	valFldIdx := uint32(len(p.file.Fields))
	p.file.Fields = append(p.file.Fields, FieldNode{
		Name:       valNameSpan,
		TypeName:   valTypeSpan,
		TypeSpan:   valSrcSpan,
		Number:     2,
		Label:      LabelOptional,
		TypeKind:   valTypeKind,
		OneofIndex: -1,
	})

	em := &p.file.Messages[entryMsgIdx]
	em.Fields = p.storeIndices([]uint32{keyFldIdx, valFldIdx})

	// Add map field to parent message
	fldIdx = uint32(len(p.file.Fields))
	p.file.Fields = append(p.file.Fields, FieldNode{
		Name:              nameSpan,
		TypeName:          entryTypeNameSpan,
		TypeSpan:          mapTypeSpan,
		Number:            int32(numVal),
		Label:             LabelRepeated,
		TypeKind:          TypeUnresolved,
		OneofIndex:        -1,
		Span:              TextSpan{Start: mapTok.Start, End: mapEnd},
		Options:           p.storeIndices(fldOptionIndices),
		CustomJSONName:    customJSON,
		DefaultVal:        defaultVal,
		NumberSpan:        numSpan,
		OptionsSpan:       optionsSpan,
		CustomJSONNameOpt: customJSONOpt,
		CustomJSONNameVal: customJSONVal,
		DefaultValSpan:    defaultValSpan,
		DefaultOpt:        defaultOpt,
	})
	return fldIdx, entryMsgIdx
}

func (p *Parser) parseOneof(parent *MessageNode, oneofIndex int32, oneofTok lexer.Token) (uint32, []uint32) {
	name, ok := p.expectIdent("oneof name")
	if !ok {
		p.recoverStatement()
		return 0, nil
	}

	openBraceTok := p.peek()
	if !p.expectPunct('{') {
		p.recoverStatement()
		return 0, nil
	}
	openBraceSpan := TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}

	ooIdx := uint32(len(p.file.Oneofs))
	p.file.Oneofs = append(p.file.Oneofs, OneofNode{
		Name:          name,
		OpenBraceSpan: openBraceSpan,
	})

	fieldIndices := make([]uint32, 0, 4)
	var oneofOptionIndices []uint32
	decls := make([]DeclRef, 0, 4)

	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct('}') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			continue
		}

		tok := p.peek()
		if p.matchIdent("option") {
			if optIdx, ok := p.parseOption(tok); ok {
				oneofOptionIndices = append(oneofOptionIndices, optIdx)
				decls = append(decls, DeclRef{Kind: DeclOption, Index: optIdx})
			}
			continue
		}

		typeSpan, typeSrcSpan := p.parseFullIdentWithSpan()
		if typeSpan.IsEmpty() {
			p.errorAt(p.peek(), "expected oneof field type")
			p.recoverStatement()
			continue
		}

		fldName, ok := p.expectIdent("oneof field name")
		if !ok {
			p.recoverStatement()
			continue
		}

		if !p.expectPunct('=') {
			p.recoverStatement()
			continue
		}

		numTok := p.peek()
		if numTok.Kind != lexer.KindInt {
			p.errorAt(numTok, "expected field number")
			p.recoverStatement()
			continue
		}
		p.next()
		numVal, _ := strconv.ParseInt(string(p.src[numTok.Start:numTok.Start+numTok.Len]), 0, 32)
		numSpan := TextSpan{Start: numTok.Start, End: numTok.Start + numTok.Len}

		var customJSON TextSpan
		var customJSONOpt TextSpan
		var customJSONVal TextSpan
		var defaultVal TextSpan
		var defaultValSpan TextSpan
		var optionsSpan TextSpan
		var defaultOpt TextSpan
		fldOptionIndices := p.parseFieldOptions(&customJSON, &customJSONOpt, &customJSONVal, &defaultVal, &defaultValSpan, &defaultOpt, &optionsSpan)
		p.expectPunct(';')

		typeKind := resolvePrimitiveType(typeSpan.String(p.src))

		fldIdx := uint32(len(p.file.Fields))
		p.file.Fields = append(p.file.Fields, FieldNode{
			Name:              fldName,
			TypeName:          typeSpan,
			TypeSpan:          typeSrcSpan,
			Number:            int32(numVal),
			Label:             LabelOptional,
			TypeKind:          typeKind,
			OneofIndex:        oneofIndex,
			Span:              TextSpan{Start: tok.Start, End: p.lastTokenEnd()},
			Options:           p.storeIndices(fldOptionIndices),
			CustomJSONName:    customJSON,
			DefaultVal:        defaultVal,
			NumberSpan:        numSpan,
			OptionsSpan:       optionsSpan,
			CustomJSONNameOpt: customJSONOpt,
			CustomJSONNameVal: customJSONVal,
			DefaultValSpan:    defaultValSpan,
			DefaultOpt:        defaultOpt,
		})
		fieldIndices = append(fieldIndices, fldIdx)
		decls = append(decls, DeclRef{Kind: DeclField, Index: fldIdx})
	}
	if !closed {
		p.errorAt(p.peek(), "expected '}'")
	}

	oo := &p.file.Oneofs[ooIdx]
	oo.Span = TextSpan{Start: oneofTok.Start, End: p.lastTokenEnd()}
	oo.Options = p.storeIndices(oneofOptionIndices)
	oo.Decls = decls

	return ooIdx, fieldIndices
}

func (p *Parser) parseExtensions(stmtStartTok lexer.Token, extRangeIndices *[]uint32) uint32 {
	stmtRangeIndices := make([]uint32, 0, 4)
	for !p.atEnd() {
		p.skipIgnored()
		startTok := p.peek()
		if startTok.Kind != lexer.KindInt {
			p.errorAt(startTok, "expected extension range start number")
			p.recoverStatement()
			break
		}
		p.next()
		startVal, _ := strconv.ParseInt(string(p.src[startTok.Start:startTok.Start+startTok.Len]), 0, 32)
		endVal := startVal + 1 // Default single number range is exclusive end = start + 1
		startSpan := TextSpan{Start: startTok.Start, End: startTok.Start + startTok.Len}
		var endSpan TextSpan

		if p.matchIdent("to") {
			endTok := p.peek()
			if p.matchIdent("max") {
				endVal = 536870912 // 0x20000000 (max tag 536870911 + 1)
				endSpan = TextSpan{Start: endTok.Start, End: endTok.Start + endTok.Len}
			} else if endTok.Kind == lexer.KindInt {
				p.next()
				ev, _ := strconv.ParseInt(string(p.src[endTok.Start:endTok.Start+endTok.Len]), 0, 32)
				endVal = ev + 1 // protoc convention is end + 1
				endSpan = TextSpan{Start: endTok.Start, End: endTok.Start + endTok.Len}
			} else {
				p.errorAt(endTok, "expected extension range end number or 'max'")
				p.recoverStatement()
				break
			}
		}

		var dummyJSON, dummyJSONOpt, dummyJSONVal, dummyDef, dummyDefSpan TextSpan
		var optSpan TextSpan
		optIndices := p.parseFieldOptions(&dummyJSON, &dummyJSONOpt, &dummyJSONVal, &dummyDef, &dummyDefSpan, nil, &optSpan)

		erEnd := startSpan.End
		if !endSpan.IsEmpty() {
			erEnd = endSpan.End
		}
		erIdx := uint32(len(p.file.ExtensionRanges))
		p.file.ExtensionRanges = append(p.file.ExtensionRanges, ExtensionRangeNode{
			Start:       int32(startVal),
			End:         int32(endVal),
			StartSpan:   startSpan,
			EndSpan:     endSpan,
			OptionsSpan: optSpan,
			Options:     p.storeIndices(optIndices),
			Span:        TextSpan{Start: startTok.Start, End: erEnd},
		})
		*extRangeIndices = append(*extRangeIndices, erIdx)
		stmtRangeIndices = append(stmtRangeIndices, erIdx)

		if p.matchPunct(',') {
			continue
		}
		break
	}
	p.expectPunct(';')

	declIdx := uint32(len(p.file.ExtensionRangeDecls))
	p.file.ExtensionRangeDecls = append(p.file.ExtensionRangeDecls, ExtensionRangeDeclNode{
		Span:         TextSpan{Start: stmtStartTok.Start, End: p.lastTokenEnd()},
		RangeIndices: p.storeIndices(stmtRangeIndices),
	})
	return declIdx
}

func (p *Parser) parseReserved(stmtStartTok lexer.Token, isEnum bool, resRanges *[]uint32, resNames *[]uint32) uint32 {
	isNames := false
	stmtItemIndices := make([]uint32, 0, 4)
	for !p.atEnd() {
		p.skipIgnored()
		tok := p.peek()

		// String name
		if tok.Kind == lexer.KindString {
			isNames = true
			unquoted, end := p.consumeStringLit()
			rawSpan := TextSpan{Start: tok.Start, End: end}
			nameSpan := p.internBytes(unquoted)
			nameIdx := uint32(len(p.file.ReservedNames))
			p.file.ReservedNames = append(p.file.ReservedNames, nameSpan)
			p.file.ReservedNameSpans = append(p.file.ReservedNameSpans, rawSpan)
			*resNames = append(*resNames, nameIdx)
			stmtItemIndices = append(stmtItemIndices, nameIdx)
		} else if tok.Kind == lexer.KindInt {
			// Number or range
			p.next()
			startSpan := TextSpan{Start: tok.Start, End: tok.Start + tok.Len}
			startVal, _ := strconv.ParseInt(string(p.src[tok.Start:tok.Start+tok.Len]), 0, 32)
			var endVal int64
			if isEnum {
				endVal = startVal // Inclusive for enums
			} else {
				endVal = startVal + 1 // Exclusive for messages
			}
			var endSpan TextSpan

			if p.matchIdent("to") {
				endTok := p.peek()
				if p.matchIdent("max") {
					endSpan = TextSpan{Start: endTok.Start, End: endTok.Start + endTok.Len}
					if isEnum {
						endVal = 536870911
					} else {
						endVal = 536870912
					}
				} else if endTok.Kind == lexer.KindInt {
					p.next()
					endSpan = TextSpan{Start: endTok.Start, End: endTok.Start + endTok.Len}
					ev, _ := strconv.ParseInt(string(p.src[endTok.Start:endTok.Start+endTok.Len]), 0, 32)
					if isEnum {
						endVal = ev
					} else {
						endVal = ev + 1
					}
				} else {
					p.errorAt(endTok, "expected reserved range end number")
					p.recoverStatement()
					break
				}
			}

			rrIdx := uint32(len(p.file.ReservedRanges))
			p.file.ReservedRanges = append(p.file.ReservedRanges, ReservedRangeNode{
				Start:     int32(startVal),
				End:       int32(endVal),
				StartSpan: startSpan,
				EndSpan:   endSpan,
				Span:      TextSpan{Start: tok.Start, End: p.lastTokenEnd()},
			})
			*resRanges = append(*resRanges, rrIdx)
			stmtItemIndices = append(stmtItemIndices, rrIdx)
		} else if tok.Kind == lexer.KindIdent {
			// Identifier name (in editions, e.g. reserved foo, bar;)
			isNames = true
			p.next()
			rawSpan := TextSpan{Start: tok.Start, End: tok.Start + tok.Len}
			nameIdx := uint32(len(p.file.ReservedNames))
			p.file.ReservedNames = append(p.file.ReservedNames, rawSpan)
			p.file.ReservedNameSpans = append(p.file.ReservedNameSpans, rawSpan)
			*resNames = append(*resNames, nameIdx)
			stmtItemIndices = append(stmtItemIndices, nameIdx)
		} else {
			p.errorAt(tok, "unexpected token in reserved statement")
			p.recoverStatement()
			break
		}

		if p.matchPunct(',') {
			continue
		}
		break
	}
	p.expectPunct(';')

	declIdx := uint32(len(p.file.ReservedDecls))
	p.file.ReservedDecls = append(p.file.ReservedDecls, ReservedDeclNode{
		Span:        TextSpan{Start: stmtStartTok.Start, End: p.lastTokenEnd()},
		IsNames:     isNames,
		ItemIndices: p.storeIndices(stmtItemIndices),
	})
	return declIdx
}

func (p *Parser) parseEnum(parent *MessageNode, enumTok lexer.Token) (uint32, bool) {
	name, ok := p.expectIdent("enum name")
	if !ok {
		p.recoverStatement()
		return 0, false
	}

	openBraceTok := p.peek()
	if !p.expectPunct('{') {
		p.recoverStatement()
		return 0, true
	}
	openBraceSpan := TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}

	enumIdx := uint32(len(p.file.Enums))
	p.file.Enums = append(p.file.Enums, EnumNode{
		Name:          name,
		OpenBraceSpan: openBraceSpan,
	})

	valIndices := make([]uint32, 0, 16)
	resRangeIndices := make([]uint32, 0, 4)
	resNameIndices := make([]uint32, 0, 4)
	optionIndices := make([]uint32, 0, 4)
	decls := make([]DeclRef, 0, 16)

	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct('}') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			continue
		}

		tok := p.peek()
		if tok.Kind != lexer.KindIdent {
			p.errorAt(tok, "unexpected token in enum body")
			p.recoverStatement()
			continue
		}

		kw := string(p.src[tok.Start : tok.Start+tok.Len])
		if kw == "option" {
			p.next()
			if optIdx, ok := p.parseOption(tok); ok {
				optionIndices = append(optionIndices, optIdx)
				decls = append(decls, DeclRef{Kind: DeclOption, Index: optIdx})
			}
			continue
		}
		if kw == "reserved" {
			p.next()
			resDeclIdx := p.parseReserved(tok, true, &resRangeIndices, &resNameIndices)
			decls = append(decls, DeclRef{Kind: DeclReserved, Index: resDeclIdx})
			continue
		}

		// Enum value: NAME = number [ [options] ];
		valTok := tok
		valName, _ := p.expectIdent("enum value name")
		if !p.expectPunct('=') {
			p.recoverStatement()
			continue
		}

		numStartPos := p.peek().Start
		isNeg := false
		if p.matchPunct('-') {
			isNeg = true
		}

		numTok := p.peek()
		if numTok.Kind != lexer.KindInt {
			p.errorAt(numTok, "expected enum value number")
			p.recoverStatement()
			continue
		}
		p.next()
		numVal, _ := strconv.ParseInt(string(p.src[numTok.Start:numTok.Start+numTok.Len]), 0, 64)
		if isNeg {
			numVal = -numVal
		}
		numSpan := TextSpan{Start: numStartPos, End: numTok.Start + numTok.Len}

		var dummyJSON, dummyJSONOpt, dummyJSONVal, dummyDef, dummyDefSpan TextSpan
		var evOptionsSpan TextSpan
		evOptIndices := p.parseFieldOptions(&dummyJSON, &dummyJSONOpt, &dummyJSONVal, &dummyDef, &dummyDefSpan, nil, &evOptionsSpan)
		p.expectPunct(';')

		vIdx := uint32(len(p.file.EnumValues))
		p.file.EnumValues = append(p.file.EnumValues, EnumValueNode{
			Name:        valName,
			Number:      int32(numVal),
			NumberSpan:  numSpan,
			Span:        TextSpan{Start: valTok.Start, End: p.lastTokenEnd()},
			OptionsSpan: evOptionsSpan,
			Options:     p.storeIndices(evOptIndices),
		})
		valIndices = append(valIndices, vIdx)
		decls = append(decls, DeclRef{Kind: DeclEnumValue, Index: vIdx})
	}
	if !closed {
		p.errorAt(p.peek(), "expected '}'")
	}

	en := &p.file.Enums[enumIdx]
	en.Span = TextSpan{Start: enumTok.Start, End: p.lastTokenEnd()}
	en.Values = p.storeIndices(valIndices)
	en.ReservedRanges = p.storeIndices(resRangeIndices)
	en.ReservedNames = p.storeIndices(resNameIndices)
	en.Options = p.storeIndices(optionIndices)
	en.Decls = decls

	return enumIdx, true
}

func (p *Parser) parseService(svcTok lexer.Token) (uint32, bool) {
	name, ok := p.expectIdent("service name")
	if !ok {
		p.recoverStatement()
		return 0, false
	}

	openBraceTok := p.peek()
	if !p.expectPunct('{') {
		p.recoverStatement()
		return 0, true
	}
	openBraceSpan := TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}

	svcIdx := uint32(len(p.file.Services))
	p.file.Services = append(p.file.Services, ServiceNode{
		Name:          name,
		OpenBraceSpan: openBraceSpan,
	})

	methodIndices := make([]uint32, 0, 8)
	optionIndices := make([]uint32, 0, 4)
	decls := make([]DeclRef, 0, 8)

	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct('}') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			continue
		}

		tok := p.peek()
		if tok.Kind != lexer.KindIdent {
			p.errorAt(tok, "unexpected token in service body")
			p.recoverStatement()
			continue
		}

		kw := string(p.src[tok.Start : tok.Start+tok.Len])
		if kw == "option" {
			p.next()
			if optIdx, ok := p.parseOption(tok); ok {
				optionIndices = append(optionIndices, optIdx)
				decls = append(decls, DeclRef{Kind: DeclOption, Index: optIdx})
			}
			continue
		}
		if kw == "rpc" {
			p.next()
			mIdx := p.parseMethod(tok)
			methodIndices = append(methodIndices, mIdx)
			decls = append(decls, DeclRef{Kind: DeclMethod, Index: mIdx})
			continue
		}

		p.errorAt(tok, fmt.Sprintf("unknown declaration '%s' in service", kw))
		p.recoverStatement()
	}
	if !closed {
		p.errorAt(p.peek(), "expected '}'")
	}

	svc := &p.file.Services[svcIdx]
	svc.Span = TextSpan{Start: svcTok.Start, End: p.lastTokenEnd()}
	svc.Methods = p.storeIndices(methodIndices)
	svc.Options = p.storeIndices(optionIndices)
	svc.Decls = decls
	return svcIdx, true
}

func (p *Parser) parseMethod(rpcTok lexer.Token) uint32 {
	name, ok := p.expectIdent("rpc method name")
	if !ok {
		p.recoverStatement()
		return 0
	}

	if !p.expectPunct('(') {
		p.recoverStatement()
		return 0
	}
	var clientStreaming bool
	var clientStreamingSpan TextSpan
	clientTok := p.peek()
	if p.matchIdent("stream") {
		clientStreaming = true
		clientStreamingSpan = TextSpan{Start: clientTok.Start, End: clientTok.Start + clientTok.Len}
	}
	inputType, inputSrcSpan := p.parseFullIdentWithSpan()
	if !p.expectPunct(')') {
		p.recoverStatement()
		return 0
	}

	if !p.matchIdent("returns") {
		p.errorAt(p.peek(), "expected 'returns'")
		p.recoverStatement()
		return 0
	}

	if !p.expectPunct('(') {
		p.recoverStatement()
		return 0
	}
	var serverStreaming bool
	var serverStreamingSpan TextSpan
	serverTok := p.peek()
	if p.matchIdent("stream") {
		serverStreaming = true
		serverStreamingSpan = TextSpan{Start: serverTok.Start, End: serverTok.Start + serverTok.Len}
	}
	outputType, outputSrcSpan := p.parseFullIdentWithSpan()
	if !p.expectPunct(')') {
		p.recoverStatement()
		return 0
	}

	optionIndices := make([]uint32, 0, 4)
	hasBody := false
	var openBraceSpan TextSpan
	openBraceTok := p.peek()
	if p.matchPunct('{') {
		hasBody = true
		openBraceSpan = TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}
		closed := false
		for !p.atEnd() {
			p.skipIgnored()
			if p.matchPunct('}') {
				closed = true
				break
			}
			if p.matchPunct(';') {
				continue
			}
			optTok := p.peek()
			if p.matchIdent("option") {
				if optIdx, ok := p.parseOption(optTok); ok {
					optionIndices = append(optionIndices, optIdx)
				}
				continue
			}
			p.errorAt(p.peek(), "unexpected token in rpc body")
			p.recoverStatement()
		}
		if !closed {
			p.errorAt(p.peek(), "expected '}'")
		}
	} else {
		p.expectPunct(';')
	}

	mIdx := uint32(len(p.file.Methods))
	p.file.Methods = append(p.file.Methods, MethodNode{
		Name:                name,
		InputType:           inputType,
		OutputType:          outputType,
		InputTypeSpan:       inputSrcSpan,
		OutputTypeSpan:      outputSrcSpan,
		ClientStreaming:     clientStreaming,
		ServerStreaming:     serverStreaming,
		ClientStreamingSpan: clientStreamingSpan,
		ServerStreamingSpan: serverStreamingSpan,
		Span:                TextSpan{Start: rpcTok.Start, End: p.lastTokenEnd()},
		OpenBraceSpan:       openBraceSpan,
		Options:             p.storeIndices(optionIndices),
		HasBody:             hasBody,
	})
	return mIdx
}

func (p *Parser) parseExtend(parent *MessageNode, extendTok lexer.Token) (uint32, []uint32) {
	extendeeSpan, extendeeSrcSpan := p.parseFullIdentWithSpan()
	if extendeeSpan.IsEmpty() {
		p.errorAt(p.peek(), "expected extendee type name")
		p.recoverStatement()
		return 0, nil
	}

	openBraceTok := p.peek()
	if !p.expectPunct('{') {
		p.recoverStatement()
		return 0, nil
	}
	openBraceSpan := TextSpan{Start: openBraceTok.Start, End: openBraceTok.Start + openBraceTok.Len}

	extIndices := make([]uint32, 0, 8)

	closed := false
	for !p.atEnd() {
		p.skipIgnored()
		if p.matchPunct('}') {
			closed = true
			break
		}
		if p.matchPunct(';') {
			continue
		}

		// Field in extend
		fldStart := p.peek()
		label := LabelOptional
		var labelSpan TextSpan
		isProto3Opt := false
		if p.matchIdent("optional") {
			label = LabelOptional
			labelSpan = TextSpan{Start: fldStart.Start, End: fldStart.Start + fldStart.Len}
			if p.file.Syntax == SyntaxProto3 {
				isProto3Opt = true
			}
		} else if p.matchIdent("required") {
			label = LabelRequired
			labelSpan = TextSpan{Start: fldStart.Start, End: fldStart.Start + fldStart.Len}
		} else if p.matchIdent("repeated") {
			label = LabelRepeated
			labelSpan = TextSpan{Start: fldStart.Start, End: fldStart.Start + fldStart.Len}
		}

		typeSpan, typeSrcSpan := p.parseFullIdentWithSpan()
		fldName, ok := p.expectIdent("extension field name")
		if !ok {
			p.recoverStatement()
			continue
		}

		if !p.expectPunct('=') {
			p.recoverStatement()
			continue
		}

		numTok := p.peek()
		numVal, _ := strconv.ParseInt(string(p.src[numTok.Start:numTok.Start+numTok.Len]), 0, 32)
		numSpan := TextSpan{Start: numTok.Start, End: numTok.Start + numTok.Len}
		p.next()

		var customJSON TextSpan
		var customJSONOpt TextSpan
		var customJSONVal TextSpan
		var defVal TextSpan
		var defValSpan TextSpan
		var optionsSpan TextSpan
		var defaultOpt TextSpan
		fldOptionIndices := p.parseFieldOptions(&customJSON, &customJSONOpt, &customJSONVal, &defVal, &defValSpan, &defaultOpt, &optionsSpan)
		p.expectPunct(';')

		typeKind := resolvePrimitiveType(typeSpan.String(p.src))

		fldIdx := uint32(len(p.file.Extensions))
		p.file.Extensions = append(p.file.Extensions, FieldNode{
			Name:              fldName,
			TypeName:          typeSpan,
			TypeSpan:          typeSrcSpan,
			Number:            int32(numVal),
			Label:             label,
			TypeKind:          typeKind,
			OneofIndex:        -1,
			Proto3Optional:    isProto3Opt,
			Span:              TextSpan{Start: fldStart.Start, End: p.lastTokenEnd()},
			Options:           p.storeIndices(fldOptionIndices),
			CustomJSONName:    customJSON,
			DefaultVal:        defVal,
			Extendee:          extendeeSpan,
			ExtendeeSpan:      extendeeSrcSpan,
			LabelSpan:         labelSpan,
			NumberSpan:        numSpan,
			OptionsSpan:       optionsSpan,
			CustomJSONNameOpt: customJSONOpt,
			CustomJSONNameVal: customJSONVal,
			DefaultValSpan:    defValSpan,
			DefaultOpt:        defaultOpt,
		})
		extIndices = append(extIndices, fldIdx)
	}
	if !closed {
		p.errorAt(p.peek(), "expected '}'")
	}

	extBlockSpan := TextSpan{Start: extendTok.Start, End: p.lastTokenEnd()}
	extBlockIdx := uint32(len(p.file.Extends))
	p.file.Extends = append(p.file.Extends, ExtendNode{
		Extendee:      extendeeSpan,
		ExtendeeSpan:  extendeeSrcSpan,
		Span:          extBlockSpan,
		OpenBraceSpan: openBraceSpan,
		FieldIndices:  p.storeIndices(extIndices),
	})

	return extBlockIdx, extIndices
}

// unescapeString strips quotes and resolves escape sequences.
func (p *Parser) unescapeString(start, length uint32) []byte {
	raw := p.src[start : start+length]
	if len(raw) < 2 {
		return raw
	}
	quote := raw[0]
	if (quote == '"' || quote == '\'') && raw[len(raw)-1] == quote {
		raw = raw[1 : len(raw)-1]
	}

	var buf bytes.Buffer
	buf.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		if b == '\\' && i+1 < len(raw) {
			i++
			switch raw[i] {
			case 'a':
				buf.WriteByte(0x07)
			case 'b':
				buf.WriteByte(0x08)
			case 'f':
				buf.WriteByte(0x0C)
			case 'n':
				buf.WriteByte(0x0A)
			case 'r':
				buf.WriteByte(0x0D)
			case 't':
				buf.WriteByte(0x09)
			case 'v':
				buf.WriteByte(0x0B)
			case '\\':
				buf.WriteByte('\\')
			case '\'':
				buf.WriteByte('\'')
			case '"':
				buf.WriteByte('"')
			case '?':
				buf.WriteByte('?')
			case 'x', 'X':
				var hexVal byte
				for h := 0; h < 2 && i+1 < len(raw) && isHex(raw[i+1]); h++ {
					i++
					hexVal = (hexVal << 4) | hexDigit(raw[i])
				}
				buf.WriteByte(hexVal)
			case '0', '1', '2', '3', '4', '5', '6', '7':
				octVal := raw[i] - '0'
				for o := 0; o < 2 && i+1 < len(raw) && raw[i+1] >= '0' && raw[i+1] <= '7'; o++ {
					i++
					octVal = (octVal << 3) | (raw[i] - '0')
				}
				buf.WriteByte(octVal)
			case 'u':
				var r rune
				for u := 0; u < 4 && i+1 < len(raw) && isHex(raw[i+1]); u++ {
					i++
					r = (r << 4) | rune(hexDigit(raw[i]))
				}
				var rBuf [utf8.UTFMax]byte
				n := utf8.EncodeRune(rBuf[:], r)
				buf.Write(rBuf[:n])
			case 'U':
				var r rune
				for u := 0; u < 8 && i+1 < len(raw) && isHex(raw[i+1]); u++ {
					i++
					r = (r << 4) | rune(hexDigit(raw[i]))
				}
				var rBuf [utf8.UTFMax]byte
				n := utf8.EncodeRune(rBuf[:], r)
				buf.Write(rBuf[:n])
			default:
				buf.WriteByte(raw[i])
			}
		} else {
			buf.WriteByte(b)
		}
	}
	return buf.Bytes()
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func hexDigit(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	}
	return 0
}

func resolvePrimitiveType(typeName string) FieldTypeKind {
	switch typeName {
	case "double":
		return TypeDouble
	case "float":
		return TypeFloat
	case "int64":
		return TypeInt64
	case "uint64":
		return TypeUint64
	case "int32":
		return TypeInt32
	case "fixed64":
		return TypeFixed64
	case "fixed32":
		return TypeFixed32
	case "bool":
		return TypeBool
	case "string":
		return TypeString
	case "bytes":
		return TypeBytes
	case "uint32":
		return TypeUint32
	case "sfixed32":
		return TypeSfixed32
	case "sfixed64":
		return TypeSfixed64
	case "sint32":
		return TypeSint32
	case "sint64":
		return TypeSint64
	default:
		return TypeUnresolved
	}
}

// postProcess performs type resolution, map entry desugaring, and proto3 optional oneof creation.
func (p *Parser) postProcess() {
	// 1. Process proto3 optional synthetic oneofs
	if p.file.Syntax == SyntaxProto3 {
		for i := range p.file.Messages {
			msg := &p.file.Messages[i]
			p.processProto3Optional(msg)
		}
	}

	// 2. Resolve symbol types within the file
	p.resolveSymbols()
}

func (p *Parser) processProto3Optional(msg *MessageNode) {
	usedNames := make(map[string]struct{})
	for i := msg.Fields.Start; i < msg.Fields.End(); i++ {
		fi := p.file.ChildIndices[i]
		if int(fi) < len(p.file.Fields) {
			usedNames[p.file.Fields[fi].Name.String(p.src)] = struct{}{}
		}
	}
	oneofs := make([]uint32, 0, msg.Oneofs.Count+2)
	for i := msg.Oneofs.Start; i < msg.Oneofs.End(); i++ {
		oi := p.file.ChildIndices[i]
		if int(oi) < len(p.file.Oneofs) {
			usedNames[p.file.Oneofs[oi].Name.String(p.src)] = struct{}{}
			oneofs = append(oneofs, oi)
		}
	}

	for i := msg.Fields.Start; i < msg.Fields.End(); i++ {
		fi := p.file.ChildIndices[i]
		if int(fi) >= len(p.file.Fields) {
			continue
		}
		fld := &p.file.Fields[fi]
		if !fld.Proto3Optional {
			continue
		}

		fName := fld.Name.String(p.src)
		ooName := "_" + fName
		for {
			if _, exists := usedNames[ooName]; !exists {
				usedNames[ooName] = struct{}{}
				break
			}
			ooName = "X" + ooName
		}

		ooSpan := p.internString(ooName)
		ooIdx := uint32(len(p.file.Oneofs))
		p.file.Oneofs = append(p.file.Oneofs, OneofNode{
			Name:        ooSpan,
			Span:        fld.Span,
			IsSynthetic: true,
		})
		fld.OneofIndex = int32(len(oneofs))
		oneofs = append(oneofs, ooIdx)
	}

	msg.Oneofs = p.storeIndices(oneofs)
}

type symbolEntry struct {
	fqn    string
	isEnum bool
}

func (p *Parser) resolveSymbols() {
	pkg := p.file.Package.String(p.src)
	var pkgPrefix string
	if pkg != "" {
		pkgPrefix = "." + pkg + "."
	} else {
		pkgPrefix = "."
	}

	// Symbol map: FQN -> symbolEntry
	symbols := make(map[string]symbolEntry)

	// Index top-level enums
	for i := p.file.TopLevelEnums.Start; i < p.file.TopLevelEnums.End(); i++ {
		ei := p.file.ChildIndices[i]
		en := &p.file.Enums[ei]
		eName := en.Name.String(p.src)
		fqn := pkgPrefix + eName
		symbols[fqn] = symbolEntry{fqn: fqn, isEnum: true}
	}

	// Index messages and nested types recursively
	var indexMessage func(m *MessageNode, prefix string)
	indexMessage = func(m *MessageNode, prefix string) {
		mName := m.Name.String(p.src)
		mFqn := prefix + mName
		symbols[mFqn] = symbolEntry{fqn: mFqn, isEnum: false}

		subPrefix := mFqn + "."
		for i := m.NestedEnums.Start; i < m.NestedEnums.End(); i++ {
			ei := p.file.ChildIndices[i]
			en := &p.file.Enums[ei]
			eName := en.Name.String(p.src)
			symbols[subPrefix+eName] = symbolEntry{fqn: subPrefix + eName, isEnum: true}
		}
		for i := m.NestedMessages.Start; i < m.NestedMessages.End(); i++ {
			mi := p.file.ChildIndices[i]
			sub := &p.file.Messages[mi]
			indexMessage(sub, subPrefix)
		}
	}

	for i := p.file.TopLevelMessages.Start; i < p.file.TopLevelMessages.End(); i++ {
		mi := p.file.ChildIndices[i]
		indexMessage(&p.file.Messages[mi], pkgPrefix)
	}

	// Now resolve fields in each message
	var resolveMessageFields func(m *MessageNode, enclosingPrefix string)
	resolveMessageFields = func(m *MessageNode, enclosingPrefix string) {
		mFqn := enclosingPrefix + m.Name.String(p.src)
		for i := m.Fields.Start; i < m.Fields.End(); i++ {
			fi := p.file.ChildIndices[i]
			fld := &p.file.Fields[fi]
			if fld.TypeKind == TypeUnresolved {
				p.resolveFieldType(fld, mFqn, pkgPrefix, symbols)
			}
		}
		for i := m.NestedMessages.Start; i < m.NestedMessages.End(); i++ {
			mi := p.file.ChildIndices[i]
			resolveMessageFields(&p.file.Messages[mi], mFqn+".")
		}
	}

	for i := p.file.TopLevelMessages.Start; i < p.file.TopLevelMessages.End(); i++ {
		mi := p.file.ChildIndices[i]
		resolveMessageFields(&p.file.Messages[mi], pkgPrefix)
	}

	// Resolve extensions
	for i := range p.file.Extensions {
		fld := &p.file.Extensions[i]
		if fld.TypeKind == TypeUnresolved {
			p.resolveFieldType(fld, pkgPrefix, pkgPrefix, symbols)
		}
		// Resolve extendee name to FQN if in this file
		extName := fld.Extendee.String(p.src)
		if !strings.HasPrefix(extName, ".") {
			cand := pkgPrefix + extName
			if _, ok := symbols[cand]; ok {
				fld.Extendee = p.internString(cand)
			}
		}
	}

	// 4. Resolve service method input/output types
	for i := range p.file.Methods {
		m := &p.file.Methods[i]
		m.InputType = p.resolveTypeName(m.InputType, pkgPrefix, symbols)
		m.OutputType = p.resolveTypeName(m.OutputType, pkgPrefix, symbols)
	}
}

func (p *Parser) resolveTypeName(typeSpan TextSpan, pkgPrefix string, symbols map[string]symbolEntry) TextSpan {
	name := typeSpan.String(p.src)
	if strings.HasPrefix(name, ".") {
		return typeSpan
	}
	candidate := pkgPrefix + name
	if _, ok := symbols[candidate]; ok {
		return p.internString(candidate)
	}
	candidate = "." + name
	if _, ok := symbols[candidate]; ok {
		return p.internString(candidate)
	}
	return typeSpan
}

func (p *Parser) resolveFieldType(fld *FieldNode, currentScope, pkgPrefix string, symbols map[string]symbolEntry) {
	typeName := fld.TypeName.String(p.src)
	if strings.HasPrefix(typeName, ".") {
		if sym, ok := symbols[typeName]; ok {
			if sym.isEnum {
				fld.TypeKind = TypeEnum
			} else {
				fld.TypeKind = TypeMessage
			}
		} else {
			fld.TypeKind = TypeMessage
		}
		return
	}

	// Search scopes from innermost to outermost
	scope := currentScope
	for {
		candidate := scope + "." + typeName
		if sym, ok := symbols[candidate]; ok {
			if sym.isEnum {
				fld.TypeKind = TypeEnum
			} else {
				fld.TypeKind = TypeMessage
			}
			fld.TypeName = p.internString(candidate)
			return
		}
		lastDot := strings.LastIndex(scope, ".")
		if lastDot <= 0 {
			break
		}
		scope = scope[:lastDot]
	}

	// Scope 2: package scope
	candidate := pkgPrefix + typeName
	if sym, ok := symbols[candidate]; ok {
		if sym.isEnum {
			fld.TypeKind = TypeEnum
		} else {
			fld.TypeKind = TypeMessage
		}
		fld.TypeName = p.internString(candidate)
		return
	}

	// Scope 3: root scope
	candidate = "." + typeName
	if sym, ok := symbols[candidate]; ok {
		if sym.isEnum {
			fld.TypeKind = TypeEnum
		} else {
			fld.TypeKind = TypeMessage
		}
		fld.TypeName = p.internString(candidate)
		return
	}

	// Unresolved in this file: keep TypeUnresolved and raw TypeName so linker resolves across files
}
