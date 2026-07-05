import { Link } from "@tanstack/react-router"
import { SparklesIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetChangelogStatus } from "@/api/changelog/changelog"
import { Badge } from "@/components/ui/badge"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function WhatsNewButton() {
  const { t } = useTranslation()
  const { data } = useGetChangelogStatus()
  const status = (data?.status === 200 && data.data.data) || undefined
  const unseen = status?.unseen_count ?? 0

  return (
    <Link
      to="/org/whats-new"
      title={t("whatsNew.title")}
      className={cn(buttonVariants({ variant: "ghost", size: "icon" }), "relative")}
    >
      <SparklesIcon className="size-5" />
      {unseen > 0 && (
        <Badge
          variant="default"
          className="absolute -end-1 -top-1 h-4 min-w-4 justify-center rounded-full px-1 text-[10px]"
        >
          {unseen > 9 ? "9+" : unseen}
        </Badge>
      )}
    </Link>
  )
}
