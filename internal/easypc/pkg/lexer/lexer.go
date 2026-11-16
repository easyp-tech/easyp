package lexer

import (
	"errors"
	"math"
	"unicode/utf8"
)

var (
	// ErrFileTooLarge is returned when input exceeds 4 GiB (MaxUint32 bytes).
	ErrFileTooLarge = errors.New("file exceeds maximum supported size of 4 GiB")
)

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isIdentContinue(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\v' || b == '\f'
}

// matchInt validates a numeric run against INT grammar rules from docs/spec/lexer.md §3.2:
//   Hex:     0[xX][0-9A-Fa-f]+
//   Octal:   0[0-7]* (note: single '0' matches this)
//   Decimal: [1-9][0-9]*
func matchInt(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	// Hex: 0[xX][0-9A-Fa-f]+
	if len(b) >= 3 && b[0] == '0' && (b[1] == 'x' || b[1] == 'X') {
		for _, c := range b[2:] {
			if !isHexDigit(c) {
				return false
			}
		}
		return true
	}
	// Octal: 0[0-7]*
	if b[0] == '0' {
		for _, c := range b[1:] {
			if c < '0' || c > '7' {
				return false
			}
		}
		return true
	}
	// Decimal: [1-9][0-9]*
	if b[0] >= '1' && b[0] <= '9' {
		for _, c := range b[1:] {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// matchFloat validates a numeric run against FLOAT grammar rules from docs/spec/lexer.md §3.2:
//   Form 1: D+ "." D* E?
//   Form 2: D+ E
//   Form 3: "." D+ E?
//   where D = [0-9], E = [eE][+-]?D+
func matchFloat(b []byte) bool {
	n := len(b)
	if n == 0 {
		return false
	}

	// Form 3: "." D+ E?
	if b[0] == '.' {
		if n < 2 || !isDigit(b[1]) {
			return false
		}
		i := 2
		for i < n && isDigit(b[i]) {
			i++
		}
		if i == n {
			return true
		}
		if b[i] == 'e' || b[i] == 'E' {
			return matchExponent(b[i:])
		}
		return false
	}

	// Form 1 (D+ "." D* E?) or Form 2 (D+ E)
	if !isDigit(b[0]) {
		return false
	}
	i := 1
	for i < n && isDigit(b[i]) {
		i++
	}
	if i == n {
		// Only digits without dot or exponent -> not a float
		return false
	}

	if b[i] == '.' {
		i++
		for i < n && isDigit(b[i]) {
			i++
		}
		if i == n {
			return true // e.g. "1." or "1.0"
		}
		if b[i] == 'e' || b[i] == 'E' {
			return matchExponent(b[i:])
		}
		return false
	}

	if b[i] == 'e' || b[i] == 'E' {
		return matchExponent(b[i:])
	}

	return false
}

// matchExponent checks [eE][+-]?[0-9]+
func matchExponent(b []byte) bool {
	n := len(b)
	if n < 2 || (b[0] != 'e' && b[0] != 'E') {
		return false
	}
	i := 1
	if i < n && (b[i] == '+' || b[i] == '-') {
		i++
	}
	if i == n || !isDigit(b[i]) {
		return false
	}
	for i < n && isDigit(b[i]) {
		i++
	}
	return i == n
}

// Lex performs lexical analysis of the input byte slice, populating ts with tokens
// and the line start table according to docs/spec/lexer.md.
func Lex(src []byte, ts *TokenStream) error {
	if len(src) > math.MaxUint32 {
		return ErrFileTooLarge
	}

	pos := 0
	// §1: If file starts with UTF-8 BOM (0xEF, 0xBB, 0xBF), skip it.
	// Token offsets are from the start of file, so first token starts at offset 3.
	if len(src) >= 3 && src[0] == 0xef && src[1] == 0xbb && src[2] == 0xbf {
		pos = 3
	}

	// Line 1 always starts at byte offset 0
	ts.Lines = append(ts.Lines, 0)
	srcLen := len(src)

	for pos < srcLen {
		b := src[pos]

		// 1. Whitespace
		if isWhitespace(b) {
			pos++
			continue
		}

		// 2. Newline (\n)
		if b == '\n' {
			pos++
			ts.Lines = append(ts.Lines, uint32(pos))
			continue
		}

		// 3. Comments or Slash
		if b == '/' {
			if pos+1 < srcLen && src[pos+1] == '/' {
				// Line comment: // until \n or EOF
				start := pos
				pos += 2
				for pos < srcLen && src[pos] != '\n' {
					pos++
				}
				commentLen := pos - start
				commentBytes := src[start:pos]
				if utf8.Valid(commentBytes) {
					ts.Push(KindLineComment, uint32(start), uint32(commentLen))
				} else {
					ts.Push(KindError, uint32(start), uint32(commentLen))
				}
				// Next iteration will process \n if present
				continue
			}

			if pos+1 < srcLen && src[pos+1] == '*' {
				// Block comment: /* ... */
				start := pos
				pos += 2
				closed := false
				for pos < srcLen {
					if src[pos] == '\n' {
						ts.Lines = append(ts.Lines, uint32(pos+1))
						pos++
						continue
					}
					if src[pos] == '*' && pos+1 < srcLen && src[pos+1] == '/' {
						pos += 2
						closed = true
						break
					}
					pos++
				}

				if !closed {
					// Unclosed block comment: from /* to EOF
					ts.Push(KindError, uint32(start), uint32(srcLen-start))
					break
				}

				commentLen := pos - start
				commentBytes := src[start:pos]
				if utf8.Valid(commentBytes) {
					ts.Push(KindBlockComment, uint32(start), uint32(commentLen))
				} else {
					ts.Push(KindError, uint32(start), uint32(commentLen))
				}
				continue
			}

			// Standalone '/' punctuator
			ts.Push(KindPunct, uint32(pos), 1)
			pos++
			continue
		}

		// 4. Identifiers and contextual keywords
		if isIdentStart(b) {
			start := pos
			pos++
			for pos < srcLen && isIdentContinue(src[pos]) {
				pos++
			}
			ts.Push(KindIdent, uint32(start), uint32(pos-start))
			continue
		}

		// 5. Numbers (INT, FLOAT, or ERROR)
		if isDigit(b) || (b == '.' && pos+1 < srcLen && isDigit(src[pos+1])) {
			start := pos
			isHexRun := false
			if src[start] == '0' && start+1 < srcLen && (src[start+1] == 'x' || src[start+1] == 'X') {
				isHexRun = true
			}

			// Step 1: Capture maximal numeric run
			// Characters in [0-9A-Za-z_.], plus '+' or '-' if preceding char is e/E and run is not hex
			for pos < srcLen {
				c := src[pos]
				if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '.' {
					pos++
				} else if (c == '+' || c == '-') && !isHexRun && pos > start && (src[pos-1] == 'e' || src[pos-1] == 'E') {
					pos++
				} else {
					break
				}
			}

			run := src[start:pos]
			runLen := uint32(pos - start)

			// Step 2: Validate against grammar
			if matchInt(run) {
				ts.Push(KindInt, uint32(start), runLen)
			} else if matchFloat(run) {
				ts.Push(KindFloat, uint32(start), runLen)
			} else {
				ts.Push(KindError, uint32(start), runLen)
			}
			continue
		}

		// 6. Strings (STRING or ERROR)
		if b == '"' || b == '\'' {
			quote := b
			start := pos
			pos++
			hasError := false
			closed := false

			for pos < srcLen {
				c := src[pos]
				if c == quote {
					pos++
					closed = true
					break
				}
				if c == '\n' {
					// Unclosed string literal before newline: token covers up to (excluding) \n
					break
				}
				if c == 0x00 {
					hasError = true
					pos++
					continue
				}
				if c == '\\' {
					pos++
					if pos >= srcLen || src[pos] == '\n' {
						hasError = true
						continue
					}
					esc := src[pos]
					switch esc {
					case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '\'', '"', '?':
						pos++
					case 'x':
						pos++
						hexCount := 0
						for hexCount < 2 && pos < srcLen && isHexDigit(src[pos]) {
							hexCount++
							pos++
						}
						if hexCount == 0 {
							hasError = true
						}
					case '0', '1', '2', '3', '4', '5', '6', '7':
						pos++
						octCount := 1
						for octCount < 3 && pos < srcLen && src[pos] >= '0' && src[pos] <= '7' {
							octCount++
							pos++
						}
					case 'u':
						pos++
						uCount := 0
						for uCount < 4 && pos < srcLen && isHexDigit(src[pos]) {
							uCount++
							pos++
						}
						if uCount != 4 {
							hasError = true
						}
					case 'U':
						pos++
						uCount := 0
						for uCount < 8 && pos < srcLen && isHexDigit(src[pos]) {
							uCount++
							pos++
						}
						if uCount != 8 {
							hasError = true
						}
					default:
						hasError = true
						pos++
					}
					continue
				}
				pos++
			}

			if !closed {
				// Unclosed string: from quote to \n or EOF
				ts.Push(KindError, uint32(start), uint32(pos-start))
			} else {
				fullStr := src[start:pos]
				if !hasError && utf8.Valid(fullStr) {
					ts.Push(KindString, uint32(start), uint32(pos-start))
				} else {
					ts.Push(KindError, uint32(start), uint32(pos-start))
				}
			}
			continue
		}

		// 7. Punctuators: ; , . : = - + ( ) { } [ ] < >
		switch b {
		case ';', ',', '.', ':', '=', '-', '+', '(', ')', '{', '}', '[', ']', '<', '>':
			ts.Push(KindPunct, uint32(pos), 1)
			pos++
			continue
		}

		// 8. Control bytes, invalid characters, or corrupted UTF-8 outside strings/comments
		if b == 0x00 {
			ts.Push(KindError, uint32(pos), 1)
			pos++
			continue
		}

		if b >= 0x80 {
			r, size := utf8.DecodeRune(src[pos:])
			if r != utf8.RuneError && size > 1 {
				// Valid multi-byte UTF-8 sequence outside string/comment -> error spanning full sequence
				ts.Push(KindError, uint32(pos), uint32(size))
				pos += size
			} else {
				// Invalid UTF-8 byte
				ts.Push(KindError, uint32(pos), 1)
				pos++
			}
			continue
		}

		// Disallowed ASCII character (@, #, $, %, ^, &, ~, ?, \, etc.)
		ts.Push(KindError, uint32(pos), 1)
		pos++
	}

	return nil
}
