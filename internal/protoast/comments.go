package protoast

// attachComments assigns every comment token to a declaration.
//
// A comment that starts on the line where a declaration (or a block's opening
// brace) ends is that declaration's trailing comment; only the first such
// comment counts. The other comments before a declaration's first token are
// its leading comments (the block directly above it, with no blank line in
// between) and detached comments (anything earlier). Comments followed by a
// token that does not start a declaration, such as a closing brace, the end of
// the file or a token inside a declaration, annotate nothing.
func (b *builder) attachComments(out *File) {
	out.owners = make(map[*Comment]*Decl)
	var pending []*Comment
	var prevEnd uint32 // end offset of the previous non-comment token
	prevLine := 0      // line of the previous non-comment token
	trailingTaken := false
	src := b.ast.Src
	for i := 0; i < b.syn.NumTokens(); i++ {
		_, start, end := b.syn.Token(i)
		if b.syn.IsComment(i) {
			c := &Comment{Raw: string(src[start:end]), Pos: b.pos(start), End: b.pos(end - 1)}
			out.Comments = append(out.Comments, c)
			if owner := b.byEnd[prevEnd]; owner != nil && !trailingTaken && len(pending) == 0 && c.Pos.Line == prevLine {
				owner.Trailing = append(owner.Trailing, c)
				out.owners[c] = owner
				trailingTaken = true
				continue
			}
			pending = append(pending, c)
			continue
		}
		if d := b.byStart[start]; d != nil && len(pending) > 0 {
			split := len(pending)
			nextLine := d.Pos.Line
			for split > 0 && nextLine-pending[split-1].End.Line <= 1 {
				split--
				nextLine = pending[split].Pos.Line
			}
			d.Detached = append(d.Detached, pending[:split]...)
			d.Leading = append(d.Leading, pending[split:]...)
			for _, c := range pending {
				out.owners[c] = d
			}
		}
		pending = nil
		prevEnd = end
		prevLine = b.pos(start).Line
		trailingTaken = false
		if b.byEnd[prevEnd] == nil && b.isSemicolon(start, end) {
			// "};" ends the block at its brace; keep the semicolon's comments with it.
			if j := b.prevSignificant(i); j >= 0 {
				_, _, braceEnd := b.syn.Token(j)
				if owner := b.byEnd[braceEnd]; owner != nil {
					b.byEnd[prevEnd] = owner
				}
			}
		}
	}
}

func (b *builder) isSemicolon(start, end uint32) bool {
	return end == start+1 && b.ast.Src[start] == ';'
}

func (b *builder) prevSignificant(i int) int {
	for j := i - 1; j >= 0; j-- {
		if !b.syn.IsComment(j) {
			return j
		}
	}
	return -1
}
