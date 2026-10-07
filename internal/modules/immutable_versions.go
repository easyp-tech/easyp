package modules

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// ErrLockedVersionChanged indicates that a previously recorded version was republished.
var ErrLockedVersionChanged = errors.New("locked version changed")

type lockedVersionSource struct {
	Source
	locked map[string]v1.LockedModule
}

func (s lockedVersionSource) Fetch(ctx context.Context, source, version string) (Fetched, error) {
	fetched, err := s.Source.Fetch(ctx, source, version)
	if err != nil {
		return Fetched{}, fmt.Errorf("Fetch: %w", err)
	}
	if err := checkLockedVersion(s.locked[source], fetched); err != nil {
		return Fetched{}, err
	}
	return fetched, nil
}

func checkLockedVersion(old v1.LockedModule, fetched Fetched) error {
	provisional := fetched.Inspection != nil && fetched.Inspection.Provisional && fetched.Lock.Hash == ""
	if semver.IsValid(old.Version) && old.Version == fetched.Lock.Version &&
		(!strings.EqualFold(old.Commit, fetched.Lock.Commit) || (!provisional && old.Hash != fetched.Lock.Hash)) {
		return fmt.Errorf("%w: %s@%s: locked commit %s hash %s, fetched commit %s hash %s; restore the published version or select a new version", ErrLockedVersionChanged, old.Source, old.Version, old.Commit, old.Hash, fetched.Lock.Commit, fetched.Lock.Hash)
	}
	return nil
}
