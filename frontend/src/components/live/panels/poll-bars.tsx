import { useTranslation } from "react-i18next"

import { resolveOptionLabel } from "../poll-labels"

interface PollBarsProps {
  options: { label: string; value: string }[]
  counts: Record<string, number>
  total: number
}

// Horizontal result bars shared by the host tally, the reveal view, the viewer
// vote modal, and the history panel.
export function PollBars({ options, counts, total }: PollBarsProps) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-2">
      {options.map((opt) => {
        const count = counts[opt.value] ?? 0
        const pct = total > 0 ? Math.round((count / total) * 100) : 0
        const label = resolveOptionLabel(opt.value, opt.label, t)
        return (
          <div key={opt.value} className="flex flex-col gap-1">
            <div className="text-muted-foreground flex items-center justify-between text-xs">
              <span>{label}</span>
              <span className="text-muted-foreground font-mono">{t("liveRoom.polls.votes", { count })}</span>
            </div>
            <div
              className="bg-muted relative h-5 overflow-hidden rounded-sm"
              role="meter"
              aria-valuenow={pct}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label={label}
            >
              <div
                className="bg-primary absolute inset-y-0 start-0 rounded-sm transition-all duration-500"
                style={{ width: `${pct}%` }}
              />
              <span className="text-primary-foreground absolute inset-0 flex items-center px-2 text-[10px] font-semibold mix-blend-normal">
                {pct}%
              </span>
            </div>
          </div>
        )
      })}
      <p className="text-muted-foreground mt-1 text-right text-[10px]">{t("liveRoom.polls.votes", { count: total })}</p>
    </div>
  )
}
