import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, FolderLockIcon, Share2Icon, SparklesIcon, UploadCloudIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/files/")({
  head: () => orgHead("org.nav.files"),
  component: RouteComponent,
})

const FEATURES = [
  { key: "upload", icon: UploadCloudIcon },
  { key: "organize", icon: FolderLockIcon },
  { key: "share", icon: Share2Icon },
] as const

// Each floating card keeps its resting rotation/offset while the float keyframe
// animates between --float-base and --float-lift (see styles.css).
const DECK = [
  {
    className: "animate-float-slow start-1 top-5 size-20 -rotate-[14deg]",
    style: {
      "--float-base": "translateY(0) rotate(-14deg)",
      "--float-lift": "translateY(-7px) rotate(-14deg)",
    },
    widths: ["w-full", "w-2/3", "w-3/4"],
    front: false,
  },
  {
    className: "animate-float end-1 top-4 size-20 rotate-[14deg]",
    style: {
      "animationDelay": "-1.6s",
      "--float-base": "translateY(0) rotate(14deg)",
      "--float-lift": "translateY(-7px) rotate(14deg)",
    },
    widths: ["w-3/4", "w-full", "w-1/2"],
    front: false,
  },
  {
    className:
      "animate-float bg-card start-1/2 top-0 size-24 -translate-x-1/2 border-primary/30 shadow-primary/10 rtl:translate-x-1/2",
    style: {
      "--float-base": "translateY(0)",
      "--float-lift": "translateY(-10px)",
    },
    widths: ["w-full", "w-5/6", "w-2/3"],
    front: true,
  },
] as const

function RouteComponent() {
  const { t } = useTranslation()
  const { orgId } = Route.useParams()
  const allowed = useOrgGuard(["media:view", "media:view_any"])
  if (!allowed) return null

  return (
    <div className="relative isolate flex min-h-[70vh] flex-col items-center justify-center overflow-hidden rounded-2xl border px-6 py-16 text-center sm:py-24">
      {/* Atmosphere: primary glow + faint grid, fading toward the edges */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_70%_55%_at_50%_0%,var(--color-primary)/12%,transparent_70%)]"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 opacity-50 [mask-image:radial-gradient(ellipse_55%_45%_at_50%_35%,black,transparent)]"
        style={{
          backgroundImage:
            "linear-gradient(to right, var(--border) 1px, transparent 1px), linear-gradient(to bottom, var(--border) 1px, transparent 1px)",
          backgroundSize: "44px 44px",
        }}
      />

      {/* Floating document deck */}
      <div aria-hidden className="relative mb-12 h-28 w-48">
        <div className="bg-primary/25 absolute inset-x-8 top-6 -z-10 h-20 rounded-full blur-2xl" />
        {DECK.map((card, i) => (
          <div
            key={i}
            className={`absolute flex flex-col gap-1.5 rounded-xl border p-2.5 ${card.front ? "" : "border-border/70 bg-card/95"} shadow-lg backdrop-blur-sm ${card.className}`}
            style={card.style as React.CSSProperties}
          >
            <div className={`mb-0.5 h-4 rounded-md ${card.front ? "bg-primary/20" : "bg-muted"}`} />
            {card.widths.map((w, j) => (
              <div key={j} className={`h-1.5 rounded-full ${card.front ? "bg-primary/15" : "bg-muted"} ${w}`} />
            ))}
          </div>
        ))}
      </div>

      {/* Badge */}
      <span className="bg-primary/10 text-primary ring-primary/20 mb-5 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium ring-1 ring-inset">
        <SparklesIcon className="size-3.5 shrink-0" />
        {t("org.files.comingSoon.badge")}
      </span>

      <h1 className="max-w-xl text-3xl font-bold tracking-tight text-balance sm:text-4xl">
        {t("org.files.comingSoon.title")}
      </h1>
      <p className="text-muted-foreground mt-3 max-w-md text-sm text-pretty sm:text-base">
        {t("org.files.comingSoon.description")}
      </p>

      {/* Feature preview */}
      <div className="mt-10 grid w-full max-w-2xl grid-cols-1 gap-3 sm:grid-cols-3">
        {FEATURES.map(({ key, icon: Icon }) => (
          <div
            key={key}
            className="bg-background/50 flex flex-col items-center gap-2 rounded-xl border p-4 text-center backdrop-blur-sm transition-colors"
          >
            <span className="bg-primary/10 text-primary flex size-9 items-center justify-center rounded-lg">
              <Icon className="size-5" />
            </span>
            <p className="text-sm font-medium">{t(`org.files.comingSoon.features.${key}.title`)}</p>
            <p className="text-muted-foreground text-xs leading-snug">
              {t(`org.files.comingSoon.features.${key}.desc`)}
            </p>
          </div>
        ))}
      </div>

      {/* Status + back */}
      <div className="mt-10 flex flex-col items-center gap-4">
        <span className="text-muted-foreground inline-flex items-center gap-2 text-xs font-medium">
          <span className="relative flex size-2">
            <span className="bg-primary/40 absolute inline-flex size-full animate-ping rounded-full" />
            <span className="bg-primary relative inline-flex size-2 rounded-full" />
          </span>
          {t("org.files.comingSoon.status")}
        </span>
        <Link to="/org/$orgId/dashboard" params={{ orgId }}>
          <Button variant="outline" size="sm">
            <ArrowLeftIcon data-icon="inline-start" className="rtl:rotate-180" />
            {t("org.files.comingSoon.back")}
          </Button>
        </Link>
      </div>
    </div>
  )
}
