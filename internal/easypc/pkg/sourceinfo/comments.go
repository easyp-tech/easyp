package sourceinfo

import (
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
)

type commentBlock struct {
	startLine     int
	endLine       int
	startPos      uint32
	isBlock       bool
	hasCodeBefore bool
	text          string
}

type codeTokenInfo struct {
	startPos  uint32
	endPos    uint32
	startLine int
	endLine   int
	isClosing bool // '}', ']', ')', ',', ';'
	comments  []commentBlock
}

// CommentTable indexes leading, detached, and trailing comments by source byte offset.
type CommentTable struct {
	leadingComments         map[uint32]string
	leadingDetachedComments map[uint32][]string
	trailingComments        map[uint32]string
	usedLeading             map[uint32]bool
	usedTrailing            map[uint32]bool
}

// BuildCommentTable processes the token stream and attaches comments using protoc donation rules.
func BuildCommentTable(ts *lexer.TokenStream, src []byte) *CommentTable {
	table := &CommentTable{
		leadingComments:         make(map[uint32]string),
		leadingDetachedComments: make(map[uint32][]string),
		trailingComments:        make(map[uint32]string),
		usedLeading:             make(map[uint32]bool),
		usedTrailing:            make(map[uint32]bool),
	}
	if ts == nil || ts.Len() == 0 {
		return table
	}

	var pendingBlocks []commentBlock
	var codeTokens []codeTokenInfo
	lastCodeTokenEndLine := -1

	for i := 0; i < ts.Len(); i++ {
		kind := ts.Kinds[i]
		start := ts.Starts[i]
		length := ts.Lens[i]

		if kind == lexer.KindLineComment || kind == lexer.KindBlockComment {
			isBlock := (kind == lexer.KindBlockComment)
			text := cleanCommentText(src, start, length, kind)
			sLine, _ := ts.LineCol(src, start)
			eLine, _ := ts.LineCol(src, start+length)
			hasCodeBefore := (lastCodeTokenEndLine == sLine)

			// Group contiguous line comments on consecutive lines only if neither has code before
			if !isBlock && !hasCodeBefore && len(pendingBlocks) > 0 {
				lastIdx := len(pendingBlocks) - 1
				prev := &pendingBlocks[lastIdx]
				if !prev.isBlock && !prev.hasCodeBefore && sLine <= prev.endLine+1 {
					prev.text += text
					prev.endLine = eLine
					continue
				}
			}

			pendingBlocks = append(pendingBlocks, commentBlock{
				startLine:     sLine,
				endLine:       eLine,
				startPos:      start,
				isBlock:       isBlock,
				hasCodeBefore: hasCodeBefore,
				text:          text,
			})
			continue
		}

		if kind == lexer.KindError {
			continue
		}

		// Code token
		sLine, _ := ts.LineCol(src, start)
		eLine, _ := ts.LineCol(src, start+length)
		lastCodeTokenEndLine = eLine
		isClosing := false
		if kind == lexer.KindPunct && length == 1 {
			b := src[start]
			if b == '}' || b == ']' || b == ')' || b == ',' || b == ';' {
				isClosing = true
			}
		}

		codeTokens = append(codeTokens, codeTokenInfo{
			startPos:  start,
			endPos:    start + length,
			startLine: sLine,
			endLine:   eLine,
			isClosing: isClosing,
			comments:  pendingBlocks,
		})
		pendingBlocks = nil
	}

	// Imaginary EOF token to capture trailing comments after last code token
	totalLines := len(ts.Lines)
	if totalLines == 0 {
		totalLines = 1
	}
	codeTokens = append(codeTokens, codeTokenInfo{
		startPos:  uint32(len(src)),
		endPos:    uint32(len(src)),
		startLine: totalLines,
		endLine:   totalLines,
		isClosing: true,
		comments:  pendingBlocks,
	})

	for idx := 0; idx < len(codeTokens); idx++ {
		curr := codeTokens[idx]
		groups := curr.comments
		if len(groups) == 0 {
			continue
		}

		prevEndLine := 0
		prevEndPos := uint32(0)
		hasPrev := (idx > 0)
		if hasPrev {
			prevEndLine = codeTokens[idx-1].endLine
			prevEndPos = codeTokens[idx-1].endPos
		}

		// 1. Maybe donate to prev
		var trail string
		var remaining []commentBlock
		hasTrail := false

		if hasPrev && groups[0].startLine <= prevEndLine+1 {
			if groups[0].startLine <= prevEndLine {
				trail = groups[0].text
				remaining = groups[1:]
				hasTrail = true
			} else if len(groups) > 1 {
				trail = groups[0].text
				remaining = groups[1:]
				hasTrail = true
			} else {
				// len(groups) == 1 && groups[0].startLine == prevEndLine + 1
				if groups[0].endLine < curr.startLine-1 {
					trail = groups[0].text
					remaining = nil
					hasTrail = true
				} else if curr.isClosing {
					if groups[0].startLine == prevEndLine && groups[0].endLine == curr.startLine {
						// Ambiguous between prev and curr
						remaining = groups
					} else {
						trail = groups[0].text
						remaining = nil
						hasTrail = true
					}
				} else {
					remaining = groups
				}
			}
		} else {
			remaining = groups
		}

		if hasTrail {
			table.trailingComments[prevEndPos] = trail
		}

		// 2. Maybe attach to curr
		if len(remaining) > 0 {
			if len(remaining) == 1 && !hasTrail && hasPrev {
				if remaining[0].startLine == prevEndLine && remaining[0].endLine == curr.startLine {
					// Ambiguous attachment: leave detached
					table.leadingDetachedComments[curr.startPos] = []string{remaining[0].text}
					continue
				}
			}

			lastIdx := len(remaining) - 1
			if remaining[lastIdx].endLine >= curr.startLine-1 {
				table.leadingComments[curr.startPos] = remaining[lastIdx].text
				if lastIdx > 0 {
					detached := make([]string, lastIdx)
					for d := 0; d < lastIdx; d++ {
						detached[d] = remaining[d].text
					}
					table.leadingDetachedComments[curr.startPos] = detached
				}
			} else {
				detached := make([]string, len(remaining))
				for d := 0; d < len(remaining); d++ {
					detached[d] = remaining[d].text
				}
				table.leadingDetachedComments[curr.startPos] = detached
			}
		}
	}

	return table
}

// TakeLeading returns the leading and leading detached comments associated with startPos, marking them used.
func (t *CommentTable) TakeLeading(startPos uint32) (string, []string) {
	if t == nil || t.usedLeading[startPos] {
		return "", nil
	}
	t.usedLeading[startPos] = true
	return t.leadingComments[startPos], t.leadingDetachedComments[startPos]
}

// TakeTrailing returns the trailing comments associated with endPos, marking them used.
func (t *CommentTable) TakeTrailing(endPos uint32) string {
	if t == nil || t.usedTrailing[endPos] {
		return ""
	}
	t.usedTrailing[endPos] = true
	return t.trailingComments[endPos]
}

func cleanCommentText(src []byte, start, length uint32, kind lexer.Kind) string {
	raw := src[start : start+length]
	if kind == lexer.KindLineComment {
		txt := string(raw)
		if strings.HasPrefix(txt, "//") {
			txt = txt[2:]
		}
		txt = strings.TrimRight(txt, "\r")
		return txt + "\n"
	}

	// Block comment: /* ... */
	if len(raw) < 4 {
		return ""
	}
	content := string(raw[2 : len(raw)-2])
	lines := strings.Split(content, "\n")
	var buf strings.Builder
	for i, l := range lines {
		l = strings.TrimRight(l, "\r")
		if i == 0 {
			buf.WriteString(l)
			continue
		}
		buf.WriteByte('\n')

		// Strip leading spaces and tabs
		j := 0
		for j < len(l) && (l[j] == ' ' || l[j] == '\t') {
			j++
		}
		if j < len(l) && l[j] == '*' {
			l = l[j+1:]
		} else if j > 0 {
			l = l[j:]
		}
		buf.WriteString(l)
	}
	return buf.String()
}
