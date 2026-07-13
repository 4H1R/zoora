import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon, CheckCircle2Icon, ClockIcon, FlagIcon, TrophyIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { OptionImageThumb } from "@/components/admin/questions/OptionImage"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { formatScore } from "@/lib/score"

import { DecorativeBackground } from "./decorations"
import { MetaCell } from "./meta-cell"
import { formatClock } from "./utils"

interface ResultScreenProps {
  quiz: Quiz
  submission: QuizSubmission
  questions?: Question[]
  backHref: string
}

export function ResultScreen({ quiz, submission, questions = [], backHref }: ResultScreenProps) {
  const { t } = useTranslation()
  const total = quiz.total_score ?? 0
  const earned = submission.total_score ?? 0
  const pct = total > 0 ? Math.round((earned / total) * 100) : 0
  const answers = submission.answers ?? []
  const totalSpent = answers.reduce((acc, a) => acc + (a.spent_seconds ?? 0), 0)
  const questionById = new Map(questions.map((q) => [q.id, q]))
  // The backend strips scores until the room window closes when the quiz opts
  // into deferred results, so early finishers can't leak them. Undefined means
  // no gating (manager view / legacy) — treat as revealed.
  const revealed = submission.results_revealed ?? true

  return (
    <div className="relative isolate flex flex-col gap-10 pb-24 pt-8">
      <DecorativeBackground />

      <div className="flex items-center justify-between">
        <Link
          to={backHref}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.session.quizzes.take.backToSession")}
        </Link>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">
          {submission.status === "graded"
            ? t("org.session.quizzes.take.result.graded")
            : t("org.session.quizzes.take.result.submitted")}
        </span>
      </div>

      <header className="flex flex-col gap-4">
        <Eyebrow>{t("org.session.quizzes.take.result.eyebrow")}</Eyebrow>
        <h1 className="text-4xl leading-tight font-semibold tracking-tight md:text-5xl">{quiz.title}</h1>
      </header>

      {!revealed && (
        <section className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-3xl px-6 py-10 text-center ring-1">
          <span className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-full [&_svg]:size-6">
            <TrophyIcon />
          </span>
          <h2 className="text-xl font-semibold tracking-tight">
            {t("org.session.quizzes.take.result.hidden.title")}
          </h2>
          <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
            {quiz.show_results
              ? t("org.session.quizzes.take.result.hidden.deferred")
              : t("org.session.quizzes.take.result.hidden.notPublished")}
          </p>
          <MetaCell
            icon={<ClockIcon className="size-4" />}
            label={t("org.session.quizzes.take.result.totalTime")}
            value={formatClock(totalSpent)}
            mono
          />
        </section>
      )}

      {revealed && (
        <section className="bg-card ring-foreground/10 grid grid-cols-2 gap-0 overflow-hidden rounded-3xl px-4 py-6 ring-1 md:grid-cols-4 md:px-6">
          <MetaCell
            icon={<TrophyIcon className="size-4" />}
            label={t("org.session.quizzes.take.result.earned")}
            value={formatScore(earned)}
            mono
          />
          <MetaCell
            icon={<FlagIcon className="size-4" />}
            label={t("org.session.quizzes.take.result.outOf")}
            value={formatScore(total)}
            mono
          />
          <MetaCell
            icon={<CheckCircle2Icon className="size-4" />}
            label={t("org.session.quizzes.take.result.percent")}
            value={`${pct}%`}
            mono
          />
          <MetaCell
            icon={<ClockIcon className="size-4" />}
            label={t("org.session.quizzes.take.result.totalTime")}
            value={formatClock(totalSpent)}
            mono
          />
        </section>
      )}

      {revealed && (
        <section className="flex flex-col gap-3">
          <Eyebrow>{t("org.session.quizzes.take.result.perQuestionEyebrow")}</Eyebrow>
          <div className="flex flex-col divide-y divide-dashed">
            {answers.map((a, i) => {
            const q = questionById.get(a.question_id)
            const selectedIds = a.selected_option_ids ?? []
            const selectedOptions = (q?.options ?? []).filter((o) =>
              selectedIds.includes(o.id ?? ""),
            )
            const imageOptions = selectedOptions.filter((o) => o.image_media_id)
            const earnedScore = a.earned_score ?? 0
            // We can only confirm a penalty was applied when the earned score is
            // negative (scores are stripped on the take endpoint). Degrade to the
            // earned score otherwise.
            const penalized = earnedScore < 0
            return (
              <div key={a.question_id} className="flex flex-col gap-2 py-3">
                <div className="flex items-center justify-between gap-4">
                  <span className="text-muted-foreground font-mono text-xs tracking-[0.2em]">
                    Q{String(i + 1).padStart(2, "0")}
                  </span>
                  <span className="text-foreground/80 ms-2 grow truncate font-mono text-xs">
                    {q?.text ?? a.question_id?.slice(0, 8)}
                  </span>
                  <span className="text-muted-foreground font-mono text-xs tabular-nums">
                    {formatClock(a.spent_seconds ?? 0)}
                  </span>
                  <span className="text-foreground font-mono text-sm font-semibold tabular-nums">
                    {formatScore(earnedScore)}
                  </span>
                </div>
                {imageOptions.length > 0 && (
                  <div className="ms-8 flex flex-wrap gap-2">
                    {imageOptions.map((o) => (
                      <OptionImageThumb key={o.id} mediaID={o.image_media_id!} />
                    ))}
                  </div>
                )}
                {penalized && (
                  <span className="text-destructive ms-8 text-xs">
                    {t("org.session.quizzes.take.penalty.breakdown")}
                  </span>
                )}
              </div>
            )
            })}
          </div>
        </section>
      )}

      <Button variant="outline" render={<Link to={backHref} />}>
        <ArrowLeftIcon className="size-4" />
        {t("org.session.quizzes.take.backToSession")}
      </Button>
    </div>
  )
}
