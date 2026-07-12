import { MutationCache, QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createRouter, RouterProvider } from "@tanstack/react-router"
import ReactDOM from "react-dom/client"

import { showPlanErrorToast } from "@/lib/plan-errors"

import { routeTree } from "./routeTree.gen"

// Import first: attaches the beforeinstallprompt listener at module load, before
// any route mounts, so the once-only event is never missed (see the store).
import "./components/pwa/install-prompt-store"
import "./i18n"
import "./styles.css"

const queryClient = new QueryClient({
  // Any mutation that hits a 402 plan/entitlement gate surfaces an upgrade toast
  // once, centrally. A mutation that renders its own plan UI (a paywall, a custom
  // message) opts out with `meta: { skipPlanErrorToast: true }`.
  mutationCache: new MutationCache({
    onError: (error, _vars, _ctx, mutation) => {
      if (mutation.meta?.skipPlanErrorToast) return
      showPlanErrorToast(error)
    },
  }),
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: false,
    },
  },
})

const router = createRouter({
  routeTree,
  context: {
    queryClient,
  },
  defaultPreload: "intent",
  // With React Query, always call the loader on preload/visit (never stale).
  defaultPreloadStaleTime: 0,
  scrollRestoration: true,
  // Render unmatched routes full-screen at the root (root.notFoundComponent)
  // instead of nested inside whatever layout partially matched.
  notFoundMode: "root",
})

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}

declare module "@tanstack/react-query" {
  interface Register {
    mutationMeta: {
      // Set on a mutation to suppress the global plan-error toast when the
      // component surfaces the 402 gate itself (custom copy or paywall UI).
      skipPlanErrorToast?: boolean
    }
  }
}

const rootElement = document.getElementById("app")!

if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}
