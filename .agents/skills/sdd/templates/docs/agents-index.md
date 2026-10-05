<!-- scope: .copilot/*, .cursor/*, .github/copilot* -->
# Agents Index Documentation Template

## What This Generates

- `AGENTS.md` in the project root — an instruction file for AI agents (Copilot, Cursor, Windsurf, Claude, etc.) explaining how to use the `.spec/` documentation directory.

## Instructions

Create `AGENTS.md` in the project root. This file instructs AI agents on how to work with the `.spec/` directory in this repository.

The file must contain the following sections:

### Section 1: What is `.spec/`

Explain that `.spec/` is a project documentation directory optimized for AI agent (LLM) consumption. It contains:
- Structured descriptions of architecture, packages, and domain
- Code and testing conventions
- Infrastructure and tooling descriptions
- Agent rules (`agent-rules.md`)

Purpose: give the agent full project context without having to read the entire source code.

### Section 2: How to Use (for agents)

Instructions for the AI agent:
- When starting work on a project — read `.spec/README.md` for the documentation map
- Before modifying code — read the relevant document from `.spec/` (e.g., before modifying API — read `ARCHITECTURE.md` and `CODE_STYLE.md`)
- Always follow rules from `agent-rules.md`
- If a document appears outdated — suggest an update

### Section 3: File Structure Convention

Describe file naming conventions in `.spec/`:
- `README.md` — index of all documents, Quick Facts, project structure, ports, commands
- `agent-rules.md` — mandatory rules for the agent (code style, naming, error handling)
- `UPPER_CASE.md` — topical documents (`ARCHITECTURE.md`, `TESTING.md`, etc.)
- `skills/` — directory for agent skills (custom scripts)
- `workflows/` — directory for agent workflows
- `prompts/` — prompts for (re)generating documentation

### Section 4: Document Categories

List the recommended document categories:
- **Core**: Architecture, Packages, Domain, Code Style
- **Development**: Tools, Testing, Files/Storage
- **Auth & Security**: OAuth, API Keys, Permissions
- **Infrastructure**: Observability, Caching, Proxy, Queues, Leader Election
- **Clients**: Frontend apps, Mobile, Bots

### Section 5: How to Maintain

Rules for keeping documentation current:
- When adding a new component — update or create the corresponding document
- When changing architecture — update `ARCHITECTURE.md`
- When adding a dependency — update `PACKAGES.md`
- `README.md` must always reflect the current list of documents
- Documents must contain code examples from the actual project, not abstract ones

### Section 6: How to Add New Document

Steps for adding a new document:
1. Create a file `TOPIC_NAME.md` in `.spec/`
2. Use this structure: title → overview → architectural diagram (ASCII) → details → code examples → configuration → testing → key files
3. Add a link in `README.md` under the appropriate category
4. If the topic is project-specific — use real code examples

## General Rules

- Format: Markdown with ASCII diagrams, tables, fenced code blocks with language tags
- Style: concise, structured, no filler. Optimized for quick scanning.
- Language: English
- All code examples must come from the actual project source, not invented
- If a section does not apply to the project, omit it
