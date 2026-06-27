import { BarChart3 } from "lucide-react"
import { useTranslation } from "react-i18next"

export function PollsPanel() {
  const { t } = useTranslation()
  return (
    <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 p-6 text-center text-zinc-500">
      <BarChart3 className="size-7 opacity-40" />
      <p className="text-sm">{t("liveRoom.polls.comingSoon")}</p>
    </div>
  )
}
