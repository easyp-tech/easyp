package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"google.golang.org/protobuf/types/pluginpb"
)

const averageGeneratedFileSize = 15 * 1024

func writeInsertionPoint(
	_ context.Context,
	insertionPointFile *pluginpb.CodeGeneratorResponse_File,
	targetFile io.Reader,
) (_ []byte, retErr error) {
	targetScanner := bufio.NewScanner(targetFile)
	match := []byte("@@protoc_insertion_point(" + insertionPointFile.GetInsertionPoint() + ")")
	postInsertionContent := bytes.NewBuffer(nil)
	postInsertionContent.Grow(averageGeneratedFileSize)
	// TODO: We should account for line terminators in the generated file. This will require
	// either that targetFile be an io.ReadSeeker and in the case of a single line file
	// two full passes over the file, or a custom implementation of bufio.Scanner.Scan()
	newline := []byte{'\n'}
	var found bool
	for i := 0; targetScanner.Scan(); i++ {
		if i > 0 {
			_, _ = postInsertionContent.Write(newline)
		}
		targetLine := targetScanner.Bytes()
		if !bytes.Contains(targetLine, match) {
			_, _ = postInsertionContent.Write(targetLine)
			continue
		}
		// Add leading whitespace to the inserted content
		whitespace := leadingWhitespace(targetLine)

		// Insert the content from the insertion point file
		insertedContentReader := strings.NewReader(insertionPointFile.GetContent())
		writeWithPrefixAndLineEnding(postInsertionContent, insertedContentReader, whitespace, newline)

		// The inserted code is placed directly above the line containing the insertion point,
		// so we include it last.
		_, _ = postInsertionContent.Write(targetLine)
		found = true
	}
	if err := targetScanner.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("could not find insertion point %q in %q", insertionPointFile.GetInsertionPoint(), insertionPointFile.GetName())
	}
	return postInsertionContent.Bytes(), nil
}

// leadingWhitespace iterates over the given byte slice and returns the leading whitespace
// as a byte slice, accounting for UTF-8 encoding.
//
//	leadingWhitespace("\u205F   foo ") -> "\u205F   "
func leadingWhitespace(buf []byte) []byte {
	leadingSize := 0
	iterBuf := buf
	for len(iterBuf) > 0 {
		r, size := utf8.DecodeRune(iterBuf)
		// Protobuf strings must be valid UTF-8
		// https://developers.google.com/protocol-buffers/docs/proto3#scalar
		// Since utf8.RuneError is not a space, we terminate and return the leading
		// valid sequence of UTF-8 whitespace.
		if !unicode.IsSpace(r) {
			out := make([]byte, leadingSize)
			copy(out, buf)
			return out
		}
		leadingSize += size
		iterBuf = iterBuf[size:]
	}
	return buf
}

// writeWithPrefixAndLineEnding iterates over each line of the given reader,
// adding the prefix to the beginning and the newline to the end.
func writeWithPrefixAndLineEnding(dst *bytes.Buffer, src io.Reader, prefix, newline []byte) {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		// Эти записи не могут завершиться ошибкой, они вызовут панику, если не смогут выделить память.
		_, _ = dst.Write(prefix)
		_, _ = dst.Write(scanner.Bytes())
		_, _ = dst.Write(newline)
	}
}
