# Migration dependency waits implementation plan

**Goal:** Preserve pinned Git revisions and Buf source-filter proofs while making
slow dependency operations visible and bounded.

## Evidence

Published `v1.0.0-nightly.20261008.1` reproduces the exact missing
`proto/protovalidate-testing/tests/example/v1/example.proto` failure on a local
Buf-v2 multi-module provider, without `easyp.lock`. It also fails when the
provider is transitive to a primary `git_repo.url` pinned with `@<commit>`.
The existing `2305e8a` fix on this branch passes both cases and retains both SHAs.
The file exists in the pinned Git tree but is excluded from the native snapshot.

The internal mirror and its timings are unavailable. A particular network stall
is not established. Git commands currently have no default deadline and capture
output until completion; cache locks can also wait without a default deadline.
These code paths account for the lack of visible diagnostics after `Use storage`.

## Changes

- [x] Add a regression with the reported multi-module/filter/path layout,
      no legacy lock, a SHA-pinned direct or transitive provider, and newer HEAD
      deleting the reported file and changing the imported contract.
- [x] Bound Git commands and cache-lock acquisition with a positive duration
      from `EASYP_GIT_TIMEOUT` (default 5m, accepted by the user).
      Caller cancellation or an earlier deadline wins. Preserve errors.Is for
      context causes, and never retry another candidate or full-history fallback
      after cancellation/deadline failure. Bound inherited command pipes too.
- [x] Add cache-scoped progress logging, supplied by the CLI's existing logger.
      Report module/revision at operation start, elapsed wait periodically, and
      completion status. Do not log the returned error twice. Preserve scoped
      historical-pin resolvers and all existing integrity/cache checks.
- [x] Use failing tests for slow Git, early cancellation, blocked locks, invalid
      timeout settings, and progress lifecycle; real temporary repositories only.
- [x] Document wait controls and distinguish archive verification from native
      source selection; changing a pin or silently dropping sources is not a fix.
- [x] Run relevant race suites, full source checks, linter and a fresh CLI wizard
      reproduction. Keep user workspaces/caches and installed binaries untouched.

Evidence is retained in the task-owned temporary directory, outside the checkout.
No release, merge or source/hash weakening is part of this change.

## Algorithm investigation requirement

The user explicitly requested investigating the hour-long operation, not merely
adding a timeout. A public upstream v1.2.2 migration completed in 2.455s: tag
lookup 0.636s, clone 1.339s, archive 0.008s. A deep/wide synthetic local mirror
with small retained non-proto files took 0.287s/1.992s/13.332s for
100/1,000/4,000 files. Profile local snapshot and legacy proof work to identify
any superlinear traversal before changing the algorithm. Preserve byte/hash,
source-ownership and filesystem-collision guarantees while optimizing proven
hot paths. Add phase debug logs and timings that distinguish Git download,
snapshot materialization, archive proof and source verification. Do not claim
that a specific internal mirror stall was reproduced.


## Final evidence

The pinned filesystem now indexes immutable trees and caches only blob sizes.
Metadata traversal does not retain decoded bodies; a weak-pointer regression
fails with the original body cache and passes with size-only caching while the
filesystem stays live. Concurrent metadata and content reads pass with race
instrumentation. Unix subprocess cleanup is tested when the direct parent exits
successfully or unsuccessfully but descendants retain output pipes. Optional
workspace-identity probing uses the same bounded runner.

An isolated paired CLI comparison, after source verification exited, used
separate cold caches and the same commands. Snapshot hashes matched in every
pair:

| Retained files | Baseline | Final |
| --- | --- | --- |
| 100 | 0.347s | 0.154s |
| 1,000 | 2.257s | 0.435s |
| 4,000 | 13.950s | 1.247s |

The 8.804s final run under concurrent race/lint verification is retained as a
contended measurement, not used for the isolated comparison. Public protovalidate
v1.2.2 resolves in a cold cache in 1.570s to commit
4baed765070e580226a0a30bf2a357cc027dc1f7.

- Go 1.26.6; go.mod and go.sum unchanged.
- Full source race suite: 869 top-level tests, 3,171 passing test actions,
  31 packages, no failures or skipped tests.
- Full standard easyp-test v1 race suite: 208 top-level tests, 968 passing
  test actions, no failures or skipped tests; fresh CLI/MCP and EASYP_SOURCE.
- Pinned direct/transitive providers, absent easyp.lock, newer provider HEAD
  deleting the excluded fixture and changing the contract: pass.
- Actual PTY wizard followed by Python generation: pass; two generated modules
  compile as Python source. Tidy preserves both pinned commits and snapshot
  hashes, adds the indirect requirement and generated lock header as expected,
  and is stable on repetition.
- Linter: zero issues. Windows runner cross-compiles; Windows process behavior
  was not exercised natively.
- Independent implementation review: three findings fixed and re-reviewed;
  no remaining correctness findings.

The internal mirror and the reported hour-long wait remain unavailable. The
confirmed algorithm overhead and timeout gaps do not establish which dominated
that particular run. Optional live_dependencies/python_grpc suites were not
repeated for this fix; the real public cold dependency and standard suites ran.
