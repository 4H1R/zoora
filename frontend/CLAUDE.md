# Frontend — Zoora

React SPA for Zoora virtual classroom platform.

## Stack

- **Framework**: React 19 + Vite + TanStack Router (file-based) + TanStack React Query
- **Styling**: Tailwind CSS 4 + shadcn/ui components + `cn()` utility from `src/lib/utils.ts`
- **Forms**: React Hook Form + Zod (prefer Zod over Yup for new forms)
- **State**: Zustand for client state, React Query for server state
- **i18n**: react-i18next — Persian (fa, RTL) and English (en, LTR)
- **Icons**: lucide-react
- **HTTP**: redaxios

## Commands

```bash
pnpm build      # Production build + type check
pnpm typecheck  # Type check only (no build)
```

> Never run `pnpm dev` — do not start the dev server.

## Performance

- Never use `useMemo` or `useCallback` — React Compiler handles memoization automatically

## Key Conventions

### UI Components

- Always use shadcn/ui components — install via `pnpm dlx shadcn@latest add <component>` before building custom ones
- Use `cn()` from `src/lib/utils.ts` for all className merging — never raw string concatenation or `clsx` directly
- shadcn components live in `src/components/ui/`; app-level components in `src/components/`
- Never use arbitrary px values (`w-[32px]`, `mt-[12px]`) — always use Tailwind spacing/sizing scale (`w-8`, `mt-3`). Use arbitrary values only for non-standard design tokens with no Tailwind equivalent

### Internationalization (i18n)

- All user-facing text must use `useTranslation()` hook — no hardcoded strings
- Translation files: `src/i18n/locales/en.json` (English) and `src/i18n/locales/fa.json` (Persian)
- Add keys to both language files when adding new text
- App supports RTL (Persian) and LTR (English) — use logical CSS properties (`ms-`, `me-`, `ps-`, `pe-`, `start`, `end`) instead of physical (`ml-`, `mr-`, `pl-`, `pr-`, `left`, `right`)
- Language config with direction info: `src/i18n/index.ts`
- Fonts: Geist (Latin), Vazirmatn (Persian)

### Routing

- File-based routing via TanStack Router — routes in `src/routes/`
- Route types auto-generated into `src/routeTree.gen.ts` — never edit this file manually
- Route-level data fetching with React Query loader patterns

### State Management

- Server state: React Query — no duplicating server data in Zustand
- Client-only state: Zustand stores in `src/stores/`
