import { SearchXIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { EmptyState } from "@/components/ui/empty-state"

/** Shown when an active search/filter returns no items (distinct from the
 * create-prompt empty state). Thin wrapper over the shared {@link EmptyState}. */
export function SectionNoResults() {
  const { t } = useTranslation()
  return (
    <EmptyState
      icon={SearchXIcon}
      title={t("org.session.controls.noResultsTitle")}
      description={t("org.session.controls.noResultsHint")}
    />
  )
}
