import { Link } from "@tanstack/react-router"
import {
  ArrowRight,
  Hand,
  MessageSquare,
  Mic,
  MicOff,
  PenLine,
  PhoneOff,
  ScreenShare,
  User,
  Users,
  Video,
} from "lucide-react"
import { motion, useReducedMotion } from "motion/react"
import { useTranslation } from "react-i18next"

import { BackgroundFX } from "@/components/background-fx"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

import { EASE_OUT, Float, scrollToSection } from "./shared"

export function Hero() {
  const { t } = useTranslation()
  const reduced = useReducedMotion()

  const enter = (delay: number) => ({
    initial: reduced ? { opacity: 0 } : { opacity: 0, y: 24, filter: "blur(6px)" },
    animate: reduced ? { opacity: 1 } : { opacity: 1, y: 0, filter: "blur(0px)" },
    transition: { duration: 0.9, delay, ease: EASE_OUT },
  })

  return (
    <section className="relative overflow-hidden pt-32 pb-16 sm:pt-40 sm:pb-24">
      <BackgroundFX />

      <div className="relative z-10 container flex flex-col items-center text-center">
        <motion.span
          {...enter(0.05)}
          className="border-border/70 bg-background/40 tracking-caps text-muted-foreground inline-flex items-center gap-2.5 rounded-full border py-1.5 ps-2 pe-3.5 font-mono text-[0.7rem] uppercase backdrop-blur-md"
        >
          <span className="relative flex size-2.5 items-center justify-center">
            <span className="bg-primary/60 absolute inline-flex size-full animate-ping rounded-full" />
            <span className="bg-primary relative inline-flex size-2 rounded-full" />
          </span>
          {t("landing.hero.badge")}
        </motion.span>

        <motion.h1
          {...enter(0.15)}
          className="font-heading mt-8 leading-[1.1] font-semibold tracking-tight text-balance"
          style={{ fontSize: "clamp(2.75rem, 7.5vw, 5.5rem)" }}
        >
          <span className="block">{t("landing.hero.titleLine1")}</span>
          <span
            className="block bg-clip-text pb-[0.15em] text-transparent"
            style={{
              backgroundImage: "linear-gradient(115deg, var(--green-400), var(--green-600) 55%, var(--green-800))",
            }}
          >
            {t("landing.hero.titleLine2")}
          </span>
        </motion.h1>

        <motion.p
          {...enter(0.25)}
          className="text-muted-foreground mt-6 max-w-xl text-base leading-relaxed text-pretty sm:text-lg"
        >
          {t("landing.hero.subtitle")}
        </motion.p>

        <motion.div {...enter(0.35)} className="mt-9 flex flex-wrap items-center justify-center gap-3">
          <Link
            to="/login"
            className={cn(
              buttonVariants({ size: "lg" }),
              "group shadow-primary/25 h-12 rounded-full px-7 text-base shadow-lg"
            )}
          >
            {t("landing.hero.ctaPrimary")}
            <ArrowRight className="transition-transform group-hover:translate-x-0.5 rtl:-scale-x-100 rtl:group-hover:-translate-x-0.5" />
          </Link>
          <button
            type="button"
            onClick={() => scrollToSection("features")}
            className={cn(buttonVariants({ variant: "outline", size: "lg" }), "h-12 rounded-full px-7 text-base")}
          >
            {t("landing.hero.ctaSecondary")}
          </button>
        </motion.div>

        <motion.p {...enter(0.45)} className="tracking-caps text-muted-foreground/70 mt-5 font-mono text-xs">
          {t("landing.hero.note")}
        </motion.p>

        <RoomMock />
      </div>
    </section>
  )
}

/** Stylized live-classroom UI — the hero's visual proof, built from the real product's parts. */
function RoomMock() {
  const { t } = useTranslation()
  const reduced = useReducedMotion()

  const tiles = [
    { tone: "var(--green-500)", speaking: true, big: true },
    { tone: "var(--neutral-400)", muted: true },
    { tone: "var(--green-700)" },
    { tone: "var(--neutral-500)" },
    { tone: "var(--green-600)", muted: true },
  ]

  return (
    <div className="relative mt-16 w-full max-w-4xl sm:mt-20" style={{ perspective: "1400px" }}>
      <motion.div
        initial={reduced ? { opacity: 0 } : { opacity: 0, y: 64, rotateX: 14 }}
        whileInView={reduced ? { opacity: 1 } : { opacity: 1, y: 0, rotateX: 0 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 1, delay: 0.2, ease: EASE_OUT }}
        className="border-border/70 bg-card/80 overflow-hidden rounded-3xl border text-start shadow-xl backdrop-blur-xl"
      >
        {/* Room top bar */}
        <div className="border-border/60 flex items-center justify-between border-b px-4 py-3 sm:px-5">
          <div className="flex items-center gap-3">
            <span className="hidden items-center gap-1.5 sm:flex" aria-hidden>
              <span className="bg-live/60 size-2.5 rounded-full" />
              <span className="bg-warning/60 size-2.5 rounded-full" />
              <span className="bg-success/60 size-2.5 rounded-full" />
            </span>
            <div>
              <p className="text-sm leading-tight font-medium">{t("landing.hero.room.title")}</p>
              <p className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                <Users className="size-3" />
                {t("landing.hero.room.participants")}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className="bg-live/10 tracking-caps text-live inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-mono text-[0.65rem] uppercase">
              <span className="animate-pulse-dot bg-live size-1.5 rounded-full" />
              {t("landing.hero.room.live")}
            </span>
            <span className="border-border/70 tracking-caps text-muted-foreground hidden items-center gap-1.5 rounded-full border px-2.5 py-1 font-mono text-[0.65rem] uppercase sm:inline-flex">
              <span className="bg-live/70 size-1.5 rounded-full" />
              {t("landing.hero.room.rec")}
            </span>
          </div>
        </div>

        {/* Video grid */}
        <div className="grid grid-cols-3 gap-2 p-3 sm:gap-2.5 sm:p-4">
          {tiles.map((tile, i) => (
            <div
              key={i}
              className={cn(
                "relative flex items-center justify-center overflow-hidden rounded-xl",
                tile.big ? "col-span-2 row-span-2 aspect-auto" : "aspect-video"
              )}
              style={{
                // oklab, not oklch: mixing green into near-achromatic zinc in oklch
                // spins the hue through zinc's 286° axis and the tiles turn blue.
                background: `linear-gradient(145deg, color-mix(in oklab, ${tile.tone} 28%, var(--muted)), color-mix(in oklab, ${tile.tone} 10%, var(--muted)))`,
              }}
            >
              {tile.speaking && !reduced ? (
                <motion.span
                  aria-hidden
                  className="ring-primary/70 absolute inset-0 rounded-xl ring-2 ring-inset"
                  animate={{ opacity: [0.9, 0.3, 0.9] }}
                  transition={{ duration: 2.2, repeat: Infinity, ease: "easeInOut" }}
                />
              ) : null}
              <span
                className={cn(
                  "bg-background/50 flex items-center justify-center rounded-full backdrop-blur-sm",
                  tile.big ? "size-14 sm:size-16" : "size-8 sm:size-9"
                )}
              >
                <User className={cn("text-foreground/70", tile.big ? "size-7" : "size-4")} />
              </span>
              <span className="bg-background/60 absolute start-1.5 bottom-1.5 flex size-5 items-center justify-center rounded-md backdrop-blur-sm">
                {tile.muted ? (
                  <MicOff className="text-muted-foreground size-3" />
                ) : (
                  <Mic className="text-primary size-3" />
                )}
              </span>
            </div>
          ))}
        </div>

        {/* Control bar */}
        <div className="border-border/60 flex items-center justify-center gap-2 border-t px-4 py-3 sm:gap-2.5">
          {[Mic, Video, ScreenShare, Hand, PenLine, MessageSquare].map((Icon, i) => (
            <span
              key={i}
              className="border-border/70 bg-background/60 text-muted-foreground flex size-9 items-center justify-center rounded-full border"
            >
              <Icon className="size-4" />
            </span>
          ))}
          <span className="bg-live ms-1.5 flex h-9 items-center justify-center rounded-full px-4 text-white">
            <PhoneOff className="size-4" />
          </span>
        </div>
      </motion.div>

      {/* Floating proof cards */}
      <Float delay={0.9} className="absolute -end-4 top-14 hidden w-auto md:block lg:-end-14">
        <div className="border-border/70 bg-card/90 flex items-center gap-2.5 rounded-2xl border px-4 py-3 text-start shadow-lg backdrop-blur-xl">
          <span className="bg-warning/15 text-warning flex size-8 items-center justify-center rounded-full">
            <Hand className="size-4" />
          </span>
          <p className="text-sm font-medium">{t("landing.hero.room.hand")}</p>
        </div>
      </Float>

      <Float delay={1.15} drift={10} className="absolute -start-4 bottom-28 hidden w-56 md:block lg:-start-16">
        <div className="border-border/70 bg-card/90 rounded-2xl border p-4 text-start shadow-lg backdrop-blur-xl">
          <p className="text-sm font-medium">{t("landing.hero.room.pollQuestion")}</p>
          <div className="mt-3 space-y-2">
            <PollBar label={t("landing.hero.room.pollYes")} pct={72} delay={1.5} />
            <PollBar label={t("landing.hero.room.pollNo")} pct={28} delay={1.65} muted />
          </div>
        </div>
      </Float>

      <Float delay={1.35} drift={7} className="absolute -end-2 bottom-6 hidden w-60 md:block lg:-end-20">
        <div className="border-border/70 bg-card/90 flex items-start gap-2.5 rounded-2xl border px-4 py-3 text-start shadow-lg backdrop-blur-xl">
          <span
            className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white"
            style={{ background: "linear-gradient(145deg, var(--green-500), var(--green-700))" }}
          >
            {t("landing.hero.room.chatName").slice(0, 1)}
          </span>
          <div>
            <p className="text-muted-foreground text-xs font-medium">{t("landing.hero.room.chatName")}</p>
            <p className="mt-0.5 text-sm leading-snug">{t("landing.hero.room.chatMsg")}</p>
          </div>
        </div>
      </Float>
    </div>
  )
}

function PollBar({ label, pct, delay, muted }: { label: string; pct: number; delay: number; muted?: boolean }) {
  const reduced = useReducedMotion()
  return (
    <div>
      <div className="flex items-center justify-between text-xs">
        <span className={muted ? "text-muted-foreground" : "font-medium"}>{label}</span>
        <span className="text-muted-foreground font-mono">{pct}%</span>
      </div>
      <div className="bg-muted mt-1 h-1.5 overflow-hidden rounded-full">
        <motion.div
          className={cn("h-full rounded-full", muted ? "bg-muted-foreground/40" : "bg-primary")}
          initial={{ width: reduced ? `${pct}%` : "0%" }}
          whileInView={{ width: `${pct}%` }}
          viewport={{ once: true }}
          transition={{ duration: 1, delay, ease: EASE_OUT }}
        />
      </div>
    </div>
  )
}
