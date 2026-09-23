package v1

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Module describes the module identity and source roots in protobuf.mod.
type Module struct {
	Name     string
	Roots    []string
	Requires []Requirement
	Replaces []Replacement
}

type Requirement struct {
	Module  string
	Version string
}

type Replacement struct {
	Module string
	Target string
}

// IsModuleManifest reports whether protobuf.mod uses the v1 module syntax.
func IsModuleManifest(raw []byte) bool {
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(stripModuleComment(line))
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "module" {
			return true
		}
		break
	}
	_, err := ParseModule(strings.NewReader(string(raw)))
	return err == nil
}

func ParseModule(r io.Reader) (Module, error) {
	var result Module
	scanner := bufio.NewScanner(r)
	block := ""
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(stripModuleComment(scanner.Text()))
		if text == "" {
			continue
		}
		if text == ")" {
			if block == "" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: unmatched closing parenthesis", line)
			}
			block = ""
			continue
		}
		if strings.HasSuffix(text, "(") {
			if block != "" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: nested block", line)
			}
			block = strings.TrimSpace(strings.TrimSuffix(text, "("))
			if block != "roots" && block != "require" && block != "replace" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: unknown block %q", line, block)
			}
			continue
		}

		directive, value := block, text
		if block == "" {
			fields := strings.Fields(text)
			directive = fields[0]
			value = strings.TrimSpace(strings.TrimPrefix(text, directive))
		}
		switch directive {
		case "module":
			if result.Name != "" || len(strings.Fields(value)) != 1 {
				return Module{}, fmt.Errorf("protobuf.mod:%d: expected one module identity", line)
			}
			result.Name = value
		case "roots":
			for _, root := range strings.Fields(value) {
				if filepath.IsAbs(root) || root == ".." || strings.HasPrefix(filepath.Clean(root), ".."+string(filepath.Separator)) {
					return Module{}, fmt.Errorf("protobuf.mod:%d: root %q leaves the module", line, root)
				}
				result.Roots = append(result.Roots, filepath.Clean(root))
			}
		case "require":
			fields := strings.Fields(value)
			if len(fields) < 1 || len(fields) > 2 {
				return Module{}, fmt.Errorf("protobuf.mod:%d: require expects module and optional version", line)
			}
			requirement := Requirement{Module: fields[0]}
			if len(fields) == 2 {
				requirement.Version = fields[1]
			}
			result.Requires = append(result.Requires, requirement)
		case "replace":
			fields := strings.Fields(value)
			if len(fields) != 3 || fields[1] != "=>" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: replace expects module => target", line)
			}
			result.Replaces = append(result.Replaces, Replacement{Module: fields[0], Target: fields[2]})
		default:
			return Module{}, fmt.Errorf("protobuf.mod:%d: unknown directive %q", line, directive)
		}
	}
	if err := scanner.Err(); err != nil {
		return Module{}, err
	}
	if block != "" {
		return Module{}, fmt.Errorf("protobuf.mod: unclosed %s block", block)
	}
	if result.Name == "" {
		return Module{}, fmt.Errorf("protobuf.mod: missing module directive")
	}
	if len(result.Roots) == 0 {
		result.Roots = []string{"."}
	}
	return result, nil
}

func stripModuleComment(line string) string {
	for i := 0; i+1 < len(line); i++ {
		if line[i] == '/' && line[i+1] == '/' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
			return line[:i]
		}
	}
	return line
}
