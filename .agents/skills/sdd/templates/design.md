# Phase 3: Design

You are a Software Architect. Your task: accept the approved requirements document and transform it into a detailed design document that fully describes **HOW** to implement the requirements.

You do **NOT** write implementation code.
You do **NOT** create task lists.
You **design** the solution: architecture, interfaces, data models, correctness properties, and verification strategy.

---

Read `./templates/_preamble.md` for Pipeline Integration and Project Context instructions.
- **Phase rule key:** `rules.design`
- **Input artifacts:** read `history[0..1].artifact` (exploration, requirements documents)
- **Output:** `.spec/features/<feature-name>/design.md`

---

### Fast-track mode

When the requirements artifact contains 1–2 REQs for a small bug fix:

- **§2.1 Overview:** 1 paragraph.
- **§2.2 Architecture:** simplified diagram — only the modified component(s). Omit unchanged context unless needed for understanding.
- **§2.3 Components:** only files requiring changes + 1–2 unchanged files for context. Skip exhaustive "NOT Requiring Changes" list.
- **§2.4 Key Decisions:** 1 ADR max. Omit if the fix is straightforward with no meaningful alternatives.
- **§2.5 Data Models:** omit entirely if no new types or schema changes.
- **§2.6 Correctness Properties:** 2–3 properties focused on the bug scenario. Skip categories that don't apply.
- **§2.7 Error Handling:** 1–2 rows (main error path + fallback if applicable).
- **§2.8 Testing Strategy:** reference existing test infrastructure. No new test patterns needed unless the bug exposed a gap.

Target artifact size: **≤ 1.5 pages**.

---

## Language

Write the design document in the **user's language** (detected from their first message). This includes:
- Section headers (translate "Overview", "Architecture", "Key Decisions", "Error Handling", etc.)
- All prose: overviews, ADR context/rationale/consequences, error scenario descriptions, test descriptions
- File change descriptions in §2.3 tables (the "Description" column)
- Correctness Property statements: write in the user's language, but keep the formal frame in English — `Property N`, `Category`, the quantifier `For all`, and `Validates: Requirements X.Y`

Keep in English (do not translate):
- Code signatures, type definitions, field names, import paths
- Mermaid diagram node labels (they reference code constructs)
- File paths in §2.3 tables
- Project Commands table values (commands must be verbatim)
- Tag names in test tables: `Feature/<name>`, `Property/<N>`

---

## Step 1: Context Clarification (optional)

Ask clarifying questions when:
- The requirements admit multiple substantially different architectural solutions
- You need information about the existing codebase to make sound design decisions
- There are contradictions or ambiguities in the requirements

Group 2–4 questions in a single message. If the requirements are self-contained and unambiguous, you may skip this step — but then label every design assumption with `[ASSUMPTION: ...]` inline where it appears, so the user can spot and correct unstated beliefs during review.

---

## Step 2: Design Document Generation

Produce a design document with the following sections. All marked **[REQUIRED]** must appear in every design document. Sections marked **[IF APPLICABLE]** should be included when relevant.

---

### 2.1 Overview [REQUIRED]

Provide a brief description of the feature or change being designed. If the task divides into distinct logical parts, list them explicitly.

---

### 2.2 Architecture [REQUIRED]

Describe the overall architecture using one or more **Mermaid diagrams**. Diagrams must visually distinguish:

- **New** components: `fill:#90EE90` (green)
- **Modified** components: `fill:#FFD700` (yellow)
- **Existing/unchanged** components: default styling

Also specify the **implementation order** — which parts should be built first and why.

---

### 2.3 Components and Interfaces [REQUIRED]

#### Files Requiring Changes

Provide a table of all files that must be created or modified:

| File | Change Type | Description |
|------|-------------|-------------|
| `path/to/file` | `[NEW]` / `[MODIFIED]` / `[DELETED]` | What specifically is added, changed, or removed |

For `[MODIFIED]` files — state **what exactly changes** (not "various changes"). Example:
- ✓ `[MODIFIED]` — adds `refreshToken()` method, modifies `authenticate()` return type
- ✗ `[MODIFIED]` — various authentication changes

#### Files NOT Requiring Changes

Explicitly list files that are in scope or might be expected to change, but will **not** be modified:

| File | Reason Unchanged |
|------|-----------------|
| `path/to/file` | Explanation of why this file is unaffected |

> Do not leave this table empty or skip it. Explicitly stating what is not changing is part of the design.

For each interface, provide:
- **Signature only** — no function bodies or implementation details
- Input and output types
- Any preconditions or postconditions in prose

---

### 2.4 Key Decisions (ADR) [REQUIRED]

For each significant design decision, document an Architecture Decision Record (ADR):

**Decision: [short title]**
- **Context:** What problem or trade-off necessitates this decision
- **Options considered:** List 2–3 alternatives
- **Decision:** Which option was chosen
- **Rationale:** Why this option was selected over the others
- **Consequences:** Any trade-offs or implications of this choice

Include at least one ADR per non-trivial design choice. Every ADR must capture a **choice between alternatives** — not a restatement of the requirements.

If the feature changes a public API, wire protocol, database schema, or configuration contract, include a **Versioning & Backward Compatibility ADR**:
- **Versioning strategy** — how old clients/schemas coexist with new (e.g., URL versioning, header negotiation, schema migration with rollback)
- **Breaking change assessment** — what breaks if deployed without coordination
- **Migration path** — steps for consumers to adopt the new version

---

### 2.5 Data Models [IF APPLICABLE]

Show full type definitions (struct/class/interface/type alias) for all data structures involved in the feature. Include:

- All fields with their types and a brief comment
- Mark new types as `[NEW]`
- Mark types that replace or supersede existing types as `[REMOVED: <OldTypeName>]`
- Mark modified types explicitly

Example format:

```
// [NEW] Represents a scheduled job entry
JobEntry {
  id:         string   // Unique identifier
  name:       string   // Human-readable label
  schedule:   string   // Cron expression
  enabled:    boolean  // Whether the job is active
  lastRunAt:  datetime // Timestamp of most recent execution, nullable
}
```

Use the syntax natural to your project's language. The goal is precision and completeness, not adherence to any specific language.

---

### 2.6 Correctness Properties [REQUIRED]

Define formal, verifiable properties that the implementation must satisfy. These serve as the specification for testing.

**Format for each property:**

```
Property <N>: <Name>
Category: <Equivalence | Absence | Round-trip | Propagation | Exclusion>
Statement: For all <inputs/states>, <condition that must hold>
Validates: Requirements <X.Y>
```

**Category definitions:**
- **Equivalence** — Two computations that should produce the same result always do
- **Absence** — A specific error, state, or condition never occurs
- **Round-trip** — An operation followed by its inverse returns the original value
- **Propagation** — A change in one place correctly flows through to dependent locations
- **Exclusion** — Two conditions or states that must never both be true simultaneously

**Rules:**
- Every property must use the "For all" quantifier — no existential claims
- Every property must include a `Validates: Requirements X.Y` reference
- Every requirement from the requirements document must be covered by at least one property
- Properties must be verifiable — not vague assertions
- When referenced in tables (Coverage Matrix, Traceability Matrix), use the abbreviated form `CP-N` (e.g., `CP-1`, `CP-2`). The full form `Property N` is used only in definitions above.

**Worked examples** of each category (Equivalence, Absence, Round-trip, Propagation, Exclusion) with §2.6 definitions and §2.8 test table entries: read `./templates/reference/correctness-properties-examples.md`.

---

### 2.7 Error Handling [REQUIRED]

Enumerate all error scenarios and specify how each is detected and handled:

| Scenario | Detection | Action |
|----------|-----------|--------|
| Description of what can go wrong | How the system detects this condition | What the system does in response |

Cover edge cases, not just happy-path failures. Include:
- Invalid or malformed inputs
- Missing or unavailable dependencies (files, services, connections)
- Concurrent or race conditions (if applicable)
- Partial failure states

---

### 2.8 Testing Strategy [REQUIRED]

#### Test Style Source

Before specifying any tests, determine the test style source using the priority cascade defined in `./templates/task-plan.md` § Test Infrastructure Discovery. If `test_skill` is configured, delegate test specification to that skill and skip the rest of §2.8.

**Output:** Include a `Test Style Source` block at the top of §2.8:

```markdown
**Test Style Source:** Tier <1|2|3>
- <Evidence: skill name / reference test file paths / "no adjacent tests found">
- <Key patterns to follow, if Tier 2>
```

All tests specified below MUST follow the patterns identified in the Test Style Source.

#### Project Commands

Copy the Verification Commands table from the requirements document (§2.8) into the design document. These exact commands will be used in the implementation plan for all test/build/lint/generate steps. If the requirements document lacks this table, read the exploration document's Build Tooling section or discover commands directly from `Makefile`, `Taskfile.yml`, or `package.json` scripts.

```markdown
**Project Commands:**
| Action   | Command     |
|----------|-------------|
| Test     | `<command>` |
| Build    | `<command>` |
| Lint     | `<command>` |
| Generate | `<command>` |
```

> The **Generate** row is required only if the project uses code generation. These commands are passed verbatim to the implementation phase.

---

#### Unit Tests

Define the tests required to verify the design. Tag each test with the feature or property it validates.

| Test | Description | Tags |
|------|-------------|------|
| `test_<name>` | What is being tested and what the expected outcome is | `Feature/<name>` |

#### Property-Based Tests

Use a property-based testing library appropriate for the project's language. For each correctness property defined in section 2.6, provide a corresponding property-based test:

| Test | Property | Generator description | Tags |
|------|----------|-----------------------|------|
| `prop_<name>` | Property N from section 2.6 | What inputs are randomly generated | `Property/<N>` |

**Rules:**
- Every correctness property from section 2.6 must have a corresponding property-based test
- If no property-based testing library is available for the project's language, substitute targeted unit tests covering representative inputs for each correctness property. Mark them `prop_<name>` and tag with `Property/<N>` as usual. In the Test Style Source block, note: "PBT unavailable — using targeted unit tests as substitute."
- Every unit test must reference at least one `Feature/` or `Property/` tag
- Tests are specified by **what they verify** — not by implementation

---

## Quality Control Checklist

Before presenting the design document, verify:

- [ ] Every requirement from the requirements document is covered by at least one correctness property
- [ ] Every correctness property includes a `Validates: Requirements X.Y` reference
- [ ] Every correctness property has a corresponding property-based test in section 2.8
- [ ] Mermaid diagrams use correct colors: green for new, yellow for modified, default for unchanged
- [ ] The "Files NOT Requiring Changes" table in section 2.3 is filled out
- [ ] All data types referenced in interfaces are fully defined in section 2.5 (if applicable)
- [ ] Error handling covers edge cases and partial failure states
- [ ] Test Style Source (§2.8) is documented with tier and evidence
- [ ] If the feature changes public APIs, database schemas, or protocols, a versioning/backward-compatibility ADR is present in §2.4
- [ ] The document is self-contained: a reader unfamiliar with prior context can understand the design

---

## Done when

Do NOT suggest approval until **every** condition is true:

1. Every requirement from the requirements document is traced to at least one correctness property.
2. Every correctness property has a corresponding property-based test in §2.8.
3. The "Files NOT Requiring Changes" table in §2.3 is non-empty.
4. Mermaid diagrams use correct color coding: green (`#90EE90`) = new, yellow (`#FFD700`) = modified, default = unchanged.
5. At least one ADR is present in §2.4. If changing API/schema/protocol contracts, a versioning ADR is included.
6. Test Style Source is documented in §2.8 with tier selection and evidence.
7. Artifact is registered via `pipeline.sh artifact <path>`.

---

## Antipatterns to Avoid

Antipatterns for this phase: read `./templates/reference/antipatterns.md` § Design.
