import { MutationCache, QueryClient } from "@tanstack/react-query"
import { createRouter } from "@tanstack/react-router"
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query"

import { showPlanErrorToast } from "@/lib/plan-errors"

import { routeTree } from "./routeTree.gen"

// getRouter is the single router+query factory TanStack Start calls on both the
// server (build-time prerender) and the client. A fresh QueryClient per call
// keeps prerender passes isolated; the SSR-query integration dehydrates on the
// server and hydrates on the client so route loaders share one cache.
export function getRouter() {
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

  setupRouterSsrQueryIntegration({
    router,
    queryClient,
  })

  return router
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>
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
