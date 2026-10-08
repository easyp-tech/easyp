// Package migration plans explicit, conservative EasyP v0 to v1 migrations.
// Build is read-only unless ResolveLock explicitly permits repository/cache
// access. It never invokes generation plugins or expands environment variables.
// Apply stages all outputs and byte-identical legacy backups before replacing
// anything, rechecks observed files, and rolls back ordinary application errors.
//
// Multiple filesystem renames are not a crash-atomic transaction. Process or
// machine failure during Apply may leave a partially migrated set; the legacy
// .v0.bak files and immutable easyp.lock are the recovery inputs. A failed
// rollback retains its private staging directory and reports its location.
// Noncooperating writers cannot be excluded between the final state check and
// rename; callers should stop other writers while applying a migration.
package migration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Options identifies the legacy project and explicitly authorizes optional
// dependency resolution. Repository is unused unless ResolveLock is true.
type Options struct {
	Dir         string
	Module      string
	ResolveLock bool
	Repository  Repository
}

// Output is one complete candidate file, including historical backups.
type Output struct {
	Name      string
	Content   []byte
	Mode      os.FileMode
	Unchanged bool
}

// Plan owns immutable candidate contents and the observed source state.
// Use Outputs to preview, then Apply to opt into writing that exact plan.
// A Plan is not safe for concurrent use.
type Plan struct {
	tx        *transaction
	outputs   []Output
	warnings  []string
	blocked   string
	alreadyV1 bool
	inputs    []legacyDirectory
	roots     []string
	packages  []string
	paths     []string
	sources   map[string]string
	git       *gitSelectionProof
}

// Outputs returns copies of all candidate files; mutating them cannot alter Apply.
func (p *Plan) Outputs() []Output {
	result := slices.Clone(p.outputs)
	for i := range result {
		result[i].Content = bytes.Clone(result[i].Content)
	}
	return result
}

// Warnings returns migration caveats, including prerequisites for application.
func (p *Plan) Warnings() []string { return slices.Clone(p.warnings) }

// AlreadyV1 reports a validated native project for which migration is a no-op.
func (p *Plan) AlreadyV1() bool { return p.alreadyV1 }

// NeedsLockResolution reports whether dependency verification must be explicitly
// authorized before this preview can be applied. It never performs resolution.
func (p *Plan) NeedsLockResolution() bool { return p.blocked != "" }

// ValidateModuleIdentity checks a proposed local identity without reading or
// writing project files. Interactive callers can correct input before planning.
func ValidateModuleIdentity(identity string) error { return validIdentity(identity) }

// CheckUnchanged verifies that a preview still describes the observed project.
// It performs no writes or dependency access and does not replace Apply's checks.
func (p *Plan) CheckUnchanged() (resultErr error) {
	if err := p.verifySourceSelection(); err != nil {
		return fmt.Errorf("verifySourceSelection: %w", err)
	}
	root, err := p.tx.openRoot()
	if err != nil {
		return fmt.Errorf("openRoot: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, closeRoot(root)) }()
	if err := p.tx.verify(root); err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	return nil
}

// Apply writes the prepared candidates after checking all application gates and
// source state again. It never resolves or refreshes dependencies.
func (p *Plan) Apply() error {
	if p.blocked != "" {
		return fmt.Errorf("%s", p.blocked)
	}
	if err := p.verifySourceSelection(); err != nil {
		return fmt.Errorf("verifySourceSelection: %w", err)
	}
	if err := p.tx.apply(); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	return nil
}

func (p *Plan) verifySourceSelection() error {
	if len(p.inputs) > 0 {
		selection, err := proveLocalSelection(p.tx.requestedRoot, p.inputs, p.roots)
		if err != nil {
			return fmt.Errorf("proveLocalSelection: %w", err)
		}
		if !maps.Equal(selection.files, p.sources) || !slices.Equal(selection.packages, p.packages) || !slices.Equal(selection.paths, p.paths) {
			return fmt.Errorf("local .proto source selection changed since planning; preview again")
		}
	}
	if p.git != nil {
		bindings, err := p.proveGitSourceBindings(context.Background(), false)
		if err != nil {
			return fmt.Errorf("proveGitSourceBindings: %w", err)
		}
		if !equalMigrationBindings(bindings, p.git.bindings) {
			return fmt.Errorf("selected/reachable protobuf source bindings changed since planning; preview again")
		}
	}
	return nil
}

// Build prepares policy, generation, manifest and (when authorized) lock files.
// Unknown or unrepresentable legacy behavior is an error, never a lossy guess.
func Build(ctx context.Context, options Options) (*Plan, error) {
	if err := validIdentity(options.Module); err != nil {
		return nil, fmt.Errorf("validIdentity: %w", err)
	}
	if options.Dir == "" {
		options.Dir = "."
	}
	tx, err := newTransaction(options.Dir)
	if err != nil {
		return nil, fmt.Errorf("newTransaction: %w", err)
	}
	p := &Plan{tx: tx}
	for _, name := range []string{v1.PolicyFile, v1.GenerateFile, v1.ModuleFile, v1.LockFile, "easyp.lock", "easyp.yaml.v0.bak", "protobuf.mod.v0.bak"} {
		capture := tx.captureInput
		if strings.HasSuffix(name, ".v0.bak") {
			capture = tx.capture
		}
		if _, err := capture(name); err != nil {
			return nil, fmt.Errorf("capture: %w", err)
		}
	}
	policy := tx.expected[v1.PolicyFile]
	if !policy.exists {
		return nil, fmt.Errorf("easyp.yaml is missing from %s", options.Dir)
	}
	node, err := document(policy.data)
	if err != nil {
		return nil, fmt.Errorf("document: %w", err)
	}
	if nativePolicy(node) {
		if err := p.checkNative(options.Module); err != nil {
			return nil, fmt.Errorf("checkNative: %w", err)
		}
		p.alreadyV1 = true
		p.warnings = append(p.warnings, "Already v1: no files changed and no dependencies refreshed.")
		return p, nil
	}
	cfg, legacyWarnings, err := parseLegacy(policy.data)
	if err != nil {
		return nil, fmt.Errorf("parseLegacy: %w", err)
	}
	p.warnings = append(p.warnings, legacyWarnings...)
	roots, inputs, err := localRoots(cfg)
	if err != nil {
		return nil, fmt.Errorf("localRoots: %w", err)
	}
	p.roots, p.inputs = roots, inputs
	p.tx.beforeApply = p.verifySourceSelection
	if err := checkManagedLocal(cfg, options.Module, len(inputs) > 0); err != nil {
		return nil, fmt.Errorf("checkManagedLocal: %w", err)
	}
	selection, err := proveLocalSelection(tx.requestedRoot, inputs, roots)
	if err != nil {
		return nil, fmt.Errorf("proveLocalSelection: %w", err)
	}
	p.sources, p.packages, p.paths = selection.files, selection.packages, selection.paths
	if len(p.packages) > 0 || len(p.paths) > 0 {
		if len(p.paths) > 0 {
			p.warnings = append(p.warnings, "Legacy directory selection is preserved through literal paths selectors for the local module. Import roots and source paths stay unchanged; files outside these paths do not become targets even when they declare the same package.")
		} else {
			p.warnings = append(p.warnings, "Legacy directory selection is preserved through exact packages selectors for the local module. Import roots and source paths stay unchanged; future files declaring those packages also participate in generation.")
		}
	}
	for _, source := range p.sources {
		if _, err := tx.captureInput(source); err != nil {
			return nil, fmt.Errorf("capture: %w", err)
		}
	}
	deps := requirements{}
	for _, dep := range cfg.Deps {
		if err := deps.add(dep, false); err != nil {
			return nil, fmt.Errorf("add: %w", err)
		}
	}
	selected, gitSelections, err := planGitSelections(cfg, options.Module, len(inputs) > 0, p.packages, p.paths, &deps)
	if err != nil {
		return nil, fmt.Errorf("planGitSelections: %w", err)
	}
	globalPackages, globalPaths := p.packages, p.paths
	if len(gitSelections) > 0 {
		globalPackages, globalPaths = nil, nil
	}
	manifest := tx.expected[v1.ModuleFile]
	nativeManifest := manifest.exists && v1.IsModuleManifest(manifest.data)
	var replacements []v1.Replacement
	if manifest.exists && !nativeManifest {
		replacements, err = parseManifest(manifest.data, &deps)
		if err != nil {
			return nil, fmt.Errorf("parseManifest: %w", err)
		}
	}
	moduleRoots := slices.Clone(roots)
	if len(moduleRoots) == 0 {
		moduleRoots = []string{"."}
	}
	module := v1.Module{Name: options.Module, Roots: moduleRoots, Requires: deps.items, Replaces: replacements}
	for _, req := range module.Requires {
		if req.Module == module.Name {
			return nil, fmt.Errorf("dependency %s has the new local module identity; reconcile this ambiguity manually", req.Module)
		}
	}
	moduleBytes := formatManifest(module, deps.indirect)
	if _, err := v1.ParseModule(bytes.NewReader(moduleBytes)); err != nil {
		return nil, fmt.Errorf("ParseModule: %w", err)
	}
	if nativeManifest && !bytes.Equal(manifest.data, moduleBytes) {
		return nil, fmt.Errorf("existing native protobuf.mod conflicts with the migration candidate; reconcile it manually (it will not be overwritten)")
	}
	policyBytes, err := convertPolicy(cfg)
	if err != nil {
		return nil, fmt.Errorf("convertPolicy: %w", err)
	}
	generateBytes, err := convertGenerate(cfg, selected, globalPackages, globalPaths)
	if err != nil {
		return nil, fmt.Errorf("convertGenerate: %w", err)
	}
	if err := validatePolicy(policyBytes); err != nil {
		return nil, fmt.Errorf("validatePolicy: %w", err)
	}
	if err := validateGenerate(generateBytes); err != nil {
		return nil, fmt.Errorf("validateGenerate: %w", err)
	}
	// A Git subdirectory may name a producer-renamed installed path. Compare an
	// existing generator with the final translated candidate after verification.
	deferGenerator := options.ResolveLock && tx.expected[v1.GenerateFile].exists && slices.ContainsFunc(selected, func(entry v1.GenerateModule) bool {
		_, git := gitSelections[entry.Module]
		return git && len(entry.Paths) > 0
	})
	if deferGenerator {
		if err := validateGenerate(tx.expected[v1.GenerateFile].data); err != nil {
			return nil, fmt.Errorf("validateGenerate: %w", err)
		}
	}
	// Preflight known local candidates before explicitly allowed cache work.
	if err := p.addOutput(v1.PolicyFile, policyBytes, true); err != nil {
		return nil, fmt.Errorf("addOutput: %w", err)
	}
	if !deferGenerator {
		if err := p.addOutput(v1.GenerateFile, generateBytes, false); err != nil {
			return nil, fmt.Errorf("addOutput: %w", err)
		}
	}
	if err := p.addOutput(v1.ModuleFile, moduleBytes, manifest.exists && !nativeManifest); err != nil {
		return nil, fmt.Errorf("addOutput: %w", err)
	}
	oldLock := tx.expected["easyp.lock"]
	pins := make(map[string]legacyPin)
	if oldLock.exists {
		pins, err = parseLegacyLock(oldLock.data)
		if err != nil {
			return nil, fmt.Errorf("parseLegacyLock: %w", err)
		}
		for _, req := range module.Requires {
			if _, exists := pins[req.Module]; !exists {
				return nil, fmt.Errorf("legacy easyp.lock is missing required dependency %s; recover its historical pin before migration", req.Module)
			}
		}
		p.warnings = append(p.warnings, "easyp.lock is retained byte-for-byte as a historical backup; it is never rewritten or removed.")
	}
	if nativeLock := tx.expected[v1.LockFile]; nativeLock.exists {
		if _, err := validateNativeLock(nativeLock.data); err != nil {
			return nil, fmt.Errorf("validateNativeLock: %w", err)
		}
	}
	needsLock := len(module.Requires) > 0 || len(pins) > 0 || len(module.Replaces) > 0
	if needsLock && !options.ResolveLock {
		p.blocked = "dependency integrity/lock verification is required before writing; preview again with --resolve-lock (this explicitly permits repository/cache access)"
		p.warnings = append(p.warnings, "protobuf.lock cannot be prepared without --resolve-lock; --write is blocked until every required pin and legacy hash is verified.")
	} else if needsLock {
		lock, fetched, err := migrateSelectionLock(ctx, module, pins, oldLock.exists, options.Repository, gitSelections)
		if err != nil {
			return nil, fmt.Errorf("migrateSelectionLock: %w", err)
		}
		p.git, err = proveGitSelections(gitSelections, fetched, selected, options.Module)
		if err != nil {
			return nil, fmt.Errorf("proveGitSelections: %w", err)
		}
		p.git.bindings, err = p.proveGitSourceBindings(ctx, true)
		if err != nil {
			return nil, fmt.Errorf("proveGitSourceBindings: %w", err)
		}
		for _, file := range p.git.bindings {
			if file.module == options.Module {
				if _, err := tx.captureInput(file.path); err != nil {
					return nil, fmt.Errorf("captureInput: %w", err)
				}
			}
		}
		for index, entry := range p.git.entries {
			if !slices.Equal(entry.Paths, selected[index].Paths) {
				p.warnings = append(p.warnings, fmt.Sprintf("Dependency %s: verified legacy sub_directory %v translates to module-relative generate.modules paths %v.", entry.Module, selected[index].Paths, entry.Paths))
			}
		}
		generateBytes, err = convertGenerate(cfg, p.git.entries, globalPackages, globalPaths)
		if err != nil {
			return nil, fmt.Errorf("convertGenerate: %w", err)
		}
		if err := validateGenerate(generateBytes); err != nil {
			return nil, fmt.Errorf("validateGenerate: %w", err)
		}
		if deferGenerator {
			if err := p.addOutput(v1.GenerateFile, generateBytes, false); err != nil {
				return nil, fmt.Errorf("addOutput: %w", err)
			}
		} else if err := p.replaceCandidate(v1.GenerateFile, generateBytes); err != nil {
			return nil, fmt.Errorf("replaceCandidate: %w", err)
		}
		content, err := yaml.Marshal(lock)
		if err != nil {
			return nil, fmt.Errorf("Marshal: %w", err)
		}
		if err := p.addOutput(v1.LockFile, content, false); err != nil {
			return nil, fmt.Errorf("addOutput: %w", err)
		}
	} else if nativeLock := tx.expected[v1.LockFile]; nativeLock.exists {
		lock, err := validateNativeLock(nativeLock.data)
		if err != nil {
			return nil, fmt.Errorf("validateNativeLock: %w", err)
		}
		if len(lock.Modules) != 0 {
			return nil, fmt.Errorf("existing native protobuf.lock contains dependencies absent from the migration candidate; it will not be overwritten")
		}
	}
	if !needsLock && !tx.expected[v1.LockFile].exists {
		content, err := yaml.Marshal(v1.Lock{Version: 1, Modules: []v1.LockedModule{}})
		if err != nil {
			return nil, fmt.Errorf("Marshal: %w", err)
		}
		if err := p.addOutput(v1.LockFile, content, false); err != nil {
			return nil, fmt.Errorf("addOutput: %w", err)
		}
	}
	p.warnings = append(p.warnings, "Relative paths and variable placeholders are preserved; no plugin is executed. Stop other writers before applying; multiple file replacements are not process-crash atomic.")
	return p, nil
}

func (p *Plan) replaceCandidate(name string, content []byte) error {
	current := p.tx.expected[name]
	if current.exists && !bytes.Equal(current.data, content) {
		return fmt.Errorf("existing %s conflicts with the verified migration candidate; it will not be overwritten", name)
	}
	for index := range p.outputs {
		if p.outputs[index].Name == name {
			p.outputs[index].Content = bytes.Clone(content)
		}
	}
	for index := range p.tx.changes {
		if p.tx.changes[index].name == name {
			p.tx.changes[index].content = bytes.Clone(content)
		}
	}
	return nil
}

func (p *Plan) addOutput(name string, content []byte, replaceLegacy bool) error {
	current, err := p.tx.capture(name)
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
		if err := p.addOutput(name+".v0.bak", current.data, false); err != nil {
			return fmt.Errorf("addOutput: %w", err)
		}
		backup := &p.outputs[len(p.outputs)-1]
		// A pre-existing backup must match both bytes and permissions exactly.
		if backup.Unchanged && backup.Mode != mode {
			return fmt.Errorf("existing %s.v0.bak has different permissions; it will not be overwritten", name)
		}
		backup.Mode = mode
		if !backup.Unchanged {
			p.tx.changes[len(p.tx.changes)-1].mode = mode
		}
	}
	p.outputs = append(p.outputs, Output{Name: name, Content: bytes.Clone(content), Mode: mode, Unchanged: unchanged})
	if !unchanged {
		p.tx.changes = append(p.tx.changes, fileChange{name: name, content: bytes.Clone(content), mode: mode})
	}
	return nil
}

func nativePolicy(node *yaml.Node) bool {
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if key == "linters" || key == "linters-settings" || key == "issues" || (key == "version" && node.Content[i+1].Value == "v1") {
			return true
		}
	}
	return false
}

func (p *Plan) checkNative(identity string) error {
	if err := validatePolicy(p.tx.expected[v1.PolicyFile].data); err != nil {
		return fmt.Errorf("validatePolicy: %w", err)
	}
	manifest := p.tx.expected[v1.ModuleFile]
	if !manifest.exists {
		return fmt.Errorf("already-v1 policy requires a native protobuf.mod; reconcile this partial migration manually")
	}
	module, err := v1.ParseModule(bytes.NewReader(manifest.data))
	if err != nil {
		return fmt.Errorf("ParseModule: %w", err)
	}
	if module.Name != identity {
		return fmt.Errorf("native protobuf.mod module is %s, not the requested identity %s", module.Name, identity)
	}
	gen := p.tx.expected[v1.GenerateFile]
	if !gen.exists {
		return fmt.Errorf("already-v1 policy requires easyp.gen.yaml; reconcile this partial migration manually")
	}
	if err := validateGenerate(gen.data); err != nil {
		return fmt.Errorf("validateGenerate: %w", err)
	}
	if lock := p.tx.expected[v1.LockFile]; lock.exists {
		parsed, err := validateNativeLock(lock.data)
		if err != nil {
			return fmt.Errorf("validateNativeLock: %w", err)
		}
		if err := modules.ValidateRequirements(modules.RemoteRequirements(module), parsed); err != nil {
			return fmt.Errorf("ValidateRequirements: %w", err)
		}
	} else if len(modules.RemoteRequirements(module)) > 0 {
		return fmt.Errorf("native protobuf.lock is missing; repair this partial migration explicitly before proceeding")
	}
	return nil
}
