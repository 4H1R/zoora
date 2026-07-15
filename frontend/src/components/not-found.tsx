import { Link, useRouter } from "@tanstack/react-router"
import { ArrowLeft, Home } from "lucide-react"
import { useTranslation } from "react-i18next"

import { StatusGlyph, StatusScreen } from "@/components/status-screen"
import { Button } from "@/components/ui/button"
import { useSeo } from "@/hooks/use-seo"

/** Rendered for unmatched routes via the root route's `notFoundComponent`. */
export function NotFound() {
  const { t } = useTranslation()
  const router = useRouter()

  useSeo("errorPages.notFound.title", "errorPages.notFound.description")

  return (
    <StatusScreen tone="brand">
      <StatusGlyph code="404" tone="brand" />

      <h1
        className="animate-reveal font-heading mt-6 leading-[1.12] font-semibold tracking-tight text-balance"
        style={{ animationDelay: "340ms", fontSize: "clamp(1.75rem, 5vw, 2.75rem)" }}
      >
        {t("errorPages.notFound.title")}
      </h1>

      <p className="text-muted-foreground mt-3 max-w-sm text-sm">{t("errorPages.notFound.description")}</p>

      <div
        className="animate-reveal mt-9 flex flex-col items-center gap-3 sm:flex-row"
        style={{ animationDelay: "470ms" }}
      >
        <Button size="lg" render={<Link to="/" />}>
          <Home />
          {t("errorPages.notFound.home")}
        </Button>
        <Button variant="outline" size="lg" onClick={() => router.history.back()}>
          <ArrowLeft className="rtl:rotate-180" />
          {t("errorPages.notFound.back")}
        </Button>
      </div>
    </StatusScreen>
  )
}
