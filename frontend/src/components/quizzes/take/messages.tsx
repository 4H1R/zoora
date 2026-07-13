import type { ReactNode } from "react"

import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon, FileQuestionIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

import { DecorativeBackground } from "./decorations"

export function LoadingScreen() {
  return (
    <div className="relative isolate flex flex-col gap-8 py-16">
      <DecorativeBackground />
      <Skeleton className="h-4 w-40" />
      <Skeleton className="h-12 w-3/4" />
      <Skeleton className="h-72 w-full" />
    </div>
  )
}

interface CenterMessageProps {
  title: string
  description: string
  backHref: string
  /** Contextual glyph shown in the medallion. Defaults to a question mark. */
  icon?: ReactNode
  /** Errors tint the medallion + glow red; default is the brand accent. */
  tone?: "default" | "destructive"
}

export function CenterMessage({
  title,
  description,
  backHref,
  icon,
  tone = "default",
}: CenterMessageProps) {
  const { t } = useTranslation()
  const destructive = tone === "destructive"
  return (
    <div className="animate-in fade-in-50 relative isolate flex min-h-[70vh] flex-col items-center justify-center gap-7 text-center duration-500">
      <DecorativeBackground />

      {/* Medallion: soft glow behind a ringed tile, framed by two concentric
          halo rings — echoes the rounded-2xl + ring language of the start
          screen so the empty states feel part of the same system. */}
      <div className="animate-in fade-in zoom-in-95 relative flex size-40 items-center justify-center duration-700">
        <div
          aria-hidden
          className={cn(
            "absolute size-36 rounded-full blur-2xl",
            destructive ? "bg-destructive/20" : "bg-primary/20",
          )}
        />
        <div aria-hidden className="ring-foreground/5 absolute size-40 rounded-full ring-1" />
        <div aria-hidden className="ring-foreground/10 absolute size-28 rounded-full ring-1" />
        <div
          className={cn(
            "bg-card/80 ring-foreground/10 relative flex size-20 items-center justify-center rounded-2xl shadow-xl ring-1 backdrop-blur-sm [&_svg]:size-8",
            destructive ? "text-destructive" : "text-primary",
          )}
        >
          {icon ?? <FileQuestionIcon />}
        </div>
      </div>

      <div className="animate-in fade-in slide-in-from-bottom-2 flex flex-col items-center gap-4 delay-100 duration-500">
        <h1 className="text-3xl font-semibold tracking-tight text-balance md:text-4xl">{title}</h1>
        <p className="text-muted-foreground max-w-md text-base leading-relaxed text-pretty">
          {description}
        </p>
      </div>

      <Button
        variant="outline"
        size="lg"
        className="animate-in fade-in delay-200 duration-500"
        render={<Link to={backHref} />}
      >
        <ArrowLeftIcon className="size-4" />
        {t("org.session.quizzes.take.backToSession")}
      </Button>
    </div>
  )
}
