import { createFileRoute, Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { useGetOrg } from "@/api/auth/auth"
import GridBackground from "@/components/auth/gradient-background"
import { LoginForm } from "@/components/auth/login-form"
import { LanguageSwitcher } from "@/components/language-switcher"
import { Logo } from "@/components/logo"
import { SplashScreen } from "@/components/splash-screen"
import { StatusGlyph, StatusScreen } from "@/components/status-screen"
import { ThemeToggle } from "@/components/theme-toggle"
import i18n from "@/i18n"
import { isAdminHost } from "@/lib/tenant"

export const Route = createFileRoute("/_guest/login")({
  validateSearch: (search: Record<string, unknown>) =>
    typeof search.redirect === "string" ? { redirect: search.redirect } : {},
  head: () => {
    const title = `${i18n.t("login.title")} — ${i18n.t("common.brandName")}`
    return { meta: [{ title }, { name: "description", content: title }] }
  },
  component: LoginComponent,
})

// Tenant hosts only render the login form once their org resolves and is
// active/trial. The admin host skips the lookup entirely.
const LIVE_STATUSES = ["active", "trial"]

function LoginComponent() {
  const { t } = useTranslation()
  const admin = isAdminHost()
  const { data, isLoading } = useGetOrg({ query: { enabled: !admin } })

  if (!admin) {
    if (isLoading) return <SplashScreen />
    const org = (data?.status === 200 && data.data.data) || undefined
    if (!org || (org.status && !LIVE_STATUSES.includes(org.status))) {
      return (
        <StatusScreen tone="alert">
          <StatusGlyph code="404" tone="alert" />
          <h1
            className="animate-reveal font-heading mt-6 leading-[1.12] font-semibold tracking-tight text-balance"
            style={{ animationDelay: "340ms", fontSize: "clamp(1.75rem, 5vw, 2.75rem)" }}
          >
            {t("login.workspace.notFoundTitle")}
          </h1>
          <p className="text-muted-foreground mt-3 max-w-sm text-sm">{t("login.workspace.notFoundDescription")}</p>
        </StatusScreen>
      )
    }
  }

  return (
    <div className="bg-muted/50 relative min-h-svh">
      <GridBackground />
      <header className="bg-background relative z-10 flex items-center justify-between px-6 py-5.5 md:px-8">
        <Link to="/" className="flex items-center">
          <Logo className="text-xl" />
        </Link>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <div className="bg-border/50 mx-1 h-4 w-px" />
          <LanguageSwitcher />
        </div>
      </header>
      <main className="relative z-10 grid min-h-[calc(100svh-72px)] place-items-center px-5 pt-6 pb-20">
        <div className="border-border bg-background w-full max-w-105 rounded-xl border p-8 shadow-sm">
          <LoginForm />
        </div>
      </main>
    </div>
  )
}
