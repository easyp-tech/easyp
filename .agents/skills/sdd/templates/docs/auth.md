<!-- scope: **/*auth*, **/*oauth*, **/*login*, **/*session* -->
# Auth Documentation Template

## What This Generates

- `.spec/AUTH.md` (or `.spec/OAUTH.md` if the project primarily uses OAuth) — authentication and authorization documentation

## Instructions

You are a technical documentarian. Create auth/authorization documentation for the project in the `.spec/` directory.
Analyze: OAuth adapters, JWT/PASETO/session middleware, login handlers, DB migrations with auth fields.

### Step 1: Identify Authentication Mechanisms

Determine which authentication mechanisms exist in the project:
- OAuth2 providers (Google, GitHub, Apple, Yandex, Facebook, Twitter, etc.)
- Email + Password (classic registration)
- Telegram Login / Mini App initData
- API Keys
- SSO (SAML, OIDC)
- Magic Links
- 2FA / MFA

### Step 2: Create the Document

Create `.spec/OAUTH.md` if OAuth is the primary mechanism, or `.spec/AUTH.md` otherwise.

#### Structure:

##### 1. Supported Providers / Methods

Table:
| Provider/Method | Status | Validation Method |
|-----------------|--------|-------------------|
| Google | Ready / TODO | ID Token / API call / ... |

Statuses: Ready, TODO, In Progress

##### 2. Architecture (Flow)

ASCII diagram of the authentication flow:
```
Frontend → obtains token → sends to backend → validation → session creation → response
```

Show the full flow from user click to session receipt.
If multiple flows exist (OAuth, email+password, Telegram) — show a unified diagram.

##### 3. User Lookup / Creation Logic

Flowchart (Mermaid or ASCII):
- Token received → validation
- Search by provider_id → found? → login
- Search by email → found? → account linking
- Not found → create new user
- Create session → return token

##### 4. Account Linking (if applicable)
Description of the mechanism for linking OAuth accounts to an existing user.

##### 5. Authorization (if applicable)

Permission and access control model:
- Authorization model type: RBAC / ABAC / scope-based / custom
- Permission matrix table:

| Role | Action | Resource | Allowed |
|------|--------|----------|---------|
| admin | * | * | Yes |
| user | read | own profile | Yes |

- Where authorization checks happen (middleware, guard, decorator, in-handler)
- Code reference: file path and key function/method for permission checking
- How roles/permissions are stored (DB table, JWT claims, config)

##### 6. Session Management (if applicable)

Session storage and lifecycle:
- Storage mechanism: database / Redis / JWT (stateless) / cookie-based
- Session model (fields: id, user_id, token, expires_at, etc.)
- Session lifecycle:
  1. Creation (on login / OAuth callback)
  2. Validation (on each request)
  3. Refresh (token rotation)
  4. Revocation (logout, password change)
- Concurrent session policy: allow multiple / single session / limit per device
- Session cleanup: TTL, cron job, on-demand

##### 7. Token Lifecycle (if applicable)

Token flow and management:
- Token types used: access token, refresh token, ID token, API key
- Token format: JWT / PASETO / opaque / custom
- Token claims / payload structure
- Issuance flow: what triggers token creation
- Rotation strategy: how refresh tokens are rotated
- Revocation mechanism: blocklist, DB flag, Redis TTL
- Token storage on client side (see also CLIENTS.md if applicable)

##### 8. Account Operations (if applicable)

User account management flows:
- **Password reset**: flow diagram (request → email → token → new password), token TTL, rate limiting
- **Email change**: verification flow, old email notification
- **Account deletion / deactivation**: soft delete vs hard delete, data retention, cascade rules
- **Profile update**: which fields are mutable, validation rules

For each flow: endpoint reference, code path, edge cases.

##### 9. Configuration
Configuration example (config.yml, .env, environment variables).
Do not include real secrets — use placeholders.

##### 10. Per-Provider / Per-Method Implementation

For each provider/method:
- **File**: path to implementation file
- **Code**: key fragment (token validation, API call)
- **Specifics**: what distinguishes this provider

##### 11. Tracing (if applicable)
Table of tracing spans for auth operations:
| Span | Description |
|------|-------------|

##### 12. Adding a New Provider / Method
Step-by-step guide (numbered list):
1. Create adapter (code example)
2. Add configuration
3. Register in main
4. DB migration (if needed)
5. Update handler
6. Update API validation

## General Rules

- Language: English
- Do not create sections for providers/methods that do not exist in the project
- Replace real secrets with placeholders
- If the project only has email+password without OAuth — name the file `AUTH.md` and adapt the structure
- All code examples must come from the actual project
- After creating, update `.spec/README.md`: add a link under Auth & Security
