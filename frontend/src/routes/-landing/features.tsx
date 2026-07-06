import type { LucideIcon } from "lucide-react"
import type { ReactNode } from "react"

import {
  Bell,
  MessageCircle,
  NotebookPen,
  PenLine,
  PlayCircle,
  Radio,
  Send,
  ShieldCheck,
  Smartphone,
} from "lucide-react"
import { motion, useReducedMotion } from "motion/react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { Reveal, SectionHeading } from "./shared"

export function Features() {
  const { t } = useTranslation()

  return (
    <section id="features" className="scroll-mt-20 py-24 sm:py-32">
      <div className="container">
        <SectionHeading
          eyebrow={t("landing.features.eyebrow")}
          title={
            <>
              {t("landing.features.title")}{" "}
              <span
                className="bg-clip-text text-transparent"
                style={{
                  backgroundImage: "linear-gradient(115deg, var(--green-400), var(--green-700))",
                }}
              >
                {t("landing.features.titleAccent")}
              </span>
            </>
          }
          subtitle={t("landing.features.subtitle")}
        />

        <div className="mt-14 grid gap-4 sm:mt-16 md:grid-cols-6">
          <FeatureTile
            index={0}
            icon={Radio}
            title={t("landing.features.live.title")}
            desc={t("landing.features.live.desc")}
            className="md:col-span-4"
          >
            <div className="mt-5 flex flex-wrap gap-2">
              {(["tag1", "tag2", "tag3", "tag4"] as const).map((tag) => (
                <span
                  key={tag}
                  className="border-primary/25 bg-primary/8 text-primary rounded-full border px-3 py-1 text-xs font-medium"
                >
                  {t(`landing.features.live.${tag}`)}
                </span>
              ))}
            </div>
          </FeatureTile>

          <FeatureTile
            index={1}
            icon={PenLine}
            title={t("landing.features.whiteboard.title")}
            desc={t("landing.features.whiteboard.desc")}
            className="md:col-span-2"
          />

          <FeatureTile
            index={2}
            icon={ShieldCheck}
            title={t("landing.features.exams.title")}
            desc={t("landing.features.exams.desc")}
            className="md:col-span-3"
          />

          <FeatureTile
            index={3}
            icon={NotebookPen}
            title={t("landing.features.homework.title")}
            desc={t("landing.features.homework.desc")}
            className="md:col-span-3"
          />

          <FeatureTile
            index={4}
            icon={PlayCircle}
            title={t("landing.features.recordings.title")}
            desc={t("landing.features.recordings.desc")}
            className="md:col-span-2"
          />

          <FeatureTile
            index={5}
            icon={Bell}
            title={t("landing.features.notifications.title")}
            desc={t("landing.features.notifications.desc")}
            className="md:col-span-4"
          >
            <div className="mt-5 flex flex-wrap gap-2">
              {(
                [
                  ["ch1", Send],
                  ["ch2", MessageCircle],
                  ["ch3", Smartphone],
                  ["ch4", Bell],
                ] as [string, LucideIcon][]
              ).map(([key, Icon]) => (
                <span
                  key={key}
                  className="border-border/70 bg-background/50 text-muted-foreground inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium"
                >
                  <Icon className="text-primary size-3" />
                  {t(`landing.features.notifications.${key}`)}
                </span>
              ))}
            </div>
          </FeatureTile>
        </div>
      </div>
    </section>
  )
}

interface FeatureTileProps {
  index: number
  icon: LucideIcon
  title: string
  desc: string
  className?: string
  children?: ReactNode
}

function FeatureTile({ index, icon: Icon, title, desc, className, children }: FeatureTileProps) {
  const reduced = useReducedMotion()
  return (
    <Reveal delay={(index % 3) * 0.08} className={cn("h-full", className)}>
      <motion.div
        whileHover={reduced ? undefined : { y: -4 }}
        transition={{ duration: 0.25, ease: "easeOut" }}
        className="group border-border/70 bg-card/60 hover:border-primary/40 relative h-full overflow-hidden rounded-3xl border p-6 transition-colors sm:p-8"
      >
        <div
          aria-hidden
          className="pointer-events-none absolute -end-20 -top-20 size-48 rounded-full opacity-0 blur-3xl transition-opacity duration-500 group-hover:opacity-100"
          style={{ background: "color-mix(in oklch, var(--primary) 18%, transparent)" }}
        />
        <span className="bg-primary/10 text-primary relative flex size-11 items-center justify-center rounded-xl">
          <Icon className="size-5" />
        </span>
        <h3 className="font-heading relative mt-5 text-lg font-semibold tracking-tight">{title}</h3>
        <p className="text-muted-foreground relative mt-2 text-sm leading-relaxed">{desc}</p>
        {children}
      </motion.div>
    </Reveal>
  )
}
