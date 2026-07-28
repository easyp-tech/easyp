package modfile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/models"
)

const (
	// FileName is the project-root dependency declaration file.
	FileName = "protobuf.mod"

	directiveDirect = "direct"
)

var (
	errUnexpectedToken = errors.New("unexpected token")
	errUnclosedDirect  = errors.New("unclosed direct block")
	errDuplicateModule = errors.New("duplicate module")
	errEmptyDependency = errors.New("empty dependency")
	errGarbageOutside  = errors.New("content outside direct block")
	errMultipleDirect  = errors.New("multiple direct blocks")
)

// FS is the minimal filesystem surface used to read/write protobuf.mod.
type FS interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Exists(name string) bool
}

// Parse parses protobuf.mod contents and returns dependency strings (path[@version]).
// An empty file or a file with only comments/blank lines yields an empty list.
func Parse(data []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var deps []string
	seen := make(map[string]struct{})
	inDirect := false
	sawDirect := false
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(stripComment(scanner.Text()))
		if line == "" {
			continue
		}

		if !inDirect {
			opened, err := tryOpenDirect(line, sawDirect, lineNo)
			if err != nil {
				return nil, err
			}
			if opened {
				inDirect = true
				sawDirect = true
				continue
			}
			return nil, fmt.Errorf("%w %q at line %d: %w", errGarbageOutside, line, lineNo, errUnexpectedToken)
		}

		if line == ")" {
			inDirect = false
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 1 {
			return nil, fmt.Errorf("%w: want single path[@version], got %q at line %d", errUnexpectedToken, line, lineNo)
		}
		entry := fields[0]
		if entry == "" {
			return nil, fmt.Errorf("%w at line %d", errEmptyDependency, lineNo)
		}

		name := models.NewModule(entry).Name
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("%w %q at line %d", errDuplicateModule, name, lineNo)
		}
		seen[name] = struct{}{}
		deps = append(deps, entry)
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	if inDirect {
		return nil, errUnclosedDirect
	}

	return deps, nil
}

func tryOpenDirect(line string, sawDirect bool, lineNo int) (bool, error) {
	if line == directiveDirect+" (" || line == directiveDirect+"(" {
		if sawDirect {
			return false, fmt.Errorf("%w at line %d", errMultipleDirect, lineNo)
		}
		return true, nil
	}
	if strings.HasPrefix(line, directiveDirect) {
		rest := strings.TrimSpace(strings.TrimPrefix(line, directiveDirect))
		if rest == "(" {
			if sawDirect {
				return false, fmt.Errorf("%w at line %d", errMultipleDirect, lineNo)
			}
			return true, nil
		}
	}
	return false, nil
}

// Read reads protobuf.mod from fs. Missing file yields an empty dependency list.
func Read(fs FS) ([]string, error) {
	if !fs.Exists(FileName) {
		return nil, nil
	}

	fp, err := fs.Open(FileName)
	if err != nil {
		return nil, fmt.Errorf("fs.Open: %w", err)
	}
	defer func() {
		_ = fp.Close()
	}()

	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	deps, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}

	return deps, nil
}

// Format renders a protobuf.mod file for the given dependency strings.
func Format(deps []string) []byte {
	var b strings.Builder
	b.WriteString("direct (\n")
	for _, dep := range deps {
		b.WriteString("\t")
		b.WriteString(dep)
		b.WriteString("\n")
	}
	b.WriteString(")\n")
	return []byte(b.String())
}

// Write writes deps to protobuf.mod via fs.
func Write(fs FS, deps []string) error {
	fp, err := fs.Create(FileName)
	if err != nil {
		return fmt.Errorf("fs.Create: %w", err)
	}
	defer func() {
		_ = fp.Close()
	}()

	_, err = fp.Write(Format(deps))
	if err != nil {
		return fmt.Errorf("fp.Write: %w", err)
	}

	return nil
}

func stripComment(line string) string {
	hashIdx := strings.Index(line, "#")
	slashIdx := strings.Index(line, "//")

	cut := -1
	switch {
	case hashIdx >= 0 && slashIdx >= 0:
		cut = min(hashIdx, slashIdx)
	case hashIdx >= 0:
		cut = hashIdx
	case slashIdx >= 0:
		cut = slashIdx
	}
	if cut >= 0 {
		return line[:cut]
	}
	return line
}
