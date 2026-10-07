package modules

import (
	"fmt"
	"strings"
)

type importRootLayout struct {
	roots [][]string
	// An intrinsic declaration can narrow its module only while its source
	// remains selected. It cannot justify silently dropping the declaring file.
	intrinsicOwners []*importRootFile
}

type importRootNamespace struct {
	files    map[string]*importRootFile
	selected map[*importRootFile]bool
	err      error
}

func (index *importRootIndex) namespace(layout importRootLayout) importRootNamespace {
	namespace := importRootNamespace{files: make(map[string]*importRootFile), selected: make(map[*importRootFile]bool)}
	identities := make(map[string]string)
	for _, file := range index.files {
		roots := index.effectiveRoots(layout, file.module)
		for _, root := range roots {
			name, within := importRootRelative(root, file.Path)
			if !within {
				continue
			}
			namespace.selected[file] = true
			if previous, exists := namespace.files[name]; exists && previous != file {
				namespace.err = fmt.Errorf("duplicate import path %q: %s and %s", name, previous.Path, file.Path)
			}
			if previous, exists := identities[file.Identity]; exists && previous != name && len(layout.roots[file.module]) > 0 && !index.modules[file.module].fixed {
				namespace.err = fmt.Errorf("physical source %s has overlapping import aliases %q and %q", file.Path, previous, name)
			}
			namespace.files[name] = file
			identities[file.Identity] = name
		}
	}
	for owner, module := range index.modules {
		if module.inspection == nil {
			continue
		}
		for _, problem := range module.inspection.Problems {
			for _, root := range index.effectiveRoots(layout, owner) {
				_, within := importRootRelative(root, problem.Path)
				if within || root == problem.Path || strings.HasPrefix(root, problem.Path+"/") {
					namespace.err = fmt.Errorf("module %s root %q: %s: %w", module.name, root, problem.Path, problem.Err)
				}
			}
		}
	}
	return namespace
}

func (index *importRootIndex) effectiveRoots(layout importRootLayout, owner int) []string {
	if len(layout.roots[owner]) > 0 {
		return layout.roots[owner]
	}
	return index.modules[owner].roots
}

type importRootMissing struct {
	owner     string
	path      string
	intrinsic *importRootFile
}

type importRootConstraints struct {
	missing  []importRootMissing
	bindings map[string]string
}

func (index *importRootIndex) constraints(namespace importRootNamespace) importRootConstraints {
	constraints := importRootConstraints{bindings: make(map[string]string)}
	for _, file := range index.files {
		for _, name := range file.intrinsic {
			// Only an actual import and a matching file within this same
			// module provide evidence. Unknown imports remain subject to
			// existing strict validation when the file is reached.
			if !ValidProtoImportPath(name) {
				continue
			}
			selected := namespace.files[name]
			if selected != nil {
				constraints.bindings[name] = selected.Identity
				continue
			}
			if len(index.candidates[name]) > 0 {
				constraints.missing = append(constraints.missing, importRootMissing{owner: file.Path, path: name, intrinsic: file})
			}
		}
	}
	return constraints
}
