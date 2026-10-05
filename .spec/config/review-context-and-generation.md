# Context, plugin imports, mixed versions and comment suppressions

## X-19 / X-22: explicit selection and bounded discovery

Default generation selects only the closest easyp.gen.yaml in the current
or a parent directory. It no longer executes every generator discovered below
the working directory. This is an intentional v1 behavior correction.

~~~bash
# Closest consumer project only.
easyp generate

# Exactly the named consumer projects (repeatable flag).
easyp generate --project backend --project frontend

# Explicit opt-in to recursive generation below the invocation directory.
easyp generate --all
~~~

The all and project flags are mutually exclusive. Recursive discovery skips
hidden directories, easyp_vendor, node_modules and nested Git repositories.
A project in a skipped directory can still be deliberately selected with
project. Automatic selection refuses a symlink configuration file; explicit
project selection is required to use it. No name-based directory denylist is
claimed to establish a trust boundary: choosing all explicitly authorizes
recursive generation, including command plugins in the selected tree.

Workspace-relative module paths and inherited Go options use the nearest Git
repository boundary rather than assuming the current directory is the root.
The boundary supports both a .git directory and a worktree .git file. Without
Git, EasyP uses enclosing EasyP config/manifest files to establish a boundary.
The workspace flag explicitly supplies that boundary when a non-Git tree is
otherwise ambiguous. User-supplied project paths, descriptor destinations and
plugin command working directories retain invocation-relative semantics;
plugin out paths remain relative to their generation config.

~~~bash
# Inside backend/deep, find backend/easyp.gen.yaml and repository modules.
easyp generate --descriptor_set_out ./api.pb

# Explicit boundary for a non-Git tree.
easyp generate --workspace ../.. --project ..
~~~

Lint and breaking search upward for the nearest policy within the boundary.
When no ancestor policy exists, one unambiguous outermost descendant policy
may be selected. Multiple independent descendant policies require an explicit
cfg instead of an arbitrary choice. An explicit cfg keeps precedence; an
explicit root is used as the default policy lookup start. Module operations
and get use the nearest ancestor protobuf.mod within the workspace. Discovery
never crosses a nested Git boundary to borrow an unrelated parent's config.

Tests that intentionally relied on recursive generation now request all.
Single-project invocation is tested independently; no negative test was changed
to expect accidental execution of an unselected command plugin.

## X-15: per-plugin import generation and output consistency

The v0 setting is restored under its original YAML spelling, with_imports.
It defaults to false and applies to the individual plugin only. When true,
FileToGenerate includes the transitive dependency descriptors exactly once.
It is independent of descriptor export's include_imports flag. Other plugins
still receive imports as descriptors for linking, without generating them.

~~~yaml
version: v1
plugins:
  - name: go
    out: .
    with_imports: true
    opts: [module=example.com/app]
  - name: python
    out: gen/python
~~~

Do not put differently named Go packages into one output directory. For modules
whose roots both contain v1/, source_relative plus one shared out directory can
be invalid even when the protobuf module graphs are individually valid. The
following explicit layout keeps module identities, Go import paths and physical
outputs aligned. The common module exposes common/v1/common.proto; user and
order expose v1/user.proto and v1/order.proto respectively.

~~~yaml
version: v1
generate:
  modules: [user, order]
  managed:
    enabled: true
    override:
      - file_option: go_package_prefix
        value: example.com/app/gen
      - file_option: go_package_prefix
        module: example.com/user
        value: example.com/app/gen/user
      - file_option: go_package_prefix
        module: example.com/order
        value: example.com/app/gen/order
plugins:
  - name: go
    out: .
    with_imports: true
    opts: [module=example.com/app]
~~~

Outputs are gen/user/v1/user.pb.go, gen/order/v1/order.pb.go and
 gen/common/v1/common.pb.go below a Go module named example.com/app. This exact
shape is generated and compiled in easyp-test. No automatic module-name prefix
is inserted behind the user's explicit options, and no imported code is
implicitly generated when with_imports is absent.

All selected plans stage plugin responses in a shared output bucket. Identical
shared outputs are deduplicated. Different contents for one filename and
conflicting Go package declarations within one generated directory are errors
before generated files are written. Import targets are deduplicated even when
a file is both an explicit input and an import. Insertion points are preserved.
This does not promise rollback of arbitrary plugin side effects, filesystem
errors during final writes, or cleanup of old files from previous invocations.

## X-23: mixed dependency requirement kinds

Semantic versions retain the maximum-required-minimum selection rule.
An explicit commit is exact. It may coexist with semantic requirements only
when their selected tag resolves to that same commit. Different explicit
commits, or a selected tag resolving elsewhere, remain a diagnosed conflict.
No semantic order is inferred from a SHA, commit date or ancestry alone.

A versionless requirement is unconstrained when a semantic version or exact
commit is supplied elsewhere in the graph. Only a still-unconstrained module
uses the existing lock pin for tidy, or current HEAD without a pin. Update
continues refreshing unconstrained HEAD requirements. The result does not
depend on the order of root declarations. Tags remain in the lock when needed
to satisfy semantic constraints, while exact commit constraints are verified
against the selected commit.

Provisional HEAD dependency edges are rebuilt if a stronger constraint changes
the selected revision. Dependencies and errors from an ultimately unselected
HEAD are not retained. Revisions are fetched at most once per resolution,
including failed fetches. Cyclic, unstable versionless constraints are rejected
rather than choosing a graph based on traversal order. The previously added
protection against republished locked tags remains active.

## X-18: per-policy, per-file comment suppressions

Comment suppressions are enabled by default in v1. Set
linters.allow_comment_ignores to false in an effective policy to disable them.
There is no process-global switch. Rules produce diagnostics; the lint engine
applies the policy's suppressions afterwards.

~~~protobuf
// Applies to this declaration and, for a message, its block.
// easyp:disable FIELD_LOWER_SNAKE_CASE
message Example {
  string legacyField = 1;
}

// A trailing directive applies to the annotated declaration's line/block.
message Another {
  string legacyField = 1; // easyp:disable FIELD_LOWER_SNAKE_CASE
}

// Explicit bounded region; it cannot leak into another file.
// easyp:disable MESSAGE_PASCAL_CASE
message old_name {}
message another_old_name {}
// easyp:enable MESSAGE_PASCAL_CASE
~~~

Without a matching enable, disable must annotate a declaration; it is not an
implicit file-wide disable. Nested paired regions for the same rule are balanced.
Individual rule names may be comma- or whitespace-separated. Unknown rule names,
empty directives, unsupported actions and unmatched enables are errors with
source coordinates, even when suppressions are disabled. Only actual parsed
comments are interpreted, never marker-like text inside string literals.

Legacy nolint:RULE and buf:lint:ignore RULE comments remain supported as
annotations. An explanation can follow a space-delimited -- or // separator.
Rule groups are not accepted as comment rule names. Disabling one rule does
not disable unrelated checks or a neighboring module's policy.

## X-25: reusable Git objects for pinned commits

Pinned revisions use a reusable bare object store per Git remote below the v1
cache. The first attempt is a depth-one fetch of the exact commit. Temporary
checkouts borrow local objects; Fetch followed by Install no longer downloads
the same remote history twice. Repeat fetches of a known commit work without
contacting the remote, including when that remote is unavailable.

Servers may reject requests for an unadvertised historical SHA. In that case,
EasyP fetches advertised reachable history as a compatibility fallback, then
still requires the original exact commit. It never substitutes HEAD. Thus a
cold request is not guaranteed shallow on every server. Object stores use
cross-process OS locks with context cancellation; process exit releases them.
Installed contents retain lock-hash verification and corrupt installations are
rejected. Tag lookups are still refreshed so cache reuse cannot hide a retag.

The reproducible benchmark is scripts/benchmarks/git_commit_cache.py in
easyp-test. It uses a read-only loopback Git daemon and 48 revisions with a
256 KiB changing payload; all files and logs are isolated below its out path.
The daemon is always stopped in finally, and the old/new lock hashes must match.
Measured on the development Mac, not a public-network performance guarantee:

| Fixture server | First run, before / after | Warm runs, before / after | Warm remote data operations after |
|---|---|---|---|
| Historical SHA allowed | 1.040 s / 0.244 s | 0.511–0.514 s / 0.075–0.076 s | 0 |
| Historical SHA denied; fallback used | 1.107 s / 0.634 s | 0.502–0.548 s / 0.069–0.071 s | 0 |

In the shallow case the object store contained 5 reachable objects versus 146
in the complete fixture history. In the fallback case it retained all 146.
Both paths produced byte-identical lockfiles compared with the previous CLI.
