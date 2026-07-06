import { useTranslation } from "react-i18next"

import { Reveal } from "./shared"

const STATS = ["s1", "s2", "s3", "s4"] as const

export function Stats() {
  const { t } = useTranslation()

  return (
    <section className="border-border/60 bg-muted/30 border-y">
      <div className="container grid grid-cols-2 gap-x-6 gap-y-10 py-14 sm:py-16 lg:grid-cols-4">
        {STATS.map((key, i) => (
          <Reveal key={key} delay={i * 0.08} className="flex flex-col items-center gap-2 text-center">
            <span
              className="font-heading bg-clip-text text-4xl font-semibold tracking-tight text-transparent sm:text-5xl"
              style={{
                backgroundImage: "linear-gradient(150deg, var(--green-400), var(--green-700))",
              }}
            >
              {t(`landing.stats.${key}v`)}
            </span>
            <span className="text-muted-foreground max-w-40 text-sm leading-snug text-balance">
              {t(`landing.stats.${key}l`)}
            </span>
          </Reveal>
        ))}
      </div>
    </section>
  )
}
