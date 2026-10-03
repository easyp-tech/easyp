<!-- scope: **/components/*, **/ui/*, **/design-system/*, **/atoms/*, **/molecules/*, **/widgets/* -->
# Components Documentation Template

## What This Generates

- `.spec/COMPONENTS.md` — design system, component library, composition patterns, theming

## Instructions

You are a technical documentarian. Create UI component documentation for the project in the `.spec/` directory.
Analyze: component directories, shared UI libraries, theme/token files, Storybook config, and style systems.

### Step 1: Identify Component Architecture

Check for the presence of:
- **React**: `components/`, JSX/TSX files, Storybook, Radix, shadcn/ui, MUI, Ant Design, Chakra
- **Vue**: `components/`, SFC (`.vue` files), Vuetify, PrimeVue, Quasar, Headless UI
- **Angular**: `components/`, Angular Material, PrimeNG, standalone components
- **Svelte**: `lib/components/`, SvelteKit components, Skeleton UI, Flowbite
- **Flutter/Dart**: `widgets/`, `lib/ui/`, custom Widget classes, Material/Cupertino
- **Web Components**: Custom Elements, Lit, Stencil

Determine from framework files, imports, and component registration patterns.

### Step 2: Create .spec/COMPONENTS.md

#### Structure:

##### 1. Overview
- One sentence: component architecture approach
- UI library / framework used (if any)
- Design methodology: Atomic Design, BEM, component composition, compound components
- Styling approach: CSS Modules, styled-components, Tailwind, CSS-in-JS, vanilla CSS, SCSS

##### 2. Component Inventory

Master table of all shared/reusable components:

| Component | Location | Category | Props / Inputs | Used By |
|-----------|----------|----------|----------------|---------|
| `Button` | `src/components/Button.tsx` | Atom | `variant`, `size`, `disabled`, `onClick` | Throughout |
| `Modal` | `src/components/Modal.tsx` | Molecule | `isOpen`, `onClose`, `title`, `children` | `Settings`, `Confirm` |
| `DataTable` | `src/components/DataTable.tsx` | Organism | `columns`, `data`, `onSort`, `pagination` | `Users`, `Orders` |

Categorize by complexity level (Atom / Molecule / Organism / Template) or by domain (Layout / Form / Feedback / Navigation / Data Display) — whichever the project uses.

##### 3. Component API Patterns

Document the conventions used across all components:

**Props / Inputs:**
- Naming conventions (e.g., `onX` for callbacks, `isX` for booleans, `xRef` for refs)
- Required vs optional patterns
- Default values strategy
- Children / slots / content projection patterns

**Composition patterns used:**
- Compound components (e.g., `<Select><Select.Option>`)
- Render props / scoped slots
- Higher-order components / mixins
- Headless / renderless components
- Forward ref / template ref

Code example of a well-structured component from the project.

##### 4. Design Tokens / Theme

**Colors:**
| Token | Light Value | Dark Value | Usage |
|-------|-------------|------------|-------|
| `--color-primary` | `#3B82F6` | `#60A5FA` | Buttons, links, active states |
| `--color-surface` | `#FFFFFF` | `#1E1E1E` | Cards, modals, backgrounds |

**Typography:**
| Token | Value | Usage |
|-------|-------|-------|
| `--font-family-base` | `Inter, sans-serif` | Body text |
| `--font-size-sm` | `0.875rem` | Secondary text, captions |

**Spacing:**
| Token | Value |
|-------|-------|
| `--space-1` | `4px` |
| `--space-2` | `8px` |

Extract from actual theme files, CSS custom properties, Tailwind config, `ThemeData`, or token files.

**Dark mode / Theming:**
- How themes are defined and switched
- CSS custom properties, `ThemeProvider`, `prefers-color-scheme`
- Where theme tokens live (file path)

##### 5. Layout System

- Grid system: CSS Grid, Flexbox, framework grid (e.g., MUI Grid, Bootstrap Grid)
- Breakpoints:

| Name | Min Width | Typical Usage |
|------|-----------|---------------|
| `sm` | `640px` | Mobile landscape |
| `md` | `768px` | Tablet |
| `lg` | `1024px` | Desktop |

- Container / wrapper components
- Responsive patterns used (fluid, adaptive, mobile-first)

##### 6. Form Components

- Form library (React Hook Form, Formik, VeeValidate, Angular Reactive Forms)
- Validation approach (schema-based with Zod/Yup, inline, server-side)
- Error display pattern (inline, toast, summary)
- Standard form controls:

| Component | Type | Validation | Accessibility |
|-----------|------|------------|---------------|
| `TextInput` | text, email, password | required, pattern | `aria-label`, error announcement |
| `Select` | dropdown | required | keyboard navigation |
| `Checkbox` | boolean | — | `role="checkbox"` |

##### 7. Icon System

- Icon source: icon library (Lucide, Heroicons, Material Icons), custom SVGs, icon font
- How icons are used in components (component import, sprite, CSS class)
- Sizing conventions
- Where custom icons are stored

##### 8. Animation & Transitions

- Animation library (Framer Motion, GSAP, Vue Transition, CSS animations)
- Standard transitions:

| Transition | Duration | Easing | Used For |
|-----------|----------|--------|----------|
| Fade in/out | `200ms` | `ease-in-out` | Modal, tooltip |
| Slide | `300ms` | `ease-out` | Drawer, sidebar |
| Scale | `150ms` | `ease-in` | Button press, card hover |

- Reduced motion: how `prefers-reduced-motion` is respected

##### 9. Accessibility (a11y)

- WCAG target level (A, AA, AAA)
- Keyboard navigation patterns (tab order, focus trap in modals, arrow keys in lists)
- ARIA attributes used by components
- Screen reader testing approach
- Color contrast compliance

##### 10. Storybook / Component Playground

- Storybook version and config location (if applicable)
- How to run: command
- Story file naming convention (`.stories.tsx`, `.stories.vue`)
- Addons used (a11y, viewport, controls)
- How to add a new story

If no Storybook or equivalent exists, note it.

##### 11. Adding a New Component

Step-by-step guide:
1. Create component file in the correct directory
2. Define props/inputs interface
3. Implement with project conventions (styling, composition)
4. Add to exports / barrel file
5. Write story (if Storybook)
6. Write tests (unit + visual regression if applicable)
7. Document usage

## General Rules

- Language: English
- All components, props, and tokens must come from actual source code — do not invent
- If the project uses a pre-built UI library with no customization (e.g., just MUI out-of-box), document the usage patterns rather than the library internals
- For Flutter/Dart projects, adapt terminology: Widget instead of Component, `ThemeData` instead of CSS tokens
- If the project has no shared UI components (e.g., pure backend, CLI), do not generate this file — skip entirely
- After creating, update `.spec/README.md`: add a link under the appropriate section
