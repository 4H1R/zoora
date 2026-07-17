import type { ErrorComponentProps } from "@tanstack/react-router"

import { Link, useRouter } from "@tanstack/react-router"
import { Home, RotateCw } from "lucide-react"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { Eyebrow } from "@/components/eyebrow"
import { StatusGlyph, StatusScreen } from "@/components/status-screen"
import { Button } from "@/components/ui/button"
import { useSeo } from "@/hooks/use-seo"
import { reportError } from "@/lib/sentry"

/** Rendered for thrown render/loader errors via the root route's `errorComponent`. */
export function ErrorScreen({ error, reset }: ErrorComponentProps) {
  const { t } = useTranslation()
  const router = useRouter()

  useSeo("errorPages.serverError.title", "errorPages.serverError.description")

  // Report the render/loader error that tripped this boundary (no-op when
  // Sentry is disabled).
  useEffect(() => {
    reportError(error)
  }, [error])

  const handleRetry = () => {
    router.invalidate()
    reset()
  }

  return (
    <StatusScreen tone="alert">
      <StatusGlyph code="500" tone="alert" />

      <div className="animate-reveal mt-2" style={{ animationDelay: "280ms" }}>
        <Eyebrow>{t("errorPages.serverError.eyebrow")}</Eyebrow>
      </div>

      <h1
        className="animate-reveal font-heading mt-4 leading-[1.12] font-semibold tracking-tight text-balance"
        style={{ animationDelay: "340ms", fontSize: "clamp(1.75rem, 5vw, 2.75rem)" }}
      >
        {t("errorPages.serverError.title")}
      </h1>

      <p
        className="animate-reveal text-muted-foreground mt-5 max-w-md text-base leading-relaxed text-pretty"
        style={{ animationDelay: "400ms" }}
      >
        {t("errorPages.serverError.description")}
      </p>

      {import.meta.env.DEV && error?.message && (
        <details className="animate-reveal mt-7 w-full max-w-lg text-start" style={{ animationDelay: "440ms" }}>
          <summary className="tracking-caps text-muted-foreground/80 hover:text-foreground cursor-pointer font-mono text-xs uppercase transition-colors">
            {t("errorPages.serverError.details")}
          </summary>
          <pre className="border-destructive/20 bg-destructive/5 text-destructive mt-3 max-h-56 overflow-auto rounded-lg border p-4 text-start font-mono text-xs leading-relaxed whitespace-pre-wrap">
            {error.message}
          </pre>
        </details>
      )}

      <div
        className="animate-reveal mt-9 flex flex-col items-center gap-3 sm:flex-row"
        style={{ animationDelay: "490ms" }}
      >
        <Button size="lg" onClick={handleRetry}>
          <RotateCw />
          {t("errorPages.serverError.retry")}
        </Button>
        <Button variant="outline" size="lg" render={<Link to="/" />}>
          <Home />
          {t("errorPages.serverError.home")}
        </Button>
      </div>
    </StatusScreen>
  )
}
