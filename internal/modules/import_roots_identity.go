package modules

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func (source *importRootSource) sameLockedPin(entry v1.LockedModule) bool {
	old, exists := source.locked[entry.Source]
	return exists && strings.EqualFold(old.Commit, entry.Commit)
}

func (source *importRootSource) fixedRootScope(entry v1.LockedModule, fetched Fetched) bool {
	return fetched.Module.RootsFromMetadata || len(source.hints[entry.Source]) > 0 || len(source.locked[entry.Source].Roots) > 0 || source.sameLockedPin(entry) || fetched.Inspection == nil
}

func (source *importRootSource) verifyRootFetch(ctx context.Context, fetched Fetched) error {
	old, exists := source.locked[fetched.Lock.Source]
	// A semantic tag must still identify its original commit before a changed
	// scope is considered. Full-commit requests are checked by the fetch caller.
	if err := checkLockedRevision(old, fetched); err != nil {
		return err
	}
	if exists && strings.EqualFold(old.Commit, fetched.Lock.Commit) {
		provisional := fetched.Inspection != nil && fetched.Inspection.Provisional && fetched.Lock.Hash == ""
		if provisional && len(source.hints[old.Source]) == 0 {
			return nil
		}
		if fetched.Inspection != nil && fetched.Inspection.Provisional {
			return fmt.Errorf("module %s: explicit root fetch remains provisional", old.Source)
		}
		if len(source.hints[old.Source]) > 0 {
			selector, supported := source.Source.(rootSelectionSource)
			if !supported {
				return fmt.Errorf("module %s: repository cannot verify the original root scope", old.Source)
			}
			previous, err := source.lockedRootScope(ctx, selector, old)
			if err != nil {
				return fmt.Errorf("lockedRootScope: %w", err)
			}
			before, after := slices.Clone(previous.Module.Roots), slices.Clone(fetched.Module.Roots)
			slices.Sort(before)
			slices.Sort(after)
			if !slices.Equal(before, after) {
				if fetched.Lock.Hash == "" {
					return fmt.Errorf("module %s: explicit root fetch has no complete hash", old.Source)
				}
				// Immutable source bytes and the original exact scope/hash
				// have been proved. A different checked logical scope can
				// materialize aliases differently and have a different hash.
				return nil
			}
		}
		if old.Hash != fetched.Lock.Hash {
			return lockedVersionChanged(old, fetched)
		}
	}
	return checkLockedVersion(old, fetched)
}

func (source *importRootSource) lockedRootScope(ctx context.Context, selector rootSelectionSource, entry v1.LockedModule) (Fetched, error) {
	if previous, exists := source.verifiedScopes[entry.Source]; exists {
		return previous, nil
	}
	previous, err := fetchLockedRootScope(ctx, selector, entry, entry.Roots)
	if err != nil {
		return Fetched{}, fmt.Errorf("fetchLockedRootScope: %w", err)
	}
	if source.verifiedScopes == nil {
		source.verifiedScopes = make(map[string]Fetched)
	}
	source.verifiedScopes[entry.Source] = previous
	return previous, nil
}
