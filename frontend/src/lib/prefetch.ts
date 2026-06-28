/**
 * Background chunk warmer — TanStack equivalent of Laravel's
 * `Vite::prefetch(concurrency: 3)`.
 *
 * `defaultPreload: "intent"` (see main.tsx) already prefetches a route's
 * code + data on hover/touch, which covers most navigation. This warmer is
 * the eager counterpart: once the app is idle, it pulls every route chunk
 * into the browser cache in the background, capped at N concurrent fetches so
 * it never competes with the initial page load.
 *
 * autoCodeSplitting (vite.config.js) splits each route into its own chunk, so
 * importing the route module is enough to make Vite fetch + cache it.
 */

// Eager: false keeps these as lazy importers; calling one triggers the fetch.
const routeModules = import.meta.glob("../routes/**/*.tsx")

/** Drain `loaders` through `concurrency` parallel workers. */
async function warm(loaders: Array<() => Promise<unknown>>, concurrency: number) {
  let next = 0
  const worker = async () => {
    while (next < loaders.length) {
      const load = loaders[next++]
      // Swallow failures — a warm miss must never surface to the user.
      await load().catch(() => {})
    }
  }
  await Promise.all(Array.from({ length: concurrency }, worker))
}

/**
 * Warm all route chunks in the background after the app goes idle.
 * @param concurrency max parallel chunk fetches (Laravel default: 3)
 */
export function prefetchRoutes(concurrency = 3) {
  const loaders = Object.values(routeModules)
  const run = () => warm(loaders, concurrency)

  if ("requestIdleCallback" in window) {
    window.requestIdleCallback(run)
  } else {
    setTimeout(run, 2000)
  }
}
