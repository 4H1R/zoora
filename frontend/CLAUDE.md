# Frontend — Zoora

React SPA for Zoora virtual classroom platform.

## Stack

- **Framework**: React 19 + Vite + **TanStack Start** (file-based routing) + TanStack React Query
- **Styling**: Tailwind CSS 4 + shadcn/ui components + `cn()` utility from `src/lib/utils.ts`
- **Forms**: React Hook Form + Zod (prefer Zod over Yup for new forms)
- **State**: Zustand for client state, React Query for server state
- **i18n**: react-i18next — Persian (fa, RTL) and English (en, LTR)
- **Icons**: lucide-react
- **HTTP**: redaxios

## Commands

Package manager is **bun** (not pnpm). The repo root is a bun workspace — run `bun install` from the root. `bunfig.toml` carries the `minimumReleaseAge` supply-chain guard.

```bash
bun run build      # Production build (Vite → Nitro .output/) + prerenders the landing
bun run typecheck  # Type check only (no build)
bun run start      # Run the built Nitro server (.output/server/index.mjs) — prod
```

> Never run `bun run dev` — do not start the dev server.

## Rendering & deploy

- **TanStack Start**, not plain Router. Vite drives dev; `bun run build` emits a self-contained **bun-preset Nitro** server at `frontend/.output/` (`nitro({ preset: 'bun' })` in `vite.config.js`).
- **Selective SSG/SPA**: only the marketing landing `/` is prerendered *with content* (SEO-ready HTML that hydrates the SPA) — see the `prerender` filter in `vite.config.js`. The app layout routes (`_admin`, `_auth`, `_guest`) set **`ssr: false`** so Nitro serves them as client-rendered shells and never SSRs the client-only libs (LiveKit, tldraw, pdfjs, Firebase).
- `src/routes/__root.tsx` is the **document shell** (`shellComponent` + `head()`); there is no `index.html` or `main.tsx`. The router factory lives in `src/router.tsx` (`getRouter()` + `setupRouterSsrQueryIntegration`).
- **Prerender-safety**: anything `/` (the landing) touches must be SSR-safe — no bare `window`/`navigator`/`localStorage` at module or render time (see the guards in `src/lib/tenant.ts` and `src/i18n/index.ts`).
- **PWA offline was dropped** during the Start migration (vite-plugin-pwa's `generateSW` doesn't emit under Nitro). Manifest is a static `public/manifest.webmanifest`; `PWAUpdater` only unregisters the stale Workbox SW. FCM's `firebase-messaging-sw.js` is untouched. Re-add offline later via a post-build Workbox step.
- **Prod**: the frontend container runs the bun Nitro server on `:3000`; **Caddy** (the edge) terminates TLS, proxies `/api/*` → `api:8080` (WS included), and reverse-proxies everything else → `frontend:3000`. See `docker-compose.prod.yml` + `frontend/Dockerfile`.
- `tldraw` is pinned to exact `5.1.1` (license patch in `frontend/patches/`, declared via `patchedDependencies` in the **root** `package.json`) — don't loosen the range or the patch stops applying.

## Performance

- Never use `useMemo` or `useCallback` — React Compiler handles memoization automatically

## Key Conventions

### UI Components

- Always use shadcn/ui components — install via `bunx shadcn@latest add <component>` before building custom ones
- Use `cn()` from `src/lib/utils.ts` for all className merging — never raw string concatenation or `clsx` directly
- shadcn components live in `src/components/ui/`; app-level components in `src/components/`
- Never use arbitrary px values (`w-[32px]`, `mt-[12px]`) — always use Tailwind spacing/sizing scale (`w-8`, `mt-3`). Use arbitrary values only for non-standard design tokens with no Tailwind equivalent
- **Skeletons must match their UI.** Every UI that has a loading skeleton — when you change the real UI (layout, count, sizes, spacing, added/removed elements), update the matching skeleton in the same change so the two stay visually identical. Loaded state and skeleton must have the same structure/dimensions.
- **`Select` is base-ui, not Radix.** `SelectValue` renders the RAW selected value in the trigger by default — it does NOT pick up the `SelectItem`'s label. Whenever the item label differs from its value (translated labels, ids→names), the trigger will show the raw value unless you either (a) pass `items={[{ value, label }]}` to the `<Select>` Root, or (b) give `SelectValue` a function child: `<SelectValue>{(v) => t(\`x.\${v}\`)}</SelectValue>`. Bare `<SelectValue placeholder=... />` is only correct when the item's visible text equals its value.

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
- **Breadcrumbs map path segments through `t(segment)`.** `SidebarBreadcrumb` (used by `admin-breadcrumb.tsx` / org equivalent) falls back to `t(currentSegment)` when a segment isn't in its `SEGMENT_KEYS`. If a route segment name (e.g. `tutorials`) collides with a **top-level i18n key that holds an object**, `t()` returns that object and React throws "Objects are not valid as a React child … an object instead of string", breaking the breadcrumb. When adding an admin route, add its segment to `SEGMENT_KEYS` in `admin-breadcrumb.tsx` pointing at a string title key.

### State Management

- Server state: React Query — no duplicating server data in Zustand
- Client-only state: Zustand stores in `src/stores/`
