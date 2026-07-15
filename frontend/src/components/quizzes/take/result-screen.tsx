import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { Link } from "@tanstack/react-router"
import {
  ArrowLeftIcon,
  CheckCircle2Icon,
  ClockIcon,
  FlagIcon,
  LockKeyholeIcon,
  ShieldCheckIcon,
  TrophyIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { OptionImageThumb } from "@/components/admin/questions/OptionImage"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { formatScore } from "@/lib/score"

import { DecorativeBackground } from "./decorations"
import { MetaCell } from "./meta-cell"
import { SystemImage } from "./system-image"
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
    <div className="relative isolate flex flex-col gap-10 pt-8 pb-24">
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
        <section className="animate-reveal relative isolate overflow-hidden rounded-3xl px-6 py-14 text-center md:py-16">
          {/* Layered surface so the brand glow reads as depth, not a flat fill. */}
          <div aria-hidden className="bg-card ring-foreground/10 absolute inset-0 -z-10 rounded-3xl ring-1" />
          <div
            aria-hidden
            className="animate-aurora pointer-events-none absolute start-1/2 -top-24 -z-10 size-72 -translate-x-1/2 rounded-full bg-[radial-gradient(circle,var(--color-primary)/22%,transparent_70%)] blur-2xl rtl:translate-x-1/2"
          />

          <div className="mx-auto flex max-w-md flex-col items-center gap-6">
            {/* Sealed-envelope motif: a settled check under an outward confirmation pulse. */}
            <span className="relative flex size-20 items-center justify-center">
              <span className="border-primary/30 absolute inset-0 animate-ping rounded-full border [animation-duration:2.4s]" />
              <span className="bg-primary/10 absolute inset-2 rounded-full" />
              <span className="bg-primary/15 text-primary ring-primary/25 relative flex size-14 items-center justify-center rounded-full ring-1 [&_svg]:size-7">
                <CheckCircle2Icon strokeWidth={2.25} />
              </span>
            </span>

            <div className="flex flex-col items-center gap-3">
              <span className="border-primary/25 bg-primary/5 text-primary inline-flex items-center gap-1.5 rounded-full border px-3 py-1 font-mono text-[0.65rem] tracking-[0.2em] uppercase rtl:font-sans rtl:tracking-normal">
                <ShieldCheckIcon className="size-3.5" />
                {t("org.session.quizzes.take.result.hidden.sealed")}
              </span>
              <h2 className="text-2xl font-semibold tracking-tight md:text-3xl">
                {t("org.session.quizzes.take.result.hidden.title")}
              </h2>
              <p className="text-muted-foreground max-w-sm text-sm leading-relaxed text-balance">
                {quiz.show_results
                  ? t("org.session.quizzes.take.result.hidden.deferred")
                  : t("org.session.quizzes.take.result.hidden.notPublished")}
              </p>
            </div>

            <div aria-hidden className="border-foreground/15 h-px w-full max-w-xs border-t border-dashed" />

            {/* Total time as the hero stat — the one number worth keeping while results wait. */}
            <div className="flex flex-col items-center gap-2">
              <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-[0.7rem] tracking-[0.25em] uppercase rtl:font-sans rtl:tracking-normal">
                <ClockIcon className="size-3.5" />
                {t("org.session.quizzes.take.result.totalTime")}
              </span>
              <span className="text-foreground font-mono text-4xl font-semibold tabular-nums md:text-5xl">
                {formatClock(totalSpent)}
              </span>
              <span className="text-muted-foreground/70 inline-flex items-center gap-1.5 text-xs">
                <LockKeyholeIcon className="size-3" />
                {t("org.session.quizzes.take.result.hidden.awaiting")}
              </span>
            </div>
          </div>
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
              const selectedOptions = (q?.options ?? []).filter((o) => selectedIds.includes(o.id ?? ""))
              const imageOptions = selectedOptions.filter((o) => o.image_media_id || o.system_image_media_id)
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
                    {q?.system_image_media_id ? (
                      <span className="ms-2 grow">
                        <SystemImage mediaID={q.system_image_media_id} className="max-h-8 w-auto" />
                      </span>
                    ) : (
                      <span className="text-foreground/80 ms-2 grow truncate font-mono text-xs">
                        {q?.text ?? a.question_id?.slice(0, 8)}
                      </span>
                    )}
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
                        <OptionImageThumb key={o.id} mediaID={(o.image_media_id ?? o.system_image_media_id)!} />
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
