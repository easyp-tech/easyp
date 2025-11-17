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

func writeInsertionPoint(
	ctx context.Context,
	insertionPointFile *pluginpb.CodeGeneratorResponse_File,
	targetFile io.Reader,
) (_ []byte, retErr error) {
	targetScanner := bufio.NewScanner(targetFile)
	match := []byte("@@protoc_insertion_point(" + insertionPointFile.GetInsertionPoint() + ")")
	postInsertionContent := bytes.NewBuffer(nil)
	postInsertionContent.Grow(averageGeneratedFileSize)
	// TODO: Мы должны учитывать окончания строк в сгенерированном файле. Это потребует
	// либо чтобы targetFile был io.ReadSeeker и в худшем случае
	// выполнения 2 полных сканирований файла (если это одна строка), либо реализации
	// bufio.Scanner.Scan() встроенным образом
	newline := []byte{'\n'}
	var found bool
	for i := 0; targetScanner.Scan(); i++ {
		if i > 0 {
			// Эти записи не могут завершиться ошибкой, они вызовут панику, если не смогут выделить память.
			_, _ = postInsertionContent.Write(newline)
		}
		targetLine := targetScanner.Bytes()
		if !bytes.Contains(targetLine, match) {
			// Эти записи не могут завершиться ошибкой, они вызовут панику, если не смогут выделить память.
			_, _ = postInsertionContent.Write(targetLine)
			continue
		}
		// Для каждой строки в новом содержимом применяем
		// такое же количество пробелов. Это важно
		// для определенных языков, например Python.
		whitespace := leadingWhitespace(targetLine)

		// Вставляем содержимое из файла точки вставки. Обрабатываем переводы строк
		// платформо-независимым способом.
		insertedContentReader := strings.NewReader(insertionPointFile.GetContent())
		writeWithPrefixAndLineEnding(postInsertionContent, insertedContentReader, whitespace, newline)

		// Код, вставленный в этой точке, размещается непосредственно
		// над строкой, содержащей точку вставки, поэтому
		// мы включаем её последней.
		// Эти записи не могут завершиться ошибкой, они вызовут панику, если не смогут
		// выделить память
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

// leadingWhitespace проходит по заданной строке
// и возвращает подстроку начальных пробелов, если они есть,
// с учетом кодировки utf-8.
//
//	leadingWhitespace("\u205F   foo ") -> "\u205F   "
func leadingWhitespace(buf []byte) []byte {
	leadingSize := 0
	iterBuf := buf
	for len(iterBuf) > 0 {
		r, size := utf8.DecodeRune(iterBuf)
		// строки protobuf всегда должны быть валидным UTF8
		// https://developers.google.com/protocol-buffers/docs/proto3#scalar
		// Кроме того, utf8.RuneError не является пробелом, поэтому мы завершаем
		// и возвращаем начальную, валидную, последовательность пробелов UTF8.
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

// writeWithPrefixAndLineEnding проходит по каждой строке заданного reader'а,
// добавляет префикс в начало и добавляет последовательность перевода строки в конец.
func writeWithPrefixAndLineEnding(dst *bytes.Buffer, src io.Reader, prefix, newline []byte) {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		// Эти записи не могут завершиться ошибкой, они вызовут панику, если не смогут выделить память.
		_, _ = dst.Write(prefix)
		_, _ = dst.Write(scanner.Bytes())
		_, _ = dst.Write(newline)
	}
}
