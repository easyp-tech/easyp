package rules

import (
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.ModuleRule = (*PackageNoImportCycle)(nil)

// PackageNoImportCycle this rule checks that packages do not import each other in a cycle.
// The package graph covers linted files and the packages of their direct
// imports, so a cycle that passes only through dependencies is not seen.
type PackageNoImportCycle struct {
	// package -> imported package -> import statements of linted files
	edges map[string]map[string][]packageImport
}

type packageImport struct {
	path, location string
	pos            core.Position
}

// Message implements lint.Rule.
func (p *PackageNoImportCycle) Message() string {
	return "package import cycle"
}

// Validate implements lint.Rule. Lint calls Collect and Finish instead.
func (p *PackageNoImportCycle) Validate(core.ProtoInfo) ([]core.Issue, error) { return nil, nil }

// Collect implements core.ModuleRule.
func (p *PackageNoImportCycle) Collect(protoInfo core.ProtoInfo) error {
	if p.edges == nil {
		p.edges = make(map[string]map[string][]packageImport)
	}
	pkg := protoInfo.File.PackageName()
	for _, imp := range protoInfo.File.Imports {
		imported, ok := protoInfo.ImportedFiles[core.ConvertImportPath(imp.Location)]
		if !ok || imported.PackageName() == pkg {
			continue
		}
		if p.edges[pkg] == nil {
			p.edges[pkg] = make(map[string][]packageImport)
		}
		target := imported.PackageName()
		p.edges[pkg][target] = append(p.edges[pkg][target], packageImport{path: protoInfo.Path, location: imp.Location, pos: imp.Pos})
	}
	return nil
}

// Finish implements core.ModuleRule. Like buf, it reports every import of a
// package that leads into a cycle back to that package.
func (p *PackageNoImportCycle) Finish() []core.IssueInfo {
	var res []core.IssueInfo
	for _, pkg := range sortedKeys(p.edges) {
		if pkg == "" {
			continue
		}
		for _, imported := range sortedKeys(p.edges[pkg]) {
			if imported == "" {
				continue
			}
			cycle := p.cycle(imported, map[string]bool{pkg: true}, []string{pkg})
			if len(cycle) == 0 {
				continue
			}
			for _, imp := range p.edges[pkg][imported] {
				for _, issue := range core.AppendIssue(nil, p, imp.pos, strings.Join(cycle, " -> ")) {
					res = append(res, core.IssueInfo{Issue: issue, Path: imp.path})
				}
			}
		}
	}
	p.edges = nil
	return res
}

// cycle returns the package path back to path[0] through pkg, if any.
func (p *PackageNoImportCycle) cycle(pkg string, used map[string]bool, path []string) []string {
	path = append(path, pkg)
	if used[pkg] {
		if path[0] == path[len(path)-1] {
			return path
		}
		return nil
	}
	used[pkg] = true
	for _, next := range sortedKeys(p.edges[pkg]) {
		if next == "" {
			continue
		}
		if found := p.cycle(next, used, path); len(found) > 0 {
			return found
		}
	}
	delete(used, pkg)
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
