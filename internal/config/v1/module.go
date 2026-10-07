package v1

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"
)

// Module describes the module identity and source roots in protobuf.mod.
type Module struct {
	Name     string
	Roots    []string
	Requires []Requirement
	Replaces []Replacement
	// RootsFromMetadata distinguishes authoritative source roots, including the
	// native default '.', from an implicit root on pre-native dependencies.
	RootsFromMetadata bool
	// ProtoFilters carries Buf source selection relative to the dependency directory.
	// Native protobuf.mod files leave it empty.
	ProtoFilters []ProtoFileFilter
	// BSRDependencies retains Buf metadata until a backend chooses Git requirements.
	// Native protobuf.mod files leave it empty.
	BSRDependencies []BSRDependency
}

// ProtoFileFilter selects proto files below a root without changing import names.
// Includes and excludes use the same base directory as Root.
type ProtoFileFilter struct {
	Root     string
	Includes []string
	Excludes []string
}

// Requirement names a dependency and its optional version or commit.
type Requirement struct {
	Module  string
	Version string
}

// Replacement redirects a dependency to a local module directory.
type Replacement struct {
	Module string
	Target string
}

// IsModuleManifest reports whether protobuf.mod uses the v1 module syntax.
func IsModuleManifest(raw []byte) bool {
	for _, line := range strings.Split(strings.TrimPrefix(string(raw), "\ufeff"), "\n") {
		fields := strings.Fields(stripModuleComment(line))
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "module" {
			return true
		}
		break
	}
	_, err := ParseModule(bytes.NewReader(raw))
	return err == nil
}

// ParseModule reads the v1 module identity, roots and dependency directives.
// Block entries must occupy separate lines from the opening and closing parentheses.
func ParseModule(r io.Reader) (Module, error) {
	var result Module
	scanner := bufio.NewScanner(r)
	block := ""
	blockLine := 0
	requires := make(map[string]int)
	replaces := make(map[string]int)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if line == 1 {
			text = strings.TrimPrefix(text, "\ufeff")
		}
		text = strings.TrimSpace(stripModuleComment(text))
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
		fields := strings.Fields(text)
		opensBlock := text == "(" || text == "roots(" || text == "require(" || text == "replace(" ||
			(len(fields) == 2 && fields[1] == "(")
		if opensBlock {
			if block != "" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: nested block (opened at line %d)", line, blockLine)
			}
			block = strings.TrimSpace(strings.TrimSuffix(text, "("))
			if block == "direct" || block == "indirect" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: legacy %s block; run easyp migrate --module <identity> to preview conversion", line, block)
			}
			if block != "roots" && block != "require" && block != "replace" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: unknown block %q", line, block)
			}
			blockLine = line
			continue
		}

		directive := block
		if block == "" {
			directive = fields[0]
			fields = fields[1:]
		}
		if len(fields) > 0 && strings.HasPrefix(fields[0], "(") {
			if block != "" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: nested block (opened at line %d)", line, blockLine)
			}
			// Requirement metadata editing relies on block entries occupying separate lines.
			return Module{}, fmt.Errorf("protobuf.mod:%d: inline blocks are not supported; put entries on separate lines", line)
		}
		for i, field := range fields {
			if directive == "replace" && i == 1 && field == "=>" {
				continue
			}
			if !validModuleToken(field) {
				return Module{}, fmt.Errorf("protobuf.mod:%d: invalid token %q", line, field)
			}
		}
		switch directive {
		case "module":
			if result.Name != "" || len(fields) != 1 {
				return Module{}, fmt.Errorf("protobuf.mod:%d: expected one module identity", line)
			}
			if err := validateManifestModuleIdentity(fields[0], "module"); err != nil {
				return Module{}, fmt.Errorf("protobuf.mod:%d: %w", line, err)
			}
			if err := ValidateModuleVersion(fields[0], ""); err != nil {
				return Module{}, fmt.Errorf("protobuf.mod:%d: %w", line, err)
			}
			result.Name = fields[0]
		case "roots":
			if len(fields) == 0 {
				return Module{}, fmt.Errorf("protobuf.mod:%d: roots expects at least one path", line)
			}
			for _, root := range fields {
				clean := filepath.Clean(root)
				if filepath.IsAbs(root) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
					return Module{}, fmt.Errorf("protobuf.mod:%d: root %q leaves the module", line, root)
				}
				result.Roots = append(result.Roots, clean)
			}
		case "require":
			if len(fields) < 1 || len(fields) > 2 {
				return Module{}, fmt.Errorf("protobuf.mod:%d: require expects module and optional version", line)
			}
			if err := validateManifestModuleIdentity(fields[0], "require"); err != nil {
				return Module{}, fmt.Errorf("protobuf.mod:%d: %w", line, err)
			}
			if len(fields) == 2 && strings.ContainsAny(fields[1], "()") {
				return Module{}, fmt.Errorf("protobuf.mod:%d: invalid version token %q", line, fields[1])
			}
			if firstLine, ok := requires[fields[0]]; ok {
				return Module{}, fmt.Errorf("protobuf.mod:%d: duplicate require source %q (first declared at line %d)", line, fields[0], firstLine)
			}
			requires[fields[0]] = line
			requirement := Requirement{Module: fields[0]}
			if len(fields) == 2 {
				requirement.Version = fields[1]
			}
			if err := ValidateModuleVersion(requirement.Module, requirement.Version); err != nil {
				return Module{}, fmt.Errorf("protobuf.mod:%d: %w", line, err)
			}
			result.Requires = append(result.Requires, requirement)
		case "replace":
			if len(fields) != 3 || fields[1] != "=>" {
				return Module{}, fmt.Errorf("protobuf.mod:%d: replace expects module => target", line)
			}
			if err := validateManifestModuleIdentity(fields[0], "replace"); err != nil {
				return Module{}, fmt.Errorf("protobuf.mod:%d: %w", line, err)
			}
			if firstLine, ok := replaces[fields[0]]; ok {
				return Module{}, fmt.Errorf("protobuf.mod:%d: duplicate replace source %q (first declared at line %d)", line, fields[0], firstLine)
			}
			if err := ValidateModuleVersion(fields[0], ""); err != nil {
				return Module{}, fmt.Errorf("protobuf.mod:%d: %w", line, err)
			}
			replaces[fields[0]] = line
			result.Replaces = append(result.Replaces, Replacement{Module: fields[0], Target: fields[2]})
		default:
			return Module{}, fmt.Errorf("protobuf.mod:%d: unknown directive %q", line, directive)
		}
	}
	if err := scanner.Err(); err != nil {
		return Module{}, fmt.Errorf("Err: %w", err)
	}
	if block != "" {
		return Module{}, fmt.Errorf("protobuf.mod:%d: unclosed %s block", blockLine, block)
	}
	if result.Name == "" {
		if line == 0 {
			line = 1
		}
		return Module{}, fmt.Errorf("protobuf.mod:%d: missing module directive", line)
	}
	if len(result.Roots) == 0 {
		result.Roots = []string{"."}
	}
	result.RootsFromMetadata = true
	return result, nil
}

func validateManifestModuleIdentity(identity, directive string) error {
	authorityEnd := -1
	if scheme := strings.Index(identity, "://"); scheme >= 0 {
		if slash := strings.Index(identity[scheme+3:], "/"); slash >= 0 {
			authorityEnd = scheme + 3 + slash
		} else {
			authorityEnd = len(identity)
		}
	}
	if at := strings.LastIndex(identity, "@"); at > authorityEnd {
		base := identity[:at]
		suffix := identity[at+1:]
		if suffix == "" {
			return fmt.Errorf("%s module identity %q must not end with @", directive, identity)
		}
		switch directive {
		case "require":
			return fmt.Errorf("require module identity %q must not embed version %q; use require %s %s", identity, suffix, base, suffix)
		case "replace":
			return fmt.Errorf("replace module identity %q must not embed version %q; use replace %s => <local-path>", identity, suffix, base)
		default:
			return fmt.Errorf("%s module identity %q must not embed version %q", directive, identity, suffix)
		}
	}
	return nil
}

func stripModuleComment(line string) string {
	tokenStart := true
	for i, char := range line {
		if tokenStart && (char == '#' || strings.HasPrefix(line[i:], "//")) {
			return line[:i]
		}
		tokenStart = unicode.IsSpace(char)
	}
	return line
}

func validModuleToken(token string) bool {
	if token == "=>" || strings.HasPrefix(token, "(") || strings.HasPrefix(token, ")") || strings.ContainsAny(token, "\"`") {
		return false
	}
	// Parentheses within URL and path tokens are literal characters.
	for _, char := range token {
		if unicode.IsControl(char) || char == '\ufeff' {
			return false
		}
	}
	return true
}
