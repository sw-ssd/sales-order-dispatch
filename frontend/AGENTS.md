# Nikala UI & SolidJS Development Guidelines

> Nikala UI is a copy-paste component system and reactive primitives suite for SolidJS built natively for Tailwind CSS v4.

## 1. Strict SolidJS Reactivity Rules

1. **NEVER Destructure Props Directly**:
   - `const { variant, class: className } = props;` -> ❌ **FORBIDDEN** (breaks SolidJS fine-grained signal tracking).
   - `const [local, others] = splitProps(props, ["variant", "class"]);` -> ✅ **REQUIRED**.

2. **Children Inspection & Tab Hydration**:
   - ALWAYS wrap `props.children` with SolidJS's native `children(() => props.children)` memoization helper when inspecting, iterating, or rendering dynamic JSX child nodes.

3. **SSR Safety Guards**:
   - Always include `typeof window !== "undefined"` and `typeof document !== "undefined"` guards inside browser event listeners or DOM access logic.

4. **Anti-FOUC Theme Script**:
   - When using `ThemeProvider`, ALWAYS place `<ThemeScript storageKey="nikala-theme" />` synchronously inside `<head>` or root HTML before `<ThemeProvider>`.

## 2. Tailwind CSS v4 Native Design Tokens

- All components must use semantic design tokens (`bg-background`, `text-foreground`, `bg-card`, `border-border`, `bg-primary`, etc.).

## 3. Pure Copy-Paste Primitives Ownership

- UI components live in `src/components/ui/`.
- Reactive hooks live in `src/hooks/` and are imported locally:
  `import { createClipboard } from "@/hooks/create-clipboard";`
