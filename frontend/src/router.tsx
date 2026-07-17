import { MutationCache, QueryClient } from "@tanstack/react-query"
import { createRouter } from "@tanstack/react-router"
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query"
import { toast } from "sonner"

import { apiErrorMessage, isServerError } from "@/lib/api-error"
import { showPlanErrorToast } from "@/lib/plan-errors"
import { initSentry } from "@/lib/sentry"

import { routeTree } from "./routeTree.gen"

export function getRouter() {
  // Optional error reporting; no-op in SSR and when no DSN is configured.
  initSentry()

  const queryClient = new QueryClient({
    mutationCache: new MutationCache({
      onError: (error, _vars, _ctx, mutation) => {
        // Plan-gate (402) upgrade toast — unless the mutation renders its own.
        if (!mutation.meta?.skipPlanErrorToast && showPlanErrorToast(error)) return
        // Global fallback toast for the *unexpected* class only (network / 5xx):
        // call sites own their specific 4xx copy, so toasting those here would
        // double up. Opt out entirely with meta.skipErrorToast.
        if (mutation.meta?.skipErrorToast) return
        if (isServerError(error)) toast.error(apiErrorMessage(error))
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
    defaultPreloadStaleTime: 0,
    scrollRestoration: true,
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
      skipPlanErrorToast?: boolean
      // Suppress the global fallback toast (network / 5xx). Set on mutations
      // that surface their own copy for those cases.
      skipErrorToast?: boolean
    }
  }
}
