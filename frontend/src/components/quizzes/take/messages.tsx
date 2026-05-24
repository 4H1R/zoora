import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"

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
}

export function CenterMessage({ title, description, backHref }: CenterMessageProps) {
  const { t } = useTranslation()
  return (
    <div className="relative isolate flex min-h-[60vh] flex-col items-start justify-center gap-4">
      <DecorativeBackground />
      <Eyebrow>{t("org.session.quizzes.take.eyebrow")}</Eyebrow>
      <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{title}</h1>
      <p className="text-muted-foreground max-w-md text-base leading-relaxed">{description}</p>
      <Button variant="outline" render={<Link to={backHref} />}>
        <ArrowLeftIcon className="size-4" />
        {t("org.session.quizzes.take.backToSession")}
      </Button>
    </div>
  )
}
