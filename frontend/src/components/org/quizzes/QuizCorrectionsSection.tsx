import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { ClipboardListIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

import { QuizCorrectionsPanel } from "./QuizCorrectionsPanel"
import { useQuizPermissions } from "./use-quiz-permissions"

interface QuizCorrectionsSectionProps {
  classSessionId: string
}

export function QuizCorrectionsSection({ classSessionId }: QuizCorrectionsSectionProps) {
  const { t } = useTranslation()
  const { canView, canEdit } = useQuizPermissions()

  const [quizId, setQuizId] = useState<string | undefined>(undefined)

  const quizzesQ = useGetQuizzes({ class_session_id: classSessionId }, { query: { enabled: canView } })
  const quizzes: Quiz[] = (quizzesQ.data?.status === 200 && quizzesQ.data.data.data?.items) || []

  const effectiveQuizId = quizId ?? quizzes[0]?.id
  const selectedQuiz = quizzes.find((q) => q.id === effectiveQuizId)

  if (!canView || !canEdit) return null

  const isLoadingQuizzes = quizzesQ.isPending
  const noQuizzes = !isLoadingQuizzes && quizzes.length === 0

  return (
    <section id="corrections" className="relative isolate flex scroll-mt-20 flex-col gap-6 overflow-hidden rounded-3xl">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_bottom_left,var(--color-primary)/8%,transparent_55%)]"
      />

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.corrections.title")}</h2>
        </div>
        <span className="text-muted-foreground hidden font-mono text-[11px] tracking-[0.25em] uppercase md:inline">
          {selectedQuiz ? `// ${selectedQuiz.title}` : `// ${t("org.session.corrections.noQuizSelected")}`}
        </span>
      </div>

      {noQuizzes ? (
        <EmptyState
          icon={ClipboardListIcon}
          title={t("org.session.corrections.emptyTitle")}
          description={t("org.session.corrections.emptyHint")}
        />
      ) : (
        <>
          <QuizSelector
            quizzes={quizzes}
            selectedId={effectiveQuizId}
            onSelect={setQuizId}
            isLoading={isLoadingQuizzes}
          />

          <QuizCorrectionsPanel quiz={selectedQuiz} />
        </>
      )}
    </section>
  )
}

function QuizSelector({
  quizzes,
  selectedId,
  onSelect,
  isLoading,
}: {
  quizzes: Quiz[]
  selectedId?: string
  onSelect: (id: string | undefined) => void
  isLoading: boolean
}) {
  const { t } = useTranslation()
  if (isLoading) {
    return (
      <div className="flex flex-wrap gap-2">
        <Skeleton className="h-9 w-32 rounded-full" />
        <Skeleton className="h-9 w-40 rounded-full" />
        <Skeleton className="h-9 w-28 rounded-full" />
      </div>
    )
  }
  return (
    <div className="flex flex-col gap-2">
      <Eyebrow className="text-[10px]">{t("org.session.corrections.pickQuiz")}</Eyebrow>
      <div className="flex flex-wrap gap-2">
        {quizzes.map((q, i) => {
          const tile = String(i + 1).padStart(2, "0")
          const active = q.id === selectedId
          return (
            <button
              key={q.id}
              type="button"
              onClick={() => onSelect(q.id)}
              className={cn(
                "group/pill border-border dark:ring-foreground/10 inline-flex items-center gap-2.5 rounded-full border px-4 py-2 text-sm font-medium shadow-sm transition-all dark:border-0 dark:shadow-none dark:ring-1",
                active
                  ? "bg-foreground text-background border-foreground dark:ring-foreground"
                  : "bg-card text-foreground hover:border-foreground/30 dark:hover:ring-foreground/30 hover:-translate-y-0.5"
              )}
            >
              <span
                className={cn(
                  "font-mono text-[10px] tracking-[0.25em]",
                  active ? "text-background/70" : "text-muted-foreground"
                )}
              >
                /{tile}
              </span>
              <span className="line-clamp-1 max-w-[16rem]">{q.title ?? "—"}</span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
