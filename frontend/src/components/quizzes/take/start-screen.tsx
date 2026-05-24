import type {
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
} from "@/api/model"

import { Link } from "@tanstack/react-router"
import {
  AlertTriangleIcon,
  ArrowLeftIcon,
  CheckCircle2Icon,
  ClipboardListIcon,
  ClockIcon,
  FlagIcon,
  LockKeyholeIcon,
  ShuffleIcon,
  TrophyIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { formatScore } from "@/lib/score"

import { DecorativeBackground } from "./decorations"
import { MetaCell } from "./meta-cell"

interface StartScreenProps {
  quiz: Quiz
  room: QuizRoom
  totalQuestions: number
  backHref: string
  starting: boolean
  onBegin: () => void
}

export function StartScreen({ quiz, room, totalQuestions, backHref, starting, onBegin }: StartScreenProps) {
  const { t } = useTranslation()
  const shortId = (quiz.id ?? "").slice(0, 8).toUpperCase()
  const closesAt = room.ended_at ? new Date(room.ended_at).toLocaleString() : "—"

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
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">№ {shortId}</span>
      </div>

      <header className="flex flex-col gap-5">
        <Eyebrow>{t("org.session.quizzes.take.eyebrow")}</Eyebrow>
        <h1 className="max-w-4xl text-4xl leading-tight font-semibold tracking-tight text-balance md:text-5xl lg:text-6xl">
          {quiz.title}
        </h1>
        {quiz.description ? (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed md:text-lg">
            {quiz.description}
          </p>
        ) : null}
      </header>

      <section className="bg-card ring-foreground/10 grid grid-cols-2 gap-0 overflow-hidden rounded-2xl px-4 py-6 ring-1 md:grid-cols-4 md:px-6">
        <MetaCell
          icon={<ClipboardListIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.questions")}
          value={totalQuestions.toString()}
          mono
        />
        <MetaCell
          icon={<ClockIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.duration")}
          value={`${quiz.duration_minutes ?? 0} ${t("org.session.quizzes.minutesShort")}`}
          mono
        />
        <MetaCell
          icon={<TrophyIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.totalScore")}
          value={formatScore(quiz.total_score ?? 0)}
          mono
        />
        <MetaCell
          icon={<FlagIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.closesAt")}
          value={closesAt}
        />
      </section>

      <section className="flex flex-col gap-3">
        <Eyebrow>{t("org.session.quizzes.take.rules.eyebrow")}</Eyebrow>
        <ul className="text-foreground/80 max-w-2xl space-y-2 text-sm leading-relaxed">
          <li className="flex items-start gap-2">
            <ClockIcon className="text-muted-foreground mt-0.5 size-4" />
            {t("org.session.quizzes.take.rules.timer", { minutes: quiz.duration_minutes ?? 0 })}
          </li>
          <li className="flex items-start gap-2">
            {quiz.no_back_navigation ? (
              <LockKeyholeIcon className="text-muted-foreground mt-0.5 size-4" />
            ) : (
              <ArrowLeftIcon className="text-muted-foreground mt-0.5 size-4" />
            )}
            {quiz.no_back_navigation
              ? t("org.session.quizzes.take.rules.noBack")
              : t("org.session.quizzes.take.rules.backAllowed")}
          </li>
          {quiz.shuffle_questions ? (
            <li className="flex items-start gap-2">
              <ShuffleIcon className="text-muted-foreground mt-0.5 size-4" />
              {t("org.session.quizzes.take.rules.shuffle")}
            </li>
          ) : null}
          <li className="flex items-start gap-2">
            <AlertTriangleIcon className="text-muted-foreground mt-0.5 size-4" />
            {t("org.session.quizzes.take.rules.autoSubmit")}
          </li>
        </ul>
      </section>

      <div className="flex items-center gap-3">
        <Button size="lg" onClick={onBegin} disabled={starting}>
          {starting ? <Spinner className="size-4" /> : <CheckCircle2Icon className="size-4" />}
          {t("org.session.quizzes.take.begin")}
        </Button>
        <Button variant="outline" render={<Link to={backHref} />}>
          {t("common.cancel")}
        </Button>
      </div>
    </div>
  )
}
