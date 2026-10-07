# External dependency import roots

Approved contract: the 2026-10-07 discussion with the user. Extend resolution for dependencies without root metadata, preserve verified resolved roots in protobuf.lock, accept legacy Git root/sub_directory through verified migration, and retain the existing generation module selectors.

## Problem and boundary

Migration currently rejects every non-default git_repo.root/sub_directory before adding the Git input to require and generate.modules. Native dependency roots are otherwise recovered from pinned protobuf.mod, Buf or legacy EasyP metadata. A repository without roots falls back to `.`.

A root changes a protobuf filename; a generation path only selects files. Identical api/svc/v1/svc.proto bytes compile as either svc.proto or svc/v1/svc.proto with different valid roots. Therefore roots cannot be guessed from the directory name, package or common directory alone. A root hint that is not derivable from a repository must survive a cold cache as resolution metadata.

Scope is one cohesive dependency/source namespace change. Keep Go 1.26.6 and existing dependencies. No v0 command fallback, new consumer-manifest roots section, source rewrites, release tags, or modifications to the real customer repositories are involved. Legacy readers remain in dependency compatibility and migration.

## Root authority and result

Native protobuf.mod (including its implicit default `.`), Buf and legacy EasyP roots remain authoritative. Malformed metadata is an error, never permission to infer another namespace. A pre-native requirements-only manifest or a legacy policy with no root declarations does not establish roots. Track this provenance separately from the value `.`.

For a metadata-free module, roots may come from an explicit resolution hint or from uniquely consistent import/file constraints. Otherwise retain the current default `.` when imports require no adjustment. Multiple disjoint roots are supported. Inference that permits multiple valid module/file layouts must return a diagnostic naming alternatives; never choose by iteration order.

Add optional `roots` to each v1 LockedModule. Each root is a canonical portable module-relative directory, `.` is permitted, traversal/absolute/glob/backslash spellings and duplicate entries are rejected. Explicit CLI hints are canonicalized/deduplicated before writing. Keep existing source/version/commit/hash/BSR fields and source hashing. Record roots that are inferred or explicitly supplied for a metadata-free dependency; ordinary authoritative metadata need not be duplicated. A supplied hint equal to authoritative roots is harmless; a conflicting hint is an error.

The installed snapshot remains a verified immutable source tree. Cached(entry) returns a fresh effective Module whose roots reflect the entry, without writing a generated manifest into the cache. Cold installation must receive the same recorded roots before selected alias validation. Include the canonical root selection in the installed-snapshot cache identity: verification for one root scope cannot authorize a different scope containing previously ignored invalid aliases. Git objects and content hashing retain their existing policy; contexts may share object bytes but never mutable effective metadata.

## Resolution

Use only dependencies already present in the declared/resolved graph. Add an explicit bounded two-stage source inspection capability before installed snapshots are accepted. A roots-aware provisional fetch returns revision metadata plus an immutable logical proto index and path-resolution problems from that pinned tree; it does not install that provisional view. The index owns the source bytes/identities needed after temporary checkout cleanup. Metadata is still parsed strictly. For a metadata-free unresolved namespace, selected-alias failures are recorded rather than immediately applying default `.` to every alias. Graph root selection runs over the inspections, and the final cold installation replays the selected roots with strict validation. A recorded problem inside a selected root is an error; an unrelated escaping/dangling alias outside it remains unused. This prevents consumer-only evidence from arriving after an already-failed Fetch/Install.

Index logical proto paths through existing bounded source traversal, excluding hidden/vendor/nested-module boundaries. An unresolved import `svc.proto` and a logical source `api/svc/v1/svc.proto` imply candidate root `api/svc/v1`; matching is component bounded and preserves the complete import path.

Known metadata roots and explicit/previous locked root choices are fixed constraints. New fallback choices must satisfy all incoming imports and the reachable import closure without introducing duplicate import paths, physical-source aliases or inconsistent namespaces. Apply candidate roots through the shared source model, then run existing collision/import checks. Accept one consistent layout; reject none or multiple. Bound candidate exploration and report ambiguity instead of expensive unbounded search.

Intrinsic imports of a metadata-free repository can provide evidence even when a consumer has not imported it yet. Inspect import declarations without making malformed unrelated message bodies a new dependency validation gate. Unknown unrelated imports cannot introduce dependencies or invalidate an otherwise unused file; reached/selected files still receive full compiler validation. Intrinsic and consumer constraints use one root inference component. Fetch may use intrinsic evidence before strict selected-alias materialization; graph resolution can complete the choice from consumer imports.

Normal get/tidy/update write successful resolution results using the existing manifest/lock transaction. Reuse roots from an unchanged pin; explicit new hints may replace a previous fallback choice after validation. On a changed revision, revalidate the prior choice; newly authoritative metadata wins after matching verification, and an invalid choice must be re-resolved or reported before files are written. Frozen never guesses a new namespace or writes files: it uses authoritative metadata or recorded roots from verified locked contents. Local replace remains forbidden in frozen. Ordinary local overlays may infer roots in memory but never write them into the published lock.

## User entry points

Add repeatable `get --import-root <directory>` for the single requested dependency. The paths refer to its module directory, not the consumer cwd or import-root namespace. Validate paths before dependency access. Frozen get remains rejected. For a replaced dependency, validate explicit hints against the replacement's effective source view without rewriting the shared lock.

Generation already selects a required dependency by its identity in generate.modules. No new generation roots option is introduced. All module roots participate when no generation selectors are supplied. Per-module paths/packages retain their current intersection behavior, and paths remain relative to the module directory.

## Migration

Initial read-only preview accepts bounded literal Git roots and sub_directory and adds the Git URL/version to require. Git inputs become module-specific generation selections when sub_directory limits their targets. Local directory selectors belong to the local module entry rather than restricting every Git input globally. Repeated identical selections deduplicate; incompatible selections for one module are diagnosed before plugins or writes.

Dependency verification remains behind --resolve-lock or the wizard's existing explicit confirmation. Use the recorded historical pin and validate the legacy content hash; do not replace a pin with HEAD. Treat the old root as a resolution hint, and compare the legacy installed-source/import-name map with the v1 pinned snapshot. Account for legacy archive root rewriting instead of equating archive-relative names with repository-relative names. Validate sub_directory selection similarly and translate it to module-relative paths. Generation targets, reachable imports and filenames must match. Normalize equivalent spellings and remove redundant roots/paths only when equality is proved.

Show proposed corrections in the verified preview. A semantic repair changing an import name or target set must be called out and must not bypass existing apply confirmation. Unrepresentable or ambiguous layouts return a contextual diagnostic. All failures keep consumer manifests/locks byte identical. Existing backups, immutable legacy easyp.lock, source-state rechecks and rollback behavior remain.

Migration repositories get an optional roots-aware verification operation and an explicit source-access capability: compose that operation with modules.Cache or return verified legacy/v1 logical mapping data in the inspection. Existing fake repositories and root-free migration remain supported through the small existing interfaces; explicit hints require capability rather than silently being ignored. A new explicit hint replaces a previous fallback before that choice becomes a fixed constraint.

## Verification

Test parsing/schema/canonical roots; authoritative native/Buf/EasyP precedence; unique and ambiguous inference; multiple roots; overlapping names; unused malformed protos; unrelated imports; root-changing get/tidy/update; failure transaction preservation; CLI validation/frozen guards; bounded symlinks and unused invalid aliases; two consumers sharing snapshot bytes; cold-cache/frozen replay; local overlay behavior; and separate current/baseline breaking locks.

Add real Git migration fixtures with historical hashes, custom root and sub_directory, mixed local/Git selections and duplicate inputs. Compare descriptor names and FileToGenerate, generated paths/bytes and compile Go SDK outputs. Update standalone tests and docs. Build fresh CLI/MCP with Go 1.26.6, run relevant and then full source race tests, schema check, lint, and standard v1 easyp-test using EASYP_BIN/EASYP_MCP_BIN/EASYP_SOURCE from the new checkout. Review the final diff before pushing branches/PRs. Tagging a new release requires a later explicit user request.
