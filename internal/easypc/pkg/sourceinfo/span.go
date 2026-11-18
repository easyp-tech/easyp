package sourceinfo

import (
	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// MakeSpan converts a TextSpan into a 3-element or 4-element 0-based span:
// - Single-line: [line, start_col, end_col]
// - Multi-line:  [start_line, start_col, end_line, end_col]
func MakeSpan(ts *lexer.TokenStream, src []byte, span parser.TextSpan) ([4]int32, uint8) {
	if span.IsEmpty() {
		return [4]int32{0, 0, 0, 0}, 3
	}
	startLine, startCol := ts.LineCol(src, span.Start)
	endLine, endCol := ts.LineCol(src, span.End)

	sLine := int32(startLine - 1)
	sCol := int32(startCol - 1)
	eLine := int32(endLine - 1)
	eCol := int32(endCol - 1)

	if sLine == eLine {
		return [4]int32{sLine, sCol, eCol, 0}, 3
	}
	return [4]int32{sLine, sCol, eLine, eCol}, 4
}
