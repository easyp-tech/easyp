package lexer

import (
	"fmt"
	"strings"
)

// Kind represents the lexical token kind (1..8).
type Kind uint8

const (
	KindIdent        Kind = 1 // ident
	KindInt          Kind = 2 // int
	KindFloat        Kind = 3 // float
	KindString       Kind = 4 // string
	KindPunct        Kind = 5 // punct
	KindLineComment  Kind = 6 // line_comment
	KindBlockComment Kind = 7 // block_comment
	KindError        Kind = 8 // error
)

var kindNames = [...]string{
	0:                "unknown",
	KindIdent:        "ident",
	KindInt:          "int",
	KindFloat:        "float",
	KindString:       "string",
	KindPunct:        "punct",
	KindLineComment:  "line_comment",
	KindBlockComment: "block_comment",
	KindError:        "error",
}

// String returns the canonical dump name of the token kind.
func (k Kind) String() string {
	if int(k) < len(kindNames) && kindNames[k] != "" {
		return kindNames[k]
	}
	return "unknown"
}

// ParseKind parses the canonical dump name into a Kind.
func ParseKind(name string) (Kind, bool) {
	switch name {
	case "ident":
		return KindIdent, true
	case "int":
		return KindInt, true
	case "float":
		return KindFloat, true
	case "string":
		return KindString, true
	case "punct":
		return KindPunct, true
	case "line_comment":
		return KindLineComment, true
	case "block_comment":
		return KindBlockComment, true
	case "error":
		return KindError, true
	default:
		return 0, false
	}
}

// Token is a value representation for single-token inspection.
type Token struct {
	Kind  Kind
	Start uint32
	Len   uint32
}

// TokenStream stores tokens in Structure-of-Arrays (SoA) format.
// Total storage per token is exactly 1 (Kind) + 4 (Start) + 4 (Len) = 9 bytes.
type TokenStream struct {
	Kinds  []Kind   // uint8
	Starts []uint32 // byte offset from beginning of file
	Lens   []uint32 // byte length of token
	Lines  []uint32 // byte offset of each line start (Lines[0] = 0)
}

// NewTokenStream creates a new TokenStream with optional initial pre-allocations.
func NewTokenStream(tokenCap, lineCap int) *TokenStream {
	return &TokenStream{
		Kinds:  make([]Kind, 0, tokenCap),
		Starts: make([]uint32, 0, tokenCap),
		Lens:   make([]uint32, 0, tokenCap),
		Lines:  make([]uint32, 0, lineCap),
	}
}

// Reset clears the token stream while retaining underlying allocated slices.
func (ts *TokenStream) Reset() {
	ts.Kinds = ts.Kinds[:0]
	ts.Starts = ts.Starts[:0]
	ts.Lens = ts.Lens[:0]
	ts.Lines = ts.Lines[:0]
}

// Len returns the number of tokens in the stream.
func (ts *TokenStream) Len() int {
	return len(ts.Kinds)
}

// Push appends a token triple (kind, start, len) to the stream.
func (ts *TokenStream) Push(kind Kind, start, length uint32) {
	ts.Kinds = append(ts.Kinds, kind)
	ts.Starts = append(ts.Starts, start)
	ts.Lens = append(ts.Lens, length)
}

// Token returns the i-th token as a Token struct.
func (ts *TokenStream) Token(i int) Token {
	return Token{
		Kind:  ts.Kinds[i],
		Start: ts.Starts[i],
		Len:   ts.Lens[i],
	}
}

// LineCol computes 1-based line and 1-based column for a given byte offset.
// Tab characters ('\t') advance the 0-based column to the next multiple of 8:
// col0 = (col0/8 + 1) * 8. All other bytes advance col0 by 1.
func (ts *TokenStream) LineCol(src []byte, offset uint32) (int, int) {
	if offset > uint32(len(src)) {
		offset = uint32(len(src))
	}
	if len(ts.Lines) == 0 {
		return 1, 1
	}

	// Binary search in ts.Lines for greatest index where Lines[mid] <= offset
	low, high := 0, len(ts.Lines)-1
	lineIdx := 0
	for low <= high {
		mid := (low + high) / 2
		if ts.Lines[mid] <= offset {
			lineIdx = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	lineStart := ts.Lines[lineIdx]
	col0 := 0
	for _, b := range src[lineStart:offset] {
		if b == '\t' {
			col0 = (col0/8 + 1) * 8
		} else {
			col0++
		}
	}
	return lineIdx + 1, col0 + 1
}

const (
	fnvOffset64 = 0xcbf29ce484222325
	fnvPrime64  = 0x100000001b3
)

// FileChecksum computes 64-bit FNV-1a over all 9-byte LE tokens in the stream.
// Each token is hashed as:
//   byte 0:    Kind (uint8)
//   bytes 1..4: Start (uint32 Little-Endian)
//   bytes 5..8: Len (uint32 Little-Endian)
func (ts *TokenStream) FileChecksum() uint64 {
	h := uint64(fnvOffset64)
	n := len(ts.Kinds)
	for i := 0; i < n; i++ {
		// Byte 0: Kind (uint8)
		h = (h ^ uint64(ts.Kinds[i])) * fnvPrime64

		// Bytes 1..4: Start (uint32 Little-Endian)
		s := ts.Starts[i]
		h = (h ^ uint64(byte(s))) * fnvPrime64
		h = (h ^ uint64(byte(s>>8))) * fnvPrime64
		h = (h ^ uint64(byte(s>>16))) * fnvPrime64
		h = (h ^ uint64(byte(s>>24))) * fnvPrime64

		// Bytes 5..8: Len (uint32 Little-Endian)
		l := ts.Lens[i]
		h = (h ^ uint64(byte(l))) * fnvPrime64
		h = (h ^ uint64(byte(l>>8))) * fnvPrime64
		h = (h ^ uint64(byte(l>>16))) * fnvPrime64
		h = (h ^ uint64(byte(l>>24))) * fnvPrime64
	}
	return h
}

// CorpusChecksum computes 64-bit FNV-1a over an ordered sequence of 8-byte LE file checksums.
func CorpusChecksum(fileChecksums []uint64) uint64 {
	h := uint64(fnvOffset64)
	for _, fc := range fileChecksums {
		h = (h ^ uint64(byte(fc))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>8))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>16))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>24))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>32))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>40))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>48))) * fnvPrime64
		h = (h ^ uint64(byte(fc>>56))) * fnvPrime64
	}
	return h
}

// FormatDump generates the textual token dump conforming to docs/spec/lexer.md §5:
// # <path>
// <kind-name> <start> <len>
// ...
func (ts *TokenStream) FormatDump(path string) string {
	var sb strings.Builder
	sb.WriteString("# ")
	sb.WriteString(path)
	sb.WriteByte('\n')

	n := len(ts.Kinds)
	for i := 0; i < n; i++ {
		sb.WriteString(ts.Kinds[i].String())
		sb.WriteByte(' ')
		sb.WriteString(fmt.Sprintf("%d %d\n", ts.Starts[i], ts.Lens[i]))
	}
	return sb.String()
}

// KindCounts returns token counts indexed by kind name.
func (ts *TokenStream) KindCounts() map[string]int {
	counts := map[string]int{
		"ident":         0,
		"int":           0,
		"float":         0,
		"string":        0,
		"punct":         0,
		"line_comment":  0,
		"block_comment": 0,
		"error":         0,
	}
	for _, k := range ts.Kinds {
		name := k.String()
		counts[name]++
	}
	return counts
}
