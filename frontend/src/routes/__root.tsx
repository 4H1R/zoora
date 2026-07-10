import type { Language } from "@/i18n"

import { QueryClient } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { createRootRouteWithContext, HeadContent, Outlet } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { ConnectionBanner } from "@/components/connection-banner"
import { ErrorScreen } from "@/components/error-screen"
import { NotFound } from "@/components/not-found"
import { PWAUpdater } from "@/components/pwa-updater"
import { DirectionProvider } from "@/components/ui/direction"
import { Toaster } from "@/components/ui/sonner"
import { languages } from "@/i18n"
import { useThemeStore } from "@/stores/theme"

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient
}>()({
  component: RootComponent,
  errorComponent: ErrorScreen,
  notFoundComponent: NotFound,
})

function RootComponent() {
  const { i18n } = useTranslation()
  const theme = useThemeStore((s) => s.theme)

  // Detector may hand back a region locale ("en-US"); resolvedLanguage is
  // guaranteed to be one of supportedLngs, so region variants no longer fall
  // through to the "fa" default and wrongly force RTL on English.
  const resolved = i18n.resolvedLanguage ?? i18n.language
  const lang = (resolved in languages ? resolved : "en") as Language

  const dir = languages[lang].dir

  useEffect(() => {
    document.documentElement.lang = lang
    document.documentElement.dir = dir
  }, [lang, dir])

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark")
  }, [theme])

  return (
    <DirectionProvider direction={dir}>
      <HeadContent />
      <Outlet />
      <ConnectionBanner />
      <Toaster />
      <PWAUpdater />
      {/*<ReactQueryDevtools buttonPosition="bottom-left" />
      <TanStackRouterDevtools position="bottom-right" />*/}
    </DirectionProvider>
  )
}
