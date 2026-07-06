import { Link } from "@tanstack/react-router"
import { Check, Sparkles } from "lucide-react"
import { useTranslation } from "react-i18next"

import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

import { Reveal, SectionHeading } from "./shared"

const PLANS = [
  { key: "free", features: ["f1", "f2", "f3", "f4", "f5"], highlighted: false },
  { key: "pro", features: ["f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8"], highlighted: true },
  { key: "enterprise", features: ["f1", "f2", "f3", "f4", "f5", "f6"], highlighted: false },
] as const

export function Pricing() {
  const { t } = useTranslation()

  return (
    <section id="pricing" className="scroll-mt-20 py-24 sm:py-32">
      <div className="container">
        <SectionHeading
          eyebrow={t("landing.pricing.eyebrow")}
          title={t("landing.pricing.title")}
          subtitle={t("landing.pricing.subtitle")}
        />

        <div className="mx-auto mt-14 grid max-w-5xl items-stretch gap-5 sm:mt-16 md:grid-cols-3">
          {PLANS.map((plan, i) => (
            <Reveal key={plan.key} delay={i * 0.1} className="h-full">
              <div
                className={cn(
                  "relative flex h-full flex-col rounded-3xl border p-7",
                  plan.highlighted
                    ? "border-primary/50 bg-card shadow-primary/10 shadow-xl"
                    : "border-border/70 bg-card/60"
                )}
              >
                {plan.highlighted ? (
                  <span className="bg-primary text-primary-foreground absolute start-1/2 -top-3.5 inline-flex -translate-x-1/2 items-center gap-1.5 rounded-full px-3.5 py-1 text-xs font-semibold shadow-md rtl:translate-x-1/2">
                    <Sparkles className="size-3" />
                    {t("landing.pricing.pro.badge")}
                  </span>
                ) : null}

                <h3 className="font-heading text-2xl font-semibold tracking-tight">
                  {t(`landing.pricing.${plan.key}.name`)}
                </h3>
                <p className="text-muted-foreground mt-1.5 text-sm">{t(`landing.pricing.${plan.key}.tagline`)}</p>

                <ul className="mt-6 flex-1 space-y-2.5">
                  {plan.features.map((f) => (
                    <li key={f} className="flex items-start gap-2.5 text-sm">
                      <Check className="text-primary mt-0.5 size-4 shrink-0" />
                      <span>{t(`landing.pricing.${plan.key}.${f}`)}</span>
                    </li>
                  ))}
                </ul>

                <Link
                  to="/login"
                  className={cn(
                    buttonVariants({ variant: plan.highlighted ? "default" : "outline", size: "lg" }),
                    "mt-7 w-full rounded-full"
                  )}
                >
                  {t(`landing.pricing.${plan.key}.cta`)}
                </Link>
              </div>
            </Reveal>
          ))}
        </div>

        <Reveal delay={0.2}>
          <p className="tracking-caps text-muted-foreground/70 mt-8 text-center font-mono text-xs">
            {t("landing.pricing.note")}
          </p>
        </Reveal>
      </div>
    </section>
  )
}
