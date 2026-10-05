<!-- scope: Makefile, Taskfile*, scripts/*, **/ci/* -->
# Development Documentation Template

## What This Generates

- `.spec/TOOLS.md` — reference of commands and project tools
- `.spec/TESTING.md` — project testing conventions
- `.spec/FILES.md` — file storage system description (only if applicable)

## Instructions

You are a technical documentarian. Create the development documentation in the `.spec/` directory.
Analyze the Makefile, Taskfile.yml, package.json scripts, CI/CD configs, test files, and file adapters.

Create the following files:

---

### File 1: .spec/TOOLS.md

Reference of commands and project tools.

#### Structure:

##### 0. Dev Environment Setup

Prerequisites and first-run instructions for new developers:

**Prerequisites:**
| Tool | Version | Install |
|------|---------|---------|
| Go / Node / Python | x.y | `brew install ...` / link |

Determine from: `go.mod`, `package.json` engines, `.tool-versions`, `Dockerfile` base image, README.

**First Run:**
Numbered steps to go from a fresh clone to a running dev environment:
1. Clone the repository
2. Copy env file (`cp .env.example .env`) — list required variables
3. Install dependencies (`go mod download` / `npm install`)
4. Start infrastructure (`docker-compose up -d`)
5. Run migrations (command)
6. Start the application (command)
7. Verify (health check URL or test command)

If a `Makefile` / `Taskfile` has an `init` or `setup` target, reference it.

##### 1. Overview
One sentence + important note about how to run commands (e.g., "All commands must be run via `task`" or "Use `make`" or "Use `npm run`").

##### 2. Quick Reference
Table of key commands:
| Action | Command |
|--------|---------|
| Unit tests | `...` |
| Build | `...` |
| Lint | `...` |
| Start | `...` |
| Stop | `...` |

Source commands from:
- Makefile / Taskfile.yml / package.json scripts
- CI/CD configs (`.github/workflows/`, `.gitlab-ci.yml`)
- Project README.md

##### 3. Detailed Command Groups

For each group (Testing, Building, Docker, Linting, Init, etc.):
- Command with usage example (bash code block)
- What it does (1-2 lines)
- Dependencies (what must be run before)
- Tools used

##### 4. Code Generation (if applicable)
Code generation commands:
- Proto/OpenAPI generation
- Test mocks
- Stringer/enum generators
- ORM generators (sqlc, ent, etc.)

For each: command + what it generates + configuration file.

##### 5. CI/CD Cheatsheet
Quick commands for simulating the CI pipeline locally.

##### 6. Tool Installation
Table of tools with installation commands:
| Tool | Install Command |
|------|-----------------|

Determine from go.mod tools, package.json devDependencies, Brewfile, etc.

---

### File 2: .spec/TESTING.md

Project testing conventions.

#### Structure:

Analyze existing test files in the project and identify patterns.

##### 1. Test Package Naming
Which convention: internal test package or external test package?
Example from actual code.

##### 2. Test File Structure
Full example of a typical test from the project:
- Imports
- Test case structure
- Setup / teardown
- Assertions

##### 3. Key Patterns
Identify and describe patterns (only those actually used):
- Table-driven tests (map vs slice)
- Parallel execution
- Variable capture in range loops
- Common test helpers (`start()`, `setup()` functions)
- Builder pattern for test data
- Golden files
- Snapshot testing

For each pattern: name + code example from the project.

##### 4. Mock Generation
How mocks are generated:
- Tool (mockgen, testify/mock, etc.)
- Generation command
- Where stored (alongside tests / separate directory)
- `go:generate` directive (if applicable)

##### 5. Integration Tests
- How separated from unit tests (build tags, directory, naming)
- Run command
- Test containers / test helpers for Docker
- Examples of test resource initialization

##### 6. Export Pattern (if applicable)
How private functions are tested from external test packages.
Example of `export_test.go` or equivalent.

##### 7. Commands
All testing commands:
```bash
# Unit tests
...
# Integration tests
...
# Coverage
...
# Race detector
...
```

---

### File 3: .spec/FILES.md (only if the project has file storage)

Description of the file handling system.

#### When to create:
- S3/MinIO/file adapter exists
- File upload/download functionality exists
- Avatars, media, or documents are handled

#### Structure:

##### 1. Overview
ASCII diagram of file flow: API → Application → FileStore → Storage

##### 2. FileStore Interface
Interface definition from the code.

##### 3. File Model
File structure from the domain layer.

##### 4. Storage Adapter
- Which storage (S3, MinIO, local FS, Cloudflare R2)
- Configuration
- Bucket / directory structure
- File metadata

##### 5. Specific Features
Project-specific features (avatars, limits, formats, etc.)

##### 6. Limitations & Recommendations
Current limitations and improvement recommendations.

##### 7. Environment Variables
Storage configuration from docker-compose / env files.

##### 8. Test Coverage
How the file system is tested.

## General Rules

- All code examples — from the actual project
- If a tool/pattern is not found — do not describe it
- Format: Markdown, tables, fenced code blocks with language tags
- After creating files, update `.spec/README.md`: add links under the Development section
- If `.spec/FILES.md` does not apply (no file storage in the project), skip it entirely
