import type { ReactNode } from "react"

import { Link } from "@tanstack/react-router"
import { Languages, Moon, Sun } from "lucide-react"
import { useTranslation } from "react-i18next"

import { BackgroundFX } from "@/components/background-fx"
import { Logo } from "@/components/logo"
import { languages } from "@/i18n"
import { useThemeStore } from "@/stores/theme"

/**
 * Full-bleed shell for terminal states (404, runtime error). Reuses the landing
 * page's atmosphere and chrome — clickable logo home, language + theme toggles —
 * so a dead end still feels like part of the product. `tone` tints the backdrop.
 */
export function StatusScreen({ tone = "brand", children }: { tone?: "brand" | "alert"; children: ReactNode }) {
  const { t, i18n } = useTranslation()
  const theme = useThemeStore((s) => s.theme)
  const toggleTheme = useThemeStore((s) => s.toggle)

  const lang = i18n.language in languages ? i18n.language : "fa"
  const nextLang = lang === "fa" ? "en" : "fa"

  return (
    <main className="relative flex min-h-svh flex-col overflow-hidden bg-background text-foreground">
      <BackgroundFX tone={tone} />

      <header
        className="animate-reveal relative z-10 flex items-center justify-between px-6 py-6 sm:px-10"
        style={{ animationDelay: "60ms" }}
      >
        <Link to="/" className="rounded-md outline-none transition-opacity hover:opacity-80 focus-visible:ring-3 focus-visible:ring-ring/50">
          <Logo className="text-lg" />
        </Link>
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={() => i18n.changeLanguage(nextLang)}
            aria-label={t("comingSoon.toggleLang")}
            className="inline-flex h-9 items-center gap-1.5 rounded-full border border-border/70 bg-background/40 px-3 font-mono text-xs tracking-caps text-muted-foreground backdrop-blur-md transition-colors hover:border-primary/40 hover:text-foreground"
          >
            <Languages className="size-3.5" />
            {languages[nextLang].label}
          </button>
          <button
            type="button"
            onClick={toggleTheme}
            aria-label={t("comingSoon.toggleTheme")}
            className="inline-flex size-9 items-center justify-center rounded-full border border-border/70 bg-background/40 text-muted-foreground backdrop-blur-md transition-colors hover:border-primary/40 hover:text-foreground"
          >
            {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
          </button>
        </div>
      </header>

      <section className="relative z-10 flex flex-1 flex-col items-center justify-center px-6 py-10 text-center sm:px-10">
        <div className="flex w-full max-w-2xl flex-col items-center">{children}</div>
      </section>

      <footer
        className="animate-reveal relative z-10 flex items-center justify-center px-6 py-6 font-mono text-[0.7rem] tracking-caps text-muted-foreground/60 uppercase sm:justify-between sm:px-10"
        style={{ animationDelay: "700ms" }}
      >
        <span>© 2026 {t("common.brandName")}</span>
        <span className="hidden sm:inline">{t("comingSoon.footerNote")}</span>
      </footer>
    </main>
  )
}

/**
 * Oversized status code rendered as the hero graphic — a layered gradient glyph
 * with a soft blurred echo behind it for depth.
 */
export function StatusGlyph({ code, tone = "brand" }: { code: string; tone?: "brand" | "alert" }) {
  const gradient =
    tone === "alert"
      ? "linear-gradient(115deg, var(--green-300), var(--destructive) 60%, var(--green-700))"
      : "linear-gradient(115deg, var(--green-400), var(--green-600) 55%, var(--green-800))"

  return (
    <div
      className="animate-reveal relative select-none"
      style={{ animationDelay: "180ms" }}
      aria-hidden
    >
      {/* blurred echo */}
      <span
        className="absolute inset-0 bg-clip-text leading-none font-heading font-bold tracking-tight text-transparent opacity-40 blur-2xl"
        style={{ backgroundImage: gradient, fontSize: "clamp(7rem, 26vw, 16rem)" }}
      >
        {code}
      </span>
      {/* foreground glyph */}
      <span
        className="relative block bg-clip-text leading-none font-heading font-bold tracking-tight text-transparent"
        style={{ backgroundImage: gradient, fontSize: "clamp(7rem, 26vw, 16rem)" }}
      >
        {code}
      </span>
    </div>
  )
}
