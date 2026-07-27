<!-- scope: **/routes/*, **/router/*, **/pages/*, **/screens/*, **/navigation/* -->
# Routing Documentation Template

## What This Generates

- `.spec/ROUTING.md` — route definitions, guards, layouts, navigation patterns, SSR/SSG strategy

## Instructions

You are a technical documentarian. Create routing and navigation documentation for the project in the `.spec/` directory.
Analyze: router configuration, route definitions, page/screen components, navigation guards/middleware, layout wrappers, and deep link configuration.

### Step 1: Identify Routing Framework

Check for the presence of:
- **React**: React Router, TanStack Router, Next.js App Router / Pages Router, Remix, Wouter
- **Vue**: Vue Router, Nuxt routing (file-based)
- **Angular**: Angular Router (@angular/router)
- **Svelte**: SvelteKit file-based routing
- **Flutter/Dart**: GoRouter, Navigator 2.0, auto_route, Beamer
- **React Native**: React Navigation, Expo Router
- **Solid**: Solid Router, SolidStart file-based routing

Determine from imports, configuration files, and directory structure (file-based routing uses `pages/`, `app/`, `routes/`).

### Step 2: Create .spec/ROUTING.md

#### Structure:

##### 1. Overview
- One sentence: routing framework and approach
- Routing type: code-based, file-based, or hybrid
- Rendering strategy: CSR, SSR, SSG, ISR, or mixed
- ASCII diagram of the route tree (high-level):

```
/
├── / (home)                    [public]
├── /auth
│   ├── /login                  [public]
│   └── /register               [public]
├── /dashboard                  [auth required]
│   ├── /settings               [auth required]
│   └── /profile                [auth required]
├── /admin/*                    [admin role]
└── /api/* (API routes)         [server only]
```

##### 2. Route Table

Complete table of all routes:

| Path | Component / Page | Auth | Roles | Layout | Method | Notes |
|------|-----------------|------|-------|--------|--------|-------|
| `/` | `HomePage` | No | — | `MainLayout` | GET | Landing page |
| `/login` | `LoginPage` | No | — | `AuthLayout` | GET | Redirects if authenticated |
| `/dashboard` | `DashboardPage` | Yes | `user` | `DashboardLayout` | GET | — |
| `/admin/users` | `AdminUsersPage` | Yes | `admin` | `AdminLayout` | GET | — |
| `/api/users` | — | Yes | — | — | GET, POST | API route (server only) |

Extract from actual router configuration, file-based route directories, or route decorators.

##### 3. Route Parameters & Query Strings

| Route Pattern | Param | Type | Validation | Example |
|--------------|-------|------|------------|---------|
| `/users/:id` | `id` | UUID | Regex / Zod | `/users/abc-123` |
| `/posts/:slug` | `slug` | string | — | `/posts/hello-world` |
| `/search?q=` | `q` | string | URL encoded | `/search?q=hello` |

Document how params are parsed and validated (middleware, loader, schema).

##### 4. Guards / Middleware / Interceptors

Navigation guards that run before route access:

| Guard | Applied To | Check | On Failure |
|-------|-----------|-------|------------|
| `authGuard` | `/dashboard/*`, `/admin/*` | Token exists and valid | Redirect to `/login` |
| `roleGuard` | `/admin/*` | User has `admin` role | Redirect to `/403` |
| `onboardingGuard` | `/dashboard/*` | Profile is complete | Redirect to `/onboarding` |

For each guard:
- **Location**: file path
- **Execution order**: in what order do guards run
- **How applied**: route meta, middleware chain, HOC wrapper, per-route config

Code example of a typical guard from the project.

##### 5. Layouts & Nesting

Layout hierarchy:

```
RootLayout (html, body, providers)
├── MainLayout (header, footer, nav)
│   ├── HomePage
│   └── AboutPage
├── AuthLayout (centered card, no nav)
│   ├── LoginPage
│   └── RegisterPage
├── DashboardLayout (sidebar, topbar)
│   ├── DashboardPage
│   └── SettingsPage
└── AdminLayout (admin sidebar)
    └── AdminPages
```

For each layout:
- **Location**: file path
- **Provides**: what UI chrome it adds (header, sidebar, footer)
- **Context / Providers**: what data it fetches or provides to children (user, theme, permissions)

##### 6. Data Loading

How data is loaded for routes:

| Pattern | Framework Feature | When |
|---------|-------------------|------|
| Server loader | `loader()` (Remix), `getServerSideProps` (Next), `load()` (SvelteKit) | Before render, server-side |
| Client fetch | `useEffect`, `onMounted`, `initState` | After mount, client-side |
| Suspense / streaming | React Suspense, deferred data | Progressive loading |
| Static generation | `getStaticProps`, `prerender` | Build time |

For each significant route, document:
- What data is loaded and from where (API endpoint, database, cache)
- Loading state handling (skeleton, spinner, placeholder)
- Error state handling (error boundary, fallback component)

##### 7. Navigation Patterns

How navigation is triggered:

| Pattern | Implementation | Example |
|---------|---------------|---------|
| Declarative link | `<Link to="/about">` | Menu items, breadcrumbs |
| Programmatic | `router.push("/dashboard")` | After form submit, auth redirect |
| Back / history | `router.back()`, `history.back()` | Cancel buttons |
| Replace (no history entry) | `router.replace("/login")` | Auth redirects |
| External | `window.open(url)` | Third-party links |

##### 8. Deep Links & Universal Links (Mobile / PWA)

If the project supports deep linking:
- URL scheme: `myapp://path` or `https://myapp.com/path`
- Configuration file: `AndroidManifest.xml`, `Info.plist`, `apple-app-site-association`, `assetlinks.json`
- Route-to-screen mapping
- Deferred deep links (install → open at correct screen)
- Testing: how to test deep links locally

If not applicable, skip this section.

##### 9. Error & Special Routes

| Route | Purpose | Component |
|-------|---------|-----------|
| `404` / `*` | Not found | `NotFoundPage` |
| `403` | Forbidden | `ForbiddenPage` |
| `500` | Server error | `ErrorPage` |
| `/maintenance` | Maintenance mode | `MaintenancePage` |

How errors are caught: error boundaries, `errorElement` (React Router), `+error.svelte`, `error.tsx`.

##### 10. SEO & Metadata

- How page titles and meta descriptions are set per route (Head component, `generateMetadata`, `useMeta`)
- Open Graph / social sharing meta tags
- Canonical URLs
- Sitemap generation (static, dynamic, `sitemap.xml`)
- `robots.txt` configuration

##### 11. Route Testing

- How routes are tested (navigation tests, guard tests, loader tests)
- Testing framework and utilities (Testing Library, Playwright, Cypress)
- How to test protected routes (mock auth, test tokens)
- Code example from the project

##### 12. Adding a New Route

Step-by-step guide:
1. Create page/screen component
2. Add route definition (config entry or file in `pages/`)
3. Apply guard/middleware if needed
4. Assign layout
5. Add data loader (if SSR/SSG)
6. Set page metadata (title, description)
7. Add to navigation (menu, breadcrumbs)
8. Write tests

## General Rules

- Language: English
- All routes, guards, and layouts must come from actual source code — do not invent
- If the project uses file-based routing, document the directory structure as the primary route definition source
- For API-only projects (no pages/screens), do not generate this file — skip entirely
- For Flutter/Dart, adapt terminology: Screen instead of Page, Navigator/GoRouter instead of React Router
- After creating, update `.spec/README.md`: add a link under the appropriate section
