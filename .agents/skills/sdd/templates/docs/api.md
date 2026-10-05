<!-- scope: **/routes/*, **/handlers/*, **/api/*, *.proto, **/openapi* -->
# API Documentation Template

## What This Generates

- `.spec/API.md` — API endpoint reference, conventions, middleware, and error format

## Instructions

You are a technical documentarian. Create API documentation for the project in the `.spec/` directory.
Analyze: route definitions, handler files, middleware chain, proto/OpenAPI specs, request/response types, error handling.

### Step 1: Identify API Surface

Determine the API type(s):
- **REST**: route files, handler functions, HTTP methods
- **gRPC**: `.proto` files, generated code, service definitions
- **GraphQL**: schema files, resolvers
- **WebSocket**: upgrade handlers, event definitions

For each API surface, identify:
- Router / framework (e.g., `chi`, `gin`, `echo`, `Express`, `FastAPI`, `net/http`)
- Base path / port
- Authentication method (Bearer token, API key, cookie)

### Step 2: Create .spec/API.md

#### Structure:

##### 1. Overview
- One sentence: API type + framework
- Base URL pattern (e.g., `http://localhost:8080/api/v1`)
- Authentication method summary

##### 2. Middleware Stack

Ordered list of middleware in the request processing pipeline:
| Order | Middleware | Purpose |
|-------|-----------|---------|
| 1 | Request ID | Assigns unique ID to each request |
| 2 | Logger | Structured request logging |
| 3 | Recovery | Panic recovery |
| 4 | Auth | Token validation |
| 5 | ... | ... |

Determine order from the actual router setup code.
Note which middleware is global vs route-specific.

##### 3. Endpoint Reference

Group endpoints by resource/domain:

**Users**
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/users` | No | Create user |
| GET | `/api/v1/users/:id` | Yes | Get user by ID |

For each group, include:
- All endpoints with method, path, auth requirement, description
- Source file reference (handler file path)

If there are many endpoints (>20), create a summary table first, then detail sections per resource.

##### 4. Request / Response Conventions

Describe the standard patterns:
- **Request envelope**: is there a wrapper? (`{"data": ...}` vs flat)
- **Response envelope**: standard response format
  ```json
  {
    "data": { ... },
    "error": null
  }
  ```
- **Pagination**: pattern used (offset/limit, cursor, page/per_page)
  - Request parameters
  - Response metadata (`total`, `next_cursor`, `has_more`)
- **Sorting**: parameter name and format (e.g., `?sort=name,-created_at`)
- **Filtering**: parameter patterns (e.g., `?status=active&role=admin`)

Code examples from the actual project.

##### 5. Status Codes & Error Format

**Standard status codes used:**
| Status | When Used |
|--------|-----------|
| 200 | Successful response |
| 201 | Resource created |
| 400 | Validation error |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not found |
| 500 | Internal error |

**Error response format:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": [...]
  }
}
```

Show the actual error response structure from the project.
Reference the error mapping code (where domain errors become HTTP errors).

##### 6. Versioning Strategy (if applicable)
- How API versions are specified: URL prefix (`/v1/`), header (`Accept-Version`), content type
- Current version(s)
- Deprecation policy

##### 7. Rate Limiting (if applicable)
- Rate limit values (requests per second/minute)
- Rate limit headers returned (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`)
- Per-user vs per-IP vs per-API-key
- Configuration reference

##### 8. Proto / OpenAPI / GraphQL Schema (if applicable)

For **gRPC**:
- `.proto` file locations
- Service definitions (list of RPCs per service)
- Code generation command
- Client stub usage example

For **OpenAPI**:
- Spec file location (e.g., `api/openapi.yaml`)
- How spec is generated (manual vs auto-generated)
- Swagger UI URL (if served)
- Code generation command (if applicable)

For **GraphQL**:
- Schema file location
- Key queries and mutations
- Resolver structure

##### 9. Validation
- Validation library / approach (struct tags, middleware, manual)
- Where validation happens (handler, middleware, domain layer)
- Common validation rules used
- Custom validators (if any)

Code example of a validated request from the project.

## General Rules

- Language: English
- All endpoints must come from actual route definitions, not assumed
- Request/response examples must use realistic but non-sensitive data
- If the project has no API (library, CLI-only), do not generate this file
- Group endpoints logically by resource, not by HTTP method
- After creating, update `.spec/README.md`: add a link under the appropriate section
