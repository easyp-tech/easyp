package v1

import (
	"bytes"
	"fmt"
	"io"

	"github.com/a8m/envsubst"
)

func expandConfigYAML(r io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return expandConfigBytes(raw)
}

func expandConfigBytes(raw []byte) ([]byte, error) {
	// Keep YAML comments out of envsubst while preserving the original source.
	// Values in quoted strings and block scalars still undergo substitution.
	expanded := make([]byte, 0, len(raw))
	segmentStart := 0
	var quote byte
	var blockIndent, headerIndent int
	inBlock := false
	for lineStart := 0; lineStart < len(raw); {
		lineEnd := len(raw)
		if newline := bytes.IndexByte(raw[lineStart:], '\n'); newline >= 0 {
			lineEnd = lineStart + newline
		}
		nextLine := lineEnd
		if nextLine < len(raw) {
			nextLine++
		}
		line := raw[lineStart:lineEnd]
		indent := leadingSpaces(line)
		trimmed := bytes.TrimSpace(line)
		if inBlock {
			switch {
			case len(trimmed) == 0:
				lineStart = nextLine
				continue
			case blockIndent == 0 && indent > headerIndent:
				blockIndent = indent
				lineStart = nextLine
				continue
			case blockIndent > 0 && indent >= blockIndent:
				lineStart = nextLine
				continue
			case trimmed[0] != '#':
				inBlock = false
			}
		}

		comment := yamlCommentStart(line, &quote)
		code := line
		if comment >= 0 {
			code = line[:comment]
			part, err := envsubst.Bytes(raw[segmentStart : lineStart+comment])
			if err != nil {
				return nil, fmt.Errorf("envsubst.Bytes: %w", err)
			}
			expanded = append(expanded, part...)
			expanded = append(expanded, raw[lineStart+comment:nextLine]...)
			segmentStart = nextLine
		}
		if quote == 0 {
			if explicitIndent, ok := blockScalarHeader(code); ok {
				inBlock = true
				headerIndent = indent
				blockIndent = 0
				if explicitIndent > 0 {
					blockIndent = headerIndent + explicitIndent
				}
			}
		}
		lineStart = nextLine
	}
	part, err := envsubst.Bytes(raw[segmentStart:])
	if err != nil {
		return nil, fmt.Errorf("envsubst.Bytes: %w", err)
	}
	expanded = append(expanded, part...)
	return expanded, nil
}

func yamlCommentStart(line []byte, quote *byte) int {
	for i := 0; i < len(line); i++ {
		switch *quote {
		case '\'':
			if line[i] == '\'' {
				if i+1 < len(line) && line[i+1] == '\'' {
					i++
				} else {
					*quote = 0
				}
			}
		case '"':
			if line[i] == '\\' {
				i++
			} else if line[i] == '"' {
				*quote = 0
			}
		default:
			switch line[i] {
			case '\'', '"':
				if yamlQuoteStarts(line[:i]) {
					*quote = line[i]
				}
			case '#':
				if i == 0 || line[i-1] == ' ' || line[i-1] == '\t' {
					return i
				}
			}
		}
	}
	return -1
}

func yamlQuoteStarts(prefix []byte) bool {
	prefix = bytes.TrimRight(prefix, " \t")
	if len(prefix) == 0 {
		return true
	}
	switch prefix[len(prefix)-1] {
	case ':', '-', '[', '{', ',':
		return true
	default:
		return false
	}
}

func blockScalarHeader(code []byte) (int, bool) {
	code = bytes.TrimSpace(code)
	separator := bytes.LastIndexByte(code, ':')
	if separator < 0 && len(code) > 1 && code[0] == '-' && (code[1] == ' ' || code[1] == '\t') {
		separator = 0
	}
	if separator < 0 {
		return 0, false
	}
	value := bytes.TrimSpace(code[separator+1:])
	if len(value) == 0 || (value[0] != '|' && value[0] != '>') {
		return 0, false
	}
	indent := 0
	for _, indicator := range value[1:] {
		switch {
		case indicator == '+' || indicator == '-':
		case indicator >= '1' && indicator <= '9' && indent == 0:
			indent = int(indicator - '0')
		default:
			return 0, false
		}
	}
	return indent, true
}

func leadingSpaces(line []byte) int {
	for i, b := range line {
		if b != ' ' {
			return i
		}
	}
	return len(line)
}
