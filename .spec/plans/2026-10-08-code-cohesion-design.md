# EasyP v1 code cohesion refactor — proposed design

Status: approved by the user on 2026-10-08; independent design review approved.
Source baseline: 898bb664e273dc8a91fbb85eb7c48dc5dfd25d95, easyp PR #236 (open, CI test passed).
Scope: the recently added import-root resolution, checked tidy import repairs, migration selection proof, and their adjacent write paths.

## Objective

Make each unit understandable from its inputs, result and one responsibility. Reduce direct access to another component's mutable state. Keep today's CLI, configuration, resolver, lock/cache integrity and generation behavior.

## Evidence from the current source

- internal/modules/operations.go:51 returns both the lock and *importRootSource from resolveV1Graph, with positional preserveHeads/hints/tidy arguments.
- internal/modules/import_roots_resolution.go:21 owns repository access, old pins, hints, current fetched revisions, verified old scopes and a tidy mode flag. Finalization mixes scoped fetch, integrity proof, root inference and namespace transition policy.
- internal/modules/import_rewrites.go:40 coordinates input capture, cache boundaries, resolution, namespace bindings, edits, compilation, manifest augmentation and commit. Its binding planner is a method on importRootSource at line 269.
- internal/modules/import_rewrites_validation.go:37 reads the resolver's fetched map directly; the source view also manages dependency metadata observations, previous inventories, proposed bytes and transaction child scopes.
- Several import rewrite helpers directly read tx.expected or append tx.inputs. Callers therefore know how observations and ownership are stored.
- internal/migration/migration.go:144 Build interleaves capture, conversion, candidate preflight, historical dependency proof, source proof and final rendering. Plan exposes the local selection proof as five separate fields and manually synchronizes outputs with transaction changes at line 380.
- internal/adapters/gitmodules/migration_mapping.go:24 receives native/selection booleans, while roots-aware migration distinguishes nil from a non-nil empty roots slice. These choices are correct but insufficiently explicit at call sites.

These are maintainability observations, not newly confirmed behavior bugs. File length alone is not a refactoring criterion.

## Alternatives

1. Recommended: incrementally introduce cohesive private collaborators and explicit result values inside existing packages. This targets the observed data coupling while preserving package and public-port boundaries.
2. Rename and extract helpers only. Lower initial risk, but the resolver and transaction state still leak into callers; useful locally, insufficient for this goal.
3. Move roots, namespaces and filesystem transactions into new packages. Stronger compiler boundaries, but a larger interface/export diff and risk around snapshot proofs. Reconsider only if the first approach exposes a genuinely independent reusable domain.

## Proposed units and data flow

### Dependency resolution and revision evidence

Keep pure intrinsic root selection with its index, constraints and bounded search together. Give repository coordination one explicit request value for pin preservation, root hints and transition policy, replacing opaque boolean argument lists.

Return a concrete checked resolution result containing the selected lock and verified per-revision source observations. Only final, exact-pin/hash-verified observations can enter that result. The result owns nested inspection bytes, identities and roots; narrow operations must not expose mutable backing maps or slices. Consumers obtain those observations through narrow operations; they do not receive *importRootSource or read its fetched/locked maps.

Move old/new namespace repair planning off importRootSource. The repair planner consumes the checked current result, previous lock and the existing pinned-source capability. Keep get/update transition guards explicit in their own orchestration and keep tidy's repair eligibility checks explicit in its orchestration. Removing the tidy boolean must never make transition checks optional by accident.

Use a named revision key instead of [2]string. Preserve the current normalization and cache identity; this is not a new cache format.

### Tidy input observations, repair planning and validation

Make TidyWithReport a readable orchestration of capture, resolve, plan, validate and commit. Use concrete private units:

- Consumer input observations own the original manifest/lock, selected files, root spelling and capture timing.
- An import repair planner owns verified old/current bindings and token-only edit proposals.
- A proposed source view owns proposed bytes and the reachable-input observations required by compilation.
- The existing transaction owns recorded file states, staged changes, rechecks, commit and rollback.

Use narrow transaction operations for reading captured bytes/physical targets, attaching read-only dependency observations and proposing changes. Keep its expected/changes/inputs storage private to the transaction implementation. Do not replace bounded SourceRoots checks with a general callback framework.

Early and late cache ownership checks remain distinct named checkpoints. When the repository exposes complete cache ownership, consumer files are captured before resolution. Generic repositories capture only after new directories are known, and previous-pin fetches add boundaries before subsequent rechecks. This ordering must remain visible.

### Migration planning and proof

Group inputs/roots/packages/paths/source inventory into one local selection proof that owns construction and rechecking. Keep the verified Git selection proof separate.

Extract cohesive migration build stages: capture and parse; derive local and Git selections; render/preflight known candidates; optionally verify historical dependencies and namespaces; finalize outputs. Separate literal selection translation from filesystem namespace observation and source-binding/compilation proof.

Keep Plan as the prepared output plus immutable verification evidence and application gate. Candidate replacement must go through one transaction operation that updates the staged write and exposed preview together; callers cannot manually edit p.tx.changes.

In the Git adapter, convert public roots and native/selection parameters once into a named internal migration request. Preserve the current nil versus non-nil-empty roots semantics, and preserve native initial-selection/archive verification behavior. No public Repository method or Fetched contract needs to change for this step.

## Invariants

- No new CLI/config options, manifest directives, lock/hash/cache formats, dependency upgrades or Go version changes.
- Authoritative roots, intrinsic-only inference, ambiguity/limits, mixed-layout default retention, exact historical pins/hashes and cold/frozen replay remain identical.
- Needed deleted/ambiguous contracts fail before writing. Import literal semantics, comments, CRLF, permissions, aliases, SDK filenames and generation selections stay unchanged.
- Get/update retain namespace-adoption guards. Tidy retains ephemeral local replacement validation and reports only committed repairs.
- Keep source/metadata/absence/alias identity evidence through the last pre-commit recheck. Preserve rollback and recovery behavior.
- Preserve verifyCapturedInputs -> cache verification -> verifyCapturedInputs around potentially repairing cache operations. Retain the post-staging transaction recheck and migration beforeApply after staging, before destination replacement.
- Migration previews remain read-only without explicit resolution, do not run plugins, and retain byte-identical legacy backups and lock.
- Preserve errors.Is/errors.As causes, useful module/pin/root/path context and CLI exit behavior. Add only repository-convention callee wrappers when extraction requires them.

## Validation and acceptance

Run relevant existing race regressions after each small extraction; add a behavioral regression only for a demonstrated uncovered boundary, not to test the new structure itself.

At completion use Go 1.26.6 and -mod=readonly for the full source race suite, vet and pinned lint. Rebuild CLI/MCP and run the full standard v1 standalone suite with all three EASYP variables, including SDK compilation, real Git migration, tidy source edits, cold/frozen replay and imported-contract breaking checks. Check generated schemas for drift. Run opt-in live/Python cases when affected paths or unresolved evidence justify them.

Review the final dependency flow, not only tests: tidy no longer accepts resolver machinery; only the transaction implementation accesses its state maps; migration Plan owns named local/Git proofs; top-level operations expose the ordered phases and early errors. No generic service/repository hierarchy or unrelated project-wide rename.

Apply on a separate refactoring branch derived from the verified baseline; preserve the feature PR. Commit and push the finished authorized work and create a reviewable PR after independent spec/quality gates and passing checks. No merge or release tag is part of this request.
