import { ChevronDown } from "lucide-react"
import { AnimatePresence, motion } from "motion/react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { EASE_OUT, Reveal, SectionHeading } from "./shared"

const ITEMS = [1, 2, 3, 4, 5] as const

export function Faq() {
  const { t } = useTranslation()
  const [open, setOpen] = useState<number | null>(1)

  return (
    <section id="faq" className="border-border/60 bg-muted/30 scroll-mt-20 border-t py-24 sm:py-32">
      <div className="container">
        <SectionHeading eyebrow={t("landing.faq.eyebrow")} title={t("landing.faq.title")} />

        <div className="mx-auto mt-12 max-w-3xl space-y-3 sm:mt-14">
          {ITEMS.map((n, i) => {
            const isOpen = open === n
            return (
              <Reveal key={n} delay={i * 0.06}>
                <div
                  className={cn(
                    "bg-card/60 overflow-hidden rounded-2xl border transition-colors",
                    isOpen ? "border-primary/40" : "border-border/70"
                  )}
                >
                  <button
                    type="button"
                    onClick={() => setOpen(isOpen ? null : n)}
                    aria-expanded={isOpen}
                    className="flex w-full items-center justify-between gap-4 px-5 py-4 text-start sm:px-6"
                  >
                    <span className="font-medium">{t(`landing.faq.q${n}`)}</span>
                    <motion.span
                      animate={{ rotate: isOpen ? 180 : 0 }}
                      transition={{ duration: 0.3, ease: EASE_OUT }}
                      className={cn("shrink-0", isOpen ? "text-primary" : "text-muted-foreground")}
                    >
                      <ChevronDown className="size-4" />
                    </motion.span>
                  </button>
                  <AnimatePresence initial={false}>
                    {isOpen ? (
                      <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: "auto", opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        transition={{ duration: 0.35, ease: EASE_OUT }}
                      >
                        <p className="text-muted-foreground px-5 pb-5 text-sm leading-relaxed sm:px-6">
                          {t(`landing.faq.a${n}`)}
                        </p>
                      </motion.div>
                    ) : null}
                  </AnimatePresence>
                </div>
              </Reveal>
            )
          })}
        </div>
      </div>
    </section>
  )
}
