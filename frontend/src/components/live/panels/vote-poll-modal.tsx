import { BarChart3 } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetPollsId } from "@/api/polls/polls"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

import { resolveOptionLabel } from "../poll-labels"
import type { LivePoll, PollResults } from "../use-room-polls"
import { PollBars } from "./poll-bars"

interface VotePollModalProps {
  activePoll: LivePoll | null
  results: PollResults | null
  hasAnswered: boolean
  isPending: boolean
  onVote: (value: string) => void
}

// Center-screen popup shown to viewers the moment a poll is launched. Persists
// independent of the side panel / tabs so students can't miss it.
export function VotePollModal({ activePoll, results, hasAnswered, isPending, onVote }: VotePollModalProps) {
  const { t } = useTranslation()
  // Auto-close for the student once they've voted; reopen only if the host reveals results.
  const open = activePoll !== null && (results !== null || !hasAnswered)

  // While the vote prompt is up, poll the poll's lifecycle state: the room can
  // finish server-side (no-host auto-close) and close the poll, making answers
  // 409. Gate the buttons on that instead of letting the request be rejected.
  const showVote = open && results === null && !hasAnswered
  const pollQuery = useGetPollsId(activePoll?.pollId ?? "", {
    query: { enabled: showVote && Boolean(activePoll?.pollId), refetchInterval: 5000 },
  })
  const isClosed = pollQuery.data?.status === 200 && pollQuery.data.data.data?.closed_at != null

  return (
    <Dialog open={open}>
      <DialogContent showCloseButton={false} className="sm:max-w-md">
        {activePoll && (
          <>
            <DialogHeader>
              <DialogTitle>{activePoll.name}</DialogTitle>
            </DialogHeader>

            {results ? (
              // Host revealed final results
              <div className="flex flex-col gap-3">
                <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("liveRoom.polls.results")}
                </p>
                <PollBars options={activePoll.options} counts={results.counts} total={results.total} />
              </div>
            ) : hasAnswered ? (
              // Voted, waiting for the host to reveal
              <div className="flex flex-col items-center gap-2 py-6 text-center text-muted-foreground">
                <BarChart3 className="size-7 opacity-40" />
                <p className="text-sm">{t("liveRoom.polls.submitted")}</p>
              </div>
            ) : (
              // Cast a vote
              <div className="flex flex-col gap-2">
                {isClosed && (
                  <p className="text-xs text-muted-foreground">{t("liveRoom.polls.closed")}</p>
                )}
                {activePoll.options.map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    disabled={isPending || isClosed}
                    onClick={() => onVote(opt.value)}
                    className="w-full rounded-md border border-border bg-muted px-4 py-2.5 text-start text-sm text-foreground transition-colors hover:border-primary hover:bg-primary/10 disabled:opacity-50 disabled:hover:border-border disabled:hover:bg-muted"
                  >
                    {resolveOptionLabel(opt.value, opt.label, t)}
                  </button>
                ))}
              </div>
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
