package generation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/scanner"

	"github.com/easyp-tech/easyp/internal/modules"
)

// selectedPackageFiles reads package declarations without compiling unrelated
// files. Malformed selected files and required imports are rejected by the real
// compiler later; an unrelated broken declaration is not a generation target.
func selectedPackageFiles(ctx context.Context, selected v1GenerationModule, packages []string) ([]string, map[string]bool, error) {
	roots, err := modules.ModuleSources(selected.directory, selected.module)
	if err != nil {
		return nil, nil, fmt.Errorf("ModuleSources: %w", err)
	}
	requested := make(map[string]bool, len(packages))
	for _, name := range packages {
		requested[name] = true
	}
	matched := make(map[string]bool)
	seen := make(map[string]bool)
	var files []string
	for _, root := range roots {
		err := modules.WalkProtoFiles(root.Path, func(path string) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("Err: %w", err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("ReadFile: %w", err)
			}
			name := sourcePackage(raw)
			if !requested[name] {
				return nil
			}
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			relative = filepath.ToSlash(relative)
			matched[name] = true
			if !seen[relative] {
				seen[relative] = true
				files = append(files, relative)
			}
			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	slices.Sort(files)
	return files, matched, nil
}

// sourcePackage lexes only top-level package declarations. Scanner tokens keep
// strings and comments opaque, so an option containing "package" cannot select
// a file. It intentionally does not validate unrelated message definitions.
func sourcePackage(raw []byte) string {
	var lex scanner.Scanner
	lex.Init(bytes.NewReader(raw))
	lex.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanChars | scanner.ScanComments | scanner.SkipComments
	lex.Error = func(*scanner.Scanner, string) {}
	depth := 0
	start := true
	for token := lex.Scan(); token != scanner.EOF; token = lex.Scan() {
		if depth == 0 && start && token == scanner.Ident && lex.TokenText() == "package" {
			var parts []string
			for {
				if lex.Scan() != scanner.Ident {
					return ""
				}
				parts = append(parts, lex.TokenText())
				switch lex.Scan() {
				case ';':
					return strings.Join(parts, ".")
				case '.':
					continue
				default:
					return ""
				}
			}
		}
		switch token {
		case '{', '(', '[':
			depth++
			start = false
		case '}', ')', ']':
			if depth > 0 {
				depth--
			}
			start = depth == 0
		case ';':
			if depth == 0 {
				start = true
			}
		default:
			if depth == 0 {
				start = false
			}
		}
	}
	return ""
}
