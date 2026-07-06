import { useTranslation } from "react-i18next"

import { Logo } from "@/components/logo"

import { scrollToSection } from "./shared"

const SECTIONS = ["features", "how", "pricing", "faq"] as const

export function LandingFooter() {
  const { t } = useTranslation()

  return (
    <footer className="border-border/60 border-t">
      <div className="container flex flex-col items-center gap-8 py-12 sm:flex-row sm:justify-between">
        <div className="flex flex-col items-center gap-2 sm:items-start">
          <Logo className="text-lg" />
          <p className="text-muted-foreground text-sm">{t("landing.footer.note")}</p>
        </div>
        <nav className="flex flex-wrap items-center justify-center gap-x-5 gap-y-2">
          {SECTIONS.map((id) => (
            <button
              key={id}
              type="button"
              onClick={() => scrollToSection(id)}
              className="text-muted-foreground hover:text-foreground text-sm transition-colors"
            >
              {t(`landing.nav.${id}`)}
            </button>
          ))}
        </nav>
      </div>
      <div className="border-border/40 border-t">
        <div className="tracking-caps text-muted-foreground/60 container flex flex-col items-center justify-between gap-2 py-5 font-mono text-[0.7rem] uppercase sm:flex-row">
          <span>© 2026 {t("common.brandName")}</span>
          <span>{t("landing.footer.rights")}</span>
        </div>
      </div>
    </footer>
  )
}
