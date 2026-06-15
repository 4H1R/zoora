import type { GithubCom4H1RZooraInternalDomainPracticeSubmission as PracticeSubmission } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetPracticesIdSubmissionsQueryKey,
  getGetPracticesSubmissionsSubmissionIdQueryKey,
  usePutPracticesSubmissionsSubmissionIdGrade,
} from "@/api/practices/practices"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { formatScore } from "@/lib/score"

interface PracticeGradeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  submission: PracticeSubmission | null
  practiceId?: string
  maxScore?: number
}

export function PracticeGradeDialog({
  open,
  onOpenChange,
  submission,
  practiceId,
  maxScore,
}: PracticeGradeDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [score, setScore] = useState<number>(0)
  const [comment, setComment] = useState("")

  useEffect(() => {
    if (open) {
      setScore(submission?.score ?? 0)
      setComment(submission?.teacher_comment ?? "")
    }
  }, [open, submission])

  const gradeMutation = usePutPracticesSubmissionsSubmissionIdGrade({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.practiceScores.dialog.saveSuccess"))
        if (submission?.id) {
          queryClient.invalidateQueries({
            queryKey: getGetPracticesSubmissionsSubmissionIdQueryKey(submission.id),
          })
        }
        if (practiceId) {
          queryClient.invalidateQueries({
            queryKey: getGetPracticesIdSubmissionsQueryKey(practiceId),
          })
        }
        onOpenChange(false)
      },
      onError: () => {
        toast.error(t("org.session.practiceScores.dialog.saveFailed"))
      },
    },
  })

  const handleSave = () => {
    if (!submission?.id) return
    gradeMutation.mutate({
      submissionId: submission.id,
      data: { score, teacher_comment: comment },
    })
  }

  const overMax = maxScore != null && maxScore > 0 && score > maxScore

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[calc(100%-2rem)] !max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex flex-wrap items-center gap-2">
            <span>{t("org.session.practiceScores.dialog.title")}</span>
            {submission?.user?.name && (
              <Badge variant="secondary" className="text-xs font-normal">
                {submission.user.name}
              </Badge>
            )}
          </DialogTitle>
          <DialogDescription>{t("org.session.practiceScores.dialog.description")}</DialogDescription>
        </DialogHeader>

        <div className="flex max-h-[60vh] flex-col gap-4 overflow-y-auto pe-1">
          <div>
            <div className="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
              {t("org.session.practiceScores.dialog.answer")}
            </div>
            <div className="bg-muted/30 min-h-[6rem] rounded-md border px-3 py-2 text-sm whitespace-pre-wrap break-words">
              {submission?.content?.trim()
                ? submission.content
                : <span className="text-muted-foreground italic">{t("org.session.practiceScores.dialog.noAnswer")}</span>}
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-muted-foreground text-xs font-medium uppercase tracking-wide">
              {t("org.session.practiceScores.dialog.score")}
            </label>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                step="0.5"
                min={0}
                max={maxScore && maxScore > 0 ? maxScore : undefined}
                value={score}
                onChange={(e) => setScore(Number(e.target.value))}
                className="h-9 w-32 text-end tabular-nums"
                data-invalid={overMax || undefined}
              />
              {maxScore != null && maxScore > 0 && (
                <span className="text-muted-foreground text-sm tabular-nums">
                  / {formatScore(maxScore)}
                </span>
              )}
            </div>
            {overMax && (
              <span className="text-destructive text-xs">
                {t("org.session.practiceScores.dialog.overMax", { max: formatScore(maxScore) })}
              </span>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-muted-foreground text-xs font-medium uppercase tracking-wide">
              {t("org.session.practiceScores.dialog.comment")}
            </label>
            <Textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder={t("org.session.practiceScores.dialog.commentPlaceholder")}
              rows={4}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={gradeMutation.isPending}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSave} disabled={gradeMutation.isPending || overMax || !submission?.id}>
            {t("org.session.practiceScores.dialog.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
