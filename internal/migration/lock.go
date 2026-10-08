package migration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Repository verifies the v0 installed-tree hash (when supplied), then returns
// independently hashed v1 tracked contents and metadata for that same commit.
// It is called only when Options.ResolveLock explicitly permits resolution.
type Repository interface {
	FetchMigration(context.Context, string, string, string) (modules.Fetched, error)
}

type legacyPin struct{ source, version, hash string }

func parseLegacyLock(raw []byte) (map[string]legacyPin, error) {
	pins := make(map[string]legacyPin)
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) != 3 {
			return nil, fmt.Errorf("easyp.lock:%d: expected SOURCE VERSION h1:HASH; a native v1 lock is only accepted as an unchanged protobuf.lock", line)
		}
		if err := validIdentity(fields[0]); err != nil {
			return nil, fmt.Errorf("validIdentity: %w", err)
		}
		legacyVersion := fields[1]
		// v0 could persist the peeled-ref line from an annotated Git tag.
		// Only full SemVer tags qualify; pseudo-shaped names must retain their
		// tag semantics rather than entering legacy pseudo-to-commit conversion.
		if tag, peeled := strings.CutSuffix(legacyVersion, "^{}"); peeled && semver.IsValid(tag) && fullSemver.MatchString(tag) && !legacyPseudo.MatchString(tag) {
			legacyVersion = tag
		}
		version, err := migrationVersion(legacyVersion)
		if err != nil {
			return nil, fmt.Errorf("migrationVersion: %w", err)
		}
		if !strings.HasPrefix(fields[2], "h1:") {
			return nil, fmt.Errorf("easyp.lock:%d: invalid legacy hash", line)
		}
		digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fields[2], "h1:"))
		if err != nil || len(digest) != 32 {
			return nil, fmt.Errorf("easyp.lock:%d: invalid legacy hash", line)
		}
		if _, ok := pins[fields[0]]; ok {
			return nil, fmt.Errorf("easyp.lock:%d: duplicate dependency %s", line, fields[0])
		}
		pins[fields[0]] = legacyPin{source: fields[0], version: version, hash: fields[2]}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Err: %w", err)
	}
	return pins, nil
}

func migrateLock(ctx context.Context, module v1.Module, pins map[string]legacyPin, hadLock bool, repository Repository) (v1.Lock, error) {
	lock, _, err := migrateSelectionLock(ctx, module, pins, hadLock, repository, nil)
	return lock, err
}

func migrateSelectionLock(ctx context.Context, module v1.Module, pins map[string]legacyPin, hadLock bool, repository Repository, selections map[string]gitModuleSelection) (v1.Lock, map[string]modules.Fetched, error) {
	if len(module.Replaces) > 0 {
		return v1.Lock{}, nil, fmt.Errorf("full lock migration with local replacements requires manual migration; the preview preserves replacement structure")
	}
	if repository == nil {
		return v1.Lock{}, nil, fmt.Errorf("--resolve-lock requires a dependency repository")
	}
	source := migrationSource{repository: repository, metadata: make(map[migrationRevision]v1.Module), fetched: make(map[migrationRevision]modules.Fetched), selections: selections}
	for _, selection := range selections {
		if selection.needsProof() {
			if _, ok := repository.(RootsRepository); !ok {
				return v1.Lock{}, nil, fmt.Errorf("generate.inputs[%d].git_repo %s requires a roots-aware migration repository; its explicit root/sub_directory cannot be ignored", selection.index, selection.name)
			}
		}
	}
	if !hadLock {
		lock, err := modules.Resolve(ctx, module, source, nil)
		if err != nil {
			return v1.Lock{}, nil, fmt.Errorf("Resolve: %w", err)
		}
		if err := validateRuntimeLock(module, lock, source.metadata); err != nil {
			return v1.Lock{}, nil, fmt.Errorf("validateRuntimeLock: %w", err)
		}
		return lock, source.selectedFetched(lock), nil
	}
	// Never run the normal version-selection algorithm on historical pins: every
	// legacy entry is verified at its recorded SHA/tag, including indirect pins.
	names := make([]string, 0, len(pins))
	for name := range pins {
		names = append(names, name)
	}
	slices.Sort(names)
	resolved := make(map[string]modules.Fetched, len(pins))
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{}}
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return v1.Lock{}, nil, fmt.Errorf("Err: %w", err)
		}
		pin := pins[name]
		fetched, err := source.fetchMigration(ctx, name, pin.version, pin.hash)
		if err != nil {
			return v1.Lock{}, nil, fmt.Errorf("fetchMigration: %w", err)
		}
		if err := checkFetched(name, pin.version, fetched); err != nil {
			return v1.Lock{}, nil, fmt.Errorf("checkFetched: %w", err)
		}
		resolved[name] = fetched
		source.remember(fetched)
		lock.Modules = append(lock.Modules, fetched.Lock)
	}
	requirements := slices.Clone(module.Requires)
	for _, name := range names {
		requirements = append(requirements, resolved[name].Module.Requires...)
	}
	verifiedVersions := make(map[string]string)
	for _, requirement := range requirements {
		pin, ok := resolved[requirement.Module]
		if !ok {
			return v1.Lock{}, nil, fmt.Errorf("legacy easyp.lock is missing required dependency %s; recover its historical pin before migration (HEAD will not be selected)", requirement.Module)
		}
		version, err := checkPinRequirement(ctx, requirement, pin.Lock, source)
		if err != nil {
			return v1.Lock{}, nil, fmt.Errorf("checkPinRequirement: %w", err)
		}
		previous := verifiedVersions[requirement.Module]
		if version != "" && (previous == "" || semver.Compare(version, previous) > 0 || semver.Compare(version, previous) == 0 && version > previous) {
			verifiedVersions[requirement.Module] = version
		}
	}
	// Every comparison above uses the original historical pin. Assign labels only
	// after all required tags are verified, so a higher tag cannot hide a moved
	// lower tag. The historical commit and its freshly verified hash stay intact.
	for i := range lock.Modules {
		if version := verifiedVersions[lock.Modules[i].Source]; version != "" {
			lock.Modules[i].Version = version
		}
	}
	if err := validateRuntimeLock(module, lock, source.metadata); err != nil {
		return v1.Lock{}, nil, fmt.Errorf("validateRuntimeLock: %w", err)
	}
	return lock, source.selectedFetched(lock), nil
}

type migrationRevision struct{ source, commit, hash string }

type migrationSource struct {
	repository Repository
	metadata   map[migrationRevision]v1.Module
	fetched    map[migrationRevision]modules.Fetched
	selections map[string]gitModuleSelection
}

func revisionKey(entry v1.LockedModule) migrationRevision {
	return migrationRevision{source: entry.Source, commit: strings.ToLower(entry.Commit), hash: entry.Hash}
}

func (s migrationSource) remember(fetched modules.Fetched) {
	key := revisionKey(fetched.Lock)
	if _, exists := s.metadata[key]; !exists {
		s.metadata[key] = fetched.Module
		if s.fetched != nil {
			s.fetched[key] = fetched
		}
	}
}

func (s migrationSource) selectedFetched(lock v1.Lock) map[string]modules.Fetched {
	result := make(map[string]modules.Fetched, len(lock.Modules))
	for _, entry := range lock.Modules {
		fetched := s.fetched[revisionKey(entry)]
		fetched.Lock = entry
		result[entry.Source] = fetched
	}
	return result
}

func (s migrationSource) fetchMigration(ctx context.Context, source, version, hash string) (modules.Fetched, error) {
	if selection, ok := s.selections[source]; ok && selection.needsProof() {
		repository, ok := s.repository.(RootsRepository)
		if !ok {
			return modules.Fetched{}, fmt.Errorf("dependency %s requires a roots-aware migration repository", source)
		}
		fetched, err := repository.FetchMigrationWithRoots(ctx, source, version, hash, selection.roots())
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("FetchMigrationWithRoots: %w", err)
		}
		return fetched, nil
	}
	fetched, err := s.repository.FetchMigration(ctx, source, version, hash)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("FetchMigration: %w", err)
	}
	return fetched, nil
}

func (s migrationSource) Fetch(ctx context.Context, source, version string) (modules.Fetched, error) {
	version, err := migrationVersion(version)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("migrationVersion: %w", err)
	}
	fetched, err := s.fetchMigration(ctx, source, version, "")
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("fetchMigration: %w", err)
	}
	if err := checkFetched(source, version, fetched); err != nil {
		return modules.Fetched{}, fmt.Errorf("checkFetched: %w", err)
	}
	s.remember(fetched)
	return fetched, nil
}

func validateRuntimeLock(module v1.Module, lock v1.Lock, metadata map[migrationRevision]v1.Module) error {
	if err := lock.Validate(); err != nil {
		return fmt.Errorf("Validate: %w", err)
	}
	if err := validateRuntimeRequirements(module, lock); err != nil {
		return fmt.Errorf("validateRuntimeRequirements: %w", err)
	}
	for _, entry := range lock.Modules {
		dependency, ok := metadata[revisionKey(entry)]
		if !ok {
			return fmt.Errorf("selected dependency %s at %s has no verified metadata", entry.Source, entry.Commit)
		}
		if err := validateRuntimeRequirements(dependency, lock); err != nil {
			return fmt.Errorf("validateRuntimeRequirements: %w", err)
		}
	}
	return nil
}

func validateRuntimeRequirements(module v1.Module, lock v1.Lock) error {
	err := modules.ValidateRequirements(module.Requires, lock)
	if err == nil {
		return nil
	}
	for _, requirement := range module.Requires {
		if legacyPseudo.MatchString(requirement.Version) {
			return fmt.Errorf("ValidateRequirements: %s retains legacy pseudo-version %s@%s; manual migration is required to replace this metadata requirement with a full commit or verified semantic tag: %w", module.Name, requirement.Module, requirement.Version, err)
		}
	}
	return fmt.Errorf("ValidateRequirements: %s metadata is incompatible with the native lock; manual migration is required: %w", module.Name, err)
}

func checkFetched(source, version string, fetched modules.Fetched) error {
	if fetched.Inspection != nil && fetched.Inspection.Provisional {
		return fmt.Errorf("dependency %s still has provisional sources; roots and integrity must be finalized before migration", source)
	}
	if fetched.Lock.Source != source || fetched.Module.Name != source {
		return fmt.Errorf("resolved dependency identity differs from %s", source)
	}
	if version != "" && fetched.Lock.Version != version {
		return fmt.Errorf("resolved version for %s differs from historical pin %s", source, version)
	}
	if v1.IsCommitRef(version) && !strings.EqualFold(version, fetched.Lock.Commit) {
		return fmt.Errorf("resolved commit for %s differs from historical pin %s", source, version)
	}
	if len(fetched.Module.Replaces) > 0 {
		return fmt.Errorf("dependency %s has local replacements; full lock migration requires manual migration", source)
	}
	for _, requirement := range fetched.Module.Requires {
		if err := validIdentity(requirement.Module); err != nil {
			return fmt.Errorf("validIdentity: %w", err)
		}
		if _, err := migrationVersion(requirement.Version); err != nil {
			return fmt.Errorf("migrationVersion: %w", err)
		}
	}
	if err := (v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}).Validate(); err != nil {
		return fmt.Errorf("Validate: %w", err)
	}
	return nil
}

func checkPinRequirement(ctx context.Context, req v1.Requirement, pin v1.LockedModule, source migrationSource) (string, error) {
	version, err := migrationVersion(req.Version)
	if err != nil {
		return "", fmt.Errorf("migrationVersion: %w", err)
	}
	if version == "" {
		return "", nil
	}
	if v1.IsCommitRef(version) {
		if !strings.EqualFold(version, pin.Commit) {
			return "", fmt.Errorf("incompatible requirement %s@%s: historical commit is %s", req.Module, version, pin.Commit)
		}
		return "", nil
	}
	if semver.IsValid(pin.Version) {
		if semver.Compare(pin.Version, version) < 0 {
			return "", fmt.Errorf("incompatible requirement %s@%s: historical version is %s", req.Module, version, pin.Version)
		}
		return "", nil
	}
	// A SHA-only pin has no semantic ordering. Only a tag resolving to that exact
	// commit proves compatibility; do not infer ancestry or silently repin it.
	tag, err := source.Fetch(ctx, req.Module, version)
	if err != nil {
		return "", fmt.Errorf("Fetch: %w", err)
	}
	if !strings.EqualFold(tag.Lock.Commit, pin.Commit) {
		return "", fmt.Errorf("incompatible requirement %s@%s: tag differs from historical commit %s; reconcile the pins manually", req.Module, version, pin.Commit)
	}
	if tag.Lock.Hash != pin.Hash {
		return "", fmt.Errorf("incompatible requirement %s@%s: content hash differs at historical commit %s; reconcile the pins manually", req.Module, version, pin.Commit)
	}
	return version, nil
}
