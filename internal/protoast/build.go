package protoast

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
	"github.com/easyp-tech/easyp/internal/easypc/protoc/syntax"
)

// Read parses r as the file at path.
func Read(path string, r io.Reader) (*File, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return Parse(path, src)
}

// Parse builds the lint view of src. It fails only on syntax errors; semantic
// diagnostics such as duplicate field numbers are left to the compiler.
func Parse(path string, src []byte) (*File, error) {
	f, err := syntax.Parse(path, src)
	if err != nil {
		return nil, err
	}
	return Build(f), nil
}

// Build converts a parsed file.
func Build(f *syntax.File) *File {
	b := &builder{
		syn:     f,
		ast:     f.AST,
		byStart: make(map[uint32]*Decl),
		byEnd:   make(map[uint32]*Decl),
	}
	out := &File{Path: f.AST.Path, Syntax: syntaxName(f.AST.Syntax)}
	b.file(out)
	b.attachComments(out)
	return out
}

func syntaxName(kind parser.SyntaxKind) string {
	switch kind {
	case parser.SyntaxProto3:
		return "proto3"
	case parser.SyntaxEditions:
		return "editions"
	default:
		return "proto2"
	}
}

type builder struct {
	syn *syntax.File
	ast *parser.ASTFile
	// byStart maps a declaration's first token offset to it; byEnd maps the end
	// offset of its last token, and of its opening brace for blocks.
	byStart map[uint32]*Decl
	byEnd   map[uint32]*Decl
}

func (b *builder) file(out *File) {
	a := b.ast
	for _, ref := range a.Decls {
		switch ref.Kind {
		case parser.DeclSyntax:
			out.SyntaxSpecified = true
			b.statement(a.SyntaxSpan)
		case parser.DeclPackage:
			pkg := &Package{Decl: b.decl(a.PackageSpan, parser.TextSpan{}), Name: b.tokens(a.Package)}
			pkg.NamePos = b.namePos(a.Package)
			b.register(&pkg.Decl, a.PackageSpan, parser.TextSpan{})
			out.Packages = append(out.Packages, pkg)
		case parser.DeclImport:
			out.Imports = append(out.Imports, b.importDecl(&a.Imports[ref.Index]))
		case parser.DeclOption:
			out.Options = append(out.Options, b.optionStmt(ref.Index))
		case parser.DeclMessage:
			out.Messages = append(out.Messages, b.message(&a.Messages[ref.Index]))
		case parser.DeclEnum:
			out.Enums = append(out.Enums, b.enum(&a.Enums[ref.Index]))
		case parser.DeclService:
			out.Services = append(out.Services, b.service(&a.Services[ref.Index]))
		case parser.DeclExtend:
			out.Extends = append(out.Extends, b.extend(&a.Extends[ref.Index]))
		}
	}
}

// decl returns the positions of a declaration spanning span.
func (b *builder) decl(span, _ parser.TextSpan) Decl {
	return Decl{Pos: b.pos(span.Start), End: b.pos(span.End - 1)}
}

// register makes d known to the comment pass. It runs after the declaration is
// placed in the tree so the map keeps a pointer to the final value.
func (b *builder) register(d *Decl, span, open parser.TextSpan) {
	if span.IsEmpty() || b.syn.IsSynthetic(span) {
		return
	}
	b.byStart[span.Start] = d
	b.byEnd[span.End] = d
	if !open.IsEmpty() {
		b.byEnd[open.End] = d
	}
}

func (b *builder) importDecl(n *parser.ImportNode) *Import {
	imp := &Import{
		Decl:   b.decl(n.Span, parser.TextSpan{}),
		Path:   b.syn.Text(n.Path),
		Public: n.Kind == parser.ImportPublic,
		Weak:   n.Kind == parser.ImportWeak,
	}
	b.eachToken(n.Span, func(kind syntax.TokenKind, text string) bool {
		if kind == syntax.KindString {
			imp.Location = text
			return false
		}
		return true
	})
	b.register(&imp.Decl, n.Span, parser.TextSpan{})
	return imp
}

func (b *builder) optionStmt(idx uint32) *Option {
	n := &b.ast.Options[idx]
	opt := &Option{
		Decl:  b.decl(n.Span, parser.TextSpan{}),
		Name:  b.tokens(n.NameSpan),
		Value: b.constant(n.ValRawSpan),
	}
	b.register(&opt.Decl, n.Span, parser.TextSpan{})
	return opt
}

// optionList converts the entries of a [...] list in source order, including
// json_name and default, which the parser keeps outside the option slab.
func (b *builder) optionList(indices parser.Span32, extra ...parser.TextSpan) []*Option {
	var opts []*Option
	for _, idx := range b.children(indices) {
		n := &b.ast.Options[idx]
		opts = append(opts, &Option{
			Decl:  Decl{Pos: b.pos(n.Span.Start), End: b.pos(n.Span.End - 1)},
			Name:  b.tokens(n.NameSpan),
			Value: b.constant(n.ValRawSpan),
		})
	}
	for _, span := range extra {
		if span.IsEmpty() {
			continue
		}
		name, value := b.splitOptionItem(span)
		opts = append(opts, &Option{
			Decl:  Decl{Pos: b.pos(span.Start), End: b.pos(span.End - 1)},
			Name:  name,
			Value: value,
		})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Pos.Offset < opts[j].Pos.Offset })
	return opts
}

func (b *builder) message(n *parser.MessageNode) *Message {
	msg := &Message{Decl: b.decl(n.Span, n.OpenBraceSpan), Name: b.syn.Text(n.Name)}
	msg.NamePos = b.namePos(n.Name)
	b.messageBody(n, msg)
	b.register(&msg.Decl, n.Span, n.OpenBraceSpan)
	return msg
}

func (b *builder) messageBody(n *parser.MessageNode, msg *Message) {
	a := b.ast
	for _, ref := range n.Decls {
		switch ref.Kind {
		case parser.DeclField:
			fn := &a.Fields[ref.Index]
			if fn.OneofIndex >= 0 && !fn.Proto3Optional {
				continue // listed under its oneof
			}
			fld := b.field(fn)
			switch {
			case fn.TypeKind == parser.TypeGroup:
				msg.Groups = append(msg.Groups, fld)
				if group := b.groupMessage(n, fn); group != nil {
					b.messageBody(group, &Message{})
				}
			case fld.KeyType != "":
				msg.Maps = append(msg.Maps, fld)
			default:
				msg.Fields = append(msg.Fields, fld)
			}
		case parser.DeclMessage:
			msg.Messages = append(msg.Messages, b.message(&a.Messages[ref.Index]))
		case parser.DeclEnum:
			msg.Enums = append(msg.Enums, b.enum(&a.Enums[ref.Index]))
		case parser.DeclExtend:
			msg.Extends = append(msg.Extends, b.extend(&a.Extends[ref.Index]))
		case parser.DeclOneof:
			msg.Oneofs = append(msg.Oneofs, b.oneof(&a.Oneofs[ref.Index]))
		case parser.DeclOption:
			msg.Options = append(msg.Options, b.optionStmt(ref.Index))
		case parser.DeclExtensionRange:
			b.statement(a.ExtensionRangeDecls[ref.Index].Span)
		case parser.DeclReserved:
			b.statement(a.ReservedDecls[ref.Index].Span)
		}
	}
}

// statement registers a declaration that has no node in the lint view, so
// comments above it still annotate it.
func (b *builder) statement(span parser.TextSpan) {
	d := b.decl(span, parser.TextSpan{})
	b.register(&d, span, parser.TextSpan{})
}

// groupMessage finds the synthetic message that holds a group's body.
func (b *builder) groupMessage(parent *parser.MessageNode, fn *parser.FieldNode) *parser.MessageNode {
	for _, idx := range b.children(parent.NestedMessages) {
		m := &b.ast.Messages[idx]
		if m.IsGroup && m.Span == fn.Span {
			return m
		}
	}
	return nil
}

func (b *builder) field(n *parser.FieldNode) *Field {
	fld := &Field{
		Decl:    b.decl(n.Span, parser.TextSpan{}),
		Label:   b.syn.Text(n.LabelSpan),
		Name:    b.syn.Text(n.Name),
		Number:  b.tokens(n.NumberSpan),
		Options: b.optionList(n.Options, n.CustomJSONNameOpt, n.DefaultOpt),
	}
	fld.NamePos = b.namePos(n.Name)
	switch {
	case n.TypeKind == parser.TypeGroup:
		fld.Type = b.tokens(n.TypeName)
		fld.Name = b.syn.Text(n.Name)
		fld.Group = true
		fld.NamePos = b.namePos(n.TypeName)
	default:
		fld.Type = b.tokens(n.TypeSpan)
		if key, value, ok := splitMapType(fld.Type); ok {
			fld.KeyType, fld.Type = key, value
		}
	}
	b.register(&fld.Decl, n.Span, parser.TextSpan{})
	return fld
}

// splitMapType splits "map<K,V>" (whitespace already removed).
func splitMapType(t string) (string, string, bool) {
	if !strings.HasPrefix(t, "map<") || !strings.HasSuffix(t, ">") {
		return "", "", false
	}
	key, value, ok := strings.Cut(t[len("map<"):len(t)-1], ",")
	return key, value, ok
}

func (b *builder) oneof(n *parser.OneofNode) *Oneof {
	o := &Oneof{Decl: b.decl(n.Span, n.OpenBraceSpan), Name: b.syn.Text(n.Name)}
	o.NamePos = b.namePos(n.Name)
	for _, ref := range n.Decls {
		switch ref.Kind {
		case parser.DeclField:
			o.Fields = append(o.Fields, b.field(&b.ast.Fields[ref.Index]))
		case parser.DeclOption:
			o.Options = append(o.Options, b.optionStmt(ref.Index))
		}
	}
	b.register(&o.Decl, n.Span, n.OpenBraceSpan)
	return o
}

func (b *builder) enum(n *parser.EnumNode) *Enum {
	e := &Enum{Decl: b.decl(n.Span, n.OpenBraceSpan), Name: b.syn.Text(n.Name)}
	e.NamePos = b.namePos(n.Name)
	for _, ref := range n.Decls {
		switch ref.Kind {
		case parser.DeclEnumValue:
			v := &b.ast.EnumValues[ref.Index]
			ev := &EnumValue{
				Decl:      b.decl(v.Span, parser.TextSpan{}),
				Name:      b.syn.Text(v.Name),
				Number:    b.tokens(v.NumberSpan),
				NumberPos: b.namePos(v.NumberSpan),
				Options:   b.optionList(v.Options),
			}
			ev.NamePos = b.namePos(v.Name)
			b.register(&ev.Decl, v.Span, parser.TextSpan{})
			e.Values = append(e.Values, ev)
		case parser.DeclOption:
			e.Options = append(e.Options, b.optionStmt(ref.Index))
		case parser.DeclReserved:
			b.statement(b.ast.ReservedDecls[ref.Index].Span)
		}
	}
	b.register(&e.Decl, n.Span, n.OpenBraceSpan)
	return e
}

func (b *builder) service(n *parser.ServiceNode) *Service {
	s := &Service{Decl: b.decl(n.Span, n.OpenBraceSpan), Name: b.syn.Text(n.Name)}
	s.NamePos = b.namePos(n.Name)
	for _, ref := range n.Decls {
		switch ref.Kind {
		case parser.DeclMethod:
			s.RPCs = append(s.RPCs, b.rpc(&b.ast.Methods[ref.Index]))
		case parser.DeclOption:
			s.Options = append(s.Options, b.optionStmt(ref.Index))
		}
	}
	b.register(&s.Decl, n.Span, n.OpenBraceSpan)
	return s
}

func (b *builder) rpc(n *parser.MethodNode) *RPC {
	r := &RPC{
		Decl:     b.decl(n.Span, n.OpenBraceSpan),
		Name:     b.syn.Text(n.Name),
		Request:  RPCType{Type: b.tokens(n.InputTypeSpan), Stream: n.ClientStreaming, Pos: b.namePos(n.InputTypeSpan)},
		Response: RPCType{Type: b.tokens(n.OutputTypeSpan), Stream: n.ServerStreaming, Pos: b.namePos(n.OutputTypeSpan)},
	}
	r.NamePos = b.namePos(n.Name)
	for _, idx := range b.children(n.Options) {
		r.Options = append(r.Options, b.optionStmt(idx))
	}
	b.register(&r.Decl, n.Span, n.OpenBraceSpan)
	return r
}

func (b *builder) extend(n *parser.ExtendNode) *Extend {
	e := &Extend{Decl: b.decl(n.Span, n.OpenBraceSpan), Type: b.tokens(n.ExtendeeSpan)}
	for _, idx := range b.children(n.FieldIndices) {
		e.Fields = append(e.Fields, b.field(&b.ast.Extensions[idx]))
	}
	b.register(&e.Decl, n.Span, n.OpenBraceSpan)
	return e
}

// children resolves a Span32 range of ChildIndices into slab indices.
func (b *builder) children(s parser.Span32) []uint32 {
	if s.IsEmpty() {
		return nil
	}
	return b.ast.ChildIndices[s.Start:s.End()]
}

// eachToken calls fn for every non-comment token inside span until fn returns false.
func (b *builder) eachToken(span parser.TextSpan, fn func(kind syntax.TokenKind, text string) bool) {
	if span.IsEmpty() || b.syn.IsSynthetic(span) {
		return
	}
	src := b.ast.Src
	for i := b.syn.TokenAt(span.Start); i < b.syn.NumTokens(); i++ {
		kind, start, end := b.syn.Token(i)
		if start >= span.End {
			return
		}
		if b.syn.IsComment(i) {
			continue
		}
		if !fn(kind, string(src[start:end])) {
			return
		}
	}
}

// tokens returns the text of span with whitespace and comments removed.
func (b *builder) tokens(span parser.TextSpan) string {
	if b.syn.IsSynthetic(span) {
		return b.syn.Text(span)
	}
	var sb strings.Builder
	b.eachToken(span, func(_ syntax.TokenKind, text string) bool {
		sb.WriteString(text)
		return true
	})
	return sb.String()
}

// constant returns an option value as written.
func (b *builder) constant(span parser.TextSpan) string {
	var lit literal
	b.eachToken(span, func(kind syntax.TokenKind, text string) bool {
		lit.add(kind, text)
		return true
	})
	return lit.String()
}

// splitOptionItem splits "name = value" written inside a [...] list.
func (b *builder) splitOptionItem(span parser.TextSpan) (string, string) {
	var name strings.Builder
	var lit literal
	afterEquals := false
	b.eachToken(span, func(kind syntax.TokenKind, text string) bool {
		switch {
		case afterEquals:
			lit.add(kind, text)
		case text == "=":
			afterEquals = true
		default:
			name.WriteString(text)
		}
		return true
	})
	return name.String(), lit.String()
}

// literal accumulates the tokens of an option value. Adjacent string literals
// are merged into one that keeps the first literal's quote character.
type literal struct {
	parts      []string
	allStrings bool
}

func (l *literal) add(kind syntax.TokenKind, text string) {
	if len(l.parts) == 0 {
		l.allStrings = true
	}
	l.parts = append(l.parts, text)
	l.allStrings = l.allStrings && kind == syntax.KindString
}

func (l *literal) String() string {
	if len(l.parts) < 2 || !l.allStrings {
		return strings.Join(l.parts, "")
	}
	q := l.parts[0][:1]
	var sb strings.Builder
	sb.WriteString(q)
	for _, p := range l.parts {
		sb.WriteString(strings.Trim(p, q))
	}
	sb.WriteString(q)
	return sb.String()
}

// pos converts a byte offset into a Position.
func (b *builder) pos(off uint32) Position {
	if off > b.ast.SrcLen {
		off = b.ast.SrcLen
	}
	line, column := b.syn.LineCol(off)
	return Position{Offset: int(off), Line: line, Column: column}
}

// namePos is the position of a declared name, or zero for synthetic names.
func (b *builder) namePos(span parser.TextSpan) Position {
	if span.IsEmpty() || b.syn.IsSynthetic(span) {
		return Position{}
	}
	return b.pos(span.Start)
}
