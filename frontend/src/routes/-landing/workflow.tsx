import { useTranslation } from "react-i18next"

import { Reveal, SectionHeading } from "./shared"

const STEPS = ["step1", "step2", "step3"] as const

export function Workflow() {
  const { t } = useTranslation()

  return (
    <section id="how" className="border-border/60 bg-muted/30 scroll-mt-20 border-t py-24 sm:py-32">
      <div className="container">
        <SectionHeading eyebrow={t("landing.how.eyebrow")} title={t("landing.how.title")} />

        <div className="relative mt-14 grid gap-10 sm:mt-16 md:grid-cols-3 md:gap-8">
          {/* Connecting line between steps on desktop */}
          <div
            aria-hidden
            className="absolute inset-x-[16%] top-8 hidden h-px md:block"
            style={{
              background:
                "linear-gradient(to right, transparent, color-mix(in oklch, var(--primary) 45%, transparent), transparent)",
            }}
          />
          {STEPS.map((step, i) => (
            <Reveal key={step} delay={i * 0.12} className="relative flex flex-col items-center text-center">
              <span className="border-primary/30 bg-background text-primary relative flex size-16 items-center justify-center rounded-2xl border font-mono text-lg font-semibold shadow-sm">
                {t(`landing.how.${step}.n`)}
              </span>
              <h3 className="font-heading mt-6 text-xl font-semibold tracking-tight">
                {t(`landing.how.${step}.title`)}
              </h3>
              <p className="text-muted-foreground mt-2.5 max-w-70 text-sm leading-relaxed text-pretty">
                {t(`landing.how.${step}.desc`)}
              </p>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  )
}
