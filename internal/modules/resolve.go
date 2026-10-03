// Package modules coordinates v1 dependency resolution and project module files.
package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Fetched contains metadata and a reproducible identity for one requested revision.
type Fetched struct {
	Module v1.Module
	Lock   v1.LockedModule
}

// Source supplies revision metadata without exposing checkout or cache layout.
type Source interface {
	Fetch(context.Context, string, string) (Fetched, error)
}

type versionRequirements struct {
	minimum string
	commit  string
	weak    bool
}

type revisionLoader struct {
	source   Source
	fetched  map[string]Fetched
	failures map[string]error
	pins     map[string]v1.LockedModule
	local    func(string, string) (v1.Module, bool, error)
}

func (l *revisionLoader) fetch(ctx context.Context, source, version string) (Fetched, error) {
	if v1.IsCommitRef(version) {
		version = strings.ToLower(version)
	}
	key := source + "@" + version
	if err := l.failures[key]; err != nil {
		return Fetched{}, err
	}
	if result, ok := l.fetched[key]; ok {
		return result, nil
	}
	// A tag already verified at this exact commit supplies identical metadata.
	if v1.IsCommitRef(version) {
		for _, result := range l.fetched {
			if result.Lock.Source == source && strings.EqualFold(result.Lock.Commit, version) {
				result.Lock.Version = version
				l.fetched[key] = result
				return result, nil
			}
		}
	}
	result, err := l.source.Fetch(ctx, source, version)
	if err != nil {
		if l.failures == nil {
			l.failures = make(map[string]error)
		}
		l.failures[key] = err
		return Fetched{}, err
	}
	if result.Module.Name != source || result.Lock.Source != source {
		return Fetched{}, fmt.Errorf("fetched module identity differs from requested %s (manifest %s, lock %s)", source, result.Module.Name, result.Lock.Source)
	}
	if err := v1.ValidateModuleVersion(source, result.Lock.Version); err != nil {
		return Fetched{}, fmt.Errorf("ValidateModuleVersion: %w", err)
	}
	l.fetched[key] = result
	return result, nil
}

func (l *revisionLoader) head(ctx context.Context, source string) (Fetched, error) {
	version := l.pins[source].Commit
	return l.fetch(ctx, source, version)
}

// Resolve uses the highest semantic minimum. A versionless requirement adds no
// constraint when a tag or explicit commit exists. Exact commits may coexist
// with semantic requirements only when the selected tag resolves to that commit.
// Provisional HEAD edges are rebuilt if stronger requirements supersede them.
func Resolve(ctx context.Context, root v1.Module, source Source, pins map[string]v1.LockedModule) (v1.Lock, error) {
	loader := revisionLoader{source: source, fetched: make(map[string]Fetched), pins: pins}
	pass, err := loader.resolve(ctx, root)
	return pass.lock, err
}

func (loader *revisionLoader) resolve(ctx context.Context, root v1.Module) (resolutionPass, error) {
	hints := make(map[string]string)
	states := make(map[string]bool)
	for {
		if err := ctx.Err(); err != nil {
			return resolutionPass{}, err
		}
		pass := loader.resolvePass(ctx, root, hints)
		if err := ctx.Err(); err != nil {
			return resolutionPass{}, err
		}
		stable := true
		for name, used := range pass.provisional {
			if pass.selected[name] != used {
				stable = false
				break
			}
		}
		if stable {
			if pass.err != nil {
				return resolutionPass{}, pass.err
			}
			return pass, nil
		}
		names := make([]string, 0, len(pass.selected))
		for name := range pass.selected {
			names = append(names, name)
		}
		slices.Sort(names)
		var signature strings.Builder
		for _, name := range names {
			fmt.Fprintf(&signature, "%s@%s\n", name, pass.selected[name])
		}
		state := signature.String()
		if states[state] {
			return resolutionPass{}, fmt.Errorf("unstable versionless dependency constraints; declare explicit compatible versions")
		}
		states[state] = true
		hints = pass.selected
	}
}

type resolutionPass struct {
	lock        v1.Lock
	locals      []string
	provisional map[string]string
	selected    map[string]string
	err         error
}

// deferError postpones errors until the provisional graph is stable. A failed
// dependency of an unselected HEAD must not poison an otherwise valid graph.
func (p *resolutionPass) deferError(err error) {
	if p.err == nil {
		p.err = err
	}
}

func (l *revisionLoader) resolvePass(ctx context.Context, root v1.Module, hints map[string]string) resolutionPass {
	pass := resolutionPass{lock: v1.Lock{Version: 1, Modules: []v1.LockedModule{}}, provisional: make(map[string]string), selected: make(map[string]string)}
	requirements := make(map[string]*versionRequirements)
	visited := make(map[string]bool)
	localVisited := make(map[string]bool)
	queue := slices.Clone(root.Requires)
	for {
		for len(queue) > 0 {
			if err := ctx.Err(); err != nil {
				pass.deferError(err)
				return pass
			}
			requirement := queue[0]
			queue = queue[1:]
			if err := v1.ValidateModuleVersion(requirement.Module, requirement.Version); err != nil {
				pass.deferError(fmt.Errorf("ValidateModuleVersion: %w", err))
				continue
			}
			if l.local != nil {
				// The main module and replaced nodes have no remote revision.
				if requirement.Module == root.Name {
					continue
				}
				module, local, err := l.local(requirement.Module, requirement.Version)
				if err != nil {
					pass.deferError(err)
					continue
				}
				if local {
					if localVisited[requirement.Module] {
						continue
					}
					localVisited[requirement.Module] = true
					pass.locals = append(pass.locals, requirement.Module)
					queue = append(queue, module.Requires...)
					continue
				}
			}
			version := requirement.Version
			if version != "" && !v1.IsCommitRef(version) && !semver.IsValid(version) {
				pass.deferError(fmt.Errorf("require %s: expected semantic version or full Git commit, got %q", requirement.Module, version))
				continue
			}
			r := requirements[requirement.Module]
			if r == nil {
				r = &versionRequirements{}
				requirements[requirement.Module] = r
			}
			switch {
			case version == "":
				r.weak = true
				continue
			case v1.IsCommitRef(version):
				version = strings.ToLower(version)
				if r.commit != "" && r.commit != version {
					pass.deferError(fmt.Errorf("module %s has conflicting requirements %s and %s", requirement.Module, r.commit, version))
					continue
				}
				r.commit = version
			default:
				if r.minimum == "" || semver.Compare(version, r.minimum) > 0 || (semver.Compare(version, r.minimum) == 0 && version > r.minimum) {
					r.minimum = version
				}
			}
			key := requirement.Module + "@" + version
			if visited[key] {
				continue
			}
			visited[key] = true
			result, err := l.fetch(ctx, requirement.Module, version)
			if err != nil {
				pass.deferError(err)
				continue
			}
			queue = append(queue, result.Module.Requires...)
		}
		pending := make([]string, 0)
		for name, r := range requirements {
			if r.weak && r.minimum == "" && r.commit == "" {
				if _, ok := pass.provisional[name]; !ok {
					pending = append(pending, name)
				}
			}
		}
		slices.Sort(pending)
		if len(pending) == 0 {
			break
		}
		if err := ctx.Err(); err != nil {
			pass.deferError(err)
			return pass
		}
		name := pending[0]
		var result Fetched
		var err error
		used := hints[name]
		if used != "" {
			result, err = l.fetch(ctx, name, used)
		} else {
			used = l.pins[name].Commit
			result, err = l.head(ctx, name)
		}
		if err != nil {
			pass.provisional[name] = used
			pass.deferError(err)
			continue
		}
		pass.provisional[name] = result.Lock.Version
		queue = append(queue, result.Module.Requires...)
	}
	names := make([]string, 0, len(requirements))
	for name := range requirements {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		r := requirements[name]
		var result Fetched
		var err error
		switch {
		case r.minimum != "":
			pass.selected[name] = r.minimum
			result, err = l.fetch(ctx, name, r.minimum)
		case r.commit != "":
			pass.selected[name] = r.commit
			result, err = l.fetch(ctx, name, r.commit)
		default:
			pass.selected[name] = l.pins[name].Commit
			result, err = l.head(ctx, name)
		}
		if err != nil {
			pass.deferError(err)
			continue
		}
		pass.selected[name] = result.Lock.Version
		if r.commit != "" && !strings.EqualFold(result.Lock.Commit, r.commit) {
			pass.deferError(fmt.Errorf("module %s has conflicting requirements: tag %s resolves to %s, explicit commit is %s", name, r.minimum, result.Lock.Commit, r.commit))
		}
		pass.lock.Modules = append(pass.lock.Modules, result.Lock)
	}
	return pass
}
