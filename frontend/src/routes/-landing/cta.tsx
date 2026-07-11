import { Link } from "@tanstack/react-router"
import { ArrowRight } from "lucide-react"
import { useTranslation } from "react-i18next"

import { GRAIN } from "@/components/background-fx"

import { Reveal } from "./shared"

export function FinalCta() {
  const { t } = useTranslation()

  return (
    <section className="py-24 sm:py-32">
      <div className="container">
        <Reveal>
          <div
            className="relative overflow-hidden rounded-4xl px-6 py-16 text-center sm:px-12 sm:py-20"
            style={{
              background: "linear-gradient(135deg, var(--green-600), var(--green-800))",
            }}
          >
            <div
              aria-hidden
              className="pointer-events-none absolute inset-0 opacity-10 mix-blend-overlay"
              style={{ backgroundImage: GRAIN }}
            />
            <div
              aria-hidden
              className="pointer-events-none absolute start-1/2 -top-24 h-64 w-[36rem] -translate-x-1/2 rounded-full blur-3xl rtl:translate-x-1/2"
              style={{ background: "color-mix(in oklch, var(--green-300) 35%, transparent)" }}
            />

            <h2 className="font-heading relative mx-auto max-w-2xl text-3xl font-semibold tracking-tight text-balance text-white sm:text-5xl sm:leading-[1.15]">
              {t("landing.cta.title")}
            </h2>
            <p className="relative mx-auto mt-4 max-w-xl text-base leading-relaxed text-pretty text-white/80 sm:text-lg">
              {t("landing.cta.subtitle")}
            </p>
            <Link
              to="/get-started"
              className="group relative mt-9 inline-flex h-12 items-center gap-2 rounded-full bg-white px-7 text-base font-semibold text-[var(--green-800)] shadow-xl transition-transform hover:scale-[1.03] active:scale-100"
            >
              {t("landing.cta.button")}
              <ArrowRight className="size-4 transition-transform group-hover:translate-x-0.5 rtl:-scale-x-100 rtl:group-hover:-translate-x-0.5" />
            </Link>
          </div>
        </Reveal>
      </div>
    </section>
  )
}
