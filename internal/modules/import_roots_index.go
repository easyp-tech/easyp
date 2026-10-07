package modules

import (
	"fmt"
	"slices"
	"strings"
)

const maxImportRootCandidates = 65536

type importRootModule struct {
	name       string
	roots      []string
	fixed      bool
	inspection *RootInspection
}

type importRootFile struct {
	RootProtoFile
	module    int
	intrinsic []string
}

type importRootCandidate struct {
	module int
	root   string
}

type importRootIndex struct {
	modules    []importRootModule
	files      []*importRootFile
	candidates map[string][]importRootCandidate
}

func newImportRootIndex(modules []importRootModule) (*importRootIndex, error) {
	index := &importRootIndex{modules: modules, candidates: make(map[string][]importRootCandidate)}
	count := 0
	for owner, module := range modules {
		if module.inspection == nil {
			continue
		}
		for _, source := range module.inspection.Files {
			if !ValidProtoImportPath(source.Path) || source.Identity == "" {
				return nil, fmt.Errorf("module %s has invalid inspected source %q", module.name, source.Path)
			}
			file := &importRootFile{RootProtoFile: source, module: owner, intrinsic: intrinsicProtoImports(source.Path, source.Content)}
			index.files = append(index.files, file)
			if module.fixed {
				continue
			}
			components := strings.Split(source.Path, "/")
			for offset := range components {
				name := strings.Join(components[offset:], "/")
				root := "."
				if offset > 0 {
					root = strings.Join(components[:offset], "/")
				}
				candidate := importRootCandidate{module: owner, root: root}
				if slices.Contains(index.candidates[name], candidate) {
					continue
				}
				count++
				if count > maxImportRootCandidates {
					return nil, fmt.Errorf("import root candidate limit exceeded; supply explicit import roots")
				}
				index.candidates[name] = append(index.candidates[name], candidate)
			}
		}
	}
	slices.SortFunc(index.files, func(a, b *importRootFile) int {
		if order := strings.Compare(modules[a.module].name, modules[b.module].name); order != 0 {
			return order
		}
		return strings.Compare(a.Path, b.Path)
	})
	return index, nil
}

func importRootRelative(root, name string) (string, bool) {
	if root == "." {
		return name, true
	}
	return strings.TrimPrefix(name, root+"/"), strings.HasPrefix(name, root+"/")
}
