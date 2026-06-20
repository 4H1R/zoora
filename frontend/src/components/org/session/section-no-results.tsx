import { SearchXIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

/** Shown when an active search/filter returns no items (distinct from the
 * create-prompt empty state). */
export function SectionNoResults() {
  const { t } = useTranslation()
  return (
    <div className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1">
      <SearchXIcon className="text-muted-foreground size-8" />
      <h3 className="text-foreground text-lg font-semibold tracking-tight">
        {t("org.session.controls.noResultsTitle")}
      </h3>
      <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
        {t("org.session.controls.noResultsHint")}
      </p>
    </div>
  )
}
