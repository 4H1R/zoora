import type { Language } from "@/i18n"
import type { ReactNode } from "react"

import { HotkeysProvider } from "@tanstack/react-hotkeys"
import { QueryClient } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { createRootRouteWithContext, HeadContent, Outlet, Scripts } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { ConnectionBanner } from "@/components/connection-banner"
import { ErrorScreen } from "@/components/error-screen"
import { GlobalErrorListeners } from "@/components/global-error-listeners"
import { NotFound } from "@/components/not-found"
import { PWAUpdater } from "@/components/pwa-updater"
import { DirectionProvider } from "@/components/ui/direction"
import { Toaster } from "@/components/ui/sonner"
import { languages } from "@/i18n"
import { useThemeStore } from "@/stores/theme"

// Client bootstrap side effects, previously in main.tsx. Order matters:
// install-prompt-store attaches its window listener at import time (guarded to
// the browser) so Chrome's once-only beforeinstallprompt is never missed; i18n
// initializes the translator; styles.css pulls in Tailwind. All three are
// SSR-safe — imported in the shell so they run in both the prerender and the
// client bundle.
import "@/components/pwa/install-prompt-store"
import "@/i18n"
import "@/styles.css"

const OG_DESCRIPTION =
  "Run live online classes, video meetings, and recordings on one secure multi-tenant platform. Built for schools, tutors, and teams."
const OG_IMAGE = "https://app.zoora.ir/og-image.png"
const TITLE = "Zoora — Virtual Classrooms & Video Meetings"

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient
}>()({
  // Head tags formerly in index.html. TanStack Start generates the document
  // shell itself, so VitePWA's index.html injection no longer runs — the
  // manifest link is declared here explicitly (the manifest file + sw.js are
  // still emitted by vite-plugin-pwa).
  head: () => ({
    meta: [
      { charSet: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1.0" },
      { title: TITLE },
      { name: "description", content: OG_DESCRIPTION },
      { name: "theme-color", content: "#16a34a" },

      // Open Graph / Facebook / LinkedIn / WhatsApp
      { property: "og:type", content: "website" },
      { property: "og:site_name", content: "Zoora" },
      { property: "og:locale", content: "en_US" },
      { property: "og:title", content: TITLE },
      { property: "og:description", content: OG_DESCRIPTION },
      { property: "og:url", content: "https://app.zoora.ir/" },
      { property: "og:image", content: OG_IMAGE },
      { property: "og:image:secure_url", content: OG_IMAGE },
      { property: "og:image:type", content: "image/png" },
      { property: "og:image:width", content: "1200" },
      { property: "og:image:height", content: "630" },
      {
        property: "og:image:alt",
        content: "Zoora — live classes, meetings & recordings on one secure platform",
      },

      // Twitter / X
      { name: "twitter:card", content: "summary_large_image" },
      { name: "twitter:title", content: TITLE },
      { name: "twitter:description", content: OG_DESCRIPTION },
      { name: "twitter:image", content: OG_IMAGE },
      {
        name: "twitter:image:alt",
        content: "Zoora — live classes, meetings & recordings on one secure platform",
      },
    ],
    links: [
      { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" },
      { rel: "icon", type: "image/png", sizes: "32x32", href: "/favicon-32.png" },
      { rel: "apple-touch-icon", href: "/apple-touch-icon.png" },
      { rel: "manifest", href: "/manifest.webmanifest" },
    ],
  }),
  component: RootComponent,
  shellComponent: RootDocument,
  errorComponent: ErrorScreen,
  notFoundComponent: NotFound,
})

// The document shell. Rendered on the server (build-time prerender) and
// hydrated on the client. lang/dir default to English/LTR for the static
// prerender; RootComponent's effect swaps them once the client detects the
// user's language.
function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en" dir="ltr">
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  )
}

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
    <HotkeysProvider>
      <DirectionProvider direction={dir}>
        <Outlet />
        <GlobalErrorListeners />
        <ConnectionBanner />
        <Toaster />
        <PWAUpdater />
        {/*<ReactQueryDevtools buttonPosition="bottom-left" />
        <TanStackRouterDevtools position="bottom-right" />*/}
      </DirectionProvider>
    </HotkeysProvider>
  )
}
