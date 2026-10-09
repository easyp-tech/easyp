# Explicit generation config implementation plan

> Execute inline on the user-requested `fix/nightly-v1-regressions` branch.
> Use test-driven development and verification-before-completion.

**Goal:** Select `private.easyp.gen.yaml` or another native generation file
without relocating it or changing relative source/output paths.

**Architecture:** Add one CLI file selector, `--gen-config`, and forward it
through `generation.Request.GenConfig`. The existing generation pipeline owns
parsing, workspace boundaries, module selection, option inheritance, frozen
validation, descriptor preparation and plugin execution. File selection must
not change those responsibilities or introduce legacy command fallback.

**Tech stack:** Go 1.26.6, urfave/cli v2, testify, existing v1 generation engine.

## Accepted behavior

- `easyp generate --gen-config private.easyp.gen.yaml` selects exactly that file.
- Relative selector paths start at the invocation working directory; absolute
  paths work within the existing workspace boundary.
- Module selectors retain their existing module/workspace coordinates. Plugin
  output paths remain relative to the selected config's directory. Existing
  plugin executable/command path rules are unchanged.
- The file is parsed as native v1 generation config. Missing files and invalid
  configs fail with their original error causes; no standard-config fallback.
- The selector is exclusive with `--project` and `--all`; an explicitly empty
  CLI value is rejected. Only one named file is selected per invocation.
- An explicitly selected config does not trigger suggestions or discovery of
  unrelated child generators merely because it has no executable targets.
- Normal discovery, ancestor-only option inheritance and frozen mode retain
  their current behavior. Global `--cfg`/`--config` remains separate.
- Explicit selection and reading must use `Request.WorkspaceRoot` as the bound,
  preserving logical paths and rejecting both lexical and symlink escapes.

## Task 1: Add behavior regressions first

**Files:**
- Create `internal/api/generate_config_test.go`.
- Extend generation tests only where API tests cannot cover the request contract.

- [x] Add CLI tests for relative, absolute and ancestor-relative named files,
      same-directory profiles and config-relative output in another directory.
      Use real native modules and builtin Python generation; verify output
      content/location and that the standard profile did not run.
- [x] Add missing-file, invalid-v1, empty-selector and conflicting-selector
      cases. Preserve filesystem/parser errors and verify no plugin outputs.
- [x] Distinguish cwd/config/workspace coordinates, reject workspace escapes,
      prove malformed canonical siblings are ignored, and preserve canonical
      ancestor-only option inheritance.
- [x] Cover an empty explicit profile next to an unrelated child generator.
- [x] Run new tests with Go 1.26.6 and `-race -count=1`; record failures caused
      by the absent `--gen-config` flag before implementation.

## Task 2: Wire explicit selection through the existing pipeline

**Files:**
- Modify `internal/api/generate.go`.
- Modify `internal/generation/generate.go`.
- Modify `internal/generation/discovery.go`.

- [x] Register a `TakesFile` string flag and reject an explicitly empty value.
- [x] Give each Generate command fresh flag instances: parsing mutates flag
      defaults, so package-level flag pointers race across independent CLI apps.
- [x] Forward it through `Request.GenConfig`.
- [x] Check selector conflicts, normalize the explicit file path, and select it
      before normal discovery. Validate it with bounded source resolution and
      read config bytes within the requested workspace, retaining logical paths.
- [x] Treat the file selector as an explicit choice in empty-generation handling.
- [x] Run the new tests; all behavior regressions must pass.

## Task 3: Document the public CLI contract

**Files:**
- Modify `README.md`.
- Modify `.spec/CLI.md`.
- Modify `V1_RELEASE_NOTES.md`.

- [x] Show colocated `private.easyp.gen.yaml`/`public.easyp.gen.yaml` examples.
- [x] Document coordinate rules, exclusive selectors, unchanged default discovery
      and the distinction from global policy `--config`.

## Task 4: Verify and review

- [x] Run `go test -mod=readonly -race -count=1 ./internal/api ./internal/generation`.
- [x] Build a fresh CLI into a temporary directory from this branch.
- [x] On temporary copies of the full-contract reproduction, select a named Go
      profile, verify default-profile isolation, compile the SDK with Go 1.26.6
      and run the consumer. Do not migrate or regenerate the user's copies.
- [x] Run appropriate full source checks and the existing linter; preserve the
      repository's dependency/toolchain versions.
- [x] Inspect `git diff --check`, review the final diff and report actual checks.

No tag, release or merge is part of this change. Preserve the existing branch's
nightly regression fixes and all unrelated user files.

## Verification results

- Independent plan review: approved after adding explicit workspace bounds.
- New CLI regressions: 19 subcases passed with Go 1.26.6 and `-race -count=1`.
- API and generation packages passed their complete race suites.
- Full source race suite: 851 top-level tests, 3,076 passing test/subtest actions,
  zero failures or skipped tests.
- golangci-lint: `0 issues.`; `git diff --check` passed.
- Fresh CLI ran a named Go profile beside a malformed standard profile on a
  temporary full consumer project. Consumer and Git dependency generated three
  SDK files, compiled with Go 1.26.6/race and produced the expected runtime JSON.
- Independent implementation review: approved, no required fixes.

Evidence is saved outside the repository in
`/var/folders/fr/wwbw9ytn1c11rqztj4vn1w9h0000gn/T/easyp-gen-config-20261009-0jtf0_ii/verification.json`.
The results describe the verified working tree before commit. The requested
branch is retained; merging and releasing remain separate operations.
