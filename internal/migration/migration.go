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
	warnings  []string
	blocked   string
	alreadyV1 bool
	local     *localSelectionProof
	git       *gitSelectionProof
}

// Outputs returns copies of all candidate files; mutating them cannot alter Apply.
func (p *Plan) Outputs() []Output { return p.tx.outputs() }

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
	if err := p.local.recheck(p.tx.requestedRoot); err != nil {
		return fmt.Errorf("recheck: %w", err)
	}
	if p.git != nil {
		bindings, err := proveGitSourceBindings(context.Background(), p.tx.requestedRoot, p.local, p.git, false)
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
	p, cfg, err := captureMigrationProject(options)
	if err != nil {
		return nil, fmt.Errorf("captureMigrationProject: %w", err)
	}
	if p.alreadyV1 {
		return p, nil
	}
	selection, err := p.selectSources(cfg, options.Module)
	if err != nil {
		return nil, fmt.Errorf("selectSources: %w", err)
	}
	candidates, err := p.tx.preflightCandidates(cfg, p.local, selection, options.ResolveLock)
	if err != nil {
		return nil, fmt.Errorf("preflightCandidates: %w", err)
	}
	lockInputs, err := p.tx.lockPrerequisites(candidates.module)
	if err != nil {
		return nil, fmt.Errorf("lockPrerequisites: %w", err)
	}
	if lockInputs.historical {
		p.warnings = append(p.warnings, "easyp.lock is retained byte-for-byte as a historical backup; it is never rewritten or removed.")
	}
	var lock *v1.Lock
	if lockInputs.needsLock && !options.ResolveLock {
		p.blocked = "dependency integrity/lock verification is required before writing; preview again with --resolve-lock (this explicitly permits repository/cache access)"
		p.warnings = append(p.warnings, "protobuf.lock cannot be prepared without --resolve-lock; --write is blocked until every required pin and legacy hash is verified.")
	} else if lockInputs.needsLock {
		verified, fetched, err := migrateSelectionLock(ctx, candidates.module, lockInputs.pins, lockInputs.historical, options.Repository, selection.git)
		if err != nil {
			return nil, fmt.Errorf("migrateSelectionLock: %w", err)
		}
		if err := p.proveGitSources(ctx, selection, fetched); err != nil {
			return nil, fmt.Errorf("proveGitSources: %w", err)
		}
		lock = &verified
	}
	if err := p.finalizeOutputs(cfg, selection, candidates, lockInputs, lock); err != nil {
		return nil, fmt.Errorf("finalizeOutputs: %w", err)
	}
	p.warnings = append(p.warnings, "Relative paths and variable placeholders are preserved; no plugin is executed. Stop other writers before applying; multiple file replacements are not process-crash atomic.")
	return p, nil
}

func captureMigrationProject(options Options) (*Plan, legacyConfig, error) {
	tx, err := newTransaction(options.Dir)
	if err != nil {
		return nil, legacyConfig{}, fmt.Errorf("newTransaction: %w", err)
	}
	p := &Plan{tx: tx}
	for _, name := range []string{v1.PolicyFile, v1.GenerateFile, v1.ModuleFile, v1.LockFile, "easyp.lock", "easyp.yaml.v0.bak", "protobuf.mod.v0.bak"} {
		capture := tx.captureInput
		if strings.HasSuffix(name, ".v0.bak") {
			capture = tx.capture
		}
		if _, err := capture(name); err != nil {
			return nil, legacyConfig{}, fmt.Errorf("capture: %w", err)
		}
	}
	policy := tx.expected[v1.PolicyFile]
	if !policy.exists {
		return nil, legacyConfig{}, fmt.Errorf("easyp.yaml is missing from %s", options.Dir)
	}
	node, err := document(policy.data)
	if err != nil {
		return nil, legacyConfig{}, fmt.Errorf("document: %w", err)
	}
	if nativePolicy(node) {
		if err := p.checkNative(options.Module); err != nil {
			return nil, legacyConfig{}, fmt.Errorf("checkNative: %w", err)
		}
		p.alreadyV1 = true
		p.warnings = append(p.warnings, "Already v1: no files changed and no dependencies refreshed.")
		return p, legacyConfig{}, nil
	}
	cfg, warnings, err := parseLegacy(policy.data)
	if err != nil {
		return nil, legacyConfig{}, fmt.Errorf("parseLegacy: %w", err)
	}
	p.warnings = append(p.warnings, warnings...)
	return p, cfg, nil
}

func (p *Plan) selectSources(cfg legacyConfig, module string) (migrationSelections, error) {
	var err error
	p.local, err = newLocalSelectionProof(p.tx, cfg, module)
	if err != nil {
		return migrationSelections{}, fmt.Errorf("newLocalSelectionProof: %w", err)
	}
	p.tx.beforeApply = p.verifySourceSelection
	if len(p.local.selection.packages) > 0 || len(p.local.selection.paths) > 0 {
		if len(p.local.selection.paths) > 0 {
			p.warnings = append(p.warnings, "Legacy directory selection is preserved through literal paths selectors for the local module. Import roots and source paths stay unchanged; files outside these paths do not become targets even when they declare the same package.")
		} else {
			p.warnings = append(p.warnings, "Legacy directory selection is preserved through exact packages selectors for the local module. Import roots and source paths stay unchanged; future files declaring those packages also participate in generation.")
		}
	}
	deps := requirements{}
	for _, dep := range cfg.Deps {
		if err := deps.add(dep, false); err != nil {
			return migrationSelections{}, fmt.Errorf("add: %w", err)
		}
	}
	entries, git, err := planGitSelections(cfg, module, len(p.local.inputs) > 0, p.local.selection.packages, p.local.selection.paths, &deps)
	if err != nil {
		return migrationSelections{}, fmt.Errorf("planGitSelections: %w", err)
	}
	packages, paths := p.local.selection.packages, p.local.selection.paths
	if len(git) > 0 {
		packages, paths = nil, nil
	}
	return migrationSelections{entries: entries, git: git, deps: deps, packages: packages, paths: paths}, nil
}

type migrationLockInputs struct {
	pins       map[string]legacyPin
	historical bool
	needsLock  bool
}

func (t *transaction) lockPrerequisites(module v1.Module) (migrationLockInputs, error) {
	oldLock := t.expected["easyp.lock"]
	pins := make(map[string]legacyPin)
	if oldLock.exists {
		var err error
		pins, err = parseLegacyLock(oldLock.data)
		if err != nil {
			return migrationLockInputs{}, fmt.Errorf("parseLegacyLock: %w", err)
		}
		for _, req := range module.Requires {
			if _, exists := pins[req.Module]; !exists {
				return migrationLockInputs{}, fmt.Errorf("legacy easyp.lock is missing required dependency %s; recover its historical pin before migration", req.Module)
			}
		}
	}
	if nativeLock := t.expected[v1.LockFile]; nativeLock.exists {
		if _, err := validateNativeLock(nativeLock.data); err != nil {
			return migrationLockInputs{}, fmt.Errorf("validateNativeLock: %w", err)
		}
	}
	needsLock := len(module.Requires) > 0 || len(pins) > 0 || len(module.Replaces) > 0
	return migrationLockInputs{pins: pins, historical: oldLock.exists, needsLock: needsLock}, nil
}

func (p *Plan) proveGitSources(ctx context.Context, selection migrationSelections, fetched map[string]modules.Fetched) error {
	proof, err := proveGitSelections(selection.git, fetched, selection.entries)
	if err != nil {
		return fmt.Errorf("proveGitSelections: %w", err)
	}
	proof.bindings, err = proveGitSourceBindings(ctx, p.tx.requestedRoot, p.local, proof, true)
	if err != nil {
		return fmt.Errorf("proveGitSourceBindings: %w", err)
	}
	for _, file := range proof.bindings {
		if file.module == p.local.module {
			if _, err := p.tx.captureInput(file.path); err != nil {
				return fmt.Errorf("captureInput: %w", err)
			}
		}
	}
	p.git = proof
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
