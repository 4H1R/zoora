import { MutationCache, QueryClient } from "@tanstack/react-query"
import { createRouter } from "@tanstack/react-router"
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query"

import { showPlanErrorToast } from "@/lib/plan-errors"

import { routeTree } from "./routeTree.gen"

export function getRouter() {
  const queryClient = new QueryClient({
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
    }
  }
}
