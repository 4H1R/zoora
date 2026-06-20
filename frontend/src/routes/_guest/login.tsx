import { createFileRoute, Link } from "@tanstack/react-router"
import i18n from "@/i18n"

import GridBackground from "@/components/auth/gradient-background"
import { LoginForm } from "@/components/auth/login-form"
import { LanguageSwitcher } from "@/components/language-switcher"
import { Logo } from "@/components/logo"
import { ThemeToggle } from "@/components/theme-toggle"

export const Route = createFileRoute("/_guest/login")({
  head: () => {
    const title = `${i18n.t("login.title")} — ${i18n.t("common.brandName")}`
    return { meta: [{ title }, { name: "description", content: title }] }
  },
  component: LoginComponent,
})

function LoginComponent() {
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
