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
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>{label}</span>
              <span className="font-mono text-muted-foreground">
                {t("liveRoom.polls.votes", { count })}
              </span>
            </div>
            <div
              className="relative h-5 overflow-hidden rounded-sm bg-muted"
              role="meter"
              aria-valuenow={pct}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label={label}
            >
              <div
                className="absolute inset-y-0 start-0 rounded-sm bg-primary transition-all duration-500"
                style={{ width: `${pct}%` }}
              />
              <span className="absolute inset-0 flex items-center px-2 text-[10px] font-semibold text-primary-foreground mix-blend-normal">
                {pct}%
              </span>
            </div>
          </div>
        )
      })}
      <p className="mt-1 text-right text-[10px] text-muted-foreground">
        {t("liveRoom.polls.votes", { count: total })}
      </p>
    </div>
  )
}
