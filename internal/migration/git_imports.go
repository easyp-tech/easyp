package migration

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/modules"
)

// gitImportNamespaces translates already verified archive mappings and module
// roots without reading the filesystem or filtering the available namespace.
func gitImportNamespaces(proof *gitSelectionProof) (map[string]map[string]migrationProtoSource, map[string]map[string]migrationProtoSource, error) {
	legacyDependencies, currentDependencies := make(map[string]map[string]migrationProtoSource), make(map[string]map[string]migrationProtoSource)
	for _, name := range slices.Sorted(maps.Keys(proof.fetched)) {
		dependency := proof.fetched[name]
		if dependency.Inspection == nil || dependency.Inspection.LegacyFiles == nil {
			continue
		}
		files := make(map[string]modules.RootProtoFile)
		for _, file := range dependency.Inspection.Files {
			files[file.Path] = file
		}
		legacy := make(map[string]migrationProtoSource)
		for installed, logical := range dependency.Inspection.LegacyFiles {
			legacy[installed] = migrationProtoSource{module: name, path: logical, unavailable: true}
			if file, ok := files[logical]; ok {
				legacy[installed] = migrationGitSource(name, file)
			}
		}
		if selection, ok := proof.selections[name]; ok && selection.root != "." {
			for _, installed := range slices.Sorted(maps.Keys(dependency.Inspection.LegacyFiles)) {
				if !strings.HasPrefix(installed, selection.root+"/") {
					continue
				}
				alias := strings.TrimPrefix(installed, selection.root+"/")
				logical := dependency.Inspection.LegacyFiles[installed]
				legacy[alias] = migrationProtoSource{module: name, path: logical, unavailable: true}
				if file, ok := files[logical]; ok {
					legacy[alias] = migrationGitSource(name, file)
				}
			}
		}
		legacyDependencies[name] = legacy
		current, err := migrationRootNamespace(dependency)
		if err != nil {
			return nil, nil, fmt.Errorf("migrationRootNamespace: %w", err)
		}
		currentDependencies[name] = make(map[string]migrationProtoSource)
		for filename, file := range current {
			currentDependencies[name][filename] = migrationGitSource(name, file)
		}
	}
	return legacyDependencies, currentDependencies, nil
}

func legacyGitSelectionTargets(selection gitModuleSelection, dependency modules.Fetched) []string {
	var targets []string
	for installed := range dependency.Inspection.LegacyFiles {
		if !slices.ContainsFunc(selection.subdirs, func(subdir string) bool {
			return subdir == "." || installed == subdir || strings.HasPrefix(installed, subdir+"/")
		}) {
			continue
		}
		name := installed
		if selection.root != "." {
			name = strings.TrimPrefix(installed, selection.root+"/")
		}
		targets = append(targets, name)
	}
	return targets
}

func migrationGitSource(module string, file modules.RootProtoFile) migrationProtoSource {
	return migrationProtoSource{module: module, path: file.Path, identity: file.Identity, content: file.Content}
}
