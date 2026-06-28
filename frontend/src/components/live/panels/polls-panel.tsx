import { BarChart3, Plus, Trash2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetPollsIdResults, usePostPolls, usePostPollsIdAnswer } from "@/api/polls/polls"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

import type { LivePoll, PollResults } from "../use-room-polls"
import { useRoomPolls } from "../use-room-polls"

interface PollsBarProps {
  options: { label: string; value: string }[]
  counts: Record<string, number>
  total: number
}

function PollsBar({ options, counts, total }: PollsBarProps) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-2">
      {options.map((opt) => {
        const count = counts[opt.value] ?? 0
        const pct = total > 0 ? Math.round((count / total) * 100) : 0
        return (
          <div key={opt.value} className="flex flex-col gap-1">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>{opt.label}</span>
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
              aria-label={opt.label}
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

// ---------------------------------------------------------------------------
// Host: Create-poll form
// ---------------------------------------------------------------------------
type PollMode = "single" | "yesno"

interface CreatePollFormProps {
  liveId: string
  onLaunch: (poll: LivePoll) => void
}

function CreatePollForm({ liveId, onLaunch }: CreatePollFormProps) {
  const { t } = useTranslation()
  const [question, setQuestion] = useState("")
  const [mode, setMode] = useState<PollMode>("single")
  const [options, setOptions] = useState(["", ""])

  const createPoll = usePostPolls()

  function addOption() {
    if (options.length < 5) setOptions([...options, ""])
  }

  function removeOption(i: number) {
    if (options.length <= 2) return
    setOptions(options.filter((_, idx) => idx !== i))
  }

  function setOption(i: number, val: string) {
    const next = [...options]
    next[i] = val
    setOptions(next)
  }

  const yesNoOptions = [
    { label: t("liveRoom.polls.yes"), value: "yes" },
    { label: t("liveRoom.polls.no"), value: "no" },
  ]

  function handleLaunch() {
    const name = question.trim()
    if (!name) return
    const resolvedOptions = mode === "yesno" ? yesNoOptions : options.map((o, i) => ({ label: o.trim() || `${t("liveRoom.polls.option")} ${i + 1}`, value: String(i) }))

    createPoll.mutate(
      {
        data: {
          model_type: "live_session",
          model_id: liveId,
          name,
          allowed_answers_count: 1,
          options: resolvedOptions,
        },
      },
      {
        onSuccess: (res) => {
          if (res.status !== 201) return
          const id = res.data.data?.id
          if (!id) return
          onLaunch({
            pollId: id,
            name,
            options: resolvedOptions,
            allowedAnswersCount: 1,
          })
          setQuestion("")
          setMode("single")
          setOptions(["", ""])
        },
      },
    )
  }

  const canLaunch = question.trim().length >= 2 && (mode === "yesno" || options.every((o) => o.trim().length > 0))

  return (
    <div className="flex flex-col gap-3 p-3">
      {/* Question */}
      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">{t("liveRoom.polls.question")}</label>
        <Input
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          placeholder={t("liveRoom.polls.question")}
          className="border-border bg-transparent text-foreground placeholder:text-muted-foreground focus-visible:ring-ring/40"
        />
      </div>

      {/* Mode toggle */}
      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => setMode("single")}
          className={cn(
            "flex-1 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors",
            mode === "single"
              ? "border-primary bg-primary text-primary-foreground"
              : "border-border bg-muted text-muted-foreground hover:bg-accent",
          )}
        >
          {t("liveRoom.polls.singleChoice")}
        </button>
        <button
          type="button"
          onClick={() => setMode("yesno")}
          className={cn(
            "flex-1 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors",
            mode === "yesno"
              ? "border-primary bg-primary text-primary-foreground"
              : "border-border bg-muted text-muted-foreground hover:bg-accent",
          )}
        >
          {t("liveRoom.polls.yesNo")}
        </button>
      </div>

      {/* Single-choice options */}
      {mode === "single" && (
        <div className="flex flex-col gap-2">
          {options.map((opt, i) => (
            <div key={i} className="flex items-center gap-2">
              <Input
                value={opt}
                onChange={(e) => setOption(i, e.target.value)}
                placeholder={`${t("liveRoom.polls.option")} ${i + 1}`}
                className="border-border bg-transparent text-foreground placeholder:text-muted-foreground focus-visible:ring-ring/40"
              />
              {options.length > 2 && (
                <button
                  type="button"
                  onClick={() => removeOption(i)}
                  aria-label={t("liveRoom.polls.removeOption")}
                  className="flex size-8 shrink-0 items-center justify-center rounded text-muted-foreground hover:text-red-400"
                >
                  <Trash2 className="size-3.5" />
                </button>
              )}
            </div>
          ))}
          {options.length < 5 && (
            <button
              type="button"
              onClick={addOption}
              className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
            >
              <Plus className="size-3.5" />
              {t("liveRoom.polls.addOption")}
            </button>
          )}
        </div>
      )}

      <Button
        onClick={handleLaunch}
        disabled={!canLaunch || createPoll.isPending}
        className="w-full"
        size="sm"
      >
        {t("liveRoom.polls.launch")}
      </Button>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Host: Active poll (live tally)
// ---------------------------------------------------------------------------
interface HostActivePollProps {
  activePoll: LivePoll
  onReveal: (results: PollResults) => void
  onClose: () => void
}

function HostActivePoll({ activePoll, onReveal, onClose }: HostActivePollProps) {
  const { t } = useTranslation()

  const resultsQuery = useGetPollsIdResults(activePoll.pollId, {
    query: { refetchInterval: 3000 },
  })

  const resultsData = resultsQuery.data?.status === 200 ? resultsQuery.data.data.data : null
  const counts: Record<string, number> = {}
  for (const opt of activePoll.options) {
    counts[opt.value] = resultsData?.counts?.[opt.value] ?? 0
  }
  const total = resultsData?.total ?? 0

  function handleReveal() {
    onReveal({ counts, total })
  }

  return (
    <div className="flex flex-col gap-4 p-3">
      <p className="text-sm font-semibold text-foreground">{activePoll.name}</p>
      <PollsBar options={activePoll.options} counts={counts} total={total} />
      <div className="flex gap-2">
        <Button size="sm" className="flex-1" onClick={handleReveal}>
          {t("liveRoom.polls.reveal")}
        </Button>
        <Button size="sm" variant="outline" className="flex-1 border-border text-muted-foreground hover:bg-accent" onClick={onClose}>
          {t("liveRoom.polls.close")}
        </Button>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main PollsPanel
// ---------------------------------------------------------------------------
interface PollsPanelProps {
  liveId: string
  isHost: boolean
}

export function PollsPanel({ liveId, isHost }: PollsPanelProps) {
  const { t } = useTranslation()
  const { activePoll, results, hasAnswered, launchPoll, revealResults, closePoll, markAnswered } =
    useRoomPolls()

  const answerMutation = usePostPollsIdAnswer()

  function handleVote(value: string) {
    if (!activePoll) return
    answerMutation.mutate(
      { id: activePoll.pollId, data: { options: [value] } },
      { onSuccess: () => markAnswered() },
    )
  }

  // ---- HOST ----
  if (isHost) {
    if (!activePoll) {
      return (
        <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
          <p className="px-3 pt-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {t("liveRoom.polls.create")}
          </p>
          <CreatePollForm liveId={liveId} onLaunch={launchPoll} />
        </div>
      )
    }

    return (
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
        <HostActivePoll
          activePoll={activePoll}
          onReveal={revealResults}
          onClose={closePoll}
        />
      </div>
    )
  }

  // ---- VIEWER ----

  // No active poll
  if (!activePoll) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 p-6 text-center text-muted-foreground">
        <BarChart3 className="size-7 opacity-40" />
        <p className="text-sm">{t("liveRoom.polls.noActive")}</p>
      </div>
    )
  }

  // Results revealed
  if (results) {
    return (
      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-3">
        <p className="text-sm font-semibold text-foreground">{activePoll.name}</p>
        <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {t("liveRoom.polls.results")}
        </p>
        <PollsBar options={activePoll.options} counts={results.counts} total={results.total} />
      </div>
    )
  }

  // Waiting for results after answering
  if (hasAnswered) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 p-6 text-center text-muted-foreground">
        <BarChart3 className="size-7 opacity-40" />
        <p className="text-sm">{t("liveRoom.polls.submitted")}</p>
      </div>
    )
  }

  // Show voting options
  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-3">
      <p className="text-sm font-semibold text-foreground">{activePoll.name}</p>
      <div className="flex flex-col gap-2">
        {activePoll.options.map((opt) => (
          <button
            key={opt.value}
            type="button"
            disabled={answerMutation.isPending}
            onClick={() => handleVote(opt.value)}
            className="w-full rounded-md border border-border bg-muted px-4 py-2.5 text-start text-sm text-foreground transition-colors hover:border-primary hover:bg-primary/10 disabled:opacity-50"
          >
            {opt.label}
          </button>
        ))}
      </div>
    </div>
  )
}
