import { createFileRoute } from "@tanstack/react-router"
import { GraduationCap, Languages, Moon, Radio, Sun, Video } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Logo } from "@/components/logo"
import { languages } from "@/i18n"
import { useThemeStore } from "@/stores/theme"

export const Route = createFileRoute("/")({
  component: RouteComponent,
})

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const theme = useThemeStore((s) => s.theme)
  const toggleTheme = useThemeStore((s) => s.toggle)

  const lang = i18n.language in languages ? i18n.language : "fa"
  const nextLang = lang === "fa" ? "en" : "fa"

  const features = [
    { icon: Radio, label: t("comingSoon.feature1") },
    { icon: Video, label: t("comingSoon.feature2") },
    { icon: GraduationCap, label: t("comingSoon.feature3") },
  ]

  return (
    <main className="relative flex min-h-svh flex-col overflow-hidden bg-background text-foreground">
      {/* atmosphere */}
      <BackgroundFX />

      {/* top bar */}
      <header
        className="animate-reveal relative z-10 flex items-center justify-between px-6 py-6 sm:px-10"
        style={{ animationDelay: "60ms" }}
      >
        <Logo className="text-lg" />
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

      {/* hero */}
      <section className="relative z-10 flex flex-1 flex-col items-center justify-center px-6 py-10 text-center sm:px-10">
        <div className="flex w-full max-w-3xl flex-col items-center">
          <span
            className="animate-reveal inline-flex items-center gap-2.5 rounded-full border border-border/70 bg-background/40 py-1.5 ps-2 pe-3.5 font-mono text-[0.7rem] tracking-caps text-muted-foreground uppercase backdrop-blur-md"
            style={{ animationDelay: "140ms" }}
          >
            <span className="relative flex size-2.5 items-center justify-center">
              <span className="absolute inline-flex size-full animate-ping rounded-full bg-primary/60" />
              <span className="relative inline-flex size-2 rounded-full bg-primary" />
            </span>
            {t("comingSoon.status")}
          </span>

          <p
            className="animate-reveal mt-9 font-mono text-xs tracking-caps text-primary uppercase"
            style={{ animationDelay: "220ms" }}
          >
            {t("comingSoon.kicker")}
          </p>

          <h1
            className="animate-reveal mt-4 font-heading font-semibold leading-[1.12] tracking-tight text-balance"
            style={{ animationDelay: "300ms", fontSize: "clamp(2.75rem, 9vw, 6rem)" }}
          >
            <span className="block">{t("comingSoon.titleLine1")}</span>
            <span
              className="block bg-clip-text pb-[0.15em] text-transparent"
              style={{
                backgroundImage:
                  "linear-gradient(115deg, var(--green-400), var(--green-600) 55%, var(--green-800))",
              }}
            >
              {t("comingSoon.titleLine2")}
            </span>
          </h1>

          <p
            className="animate-reveal mt-7 max-w-xl text-base leading-relaxed text-muted-foreground text-pretty sm:text-lg"
            style={{ animationDelay: "380ms" }}
          >
            {t("comingSoon.subtitle")}
          </p>

          {/* features */}
          <ul
            className="animate-reveal mt-11 flex flex-wrap items-center justify-center gap-x-8 gap-y-3"
            style={{ animationDelay: "460ms" }}
          >
            {features.map(({ icon: Icon, label }) => (
              <li key={label} className="inline-flex items-center gap-2 text-sm text-muted-foreground">
                <Icon className="size-4 text-primary" />
                {label}
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* footer */}
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

const GRAIN =
  "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")"

function BackgroundFX() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 overflow-hidden">
      {/* base wash */}
      <div
        className="absolute inset-0"
        style={{
          background:
            "radial-gradient(120% 80% at 50% -10%, color-mix(in oklch, var(--primary) 14%, transparent), transparent 60%)",
        }}
      />
      {/* aurora blobs */}
      <div
        className="animate-aurora absolute -top-40 start-[-10%] size-[40rem] rounded-full opacity-50 blur-3xl"
        style={{ background: "radial-gradient(circle, var(--green-500), transparent 65%)" }}
      />
      <div
        className="animate-aurora-slow absolute -bottom-52 end-[-8%] size-[36rem] rounded-full opacity-40 blur-3xl"
        style={{ background: "radial-gradient(circle, var(--green-700), transparent 65%)" }}
      />
      {/* grid */}
      <div
        className="absolute inset-0"
        style={{
          backgroundImage:
            "linear-gradient(to right, color-mix(in oklch, var(--foreground) 5%, transparent) 1px, transparent 1px), linear-gradient(to bottom, color-mix(in oklch, var(--foreground) 5%, transparent) 1px, transparent 1px)",
          backgroundSize: "64px 64px",
          maskImage: "radial-gradient(120% 90% at 50% 0%, black, transparent 75%)",
          WebkitMaskImage: "radial-gradient(120% 90% at 50% 0%, black, transparent 75%)",
        }}
      />
      {/* grain */}
      <div
        className="absolute inset-0 opacity-[0.04] mix-blend-overlay dark:opacity-[0.06]"
        style={{ backgroundImage: GRAIN }}
      />
      {/* bottom fade */}
      <div
        className="absolute inset-x-0 bottom-0 h-40"
        style={{ background: "linear-gradient(to top, var(--background), transparent)" }}
      />
    </div>
  )
}
