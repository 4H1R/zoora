import { Link } from "@tanstack/react-router"
import { Languages, Moon, Sun } from "lucide-react"
import { motion } from "motion/react"
import { useTranslation } from "react-i18next"

import { Logo } from "@/components/logo"
import { buttonVariants } from "@/components/ui/button"
import { languages } from "@/i18n"
import { cn } from "@/lib/utils"
import { useThemeStore } from "@/stores/theme"

import { EASE_OUT, scrollToSection } from "./shared"

const SECTIONS = ["features", "how", "pricing", "faq"] as const

export function LandingNav() {
  const { t, i18n } = useTranslation()
  const theme = useThemeStore((s) => s.theme)
  const toggleTheme = useThemeStore((s) => s.toggle)

  const lang = i18n.language in languages ? i18n.language : "fa"
  const nextLang = lang === "fa" ? "en" : "fa"

  return (
    <motion.header
      initial={{ y: -24, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.7, ease: EASE_OUT }}
      className="border-border/60 bg-background/70 fixed inset-x-0 top-0 z-40 border-b backdrop-blur-xl"
    >
      <div className="container flex h-16 items-center justify-between gap-4">
        <Logo className="text-lg" />

        <nav className="hidden items-center md:flex">
          {SECTIONS.map((id) => (
            <button
              key={id}
              type="button"
              onClick={() => scrollToSection(id)}
              className="text-muted-foreground hover:text-foreground rounded-full px-3.5 py-1.5 text-sm transition-colors"
            >
              {t(`landing.nav.${id}`)}
            </button>
          ))}
        </nav>

        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={() => i18n.changeLanguage(nextLang)}
            aria-label={t("comingSoon.toggleLang")}
            className="border-border/70 bg-background/40 tracking-caps text-muted-foreground hover:border-primary/40 hover:text-foreground inline-flex h-9 items-center gap-1.5 rounded-full border px-3 font-mono text-xs transition-colors"
          >
            <Languages className="size-3.5" />
            <span className="hidden sm:inline">{languages[nextLang].label}</span>
          </button>
          <button
            type="button"
            onClick={toggleTheme}
            aria-label={t("comingSoon.toggleTheme")}
            className="border-border/70 bg-background/40 text-muted-foreground hover:border-primary/40 hover:text-foreground inline-flex size-9 items-center justify-center rounded-full border transition-colors"
          >
            {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
          </button>
          <Link to="/login" className={cn(buttonVariants({ variant: "ghost" }), "hidden rounded-full sm:inline-flex")}>
            {t("landing.nav.signIn")}
          </Link>
          <Link to="/login" className={cn(buttonVariants({ variant: "default" }), "rounded-full px-4")}>
            {t("landing.nav.getStarted")}
          </Link>
        </div>
      </div>
    </motion.header>
  )
}
