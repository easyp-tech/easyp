<!-- scope: **/store/*, **/stores/*, **/state/*, **/redux/*, **/bloc/*, **/providers/* -->
# State Management Documentation Template

## What This Generates

- `.spec/STATE.md` — state management architecture, store structure, data flow, side effects

## Instructions

You are a technical documentarian. Create state management documentation for the project in the `.spec/` directory.
Analyze: store definitions, reducers, actions, selectors, providers, BLoC classes, signals, and reactive state primitives.

### Step 1: Identify State Management Approach

Check for the presence of:
- **React**: Redux (Toolkit), Zustand, MobX, Jotai, Recoil, Valtio, XState, React Context + useReducer
- **Vue**: Pinia, Vuex
- **Angular**: NgRx, Akita, Elf, simple services with RxJS
- **Svelte**: Svelte stores (writable, readable, derived)
- **Solid**: Solid signals, createStore
- **Flutter/Dart**: BLoC, Provider, Riverpod, GetX, MobX
- **Mobile (RN)**: Redux, Zustand, MobX-State-Tree

Determine from `package.json`, `pubspec.yaml`, imports in source files.

### Step 2: Create .spec/STATE.md

#### Structure:

##### 1. Overview
- One sentence: which state management solution and why
- Architecture pattern name (Flux, CQRS, reactive, signal-based)
- ASCII diagram of the data flow:

```
User Action → Dispatch → Store/Reducer → New State → UI Re-render
                           ↓
                      Side Effects (API calls, persistence)
```

Adapt the diagram to the actual pattern used (BLoC: Event → Bloc → State → Widget; Signals: Signal → Effect → Derived).

##### 2. Store Structure

Map of all stores / slices / blocs / atoms:

| Store / Slice | Location | Purpose | Persistence |
|---------------|----------|---------|-------------|
| `authStore` | `src/stores/auth.ts` | User session, tokens | localStorage |
| `cartSlice` | `src/store/cart.ts` | Shopping cart items | None (memory) |

For each store, note whether it is:
- **Global** (app-wide singleton)
- **Scoped** (per-route, per-component tree, per-feature)
- **Local** (component-level, not shared)

##### 3. State Shape

For each significant store, document the state type/interface:

```typescript
// authStore state
interface AuthState {
  user: User | null;       // Current authenticated user
  token: string | null;    // JWT access token
  isLoading: boolean;      // Auth operation in progress
  error: string | null;    // Last auth error message
}
```

Use the project's actual language and type definitions. Include initial/default values.

##### 4. Actions / Events / Mutations

For each store, list the actions that modify state:

| Action / Event | Payload | Effect on State | Side Effects |
|----------------|---------|-----------------|--------------|
| `login` | `{email, password}` | Sets `isLoading: true` | API call to `/auth/login` |
| `loginSuccess` | `{user, token}` | Sets `user`, `token`, `isLoading: false` | Persists token |
| `logout` | — | Clears `user`, `token` | Removes token from storage |

##### 5. Selectors / Derived State / Computed

| Selector / Computed | Source | Returns | Used By |
|---------------------|--------|---------|---------|
| `isAuthenticated` | `authStore` | `boolean` | `ProtectedRoute`, `Header` |
| `cartTotal` | `cartStore` | `number` | `CartSummary`, `Checkout` |

##### 6. Side Effects

How async operations and side effects are handled:
- **Pattern**: thunks, sagas, epics, effects, listeners, BLoC events, async actions
- **API integration**: how API calls are dispatched and results handled
- **Error handling**: how API errors flow back to state
- **Optimistic updates**: are they used? Where?
- **Cancellation**: how in-flight requests are cancelled (e.g., route change, unmount)

Code example of a typical side effect from the project.

##### 7. Persistence

Which parts of state are persisted and how:

| Store | Medium | Key | Serialization | Hydration |
|-------|--------|-----|---------------|-----------|
| `auth` | localStorage | `auth_token` | JSON | On app init |
| `settings` | AsyncStorage | `user_prefs` | JSON | On app init |
| `cart` | sessionStorage | `cart_items` | JSON | On tab restore |

If no persistence is used, state it explicitly.

##### 8. Subscriptions / Reactivity

How components subscribe to state changes:
- **React**: `useSelector`, `useStore`, `useSnapshot`, custom hooks
- **Vue**: `storeToRefs`, computed properties
- **Flutter**: `BlocBuilder`, `Consumer`, `Watch`
- **Svelte**: `$store` auto-subscription

Performance considerations:
- Selector memoization
- Render optimization (shallow compare, structural sharing)
- Common pitfalls in the project

##### 9. Testing State

- How stores/slices/blocs are unit tested
- How to create test state (factories, fixtures, initial state overrides)
- How to mock stores in component tests
- Code example of a store test from the project

##### 10. Adding a New Store

Step-by-step guide for adding a new store/slice/bloc:
1. Create the state type
2. Define actions/events
3. Implement reducer/handler
4. Add selectors
5. Connect to components
6. Add tests

Reference existing stores as examples.

## General Rules

- Language: English
- All state shapes, actions, and selectors must come from actual source code
- If the project uses multiple state management solutions (e.g., Redux for global + useState for local), document all of them and explain the boundary
- If the project has no state management beyond component-local state, do not generate this file — skip entirely
- After creating, update `.spec/README.md`: add a link under the appropriate section
