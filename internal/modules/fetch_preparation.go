package modules

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/parallel"
)

// FetchRequest describes independent revision and source-scope preparation.
// Roots and Inspect preserve the root-selection source contracts when a wrapper
// prepares a request on behalf of the resolver.
type FetchRequest struct {
	Source  string
	Version string
	Roots   []string
	Inspect bool
}

// PreparedFetch completes prepared work in the resolver's original request
// order. Preparation must release temporary resources before returning; a
// provisional graph may discard the callback without consuming it.
type PreparedFetch func(context.Context) (Fetched, error)

// fetchPreparer is optional: ordinary Source implementations remain serial.
// PrepareFetch may run concurrently for independent modules. Its callback must
// defer mutable resolver, metadata-backend, and import-root state until consumed.
type fetchPreparer interface {
	PreparationSource() Source
	PrepareFetch(context.Context, FetchRequest) (PreparedFetch, error)
}

const moduleFetchLimit = 8

type preparedRevision struct {
	complete PreparedFetch
	err      error
}

func sourceFetchPreparer(source Source) (fetchPreparer, bool) {
	preparer, supported := source.(fetchPreparer)
	if !supported || !reflect.ValueOf(source).Comparable() {
		return nil, false
	}
	// Promoted methods on an embedding wrapper do not advertise safety for
	// that wrapper's overridden Fetch/root methods. It must opt in itself.
	if preparer.PreparationSource() != source {
		return nil, false
	}
	return preparer, true
}

func (loader *revisionLoader) prepareFetches(ctx context.Context, queue []v1.Requirement, requirements map[string]*versionRequirements, visited map[string]bool) error {
	preparer, supported := sourceFetchPreparer(loader.source)
	// Local replacement lookup can itself discover stateful graph edges. Keep
	// its requests serial until that lookup has selected a remote revision.
	if !supported || loader.local != nil {
		return nil
	}
	var requests []FetchRequest
	seen := make(map[string]bool)
	for _, requirement := range queue {
		version := requirement.Version
		// Versionless nodes must wait until stronger requirements are known.
		if version == "" || v1.ValidateModuleVersion(requirement.Module, version) != nil {
			continue
		}
		if v1.IsCommitRef(version) {
			version = strings.ToLower(version)
			if previous := requirements[requirement.Module]; previous != nil && previous.commit != "" && previous.commit != version {
				continue
			}
		} else if !semver.IsValid(version) {
			continue
		}
		// One revision per module preserves tag-to-commit reuse and avoids
		// fetching a later conflicting exact requirement speculatively.
		if seen[requirement.Module] {
			continue
		}
		seen[requirement.Module] = true
		key := requirement.Module + "@" + version
		if visited[key] || loader.failures[key] != nil {
			continue
		}
		if _, exists := loader.prepared[key]; exists {
			continue
		}
		if _, exists := loader.fetched[key]; exists {
			continue
		}
		if v1.IsCommitRef(version) {
			known := false
			for _, fetched := range loader.fetched {
				if fetched.Lock.Source == requirement.Module && strings.EqualFold(fetched.Lock.Commit, version) {
					known = true
					break
				}
			}
			if known {
				continue
			}
		}
		requests = append(requests, FetchRequest{Source: requirement.Module, Version: version})
	}
	results := make([]preparedRevision, len(requests))
	err := parallel.Run(ctx, len(requests), moduleFetchLimit, func(index int) error {
		complete, err := preparer.PrepareFetch(ctx, requests[index])
		results[index] = preparedRevision{complete: complete, err: err}
		// A provisional graph can discard this revision. Report its error only
		// when the serial solver actually consumes the corresponding request.
		return nil
	})
	if err != nil {
		return fmt.Errorf("Run: %w", err)
	}
	if len(requests) > 0 && loader.prepared == nil {
		loader.prepared = make(map[string]preparedRevision)
	}
	for index, request := range requests {
		loader.prepared[request.Source+"@"+request.Version] = results[index]
	}
	return nil
}
