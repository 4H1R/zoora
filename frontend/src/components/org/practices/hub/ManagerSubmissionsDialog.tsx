import type {
  GithubCom4H1RZooraInternalDomainPracticeRoomView as PracticeRoomView,
  GithubCom4H1RZooraInternalDomainPracticeSubmission as PracticeSubmission,
} from "@/api/model"

import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetPracticesIdSubmissions } from "@/api/practices/practices"
import { PracticeGradeDialog } from "@/components/org/practices/PracticeGradeDialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  practice: PracticeRoomView | null
}

export function ManagerSubmissionsDialog({ open, onOpenChange, practice }: Props) {
  const { t, i18n } = useTranslation()
  const [gradeTarget, setGradeTarget] = useState<PracticeSubmission | null>(null)

  const { data, isLoading } = useGetPracticesIdSubmissions(
    practice?.id ?? "",
    { order_by: "submitted_at", order_dir: "desc" },
    { query: { enabled: open && !!practice?.id } }
  )

  const subsData = (data?.status === 200 && data.data.data) || undefined
  const submissions = subsData?.items ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[calc(100%-2rem)] !max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("org.practices.manager.submissionsTitle")}</DialogTitle>
          <DialogDescription>{practice?.title}</DialogDescription>
        </DialogHeader>

        <div className="flex max-h-[60vh] flex-col divide-y overflow-y-auto pe-1">
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 py-3">
                <Skeleton className="size-8 rounded-full" />
                <Skeleton className="h-4 w-40 flex-1" />
                <Skeleton className="h-6 w-16" />
              </div>
            ))
          ) : submissions.length === 0 ? (
            <p className="text-muted-foreground py-10 text-center text-sm">
              {t("org.practices.manager.noSubmissions")}
            </p>
          ) : (
            submissions.map((sub) => {
              const graded = sub.score != null
              return (
                <div key={sub.id} className="flex items-center gap-3 py-3">
                  <div className="flex min-w-0 flex-1 flex-col">
                    <span className="truncate text-sm font-medium">{sub.user?.name ?? "—"}</span>
                    <span className="text-muted-foreground text-xs">
                      {formatSessionDate(sub.submitted_at, i18n.language, "short")}
                    </span>
                  </div>
                  {graded ? (
                    <span className="text-sm font-semibold tabular-nums">
                      {formatScore(sub.score)}
                      <span className="text-muted-foreground font-normal">
                        {" / "}
                        {formatScore(practice?.max_score ?? 0)}
                      </span>
                    </span>
                  ) : (
                    <Badge variant="secondary">{t("org.practices.status.submitted")}</Badge>
                  )}
                  <Button size="sm" variant="outline" onClick={() => setGradeTarget(sub)}>
                    {t("org.practices.actions.grade")}
                  </Button>
                </div>
              )
            })
          )}
        </div>
      </DialogContent>

      <PracticeGradeDialog
        open={!!gradeTarget}
        onOpenChange={(o) => !o && setGradeTarget(null)}
        submission={gradeTarget}
        practiceId={practice?.id}
        maxScore={practice?.max_score}
      />
    </Dialog>
  )
}
