<!-- generated: 2026-09-30, template: bootstrap.md -->
# Agent Rules — EasyP

Mandatory rules for AI agents. Prefer these over generic Go guides when they conflict. Expanded skills: [go-code-style](../.agents/skills/go-code-style/SKILL.md), [go-testing](../.agents/skills/go-testing/SKILL.md). Dependency details: [config/dependency.md](./config/dependency.md). The bundled skills describe this CLI; the legacy epctl-commands skill name is retained for compatibility, not as an instruction to use the separate service CLI.

## Code Style

- Wrap errors as <code>fmt.Errorf("&lt;called_function&gt;: %w", err)</code> — function/method name only, **no package prefix**.
- Do not assign inside <code>if</code> conditions except <code>:=</code> for block-scoped vars; assign then check on separate lines.
- Never bare <code>defer x.Close()</code> — always wrap Close and log/handle the error.
- No inline comments on <code>if</code> / <code>for</code> / <code>return</code> lines; comments go above if needed.
- Imports: stdlib → third-party → project (repository convention; current lint config does not enable <code>gci</code>); comments and godoc in English; exported symbols need godoc starting with the name.
- Interfaces: context first, error last, no <code>I</code> prefix; prefer defining interfaces in the consumer package.
- Enums: reserve zero with <code>_ = iota</code> so unset values are never silently valid.

## Naming Conventions

- Files: colocated tests as <code>&lt;file&gt;_test.go</code>; lint rules as <code>internal/rules/&lt;rule&gt;.go</code>.
- Types: exported domain/entities and adapter implementations; keep config structs unexported when local.
- Struct tags: follow existing public config spellings, usually <code>snake_case</code>; retain established hyphenated keys such as <code>linters-settings</code> and <code>exclude-rules</code>.
- Do not invent parallel agent instruction files (<code>.cursor/rules/</code>, <code>CLAUDE.md</code>); root <code>AGENTS.md</code> + <code>.spec/</code> are the source of truth.

## Error Handling

- Wrap at every call site with <code>%w</code>; use <code>errors.Is</code> / <code>errors.As</code> for sentinels. The lint configuration exempts identifier-style exported callee labels from ST1005, not capitalized prose errors.
- Reuse existing errors in their owning packages: core parsing/breaking types, <code>modules.ErrLockedVersionChanged</code>, <code>v1.ErrLegacyConfiguration</code>, and contextual migration errors. See [ERRORS.md](./ERRORS.md); do not invent duplicate sentinels.
- CLI: inspect the actual handler and process entrypoint before assigning exit codes. Local replacements preserve the published lock; explicit frozen mode verifies only the published graph and unknown-import discovery is intentionally excluded; there is no universal package-manager sentinel-to-exit mapper.

## Testing

- Use <code>testify</code> (<code>require</code> for fatal preconditions, <code>assert</code> for non-fatal checks).
- New slice-based table tests require a <code>name</code> field. Use <code>t.Parallel()</code> at top level and in isolated subtests; tests that change process cwd/environment must stay sequential. Existing rule tests also use named map keys.
- No shared mutable mocks/state across parallel sub-tests — construct deps per case.
- Prefer <code>require.ErrorIs</code> for expected errors; run with <code>-race</code> (<code>task test</code>).
- Update test doubles when consumer interfaces change. Optional <code>task mocks</code> uses <code>internal/core</code> for <code>Rule</code>/<code>CurrentProjectGitWalker</code> and <code>internal/adapters/console</code> for <code>Console</code>. Preserve handwritten doubles and compile any generated mocks before using them; see [TESTING.md](./TESTING.md).

## Dependencies

- Native v1 proto dependencies come from <code>protobuf.mod</code> <code>require</code> directives, never removed <code>generate.inputs</code>. Lock with <code>protobuf.lock</code>; cache under <code>EASYPPATH</code> (default <code>~/.easyp</code>). Legacy inputs are read only by dependency compatibility/migration adapters.
- Vendor directory is <code>easyp_vendor</code>, not <code>vendor/</code>.
- Use <code>easyp mod download</code> (lock-first) vs <code>easyp mod update</code> (refresh from <code>protobuf.mod</code>); commit <code>protobuf.lock</code> for CI reproducibility.
- No remotes/mirrors/auth fields in <code>easyp.yaml</code> — auth via system git.
- Regenerate schemas via <code>task schema:generate</code>: <code>schemas/easyp-v1.schema.json</code>, <code>schemas/easyp.schema.json</code>, <code>schemas/easyp.gen-v1.schema.json</code>, <code>schemas/easyp.gen.schema.json</code>, <code>schemas/protobuf.lock-v1.schema.json</code>, <code>schemas/protobuf.lock.schema.json</code>. Never hand-edit these outputs.
- Use the implemented descriptor-based FILE/PACKAGE/WIRE_JSON/WIRE catalog; omitted or empty categories default to FILE for v1 policies; use [CLI.md](./CLI.md) for current v1 behavior and the migration wizard.

## Formatting

- Format with <code>gofmt</code> / project tooling; <code>task lint</code> attempts <code>lint:go</code> and <code>lint:docker</code> and reports either failure. Hadolint checks root <code>Dockerfile</code> and requires Docker; <code>task lint:go</code> runs independently.
- Do not commit secrets, generated dependency directories, <code>coverage.out</code>, or the local <code>easyp</code> binary in the repo root. The docs site is maintained in a separate repository.
- After behavior changes, run tests for touched packages; after schema changes, run <code>task schema:check</code>.

## Quick Checklist

- [ ] Error wrap = callee name only; no bare <code>defer Close()</code>
- [ ] Tests: named cases, <code>t.Parallel</code>, no shared mutable mocks
- [ ] Schema JSON regenerated, not hand-edited; vendor = <code>easyp_vendor</code>

Shared producer policies use section-scoped <code>extends</code> from bounded local files or already declared/locked modules. See [.spec/config/policy-extends.md](config/policy-extends.md). No standalone policy versions or unpinned reference downloads are allowed.
