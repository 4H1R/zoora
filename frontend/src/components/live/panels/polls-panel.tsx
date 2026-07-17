import type { LivePoll, PollResults, RoomPolls } from "../use-room-polls"

import { BarChart3, Check, ChevronDown, History, Plus, Trash2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { useGetPolls, useGetPollsId, useGetPollsIdResults, usePostPolls } from "@/api/polls/polls"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

import { PollBars } from "./poll-bars"

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
    const resolvedOptions =
      mode === "yesno"
        ? yesNoOptions
        : options.map((o, i) => ({ label: o.trim() || `${t("liveRoom.polls.option")} ${i + 1}`, value: String(i) }))

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
          if (res.status !== 201) {
            toast.error(t("liveRoom.polls.launchError"))
            return
          }
          const id = res.data.data?.id
          if (!id) {
            toast.error(t("liveRoom.polls.launchError"))
            return
          }
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
        onError: () => toast.error(t("liveRoom.polls.launchError")),
      }
    )
  }

  const canLaunch = question.trim().length >= 2 && (mode === "yesno" || options.every((o) => o.trim().length > 0))

  return (
    <div className="flex flex-col gap-3 p-3">
      {/* Question */}
      <div className="flex flex-col gap-1.5">
        <label className="text-muted-foreground text-xs font-medium">{t("liveRoom.polls.question")}</label>
        <Input
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          placeholder={t("liveRoom.polls.question")}
          className="border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-ring/40 bg-transparent"
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
              : "border-border bg-muted text-muted-foreground hover:bg-accent"
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
              : "border-border bg-muted text-muted-foreground hover:bg-accent"
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
                className="border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-ring/40 bg-transparent"
              />
              {options.length > 2 && (
                <button
                  type="button"
                  onClick={() => removeOption(i)}
                  aria-label={t("liveRoom.polls.removeOption")}
                  className="text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded hover:text-red-400"
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
              className="text-muted-foreground hover:text-foreground flex items-center gap-1.5 text-xs"
            >
              <Plus className="size-3.5" />
              {t("liveRoom.polls.addOption")}
            </button>
          )}
        </div>
      )}

      <Button onClick={handleLaunch} disabled={!canLaunch || createPoll.isPending} className="w-full" size="sm">
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
  revealed: boolean
  onReveal: (results: PollResults) => void
  onClose: () => void
}

function HostActivePoll({ activePoll, revealed, onReveal, onClose }: HostActivePollProps) {
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
      <div className="flex items-center justify-between gap-2">
        <p className="text-foreground text-sm font-semibold">{activePoll.name}</p>
        {revealed && (
          <span className="flex shrink-0 items-center gap-1.5 rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
            <span className="size-1.5 rounded-full bg-emerald-500" />
            {t("liveRoom.polls.liveToStudents")}
          </span>
        )}
      </div>
      <PollBars options={activePoll.options} counts={counts} total={total} />
      <div className="flex gap-2">
        <Button size="sm" className="flex-1" onClick={handleReveal} disabled={revealed}>
          {revealed ? (
            <>
              <Check className="size-4" />
              {t("liveRoom.polls.resultsShared")}
            </>
          ) : (
            t("liveRoom.polls.reveal")
          )}
        </Button>
        <Button
          size="sm"
          variant="outline"
          className="border-border text-muted-foreground hover:bg-accent flex-1"
          onClick={onClose}
        >
          {t("liveRoom.polls.close")}
        </Button>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Host: Past-polls history (review results of closed polls)
// ---------------------------------------------------------------------------
interface HistoryRowProps {
  id: string
  name: string
  options: { label: string; value: string }[]
}

function HistoryPollRow({ id, name, options }: HistoryRowProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const resultsQuery = useGetPollsIdResults(id, { query: { enabled: open } })
  const data = resultsQuery.data?.status === 200 ? resultsQuery.data.data.data : null
  const counts: Record<string, number> = {}
  for (const opt of options) counts[opt.value] = data?.counts?.[opt.value] ?? 0
  const total = data?.total ?? 0

  return (
    <div className="border-border rounded-md border">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="text-foreground hover:bg-accent flex w-full items-center justify-between gap-2 px-3 py-2 text-start text-sm"
      >
        <span className="truncate">{name}</span>
        <ChevronDown
          className={cn("text-muted-foreground size-4 shrink-0 transition-transform", open && "rotate-180")}
        />
      </button>
      {open && (
        <div className="border-border border-t p-3">
          {resultsQuery.isLoading ? (
            <p className="text-muted-foreground text-xs">{t("common.loading")}</p>
          ) : (
            <PollBars options={options} counts={counts} total={total} />
          )}
        </div>
      )}
    </div>
  )
}

interface PollHistoryProps {
  liveId: string
  activePollId?: string
}

function PollHistory({ liveId, activePollId }: PollHistoryProps) {
  const { t } = useTranslation()

  const query = useGetPolls(
    { model_type: "live_session", model_id: liveId, order_by: "created_at", order_dir: "desc" },
    { query: { refetchInterval: 10000 } }
  )

  const polls = query.data?.status === 200 ? (query.data.data.data?.items ?? []) : []
  // Exclude the currently-running poll — it has its own live view above.
  const past = polls.filter((p) => p.id && p.id !== activePollId)

  if (past.length === 0) return null

  return (
    <div className="flex flex-col gap-2 px-3 pb-3">
      <p className="text-muted-foreground flex items-center gap-1.5 pt-2 text-xs font-semibold tracking-wider uppercase">
        <History className="size-3.5" />
        {t("liveRoom.polls.history")}
      </p>
      {past.map((p) => (
        <HistoryPollRow
          key={p.id}
          id={p.id!}
          name={p.name ?? "—"}
          options={(p.options ?? []).map((o) => ({ label: o.label ?? "", value: o.value ?? "" }))}
        />
      ))}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Viewer: Active poll vote
// ---------------------------------------------------------------------------
interface ViewerVoteProps {
  activePoll: LivePoll
  onVote: (value: string) => void
  answerPending: boolean
}

function ViewerVote({ activePoll, onVote, answerPending }: ViewerVoteProps) {
  const { t } = useTranslation()

  // The room can finish server-side (e.g. no-host auto-close) while this vote
  // view is still open, which closes the poll and makes any answer 409. Poll the
  // poll's lifecycle state so we gate the buttons before the request is rejected.
  const pollQuery = useGetPollsId(activePoll.pollId, { query: { refetchInterval: 5000 } })
  const isClosed = pollQuery.data?.status === 200 && pollQuery.data.data.data?.closed_at != null

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-3">
      <p className="text-foreground text-sm font-semibold">{activePoll.name}</p>
      {isClosed && <p className="text-muted-foreground text-xs">{t("liveRoom.polls.closed")}</p>}
      <div className="flex flex-col gap-2">
        {activePoll.options.map((opt) => (
          <button
            key={opt.value}
            type="button"
            disabled={answerPending || isClosed}
            onClick={() => onVote(opt.value)}
            className="border-border bg-muted text-foreground hover:border-primary hover:bg-primary/10 disabled:hover:border-border disabled:hover:bg-muted w-full rounded-md border px-4 py-2.5 text-start text-sm transition-colors disabled:opacity-50"
          >
            {opt.label}
          </button>
        ))}
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
  polls: RoomPolls
  onVote: (value: string) => void
  answerPending: boolean
}

export function PollsPanel({ liveId, isHost, polls, onVote, answerPending }: PollsPanelProps) {
  const { t } = useTranslation()
  const { activePoll, results, hasAnswered, launchPoll, revealResults, closePoll } = polls

  // ---- HOST ----
  if (isHost) {
    return (
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
        {activePoll ? (
          <HostActivePoll
            activePoll={activePoll}
            revealed={results !== null}
            onReveal={revealResults}
            onClose={closePoll}
          />
        ) : (
          <>
            <p className="text-muted-foreground px-3 pt-3 text-xs font-semibold tracking-wider uppercase">
              {t("liveRoom.polls.create")}
            </p>
            <CreatePollForm liveId={liveId} onLaunch={launchPoll} />
          </>
        )}
        <PollHistory liveId={liveId} activePollId={activePoll?.pollId} />
      </div>
    )
  }

  // ---- VIEWER ----

  // No active poll
  if (!activePoll) {
    return (
      <div className="text-muted-foreground flex min-h-0 flex-1 flex-col items-center justify-center gap-2 p-6 text-center">
        <BarChart3 className="size-7 opacity-40" />
        <p className="text-sm">{t("liveRoom.polls.noActive")}</p>
      </div>
    )
  }

  // Results revealed
  if (results) {
    return (
      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-3">
        <p className="text-foreground text-sm font-semibold">{activePoll.name}</p>
        <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          {t("liveRoom.polls.results")}
        </p>
        <PollBars options={activePoll.options} counts={results.counts} total={results.total} />
      </div>
    )
  }

  // Waiting for results after answering
  if (hasAnswered) {
    return (
      <div className="text-muted-foreground flex min-h-0 flex-1 flex-col items-center justify-center gap-2 p-6 text-center">
        <BarChart3 className="size-7 opacity-40" />
        <p className="text-sm">{t("liveRoom.polls.submitted")}</p>
      </div>
    )
  }

  // Active poll: vote (also surfaced as a modal popup)
  return <ViewerVote activePoll={activePoll} onVote={onVote} answerPending={answerPending} />
}
