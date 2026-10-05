<!-- scope: **/*error*, **/*errs* -->
# Errors Documentation Template

## What This Generates

- `.spec/ERRORS.md` — error architecture, business error catalog, wrapping conventions, and API error format

## Instructions

You are a technical documentarian. Create error handling documentation for the project in the `.spec/` directory.
Analyze: error type definitions, error packages, HTTP/gRPC error mapping, middleware error handlers, domain error constants.

### Step 1: Identify Error Patterns

Determine how the project handles errors:
- **Custom error types**: struct-based errors with codes, domain error packages
- **Sentinel errors**: `var ErrNotFound = errors.New(...)` / constants
- **Error wrapping**: `fmt.Errorf("...: %w", err)` / custom wrap functions
- **Error mapping**: domain → HTTP status, domain → gRPC code
- **Error response format**: JSON structure returned to API consumers

### Step 2: Create .spec/ERRORS.md

#### Structure:

##### 1. Error Architecture

Layered error model showing how errors flow through the application:
```
┌─────────────────────────────────────────────────┐
│  API Layer                                       │
│  Maps domain errors → HTTP status / gRPC code   │
│  Formats error response JSON                     │
├─────────────────────────────────────────────────┤
│  Application Layer                               │
│  Returns domain errors (ErrNotFound, etc.)       │
│  Wraps adapter errors with context               │
├─────────────────────────────────────────────────┤
│  Adapter Layer                                   │
│  Wraps infrastructure errors (DB, network)       │
│  Converts to domain errors where appropriate     │
├─────────────────────────────────────────────────┤
│  Infrastructure                                  │
│  Raw errors (sql.ErrNoRows, net.Error, etc.)     │
└─────────────────────────────────────────────────┘
```

Describe the error propagation direction and rules:
- Which layer creates errors
- Which layer wraps errors
- Which layer maps/translates errors
- Where errors are logged vs returned

##### 2. Business Error Catalog

Table of all defined business/domain errors:

| Code | Name | HTTP Status | gRPC Code | Description |
|------|------|-------------|-----------|-------------|
| `NOT_FOUND` | `ErrNotFound` | 404 | `NotFound` | Resource does not exist |
| `ALREADY_EXISTS` | `ErrAlreadyExists` | 409 | `AlreadyExists` | Resource already exists |
| `VALIDATION` | `ErrValidation` | 400 | `InvalidArgument` | Input validation failed |
| `UNAUTHORIZED` | `ErrUnauthorized` | 401 | `Unauthenticated` | Authentication required |
| `FORBIDDEN` | `ErrForbidden` | 403 | `PermissionDenied` | Insufficient permissions |

Group by domain area if many errors exist. Include the file path where each error is defined.

##### 3. Error Wrapping Convention

Rules for wrapping errors at each layer boundary:

**Adapter → Application:**
```
// Example pattern (use actual code from the project)
func (r *repo) GetUser(ctx context.Context, id UserID) (*User, error) {
    row := r.db.QueryRow(ctx, query, id)
    if err := row.Scan(...); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    ...
}
```

**Application → API:**
```
// Example pattern (use actual code from the project)
func handleError(err error) (int, ErrorResponse) {
    switch {
    case errors.Is(err, app.ErrNotFound):
        return 404, ErrorResponse{Code: "NOT_FOUND", ...}
    ...
    }
}
```

Show the actual wrapping pattern used in the project. Include file references.

##### 4. Error Response Format

The standard error response structure returned by the API:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address"
      }
    ]
  }
}
```

Document:
- Envelope structure
- Required vs optional fields
- How validation errors include field-level details
- How the code maps to the Business Error Catalog
- Content-Type header

##### 5. Sentinel Errors vs Error Types

When to use each approach in this project:

**Sentinel errors** (simple, comparable with `errors.Is`):
- When to use: identity-only errors without extra data
- Examples from the project

**Typed errors** (struct implementing `error` interface):
- When to use: errors carrying extra context (field name, entity ID, etc.)
- Examples from the project

**Error interfaces** (if applicable):
- Custom error interfaces for behavior-based checking (`errors.As`)
- Examples from the project

##### 6. Retry Policy (if applicable)

Which errors are retryable and how:

| Error Category | Retryable | Strategy |
|----------------|-----------|----------|
| Network timeout | Yes | Exponential backoff, max 3 |
| DB connection lost | Yes | Reconnect + retry once |
| Validation error | No | — |
| Not found | No | — |
| Rate limited | Yes | Respect Retry-After header |

Include retry logic implementation reference if it exists in the project.

##### 7. Error Logging

Rules for error logging:
- Which errors are logged and at which level (error, warn, info)
- What context is included (request ID, user ID, stack trace)
- Where errors are logged (handler layer, middleware, or both)
- How to avoid double-logging (log at the top, wrap at the bottom)

## General Rules

- Language: English
- All error types and codes must come from actual code — do not invent errors
- Code examples must be from the project, showing real wrapping and mapping patterns
- If the project has no custom error handling (uses raw errors only), document that and keep the file minimal
- Do not duplicate the full error catalog if it already exists in `DOMAIN.md` §3 — cross-reference instead, but add HTTP/gRPC mapping columns
- After creating, update `.spec/README.md`: add a link under the appropriate section
