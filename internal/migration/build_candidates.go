package migration

import (
	"bytes"
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type migrationCandidates struct {
	module         v1.Module
	deferGenerator bool
}

// preflightCandidates rejects known metadata/output conflicts before any
// explicitly permitted dependency work. Only a generator that needs verified
// archive-path translation can defer its candidate comparison.
func (t *transaction) preflightCandidates(cfg legacyConfig, local *localSelectionProof, selection migrationSelections, resolveLock bool) (migrationCandidates, error) {
	deps := selection.deps
	manifest := t.expected[v1.ModuleFile]
	nativeManifest := manifest.exists && v1.IsModuleManifest(manifest.data)
	var replacements []v1.Replacement
	if manifest.exists && !nativeManifest {
		var err error
		replacements, err = parseManifest(manifest.data, &deps)
		if err != nil {
			return migrationCandidates{}, fmt.Errorf("parseManifest: %w", err)
		}
	}
	moduleRoots := slices.Clone(local.roots)
	if len(moduleRoots) == 0 {
		moduleRoots = []string{"."}
	}
	module := v1.Module{Name: local.module, Roots: moduleRoots, Requires: deps.items, Replaces: replacements}
	for _, req := range module.Requires {
		if req.Module == module.Name {
			return migrationCandidates{}, fmt.Errorf("dependency %s has the new local module identity; reconcile this ambiguity manually", req.Module)
		}
	}
	moduleBytes := formatManifest(module, deps.indirect)
	if _, err := v1.ParseModule(bytes.NewReader(moduleBytes)); err != nil {
		return migrationCandidates{}, fmt.Errorf("ParseModule: %w", err)
	}
	if nativeManifest && !bytes.Equal(manifest.data, moduleBytes) {
		return migrationCandidates{}, fmt.Errorf("existing native protobuf.mod conflicts with the migration candidate; reconcile it manually (it will not be overwritten)")
	}
	policyBytes, err := convertPolicy(cfg)
	if err != nil {
		return migrationCandidates{}, fmt.Errorf("convertPolicy: %w", err)
	}
	generateBytes, err := convertGenerate(cfg, selection.entries, selection.packages, selection.paths)
	if err != nil {
		return migrationCandidates{}, fmt.Errorf("convertGenerate: %w", err)
	}
	if err := validatePolicy(policyBytes); err != nil {
		return migrationCandidates{}, fmt.Errorf("validatePolicy: %w", err)
	}
	if err := validateGenerate(generateBytes); err != nil {
		return migrationCandidates{}, fmt.Errorf("validateGenerate: %w", err)
	}
	deferGenerator := resolveLock && t.expected[v1.GenerateFile].exists && slices.ContainsFunc(selection.entries, func(entry v1.GenerateModule) bool {
		_, git := selection.git[entry.Module]
		return git && len(entry.Paths) > 0
	})
	if deferGenerator {
		if err := validateGenerate(t.expected[v1.GenerateFile].data); err != nil {
			return migrationCandidates{}, fmt.Errorf("validateGenerate: %w", err)
		}
	}
	if err := t.addOutput(v1.PolicyFile, policyBytes, true); err != nil {
		return migrationCandidates{}, fmt.Errorf("addOutput: %w", err)
	}
	if !deferGenerator {
		if err := t.addOutput(v1.GenerateFile, generateBytes, false); err != nil {
			return migrationCandidates{}, fmt.Errorf("addOutput: %w", err)
		}
	}
	if err := t.addOutput(v1.ModuleFile, moduleBytes, manifest.exists && !nativeManifest); err != nil {
		return migrationCandidates{}, fmt.Errorf("addOutput: %w", err)
	}
	return migrationCandidates{module: module, deferGenerator: deferGenerator}, nil
}

func (p *Plan) finalizeOutputs(cfg legacyConfig, selection migrationSelections, candidates migrationCandidates, lockInputs migrationLockInputs, lock *v1.Lock) error {
	if lock != nil {
		for index, entry := range p.git.entries {
			if !slices.Equal(entry.Paths, selection.entries[index].Paths) {
				p.warnings = append(p.warnings, fmt.Sprintf("Dependency %s: verified legacy sub_directory %v translates to module-relative generate.modules paths %v.", entry.Module, selection.entries[index].Paths, entry.Paths))
			}
		}
		generateBytes, err := convertGenerate(cfg, p.git.entries, selection.packages, selection.paths)
		if err != nil {
			return fmt.Errorf("convertGenerate: %w", err)
		}
		if err := validateGenerate(generateBytes); err != nil {
			return fmt.Errorf("validateGenerate: %w", err)
		}
		if candidates.deferGenerator {
			if err := p.tx.addOutput(v1.GenerateFile, generateBytes, false); err != nil {
				return fmt.Errorf("addOutput: %w", err)
			}
		} else if err := p.tx.replaceCandidate(v1.GenerateFile, generateBytes); err != nil {
			return fmt.Errorf("replaceCandidate: %w", err)
		}
		content, err := yaml.Marshal(lock)
		if err != nil {
			return fmt.Errorf("Marshal: %w", err)
		}
		if err := p.tx.addOutput(v1.LockFile, content, false); err != nil {
			return fmt.Errorf("addOutput: %w", err)
		}
		return nil
	}
	if lockInputs.needsLock {
		return nil
	}
	if nativeLock := p.tx.expected[v1.LockFile]; nativeLock.exists {
		lock, err := validateNativeLock(nativeLock.data)
		if err != nil {
			return fmt.Errorf("validateNativeLock: %w", err)
		}
		if len(lock.Modules) != 0 {
			return fmt.Errorf("existing native protobuf.lock contains dependencies absent from the migration candidate; it will not be overwritten")
		}
		return nil
	}
	content, err := yaml.Marshal(v1.Lock{Version: 1, Modules: []v1.LockedModule{}})
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}
	if err := p.tx.addOutput(v1.LockFile, content, false); err != nil {
		return fmt.Errorf("addOutput: %w", err)
	}
	return nil
}

func (t *transaction) outputs() []Output {
	result := slices.Clone(t.candidates)
	for i := range result {
		result[i].Content = bytes.Clone(result[i].Content)
	}
	return result
}

func (t *transaction) replaceCandidate(name string, content []byte) error {
	current := t.expected[name]
	if current.exists && !bytes.Equal(current.data, content) {
		return fmt.Errorf("existing %s conflicts with the verified migration candidate; it will not be overwritten", name)
	}
	for _, candidate := range t.candidates {
		if candidate.Name == name {
			t.proposeCandidate(name, content, candidate.Mode)
			return nil
		}
	}
	return fmt.Errorf("candidate %q was not prepared during planning", name)
}

func (t *transaction) addOutput(name string, content []byte, replaceLegacy bool) error {
	current, err := t.capture(name)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	if len(current.links) > 0 && !replaceLegacy && !bytes.Equal(current.data, content) {
		return fmt.Errorf("destination %q must be a regular file", name)
	}
	mode := os.FileMode(0o644)
	if current.exists {
		mode = current.mode
	}
	unchanged := current.exists && bytes.Equal(current.data, content)
	if current.exists && !unchanged && !replaceLegacy {
		return fmt.Errorf("existing %s conflicts with the migration candidate; it will not be overwritten", name)
	}
	if replaceLegacy && !unchanged {
		if err := t.addBackupOutput(name+".v0.bak", current); err != nil {
			return fmt.Errorf("addBackupOutput: %w", err)
		}
	}
	t.proposeCandidate(name, content, mode)
	return nil
}

func (t *transaction) addBackupOutput(name string, original snapshot) error {
	backup, err := t.capture(name)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	if backup.exists && !bytes.Equal(backup.data, original.data) {
		return fmt.Errorf("existing %s conflicts with the migration candidate; it will not be overwritten", name)
	}
	// A pre-existing backup must match both bytes and permissions exactly.
	if backup.exists && backup.mode != original.mode {
		return fmt.Errorf("existing %s has different permissions; it will not be overwritten", name)
	}
	t.proposeCandidate(name, original.data, original.mode)
	return nil
}

// proposeCandidate owns both representations of a candidate. Preview buffers
// and staged buffers are independent, and replacement retains their ordering.
func (t *transaction) proposeCandidate(name string, content []byte, mode os.FileMode) {
	current := t.expected[name]
	unchanged := current.exists && bytes.Equal(current.data, content)
	output := Output{Name: name, Content: bytes.Clone(content), Mode: mode, Unchanged: unchanged}
	index := slices.IndexFunc(t.candidates, func(candidate Output) bool { return candidate.Name == name })
	if index < 0 {
		t.candidates = append(t.candidates, output)
	} else {
		t.candidates[index] = output
	}
	index = slices.IndexFunc(t.changes, func(change fileChange) bool { return change.name == name })
	if unchanged {
		if index >= 0 {
			t.changes = slices.Delete(t.changes, index, index+1)
		}
		return
	}
	change := fileChange{name: name, content: bytes.Clone(content), mode: mode}
	if index < 0 {
		t.changes = append(t.changes, change)
	} else {
		t.changes[index] = change
	}
}
