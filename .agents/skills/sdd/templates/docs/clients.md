<!-- scope: **/client*, **/sdk/* -->
# Clients Documentation Template

## What This Generates

- `.spec/CLIENTS.md` — overview of all client applications
- `.spec/<CLIENT_NAME>.md` — one file per significant client (e.g., `FRONTEND.md`, `TELEGRAM.md`, `MOBILE.md`, `CLI.md`)

## Instructions

You are a technical documentarian. Create client application documentation for the project in the `.spec/` directory.
Analyze directories: `clients/`, `frontend/`, `web/`, `mobile/`, `app/`, `bot/`, `cli/`, etc.

### Step 1: Identify Client Applications

Check for the presence of:
- **Web SPA**: React, Vue, Angular, Svelte, Next.js, Nuxt
- **Mobile**: React Native, Flutter, Swift, Kotlin
- **Telegram**: Mini App, Bot
- **Desktop**: Electron, Tauri
- **CLI**: cobra, click, argparse
- **Browser Extension**: Chrome, Firefox

For each found client, analyze: `package.json`, `pubspec.yaml`, `build.gradle`, etc.

### Step 2: Create .spec/CLIENTS.md (overview)

#### Structure:

##### 1. Available Clients
Table:
| Client | Location | Technology | Purpose |
|--------|----------|------------|---------|

##### 2. Shared Code
What clients share:
- API types (TypeScript interfaces, Dart models)
- API client (Axios, Dio, fetch wrapper)
- Commands for syncing shared code

**Code Generation** (if applicable):
- Source of truth: proto files, OpenAPI spec, GraphQL schema
- Generation command (e.g., `make gen-client`, `npm run codegen`)
- Output directory (e.g., `clients/shared/api/`)
- Sync workflow: when to regenerate, how to verify freshness

**API Version Management** (if applicable):
- How clients handle API versioning (URL prefix, header, content-type)
- Backward compatibility strategy (deprecation notices, feature flags)
- How multiple API versions coexist on the client side

##### 3. API Communication
ASCII diagram: how clients communicate with the backend:
```
┌──────────┐     ┌──────────┐
│  Client  │────>│ Backend  │
│  (React) │ REST│ (API)    │
└──────────┘     └──────────┘
```
Protocol (REST, gRPC, WebSocket, GraphQL), ports.

##### 4. Authentication Flow
Authentication steps from the client perspective:
1. Registration → endpoint
2. Login → endpoint → token receipt
3. Token storage (localStorage, SecureStorage, cookies)
4. Sending in header

##### 5. Adding a New Client
Options (React-based, native mobile, PWA, etc.) with step-by-step instructions.

##### 6. Development Workflow
Commands for running the full stack (backend + clients).

### Step 3: Create Per-Client Documents

Create a separate document if the client:
- Has its own technology stack
- Contains > 10 files
- Has specific logic (Telegram SDK, TON Connect, etc.)

#### Per-Client Template: .spec/{CLIENT_NAME}.md

Example names: `FRONTEND.md`, `TELEGRAM.md`, `MOBILE.md`, `CLI.md`

##### 1. Overview
One sentence + link to base template (if applicable).
Path: `clients/web/` or equivalent.

##### 2. Tech Stack
Table:
| Technology | Purpose |
|------------|---------|

Determine from `package.json` / `pubspec.yaml` / `build.gradle`.

##### 3. Directory Structure
ASCII tree (2 levels):
```
clients/web/
├── src/
│   ├── api/         # API client and types
│   ├── components/  # Reusable components
│   ├── pages/       # Page components
│   └── ...
└── package.json
```

##### 4. Commands
```bash
npm install          # Install dependencies
npm run dev          # Start dev server
npm run build        # Production build
...
```

##### 5. Routes / Navigation
Route table:
| Path | Component | Auth Required |
|------|-----------|---------------|

##### 6. API Client
- Configuration (base URL, interceptors)
- Authentication (how the token is stored and sent)
- Available methods (table: Method | Endpoint | Auth)

##### 7. UI Framework / Theme (if applicable)
- Which UI kit is used
- Colors, fonts, styling approach

##### 8. Platform-Specific Features

**For Telegram Mini App:**
- SDK initialization
- initData authentication
- Mock environment for local development
- Native UI components (@telegram-apps/telegram-ui)
- TON Connect (if applicable)
- Auto-login flow

**For Mobile (React Native / Flutter):**
- Platform-specific code (iOS/Android)
- Push notifications
- Deep linking
- Native modules

**For CLI:**
- Commands and subcommands
- Flags and arguments
- Configuration files

##### 9. Environment Variables
```bash
VITE_API_URL=http://localhost:10002
# or
NEXT_PUBLIC_API_URL=...
```

##### 10. Development vs Production
Differences between dev and prod modes.

##### 11. Deployment
How to deploy the client (GitHub Pages, Vercel, App Store, npm publish).

##### 12. Adding New Pages / Features
Step-by-step instructions for adding a new page/feature.

## General Rules

- If a client is trivial (< 5 files) — do not create a separate document, describe it in `CLIENTS.md`
- Determine technologies from actual `package.json` / dependencies, do not guess
- Routes must come from the actual router code (`App.tsx`, `routes.tsx`, `router.ts`)
- API methods must come from the actual API client
- After creating files, update `.spec/README.md`: add links under the Clients section
- If there are no client applications in the project, do not generate any files — skip this template entirely
