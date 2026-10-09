// Package policy resolves producer-policy extends without changing module
// ownership or acquiring unpinned dependencies.
package policy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

const maxExtendsDepth = 16

// GraphProvider supplies the consumer's verified module graph. It is injected
// so config/v1 parsing remains syntax-only and independent of module resolution.
type GraphProvider func(context.Context, string) (modules.PolicyGraph, error)

// Resolver is operation-scoped. Its caches are never shared between commands
// or module contexts.
type Resolver struct {
	WorkspaceRoot string
	Graph         GraphProvider

	loaded   map[loadKey]loadedPolicy
	graphs   map[string]modules.PolicyGraph
	resolved map[resolveKey]resolvedSection
	pinned   map[pinnedSourceKey]modules.PolicyFiles
}

// NewResolver constructs an operation-scoped policy resolver.
func NewResolver(workspaceRoot string, graph GraphProvider) *Resolver {
	return &Resolver{
		WorkspaceRoot: workspaceRoot,
		Graph:         graph,
		loaded:        make(map[loadKey]loadedPolicy),
		graphs:        make(map[string]modules.PolicyGraph),
		resolved:      make(map[resolveKey]resolvedSection),
		pinned:        make(map[pinnedSourceKey]modules.PolicyFiles),
	}
}

type loadedPolicy struct {
	path              string
	canonical         string
	policy            v1.Policy
	presence          v1.PolicyPresence
	expandEnvironment bool
	graphKey          string
}

type pinnedSourceKey struct{ boundary, graph string }

type loadKey struct {
	path              string
	expandEnvironment bool
	graph             string
}

type resolveKey struct {
	path              string
	section           string
	graph             string
	boundary          string
	expandEnvironment bool
}

type resolvedSection struct {
	linters  v1.LinterPolicy
	settings map[string]map[string]string
	breaking v1.BreakingPolicy
	chain    []string
}

// LintInput describes the nearest-section lint cascade selected by the CLI.
// Linters and settings may come from different policy files.
type LintInput struct {
	PolicyPath     string
	Policy         v1.Policy
	Presence       v1.PolicyPresence
	SettingsPath   string
	Settings       map[string]map[string]string
	SettingsSource v1.PolicyPresence
	ModuleDir      string
}

// LintResult contains the resolved lint section and its linter settings. Issues
// stay owned by the consuming nearest-section cascade and are never inherited.
type LintResult struct {
	Policy v1.Policy
	Chain  []string
}

// ResolveLint resolves linters.extends recursively, then applies the local
// linter and settings adjustments. Remote references use the checked graph of
// moduleDir; local references stay inside the workspace or verified module.
func (r *Resolver) ResolveLint(ctx context.Context, input LintInput) (LintResult, error) {
	local := input.Policy
	local.LinterSettings = input.Settings
	if local.Linters.Extends == "" {
		return LintResult{Policy: local}, nil
	}
	graph, err := r.graphForReference(ctx, input.ModuleDir, local.Linters.Extends)
	if err != nil {
		return LintResult{}, fmt.Errorf("resolve %s linters.extends %q for consuming policy %s: %w", input.PolicyPath, local.Linters.Extends, input.PolicyPath, err)
	}
	current, err := r.load(ctx, input.PolicyPath, r.WorkspaceRoot, true, input.ModuleDir)
	if err != nil {
		return LintResult{}, err
	}
	basePath, baseBoundary, reference, expandEnvironment, err := r.resolveReference(ctx, current, local.Linters.Extends, graph, r.WorkspaceRoot)
	if err != nil {
		return LintResult{}, fmt.Errorf("%s: linters.extends %q: %w", input.PolicyPath, local.Linters.Extends, err)
	}
	base, err := r.resolveLintFile(ctx, basePath, baseBoundary, input.ModuleDir, expandEnvironment, nil)
	if err != nil {
		return LintResult{}, fmt.Errorf("%s: linters.extends %q resolved to %s: %w", input.PolicyPath, reference, basePath, err)
	}
	local.Linters = v1.MergeLinterSection(base.linters, local.Linters, input.Presence, true)
	local.Linters.Extends = ""
	local.LinterSettings = v1.MergeLinterSettings(base.settings, input.Settings, input.SettingsSource, true)
	return LintResult{Policy: local, Chain: append(base.chain, input.PolicyPath)}, nil
}

// BreakingInput describes the nearest breaking section and the module whose
// files consume it. The module graph is intentionally supplied per checked
// module because a shared policy can be used by modules with different overlays.
type BreakingInput struct {
	PolicyPath string
	Policy     v1.Policy
	Presence   v1.PolicyPresence
	ModuleDir  string
}

// BreakingResult contains the resolved breaking section and provenance chain.
type BreakingResult struct {
	Policy v1.Policy
	Chain  []string
}

// ResolveBreaking resolves breaking.extends independently of lint inheritance.
func (r *Resolver) ResolveBreaking(ctx context.Context, input BreakingInput) (BreakingResult, error) {
	local := input.Policy
	if local.Breaking.Extends == "" {
		return BreakingResult{Policy: local}, nil
	}
	graph, err := r.graphForReference(ctx, input.ModuleDir, local.Breaking.Extends)
	if err != nil {
		return BreakingResult{}, fmt.Errorf("resolve %s breaking.extends %q for consuming policy %s: %w", input.PolicyPath, local.Breaking.Extends, input.PolicyPath, err)
	}
	current, err := r.load(ctx, input.PolicyPath, r.WorkspaceRoot, true, input.ModuleDir)
	if err != nil {
		return BreakingResult{}, err
	}
	basePath, baseBoundary, reference, expandEnvironment, err := r.resolveReference(ctx, current, local.Breaking.Extends, graph, r.WorkspaceRoot)
	if err != nil {
		return BreakingResult{}, fmt.Errorf("%s: breaking.extends %q: %w", input.PolicyPath, local.Breaking.Extends, err)
	}
	base, err := r.resolveBreakingFile(ctx, basePath, baseBoundary, input.ModuleDir, expandEnvironment, nil)
	if err != nil {
		return BreakingResult{}, fmt.Errorf("%s: breaking.extends %q resolved to %s: %w", input.PolicyPath, reference, basePath, err)
	}
	local.Breaking = v1.MergeBreakingSection(base.breaking, local.Breaking, input.Presence, true)
	local.Breaking.Extends = ""
	return BreakingResult{Policy: local, Chain: append(base.chain, input.PolicyPath)}, nil
}

func (r *Resolver) resolveLintFile(ctx context.Context, path, boundary string, graphKey string, expandEnvironment bool, stack []string) (resolvedSection, error) {
	loaded, err := r.load(ctx, path, boundary, expandEnvironment, graphKey)
	if err != nil {
		return resolvedSection{}, err
	}
	key := resolveKey{path: loaded.path, section: v1.PolicySectionLinters, graph: graphKey, boundary: boundary, expandEnvironment: loaded.expandEnvironment}
	if cached, ok := r.resolved[key]; ok {
		if len(stack)+len(cached.chain) > maxExtendsDepth {
			return resolvedSection{}, fmt.Errorf("policy extends exceeds maximum depth %d at %s", maxExtendsDepth, path)
		}
		return cached, nil
	}
	if err := checkCycle(loaded.canonical, v1.PolicySectionLinters, stack); err != nil {
		return resolvedSection{}, err
	}
	if !loaded.presence.Has(v1.PolicySectionLinters) {
		return resolvedSection{}, fmt.Errorf("base policy %s does not define linters", loaded.path)
	}
	stack = append(stack, loaded.canonical+"#linters")
	var result resolvedSection
	if loaded.policy.Linters.Extends != "" {
		graph, err := r.graphForReference(ctx, graphKey, loaded.policy.Linters.Extends)
		if err != nil {
			return resolvedSection{}, fmt.Errorf("%s: linters.extends %q: %w", loaded.path, loaded.policy.Linters.Extends, err)
		}
		basePath, baseBoundary, _, baseExpandEnvironment, err := r.resolveReference(ctx, loaded, loaded.policy.Linters.Extends, graph, boundary)
		if err != nil {
			return resolvedSection{}, fmt.Errorf("%s: linters.extends %q: %w", loaded.path, loaded.policy.Linters.Extends, err)
		}
		base, err := r.resolveLintFile(ctx, basePath, baseBoundary, graphKey, baseExpandEnvironment, stack)
		if err != nil {
			return resolvedSection{}, fmt.Errorf("%s: linters.extends %q resolved to %s: %w", loaded.path, loaded.policy.Linters.Extends, basePath, err)
		}
		result.linters = v1.MergeLinterSection(base.linters, loaded.policy.Linters, loaded.presence, true)
		result.settings = v1.MergeLinterSettings(base.settings, loaded.policy.LinterSettings, loaded.presence, true)
		result.chain = append(base.chain, loaded.canonical)
	} else {
		result.linters = v1.BaseLinterSection(loaded.policy.Linters, loaded.presence)
		result.settings = cloneSettings(loaded.policy.LinterSettings)
		result.chain = []string{loaded.canonical}
	}
	r.resolved[key] = result
	return result, nil
}

func (r *Resolver) resolveBreakingFile(ctx context.Context, path, boundary string, graphKey string, expandEnvironment bool, stack []string) (resolvedSection, error) {
	loaded, err := r.load(ctx, path, boundary, expandEnvironment, graphKey)
	if err != nil {
		return resolvedSection{}, err
	}
	key := resolveKey{path: loaded.path, section: v1.PolicySectionBreaking, graph: graphKey, boundary: boundary, expandEnvironment: loaded.expandEnvironment}
	if cached, ok := r.resolved[key]; ok {
		if len(stack)+len(cached.chain) > maxExtendsDepth {
			return resolvedSection{}, fmt.Errorf("policy extends exceeds maximum depth %d at %s", maxExtendsDepth, path)
		}
		return cached, nil
	}
	if err := checkCycle(loaded.canonical, v1.PolicySectionBreaking, stack); err != nil {
		return resolvedSection{}, err
	}
	if !loaded.presence.Has(v1.PolicySectionBreaking) {
		return resolvedSection{}, fmt.Errorf("base policy %s does not define breaking", loaded.path)
	}
	stack = append(stack, loaded.canonical+"#breaking")
	var result resolvedSection
	if loaded.policy.Breaking.Extends != "" {
		graph, err := r.graphForReference(ctx, graphKey, loaded.policy.Breaking.Extends)
		if err != nil {
			return resolvedSection{}, fmt.Errorf("%s: breaking.extends %q: %w", loaded.path, loaded.policy.Breaking.Extends, err)
		}
		basePath, baseBoundary, _, baseExpandEnvironment, err := r.resolveReference(ctx, loaded, loaded.policy.Breaking.Extends, graph, boundary)
		if err != nil {
			return resolvedSection{}, fmt.Errorf("%s: breaking.extends %q: %w", loaded.path, loaded.policy.Breaking.Extends, err)
		}
		base, err := r.resolveBreakingFile(ctx, basePath, baseBoundary, graphKey, baseExpandEnvironment, stack)
		if err != nil {
			return resolvedSection{}, fmt.Errorf("%s: breaking.extends %q resolved to %s: %w", loaded.path, loaded.policy.Breaking.Extends, basePath, err)
		}
		result.breaking = v1.MergeBreakingSection(base.breaking, loaded.policy.Breaking, loaded.presence, true)
		result.chain = append(base.chain, loaded.canonical)
	} else {
		result.breaking = v1.BaseBreakingSection(loaded.policy.Breaking, loaded.presence)
		result.chain = []string{loaded.canonical}
	}
	r.resolved[key] = result
	return result, nil
}

func (r *Resolver) load(ctx context.Context, path, boundary string, expandEnvironment bool, graphKey string) (loadedPolicy, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return loadedPolicy{}, fmt.Errorf("policy source %s: Abs: %w", path, err)
	}
	key := loadKey{path: absolute + "\x00" + boundary, expandEnvironment: expandEnvironment, graph: graphKey}
	if cached, ok := r.loaded[key]; ok {
		return cached, nil
	}
	relative, err := filepath.Rel(boundary, absolute)
	if err != nil || !filepath.IsLocal(relative) {
		return loadedPolicy{}, fmt.Errorf("policy source %s is outside allowed policy root %s", path, boundary)
	}
	var raw []byte
	var canonical string
	if files := r.pinned[pinnedSourceKey{boundary, graphKey}]; files != nil {
		file, err := files.Read(ctx, filepath.ToSlash(relative))
		if err != nil {
			return loadedPolicy{}, fmt.Errorf("Read: policy source %s: %w", absolute, err)
		}
		raw, canonical = file.Content, file.Canonical
	} else {
		canonical, err = filepath.EvalSymlinks(absolute)
		if err != nil {
			return loadedPolicy{}, fmt.Errorf("policy source %s: EvalSymlinks: %w", path, err)
		}
		raw, err = sourceview.ReadLocal(ctx, boundary, relative)
		if err != nil {
			return loadedPolicy{}, fmt.Errorf("read policy source %s: %w", canonical, err)
		}
	}
	var parsed v1.Policy
	if expandEnvironment {
		parsed, err = v1.ParsePolicy(bytes.NewReader(raw))
	} else {
		parsed, err = v1.ParsePolicyLiteral(bytes.NewReader(raw))
	}
	if err != nil {
		return loadedPolicy{}, fmt.Errorf("invalid policy source %s: %w", canonical, err)
	}
	presence, err := v1.ParsePolicyPresence(raw)
	if err != nil {
		return loadedPolicy{}, fmt.Errorf("policy source %s presence: %w", canonical, err)
	}
	result := loadedPolicy{
		path: absolute, canonical: canonical, policy: parsed, presence: presence,
		expandEnvironment: expandEnvironment, graphKey: graphKey,
	}
	r.loaded[key] = result
	return result, nil
}

func (r *Resolver) graph(ctx context.Context, moduleDir string) (modules.PolicyGraph, error) {
	if moduleDir == "" {
		return nil, fmt.Errorf("remote policy references require the consuming protobuf.mod and its locked graph")
	}
	key, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, err
	}
	if graph, ok := r.graphs[key]; ok {
		return graph, nil
	}
	if r.Graph == nil {
		return nil, fmt.Errorf("remote policy references require a verified module graph")
	}
	graph, err := r.Graph(ctx, key)
	if err != nil {
		return nil, err
	}
	r.graphs[key] = graph
	return graph, nil
}

func (r *Resolver) graphForReference(ctx context.Context, moduleDir, raw string) (modules.PolicyGraph, error) {
	reference, err := v1.ParsePolicyReference(raw)
	if err != nil {
		return nil, err
	}
	if reference.Local {
		return nil, nil
	}
	return r.graph(ctx, moduleDir)
}

func (r *Resolver) resolveReference(ctx context.Context, from loadedPolicy, raw string, graph modules.PolicyGraph, inheritedBoundary string) (string, string, string, bool, error) {
	reference, err := v1.ParsePolicyReference(raw)
	if err != nil {
		return "", "", "", false, err
	}
	if reference.Empty() {
		return "", "", "", false, fmt.Errorf("empty reference")
	}
	expandEnvironment := from.expandEnvironment
	var directory, boundary string
	if reference.Local {
		directory = filepath.Join(filepath.Dir(from.path), filepath.FromSlash(reference.Path))
		boundary = inheritedBoundary
	} else {
		moduleName, relative, err := matchModuleReference(reference, graph)
		if err != nil {
			return "", "", "", false, err
		}
		module := graph[moduleName]
		directory = module.Directory
		boundary = module.Directory
		expandEnvironment = false
		r.pinned[pinnedSourceKey{boundary, from.graphKey}] = module.Files
		if relative != "" {
			directory = filepath.Join(module.Directory, filepath.FromSlash(relative))
		}
	}
	if !within(directory, boundary) {
		return "", "", "", false, fmt.Errorf("reference %q selects a path outside allowed policy root %s", raw, boundary)
	}
	relative, err := filepath.Rel(boundary, directory)
	if err != nil {
		return "", "", "", false, fmt.Errorf("Rel: %w", err)
	}
	if files := r.pinned[pinnedSourceKey{boundary, from.graphKey}]; files != nil {
		file, err := files.Read(ctx, filepath.ToSlash(relative))
		if err != nil {
			return "", "", "", false, fmt.Errorf("Read: reference %q: %w", raw, err)
		}
		return filepath.Join(boundary, filepath.FromSlash(file.Path)), boundary, raw, false, nil
	}
	resolved, err := sourceview.ResolveLocal(ctx, boundary, relative)
	info := resolved.Info
	if err != nil {
		return "", "", "", false, fmt.Errorf("reference %q selects %s: %w", raw, directory, err)
	}
	if info.IsDir() {
		directory = filepath.Join(directory, v1.PolicyFile)
	}
	canonical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", "", "", false, fmt.Errorf("reference %q policy file %s: %w", raw, directory, err)
	}
	info, err = os.Stat(canonical)
	if err != nil {
		return "", "", "", false, fmt.Errorf("reference %q policy file %s: %w", raw, canonical, err)
	}
	if !info.Mode().IsRegular() {
		return "", "", "", false, fmt.Errorf("reference %q policy file %s is not regular", raw, canonical)
	}
	canonicalBoundary, err := filepath.EvalSymlinks(boundary)
	if err != nil {
		return "", "", "", false, fmt.Errorf("reference %q boundary %s: %w", raw, boundary, err)
	}
	if !within(canonical, canonicalBoundary) {
		return "", "", "", false, fmt.Errorf("reference %q resolves outside allowed policy root %s", raw, canonicalBoundary)
	}
	return directory, boundary, raw, expandEnvironment, nil
}

func matchModuleReference(reference v1.PolicyReference, graph modules.PolicyGraph) (string, string, error) {
	if reference.Fragment {
		if _, ok := graph[reference.Module]; !ok {
			return "", "", fmt.Errorf("module %s in reference %q is not declared and resolved in the consumer's protobuf.mod/protobuf.lock", reference.Module, reference.Raw)
		}
		return reference.Module, reference.Relative, nil
	}
	var candidates []string
	for _, name := range graph.SortedNames() {
		if reference.Raw == name || strings.HasPrefix(reference.Raw, name+"/") {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return "", "", fmt.Errorf("reference %q does not match a declared and resolved module identity", reference.Raw)
	}
	slices.SortFunc(candidates, func(a, b string) int {
		if len(a) > len(b) {
			return -1
		}
		if len(a) < len(b) {
			return 1
		}
		return strings.Compare(a, b)
	})
	module := candidates[0]
	if len(candidates) > 1 && len(candidates[0]) == len(candidates[1]) {
		return "", "", fmt.Errorf("reference %q is ambiguous across module identities %s and %s; use <module>#<policy-path>", reference.Raw, candidates[0], candidates[1])
	}
	relative := strings.TrimPrefix(strings.TrimPrefix(reference.Raw, module), "/")
	return module, relative, nil
}

func checkCycle(canonical, section string, stack []string) error {
	key := canonical + "#" + section
	if len(stack) >= maxExtendsDepth {
		return fmt.Errorf("policy extends exceeds maximum depth %d at %s", maxExtendsDepth, key)
	}
	if slices.Contains(stack, key) {
		return fmt.Errorf("policy extends cycle at %s through %s", key, strings.Join(stack, " -> "))
	}
	return nil
}

func within(path, root string) bool {
	if root == "" {
		return false
	}
	relative, err := filepath.Rel(root, path)
	return err == nil && filepath.IsLocal(relative)
}

func cloneSettings(settings map[string]map[string]string) map[string]map[string]string {
	if settings == nil {
		return nil
	}
	result := make(map[string]map[string]string, len(settings))
	for rule, values := range settings {
		copyValues := make(map[string]string, len(values))
		for key, value := range values {
			copyValues[key] = value
		}
		result[rule] = copyValues
	}
	return result
}
