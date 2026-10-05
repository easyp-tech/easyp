# Documentation Workflows

This file contains all documentation generation, staleness checking, and post-pipeline maintenance workflows. Referenced from `SKILL.md`.

---

## Standalone Documentation Workflow

**Trigger:** the user requests documentation generation or update **without referring to a feature**. Examples:
- *"Generate the project documentation"* / *"сгенерируй документацию"*
- *"Update the docs"* / *"обнови документацию"*
- *"Refresh AUTH.md"* / *"actualize the architecture doc"*

**DO NOT** run `pipeline.sh init <feature>` for these requests. Documentation generation is a **standalone workflow** with its own state machine — `.spec/.docs-queue.kv` — driven by `pipeline.sh docs-*` commands.

### Step 1: Detect intent

| User intent | Command |
|-------------|---------|
| Bootstrap docs from scratch (no `<docs_dir>/` yet) | `pipeline.sh docs-init --all` |
| Update stale docs only | `pipeline.sh docs-init --update` |
| Regenerate specific docs (user named them, e.g. "regenerate AUTH and API") | `pipeline.sh docs-init auth api` |
| User unsure | Run `pipeline.sh docs-check`. If `exists: false` → propose `--all`. If `stale: []` → propose `--update`. Wait for user confirmation. |

### Step 2: Initialize the queue

Run the chosen `docs-init` command. It writes `.spec/.docs-queue.kv` with the list of templates to process. If a queue already exists, the command errors — run `pipeline.sh docs-reset` first if you intend to start over.

Then verify with `pipeline.sh docs-status` (returns JSON with `total`, `completed`, `pending`, `current`, `mode`).

### Step 2.5: Evaluate Execution Strategy

Documentation templates are **independent** (no shared state) and **heavy** (each requires reading dozens of source files). This makes parallelization safe and beneficial.

**Default: SUBAGENT mode** — recommended whenever your toolset supports subagent dispatch (e.g. Claude Code `Task` tool, Cursor Composer, GitHub Copilot subagents, or equivalent).

**Fallback: SEQUENTIAL mode** — used only when subagent dispatch is unavailable.

**Skip mode selection if `total = 1`** — sequential always (overhead of dispatch not justified).

#### Subagent mode (recommended default)

You become the **controller**. Loop:

1. Read `pipeline.sh docs-status` to see pending templates.
2. **Dispatch up to 3 subagents in parallel** (rate-limit + controller context safety). Each subagent gets one template via this **context package**:
   - Path to template file (e.g. `./templates/docs/auth.md`)
   - `docs_dir` value (where to save output)
   - `rules.docs` from `.spec/config.yaml` (if set)
   - `context` from `.spec/config.yaml` (if set)
   - One- or two-sentence summary of project stack (language, framework, key patterns)
3. **Receive each subagent's report** — it should report which file(s) were created and where.
4. **Verify** each generated file (controller, never delegated):
   - File exists at the expected path
   - **Line 1** matches `<!-- generated: YYYY-MM-DD, template: <name>.md -->`
   - File is non-trivial (≥ 50 lines)
5. **If verification fails**: re-dispatch with the verification error in the context package. Maximum **2 retries per template**, then escalate to the user.
6. **If verification passes**: call `pipeline.sh docs-done <template>`.
7. Repeat from step 1 until `docs-next` reports the queue is complete.
8. Run `pipeline.sh docs-reset` to clear the queue.

**Subagent rules:**
- One subagent per template — never bundle multiple templates into one dispatch.
- Subagents do NOT interact with `pipeline.sh` — only the controller does.
- Up to 3 in parallel (no more — controller context fills up receiving multiple large outputs at once).
- Verification is always performed by the controller.

#### Sequential mode (fallback)

Loop:

1. Run `pipeline.sh docs-next` — it prints the next pending template name and path (tab-separated).
2. Read the template file from the printed path.
3. Read `rules.docs` and `context` from `.spec/config.yaml` (if set).
4. Generate the documentation file(s) following the template's instructions. Save to `<docs_dir>/`. Verify line 1 has the `<!-- generated: ... -->` metadata.
5. Run `pipeline.sh docs-done <template-name>`.
6. Repeat until `docs-next` reports the queue is complete.

**Fresh-chat hint:** after position 3, `docs-next` automatically prints a hint suggesting you start a fresh chat to avoid context exhaustion. The queue persists in `.spec/.docs-queue.kv` — in a new chat, run `pipeline.sh docs-status` to see your position and continue with `docs-next`.

### Step 3: Auto-trigger from `docs-check`

When `pipeline.sh docs-check` reports `stale: [...]` non-empty (during pre-pipeline check or on-demand), the workflow is:

1. Inform the user: *"Found N stale doc(s): X, Y, Z. Initialize update queue? Say 'update docs' or 'skip'."*
2. If user agrees → `pipeline.sh docs-init --update` → proceed to Step 2.5 (execution strategy).
3. If user declines → no action. The user can run `pipeline.sh docs-init --update` later.

The script does not auto-create the queue — explicit user consent is always required to begin a regeneration cycle.

---

## Documentation Context

The skill supports a self-documenting mechanic via a project documentation directory (default: `.spec/`, configurable via `docs_dir` in `config.yaml`).

> **Directory separation:** Project documentation files (README.md, ARCHITECTURE.md, DOMAIN.md, etc.) live in `<docs_dir>/` (default: `.spec/`). Pipeline phase artifacts (explore.md, requirements.md, design.md, etc.) live in `.spec/features/<feature>/`. **Never place project documentation into the feature directory or vice versa.**

### Pre-pipeline check

When running `pipeline.sh init <feature-name>`, before starting the Explore phase:

1. Determine the docs directory: read `docs_dir` from `.spec/config.yaml`. If not set, default to `.spec`.
2. Run `pipeline.sh docs-check` to determine if documentation exists and check freshness.
3. If the docs directory **exists and contains `README.md`**:
   - Read `README.md` for the documentation map
   - Use available docs (`ARCHITECTURE.md`, `PACKAGES.md`, etc.) as supplementary context for ALL phases
   - This is richer than `config.yaml` context and reduces the file-read budget in Explore phase
   - Check the `stale` array in docs-check output. If stale files exist, suggest: *"Some docs are outdated (<file>: <N> days old). Regenerate before starting? Say 'update docs' or 'skip'."* If user agrees, follow the **Standalone Documentation Workflow** above (`pipeline.sh docs-init --update`).
4. If the docs directory **does not exist**:
   - Suggest to the user: *"Project documentation (<docs_dir>/) not found. I can generate it to better understand your codebase. Say 'generate docs' or 'skip'."*
   - If user says **"generate docs"**: follow the **Standalone Documentation Workflow** above (`pipeline.sh docs-init --all`).
   - If user says **"skip"**: proceed with the pipeline normally — documentation is NOT required
   - **This is a soft suggestion, not a blocker.** The pipeline works without documentation.

### Stale doc regeneration workflow (legacy ad-hoc)

> **Prefer the Standalone Documentation Workflow** (top of this file) which uses the `docs-init --update` queue. The ad-hoc steps below remain as a reference for manual single-file regeneration.

When `pipeline.sh docs-check` reports stale files (or the user requests a doc update), follow these steps:

1. Parse the `docs-check` JSON output — read the `stale` array.
2. For each stale file, extract the `template` field from its freshness metadata.
3. Group stale files by template (one template may generate multiple files).
4. For each affected template:
   a. Read the template from `./templates/docs/<template>.md`.
   b. Read the existing generated file(s) as baseline — preserve project-specific content where possible.
   c. Regenerate following the template instructions.
   d. Update the freshness metadata: `<!-- generated: YYYY-MM-DD, template: <name>.md -->`.
5. Present updated files to the user for review before saving.
6. **Never auto-overwrite.** Always confirm with the user.

Use this lookup table to find the owner template for any generated file:

| Generated file | Owner template |
|----------------|----------------|
| `README.md`, `agent-rules.md` | `bootstrap.md` |
| `AGENTS.md` | `agents-index.md` |
| `ARCHITECTURE.md`, `PACKAGES.md`, `DOMAIN.md`, `CODE_STYLE.md` | `core.md` |
| `TOOLS.md`, `TESTING.md`, `FILES.md` | `development.md` |
| `ERRORS.md` | `errors.md` |
| `AUTH.md`, `OAUTH.md` | `auth.md` |
| `DATABASE.md` | `database.md` |
| `API.md` | `api.md` |
| `DEPLOYMENT.md` | `deployment.md` |
| `SECURITY.md` | `security.md` |
| `CLIENTS.md` + per-client docs | `clients.md` |
| `FEATURE_FLAGS.md` | `feature-flags.md` |
| `BACKGROUND_JOBS.md` | `background-jobs.md` |
| `<COMPONENT>.md` (infra) | `infrastructure.md` |
| `CLI.md` | `cli.md` |
| `STATE.md` | `state-management.md` |
| `EVENTS.md` | `events.md` |
| `COMPONENTS.md` | `components.md` |
| `ROUTING.md` | `routing.md` |

### Documentation generation templates

Templates for generating project documentation are in `./templates/docs/`. Read the manifest (`./templates/docs/README.md`) to discover available templates. When generating docs:
- Apply `rules.docs` from `config.yaml` (if present) as additional rules
- Apply `context` from `config.yaml` as background knowledge
- Each template is self-contained and generates one or more files in `<docs_dir>/`
- **Freshness metadata**: when generating or updating any file in `<docs_dir>/`, MUST add `<!-- generated: YYYY-MM-DD, template: <template-name>.md -->` as the **first line** of the file (before the title). This enables `pipeline.sh docs-check` to track documentation age and detect stale files.
- **Freshness metadata validation**: after saving a generated doc, verify that line 1 matches the pattern `<!-- generated: YYYY-MM-DD, template: <name>.md -->`. If the metadata is missing or malformed, fix it immediately — `docs-check` will silently skip files without valid metadata.
- **Content-aware staleness**: `pipeline.sh docs-check` uses scope metadata from templates (`<!-- scope: ... -->` first line) combined with `git log --since=<generated_date>` to determine staleness. A doc is marked stale only if (a) files matching its template's scope patterns were changed since generation **and** (b) the doc exceeds the freshness threshold. Docs whose scope shows no changes remain fresh regardless of age. If a template has no scope line, the check falls back to pure age-based staleness. The JSON output includes a `scope_changed` field (`true`/`false`/`null`) per file.

---

## Documentation Maintenance

After the pipeline reaches `phase=done` and artifacts are published, check if project documentation needs updating.

### Step 1: Identify affected docs

Read the design document §2.3 ("Files Requiring Changes" table). Match changed file paths against this pattern table:

| Changed file pattern | Affected doc | Owner template |
|----------------------|-------------|----------------|
| `*domain*`, `models/*`, `types/*`, `*entity*` | `DOMAIN.md` | `core.md` |
| new directory under `internal/`, `pkg/` | `PACKAGES.md` | `core.md` |
| `cmd/*`, new service, layer changes | `ARCHITECTURE.md` | `core.md` |
| `*_test*`, `__tests__/`, test config files | `TESTING.md` | `development.md` |
| `Makefile`, `Taskfile`, `scripts/*`, CI tool changes | `TOOLS.md` | `development.md` |
| `*error*`, `*errs*`, error codes, error types | `ERRORS.md` | `errors.md` |
| `*auth*`, `*oauth*`, `*login*`, `*session*` | `AUTH.md` / `OAUTH.md` | `auth.md` |
| `migrations/*`, `schema*`, `*_repo*`, `*_store*` | `DATABASE.md` | `database.md` |
| `*handler*`, `*route*`, `*endpoint*`, `*.proto`, `openapi*` | `API.md` | `api.md` |
| `Dockerfile`, `.github/workflows/*`, `k8s/*`, `docker-compose*` | `DEPLOYMENT.md` | `deployment.md` |
| `*redis*`, `*kafka*`, `*traefik*`, `*prometheus*`, `*nats*` | `<COMPONENT>.md` | `infrastructure.md` |
| `*client*`, `*frontend*`, `*mobile*` | `CLIENTS.md` | `clients.md` |
| `*cors*`, `*csrf*`, `*rate_limit*`, `*security*`, `*helmet*` | `SECURITY.md` | `security.md` |
| `*feature_flag*`, `*toggle*`, `*experiment*` | `FEATURE_FLAGS.md` | `feature-flags.md` |
| `*worker*`, `*job*`, `*queue*`, `*cron*`, `*scheduler*` | `BACKGROUND_JOBS.md` | `background-jobs.md` |
| new code style rule, naming convention change | `CODE_STYLE.md` | `core.md` |
| `cmd/*`, `cli/*`, `commands/*`, `*cobra*`, `*clap*`, `*click*` | `CLI.md` | `cli.md` |
| `*store*`, `*redux*`, `*bloc*`, `*provider*`, `*zustand*`, `*pinia*` | `STATE.md` | `state-management.md` |
| `*event*`, `*messaging*`, `*pubsub*`, `*subscriber*`, `*producer*`, `*consumer*` | `EVENTS.md` | `events.md` |
| `components/*`, `ui/*`, `design-system/*`, `*widget*`, `atoms/*`, `molecules/*` | `COMPONENTS.md` | `components.md` |
| `routes/*`, `router/*`, `pages/*`, `screens/*`, `navigation/*` | `ROUTING.md` | `routing.md` |

### Step 2: Filter and suggest

1. Collect unique affected docs from the pattern matches.
2. **Filter**: only suggest docs that already exist in `<docs_dir>/`. Do not suggest creating new docs post-pipeline.
3. Present to user: *"This feature touched auth and database files. Update AUTH.md and DATABASE.md? Say 'update docs' or 'skip'."*
4. If user says **"update docs"**: for each affected doc, read its owner template from `./templates/docs/`, regenerate the doc, update freshness metadata.
5. If user says **"skip"**: done, no action.
6. If the docs directory does not exist at all, suggest full generation (same as Pre-flight Checklist step 3).
7. **Never auto-update documentation.** Always ask the user first.
