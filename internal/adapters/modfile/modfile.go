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

	directiveDirect  = "direct"
	directiveReplace = "replace"
	replaceArrow     = "=>"
)

var (
	errUnexpectedToken        = errors.New("unexpected token")
	errUnclosedDirect         = errors.New("unclosed direct block")
	errUnclosedReplace        = errors.New("unclosed replace block")
	errDuplicateModule        = errors.New("duplicate module")
	errDuplicateReplace       = errors.New("duplicate replace")
	errEmptyDependency        = errors.New("empty dependency")
	errEmptyReplacePath       = errors.New("empty replace path")
	errReplaceVersionRequired = errors.New("replace version required")
	errGarbageOutside         = errors.New("content outside direct block")
	errMultipleDirect         = errors.New("multiple direct blocks")
)

// FS is the minimal filesystem surface used to read/write protobuf.mod.
type FS interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Exists(name string) bool
}

// File is the parsed contents of protobuf.mod.
type File struct {
	Direct  []string
	Replace []Replace
}

// Replace maps a module (name + version) to a local filesystem path.
type Replace struct {
	Module models.Module
	Path   string
}

type replaceKey struct {
	name    string
	version models.RequestedVersion
}

// Parse parses protobuf.mod contents.
// An empty file or a file with only comments/blank lines yields an empty File.
func Parse(data []byte) (File, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var file File
	seenDirect := make(map[string]struct{})
	seenReplace := make(map[replaceKey]struct{})
	inDirect := false
	inReplace := false
	sawDirect := false
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(stripComment(scanner.Text()))
		if line == "" {
			continue
		}

		switch {
		case inDirect:
			if line == ")" {
				inDirect = false
				continue
			}
			entry, err := parseDirectEntry(line, lineNo, seenDirect)
			if err != nil {
				return File{}, err
			}
			file.Direct = append(file.Direct, entry)
		case inReplace:
			if line == ")" {
				inReplace = false
				continue
			}
			rep, err := parseReplaceSpec(line, lineNo, seenReplace)
			if err != nil {
				return File{}, err
			}
			file.Replace = append(file.Replace, rep)
		default:
			opened, err := tryOpenDirect(line, sawDirect, lineNo)
			if err != nil {
				return File{}, err
			}
			if opened {
				inDirect = true
				sawDirect = true
				continue
			}
			if tryOpenBlock(line, directiveReplace) {
				inReplace = true
				continue
			}
			if isSingleLineReplace(line) {
				spec := strings.TrimSpace(line[len(directiveReplace):])
				rep, err := parseReplaceSpec(spec, lineNo, seenReplace)
				if err != nil {
					return File{}, err
				}
				file.Replace = append(file.Replace, rep)
				continue
			}
			return File{}, fmt.Errorf("%w %q at line %d: %w", errGarbageOutside, line, lineNo, errUnexpectedToken)
		}
	}

	err := scanner.Err()
	if err != nil {
		return File{}, fmt.Errorf("scan: %w", err)
	}
	if inDirect {
		return File{}, errUnclosedDirect
	}
	if inReplace {
		return File{}, errUnclosedReplace
	}

	return file, nil
}

func parseDirectEntry(line string, lineNo int, seen map[string]struct{}) (string, error) {
	fields := strings.Fields(line)
	if len(fields) != 1 {
		return "", fmt.Errorf("%w: want single path[@version], got %q at line %d", errUnexpectedToken, line, lineNo)
	}
	entry := fields[0]
	if entry == "" {
		return "", fmt.Errorf("%w at line %d", errEmptyDependency, lineNo)
	}

	name := models.NewModule(entry).Name
	if _, ok := seen[name]; ok {
		return "", fmt.Errorf("%w %q at line %d", errDuplicateModule, name, lineNo)
	}
	seen[name] = struct{}{}
	return entry, nil
}

func parseReplaceSpec(line string, lineNo int, seen map[replaceKey]struct{}) (Replace, error) {
	left, right, ok := strings.Cut(line, replaceArrow)
	if !ok {
		return Replace{}, fmt.Errorf("%w: want module@version => path, got %q at line %d", errUnexpectedToken, line, lineNo)
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return Replace{}, fmt.Errorf("%w at line %d", errEmptyDependency, lineNo)
	}
	if right == "" {
		return Replace{}, fmt.Errorf("%w at line %d", errEmptyReplacePath, lineNo)
	}

	fields := strings.Fields(left)
	if len(fields) != 1 {
		return Replace{}, fmt.Errorf("%w: want single module@version, got %q at line %d", errUnexpectedToken, left, lineNo)
	}

	module := models.NewModule(fields[0])
	if module.Name == "" {
		return Replace{}, fmt.Errorf("%w at line %d", errEmptyDependency, lineNo)
	}
	if module.Version.IsOmitted() {
		return Replace{}, fmt.Errorf("%w at line %d", errReplaceVersionRequired, lineNo)
	}

	key := replaceKey{name: module.Name, version: module.Version}
	if _, exists := seen[key]; exists {
		return Replace{}, fmt.Errorf("%w %q at line %d", errDuplicateReplace, fields[0], lineNo)
	}
	seen[key] = struct{}{}

	return Replace{Module: module, Path: right}, nil
}

func tryOpenDirect(line string, sawDirect bool, lineNo int) (bool, error) {
	if !tryOpenBlock(line, directiveDirect) {
		return false, nil
	}
	if sawDirect {
		return false, fmt.Errorf("%w at line %d", errMultipleDirect, lineNo)
	}
	return true, nil
}

func tryOpenBlock(line, directive string) bool {
	if line == directive+" (" || line == directive+"(" {
		return true
	}
	if strings.HasPrefix(line, directive) {
		rest := strings.TrimSpace(strings.TrimPrefix(line, directive))
		if rest == "(" {
			return true
		}
	}
	return false
}

func isSingleLineReplace(line string) bool {
	if !strings.HasPrefix(line, directiveReplace) {
		return false
	}
	rest := line[len(directiveReplace):]
	if rest == "" {
		return true
	}
	return rest[0] == ' ' || rest[0] == '\t'
}

// Read reads protobuf.mod from fs. Missing file yields an empty File.
func Read(fs FS) (File, error) {
	if !fs.Exists(FileName) {
		return File{}, nil
	}

	fp, err := fs.Open(FileName)
	if err != nil {
		return File{}, fmt.Errorf("fs.Open: %w", err)
	}
	defer func() {
		_ = fp.Close()
	}()

	data, err := io.ReadAll(fp)
	if err != nil {
		return File{}, fmt.Errorf("io.ReadAll: %w", err)
	}

	file, err := Parse(data)
	if err != nil {
		return File{}, fmt.Errorf("Parse: %w", err)
	}

	return file, nil
}

// Format renders a protobuf.mod file.
func Format(file File) []byte {
	var b strings.Builder
	if len(file.Direct) > 0 || len(file.Replace) == 0 {
		b.WriteString("direct (\n")
		for _, dep := range file.Direct {
			b.WriteString("\t")
			b.WriteString(dep)
			b.WriteString("\n")
		}
		b.WriteString(")\n")
	}
	if len(file.Replace) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("replace (\n")
		for _, rep := range file.Replace {
			b.WriteString("\t")
			b.WriteString(formatModule(rep.Module))
			b.WriteString(" ")
			b.WriteString(replaceArrow)
			b.WriteString(" ")
			b.WriteString(rep.Path)
			b.WriteString("\n")
		}
		b.WriteString(")\n")
	}
	return []byte(b.String())
}

// Write writes file to protobuf.mod via fs.
func Write(fs FS, file File) error {
	fp, err := fs.Create(FileName)
	if err != nil {
		return fmt.Errorf("fs.Create: %w", err)
	}
	defer func() {
		_ = fp.Close()
	}()

	_, err = fp.Write(Format(file))
	if err != nil {
		return fmt.Errorf("fp.Write: %w", err)
	}

	return nil
}

func formatModule(module models.Module) string {
	if module.Version.IsOmitted() {
		return module.Name
	}
	return module.Name + "@" + string(module.Version)
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
