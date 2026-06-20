import { Link, useRouter } from "@tanstack/react-router"
import { ArrowLeft, Home } from "lucide-react"
import { useTranslation } from "react-i18next"

import { StatusGlyph, StatusScreen } from "@/components/status-screen"
import { Button } from "@/components/ui/button"

/** Rendered for unmatched routes via the root route's `notFoundComponent`. */
export function NotFound() {
  const { t } = useTranslation()
  const router = useRouter()

  return (
    <StatusScreen tone="brand">
      <StatusGlyph code="404" tone="brand" />

      <h1
        className="animate-reveal mt-6 font-heading font-semibold leading-[1.12] tracking-tight text-balance"
        style={{ animationDelay: "340ms", fontSize: "clamp(1.75rem, 5vw, 2.75rem)" }}
      >
        {t("errorPages.notFound.title")}
      </h1>

      <p
        className="animate-reveal mt-5 max-w-md text-base leading-relaxed text-muted-foreground text-pretty"
        style={{ animationDelay: "400ms" }}
      >
        {t("errorPages.notFound.description")}
      </p>

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
