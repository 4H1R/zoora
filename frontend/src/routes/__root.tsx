import type { Language } from "@/i18n"

import { QueryClient } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { createRootRouteWithContext, HeadContent, Outlet } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { DirectionProvider } from "@/components/ui/direction"
import { Toaster } from "@/components/ui/sonner"
import { languages } from "@/i18n"
import { useThemeStore } from "@/stores/theme"

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient
}>()({
  component: RootComponent,
})

function RootComponent() {
  const { i18n } = useTranslation()
  const theme = useThemeStore((s) => s.theme)

  const lang = (i18n.language in languages ? i18n.language : "fa") as Language

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
      <Toaster />
      <ReactQueryDevtools buttonPosition="bottom-left" />
      <TanStackRouterDevtools position="bottom-right" />
    </DirectionProvider>
  )
}
