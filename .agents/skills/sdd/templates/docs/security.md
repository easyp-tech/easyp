<!-- scope: **/*security*, **/*crypto*, **/*cert* -->
# Security Documentation Template

## What This Generates

- `.spec/SECURITY.md` — security model, threat boundaries, input validation, secrets management, and OWASP Top 10 mapping

## Instructions

You are a technical documentarian with a security focus. Create security documentation for the project in the `.spec/` directory.
Analyze: input validation logic, authentication/authorization middleware, CORS/CSP config, rate limiting, TLS setup, secrets handling, dependency manifests, security headers.

**Important**: This template focuses on the **security perspective** of the project. It cross-references `AUTH.md` (if it exists) for detailed auth flows and `DEPLOYMENT.md` for infrastructure security, but does not duplicate them. Instead, it audits and evaluates security posture.

### Step 1: Identify Security Surface

Map the project's security surface:
- **Entry points**: HTTP endpoints, gRPC services, WebSocket handlers, CLI commands, message consumers
- **Trust boundaries**: external → API gateway → backend → database, user input → validation → business logic
- **Sensitive data**: PII fields, credentials, tokens, financial data
- **Third-party integrations**: OAuth providers, payment systems, external APIs

### Step 2: Create .spec/SECURITY.md

#### Structure:

##### 1. Security Overview

High-level security posture:
- One sentence: security model summary (e.g., "Token-based auth with RBAC, TLS everywhere, secrets in Vault")
- Trust boundary diagram (ASCII):
```
┌─────────────────────────────────────────────┐
│  External (untrusted)                        │
│  ┌─────────┐    ┌──────────┐                │
│  │ Browser │    │ Mobile   │                │
│  └────┬────┘    └────┬─────┘                │
│       └──────┬───────┘                       │
├──────────────┼───────────────────────────────┤
│  DMZ         │                               │
│  ┌───────────▼──────────┐                   │
│  │  Reverse Proxy / LB  │                   │
│  └───────────┬──────────┘                   │
├──────────────┼───────────────────────────────┤
│  Internal (trusted)                          │
│  ┌───────────▼──────────┐  ┌─────────────┐ │
│  │  Application Server  │──│  Database    │ │
│  └──────────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────┘
```
- List of security-critical components with file references

##### 2. Input Validation

Input sanitization strategy:
- **Validation layer**: where validation happens (handler, middleware, domain, or all)
- **Validation approach**: struct tags, validation library, manual checks
- **Injection prevention**:
  - SQL injection: parameterized queries? ORM? raw SQL audit
  - XSS: output encoding, template escaping, CSP
  - Command injection: any `exec` / `os.Command` usage? input sanitization
  - Path traversal: file path validation
- **Validation rules**: table of key validated inputs:

| Input | Location | Validation | Sanitization |
|-------|----------|------------|--------------|
| email | POST /users | format, length | lowercase, trim |
| file upload | POST /files | size, mime type | rename, virus scan |

Code reference: file path where validation is implemented.

##### 3. Authentication & Authorization Audit

Security-focused review of auth (cross-references `AUTH.md` if it exists):
- **Authentication strength**: password policy (min length, complexity), brute force protection (rate limit, lockout)
- **Token security**: token type, signing algorithm, expiry times, rotation policy
- **Session security**: HttpOnly cookies, Secure flag, SameSite, session fixation prevention
- **Privilege escalation prevention**: horizontal (user A accessing user B's data) and vertical (user → admin)
- **Known weaknesses or TODOs**: any `// TODO: security` comments, missing auth checks

If `AUTH.md` exists, reference it for flow details and focus here on security audit findings.

##### 4. Transport Security

TLS and network security:
- **TLS configuration**: version (1.2/1.3), cipher suites, certificate source (Let's Encrypt, self-signed, CA)
- **Internal communication**: mTLS between services? plain HTTP internally?
- **Certificate management**: how certificates are provisioned and rotated
- **Redirect policy**: HTTP → HTTPS redirect enforced?

Code/config reference for TLS setup.

##### 5. CORS & CSP

Cross-origin and content security policies:
- **CORS configuration**:
  - Allowed origins (exact list or pattern)
  - Allowed methods and headers
  - Credentials policy
  - Preflight cache duration
  - Code reference: where CORS is configured
- **Content Security Policy** (if applicable):
  - CSP directives (script-src, style-src, connect-src, etc.)
  - Report-only vs enforced
  - Where CSP header is set

##### 6. Rate Limiting & Abuse Prevention

- **Rate limits**: table of limits per endpoint or operation:

| Scope | Limit | Window | Action |
|-------|-------|--------|--------|
| Login | 5 req | 1 min | 429 + lockout |
| API global | 100 req | 1 min | 429 |
| File upload | 10 req | 1 hour | 429 |

- **Rate limit implementation**: middleware, reverse proxy, cloud WAF
- **Headers returned**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`
- **DDoS mitigation**: Cloudflare, AWS Shield, or application-level measures
- **Bot protection**: CAPTCHA, proof-of-work, behavioral analysis

##### 7. Secrets Management (audit only)

> Operational details (inventory, storage, rotation, local dev) are owned by `deployment.md` → `DEPLOYMENT.md §8 Secrets`.
> This section focuses on **security audit**: anti-patterns, gaps, and recommendations.

- **Anti-patterns to flag**: secrets in code, secrets in git history, secrets in logs, unencrypted config files
- **Gaps**: any secrets without rotation, any secrets accessible beyond need-to-know
- **Recommendations**: concrete next steps to improve secrets hygiene

##### 8. Data Protection

Data handling and privacy:
- **Encryption at rest**: database encryption, file storage encryption
- **Encryption in transit**: TLS (see §4)
- **PII handling**: which fields are PII, how they are stored, access controls
- **Data retention**: retention periods, deletion policy, right to erasure
- **Backup encryption**: are backups encrypted?
- **Compliance**: GDPR, HIPAA, SOC2, PCI-DSS — which apply and how addressed

If no compliance requirements exist, note that and skip the compliance subsection.

##### 9. Security Headers

HTTP security headers:

| Header | Value | Purpose |
|--------|-------|---------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-XSS-Protection` | `0` | Disable legacy XSS filter |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer |
| `Permissions-Policy` | `camera=(), microphone=()` | Restrict browser features |

Identify which headers are set and which are missing. Note where they are configured.

##### 10. Dependency Security

Third-party dependency audit:
- **Audit tool**: `npm audit`, `govulncheck`, `pip-audit`, `cargo audit`, Snyk, Dependabot
- **Audit command**: how to run a vulnerability scan
- **Update policy**: how often dependencies are updated, who reviews
- **Known vulnerabilities**: any current advisories (from the latest audit)
- **Lock file**: is a lock file committed (`go.sum`, `package-lock.json`, `poetry.lock`)?
- **Supply chain**: pinned versions, integrity hashes, trusted registries

##### 11. OWASP Top 10 Mapping

Map each OWASP Top 10 (2021) category to the project's mitigations:

| # | OWASP Category | Project Mitigation | Status |
|---|----------------|-------------------|--------|
| A01 | Broken Access Control | RBAC middleware, ownership checks | ✅ Mitigated |
| A02 | Cryptographic Failures | TLS 1.3, bcrypt passwords, JWT RS256 | ✅ Mitigated |
| A03 | Injection | Parameterized queries, input validation | ✅ Mitigated |
| A04 | Insecure Design | Threat modeling, code review | ⚠️ Partial |
| A05 | Security Misconfiguration | Hardened defaults, no debug in prod | ✅ Mitigated |
| A06 | Vulnerable Components | Dependabot, weekly audit | ✅ Mitigated |
| A07 | Auth Failures | Rate limiting, MFA, strong passwords | ⚠️ Partial |
| A08 | Data Integrity Failures | Signed deployments, CI/CD gates | ✅ Mitigated |
| A09 | Logging & Monitoring | Structured logging, alerting | ⚠️ Partial |
| A10 | SSRF | No user-controlled URLs / URL allowlist | ✅ Mitigated |

Statuses: ✅ Mitigated, ⚠️ Partial, ❌ Not Addressed, N/A

For each "Partial" or "Not Addressed": note what's missing and recommend next steps.

##### 12. Incident Response (if applicable)

If the project has incident response procedures or documentation:
- **Detection**: how security incidents are detected (alerts, logs, user reports)
- **Containment**: immediate steps (revoke tokens, block IPs, disable accounts)
- **Communication**: who to notify, disclosure timeline
- **Recovery**: restore from backup, re-deploy, rotate secrets
- **Post-mortem**: how lessons learned are documented

If no incident response exists, note this as a gap and suggest creating one.

## General Rules

- Language: English
- This is a **security audit** template — focus on findings, gaps, and recommendations
- Cross-reference `AUTH.md` for detailed auth flows, `DEPLOYMENT.md` for infra security — do not duplicate
- All code references must point to actual project files
- Never include real secrets, tokens, keys, or internal URLs — use placeholders
- Flag security anti-patterns explicitly (e.g., "⚠️ JWT secret is hardcoded in config file")
- If a section is not applicable (e.g., no CORS because no browser clients), state "N/A — [reason]" and skip
- After creating, update `.spec/README.md`: add a link under the Security section
