import type {
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon, CheckCircle2Icon, ClockIcon, FlagIcon, TrophyIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { formatScore } from "@/lib/score"

import { DecorativeBackground } from "./decorations"
import { MetaCell } from "./meta-cell"
import { formatClock } from "./utils"

interface ResultScreenProps {
  quiz: Quiz
  submission: QuizSubmission
  backHref: string
}

export function ResultScreen({ quiz, submission, backHref }: ResultScreenProps) {
  const { t } = useTranslation()
  const total = quiz.total_score ?? 0
  const earned = submission.total_score ?? 0
  const pct = total > 0 ? Math.round((earned / total) * 100) : 0
  const answers = submission.answers ?? []
  const totalSpent = answers.reduce((acc, a) => acc + (a.spent_seconds ?? 0), 0)

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

      <section className="flex flex-col gap-3">
        <Eyebrow>{t("org.session.quizzes.take.result.perQuestionEyebrow")}</Eyebrow>
        <div className="flex flex-col divide-y divide-dashed">
          {answers.map((a, i) => (
            <div key={a.question_id} className="flex items-center justify-between gap-4 py-3">
              <span className="text-muted-foreground font-mono text-xs tracking-[0.2em]">
                Q{String(i + 1).padStart(2, "0")}
              </span>
              <span className="text-foreground/80 ms-2 grow truncate font-mono text-xs">
                {a.question_id?.slice(0, 8)}
              </span>
              <span className="text-muted-foreground font-mono text-xs tabular-nums">
                {formatClock(a.spent_seconds ?? 0)}
              </span>
              <span className="text-foreground font-mono text-sm font-semibold tabular-nums">
                {formatScore(a.earned_score ?? 0)}
              </span>
            </div>
          ))}
        </div>
      </section>

      <Button variant="outline" render={<Link to={backHref} />}>
        <ArrowLeftIcon className="size-4" />
        {t("org.session.quizzes.take.backToSession")}
      </Button>
    </div>
  )
}
