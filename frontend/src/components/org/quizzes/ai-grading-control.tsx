import type { ErrorType } from "@/api/mutator/custom-instance"
import type { GithubCom4H1RZooraInternalDomainAIGradingJob as AIGradingJob } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { Loader2Icon, SparklesIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuizzesIdSubmissionsQueryKey,
  useGetQuizzesAiGradingJobId,
  usePostQuizzesIdAiGrading,
} from "@/api/quizzes/quizzes"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Switch } from "@/components/ui/switch"
import { useHasFeature, FEATURE } from "@/lib/entitlements"
import { cn } from "@/lib/utils"

type Mode = "apply" | "suggest"

interface AIGradingControlProps {
  quizId?: string
  disabled?: boolean
}

// AI auto-grading entry point for the corrections workbench. Gated on the org's
// FeatureAI plan entitlement — renders nothing when the plan lacks it, matching
// the nav gating. Starts a fan-out job, then polls its progress and refetches
// the submissions list when it finishes. Only descriptive answers are scored,
// and manual grades are never overwritten (enforced server-side).
export function AIGradingControl({ quizId, disabled }: AIGradingControlProps) {
  const { t } = useTranslation()
  const { enabled: hasAI } = useHasFeature(FEATURE.ai)
  const queryClient = useQueryClient()

  const [open, setOpen] = useState(false)
  const [mode, setMode] = useState<Mode>("apply")
  const [force, setForce] = useState(false)
  const [jobId, setJobId] = useState<string | null>(null)
  const handled = useRef<string | null>(null)

  const startMutation = usePostQuizzesIdAiGrading({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 202 && res.data.data?.id) {
          handled.current = null
          setJobId(res.data.data.id)
          setOpen(false)
          toast.success(t("org.session.corrections.ai.started"))
        }
      },
      onError: (err) => {
        const msg = (err as ErrorType<{ error?: string }>).response?.data?.error
        toast.error(msg || t("org.session.corrections.ai.failed"))
      },
    },
  })

  const jobQ = useGetQuizzesAiGradingJobId(jobId ?? "", {
    query: {
      enabled: !!jobId,
      refetchInterval: (q) => {
        const data = q.state.data
        const status = data?.status === 200 ? data.data.data?.status : undefined
        return status === "pending" || status === "running" ? 1500 : false
      },
    },
  })
  const job: AIGradingJob | undefined = jobQ.data?.status === 200 ? jobQ.data.data.data : undefined

  // React to terminal job states exactly once per job.
  useEffect(() => {
    if (!jobId || !job || handled.current === jobId) return
    if (job.status !== "completed" && job.status !== "failed") return
    handled.current = jobId
    if (quizId) {
      queryClient.invalidateQueries({ queryKey: getGetQuizzesIdSubmissionsQueryKey(quizId) })
    }
    if (job.status === "completed") {
      toast.success(t("org.session.corrections.ai.completed", { done: job.done ?? 0, total: job.total ?? 0 }))
    } else {
      toast.error(job.error || t("org.session.corrections.ai.failed"))
    }
    setJobId(null)
  }, [jobId, job, quizId, queryClient, t])

  if (!hasAI) return null

  const running = !!jobId
  const done = job?.done ?? 0
  const total = job?.total ?? 0

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            size="sm"
            variant="outline"
            disabled={disabled || !quizId || running || startMutation.isPending}
            className={cn(
              "border-violet-500/40 bg-violet-500/5 text-violet-700 hover:bg-violet-500/10 hover:text-violet-800",
              "dark:border-violet-400/30 dark:bg-violet-400/10 dark:text-violet-300 dark:hover:bg-violet-400/20 dark:hover:text-violet-200",
              "aria-expanded:bg-violet-500/10"
            )}
          >
            {running ? (
              <>
                <Loader2Icon data-icon="inline-start" className="animate-spin" />
                <span className="tabular-nums">
                  {t("org.session.corrections.ai.running")} {total > 0 ? `${done}/${total}` : ""}
                </span>
              </>
            ) : (
              <>
                <SparklesIcon data-icon="inline-start" />
                {t("org.session.corrections.ai.trigger")}
              </>
            )}
          </Button>
        }
      />
      <PopoverContent align="end" className="w-80 gap-3">
        <div className="flex items-center gap-2">
          <span className="flex size-7 items-center justify-center rounded-lg bg-violet-500/10 text-violet-600 dark:text-violet-400">
            <SparklesIcon className="size-4" />
          </span>
          <div className="min-w-0">
            <div className="text-sm font-semibold">{t("org.session.corrections.ai.title")}</div>
            <p className="text-muted-foreground text-xs leading-snug">{t("org.session.corrections.ai.description")}</p>
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <span className="text-muted-foreground text-[10px] font-medium tracking-wide uppercase">
            {t("org.session.corrections.ai.modeLabel")}
          </span>
          <div className="grid grid-cols-2 gap-1.5">
            <ModeCard
              active={mode === "apply"}
              title={t("org.session.corrections.ai.modeApply")}
              hint={t("org.session.corrections.ai.modeApplyHint")}
              onSelect={() => setMode("apply")}
            />
            <ModeCard
              active={mode === "suggest"}
              title={t("org.session.corrections.ai.modeSuggest")}
              hint={t("org.session.corrections.ai.modeSuggestHint")}
              onSelect={() => setMode("suggest")}
            />
          </div>
        </div>

        <label className="flex cursor-pointer items-start justify-between gap-3 rounded-lg border p-2.5">
          <span className="min-w-0">
            <span className="block text-xs font-medium">{t("org.session.corrections.ai.force")}</span>
            <span className="text-muted-foreground block text-[11px] leading-snug">
              {t("org.session.corrections.ai.forceHint")}
            </span>
          </span>
          <Switch checked={force} onCheckedChange={setForce} />
        </label>

        <Button
          size="sm"
          disabled={startMutation.isPending || !quizId}
          onClick={() => quizId && startMutation.mutate({ id: quizId, data: { mode, force } })}
          className="bg-violet-600 text-white hover:bg-violet-600/90 dark:bg-violet-500 dark:hover:bg-violet-500/90"
        >
          {startMutation.isPending ? (
            <Loader2Icon data-icon="inline-start" className="animate-spin" />
          ) : (
            <SparklesIcon data-icon="inline-start" />
          )}
          {t("org.session.corrections.ai.start")}
        </Button>
      </PopoverContent>
    </Popover>
  )
}

function ModeCard({
  active,
  title,
  hint,
  onSelect,
}: {
  active: boolean
  title: string
  hint: string
  onSelect: () => void
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={cn(
        "flex flex-col gap-0.5 rounded-lg border p-2 text-start transition-colors",
        active
          ? "border-violet-500/60 bg-violet-500/10 text-violet-800 dark:text-violet-200"
          : "border-border hover:bg-muted text-muted-foreground"
      )}
    >
      <span className="text-xs font-semibold">{title}</span>
      <span className="text-[10px] leading-snug opacity-80">{hint}</span>
    </button>
  )
}
