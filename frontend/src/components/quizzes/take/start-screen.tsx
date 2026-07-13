import type {
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
} from "@/api/model"
import type { ReactNode } from "react"

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
import { useFormatDate } from "@/lib/format-date"
import { formatScore } from "@/lib/score"
import { cn } from "@/lib/utils"

import { DecorativeBackground } from "./decorations"
import { MetaCell } from "./meta-cell"

// One rule in the pre-flight checklist: an icon chip + line of copy, in a shared
// row shell so every rule (and the destructive penalty note) lines up.
function RuleRow({
  icon,
  destructive = false,
  children,
}: {
  icon: ReactNode
  destructive?: boolean
  children: ReactNode
}) {
  return (
    <li
      className={cn(
        "flex items-center gap-3 px-4 py-3 text-sm leading-relaxed",
        destructive ? "text-destructive" : "text-foreground/80",
      )}
    >
      <span
        className={cn(
          "flex size-8 shrink-0 items-center justify-center rounded-lg ring-1 [&_svg]:size-4",
          destructive
            ? "bg-destructive/10 text-destructive ring-destructive/20"
            : "bg-foreground/5 text-muted-foreground ring-foreground/10",
        )}
      >
        {icon}
      </span>
      {children}
    </li>
  )
}

interface StartScreenProps {
  quiz: Quiz
  room: QuizRoom
  totalQuestions: number
  hasNegativeMarking: boolean
  backHref: string
  starting: boolean
  locating?: boolean
  onBegin: () => void
}

export function StartScreen({
  quiz,
  room,
  totalQuestions,
  hasNegativeMarking,
  backHref,
  starting,
  locating = false,
  onBegin,
}: StartScreenProps) {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const closesAt = formatDate(room.ended_at, "datetime")

  return (
    <div className="relative isolate flex flex-col gap-10 pb-24 pt-8">
      <DecorativeBackground />

      <div className="flex items-center">
        <Link
          to={backHref}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.session.quizzes.take.backToSession")}
        </Link>
      </div>

      <header className="flex flex-col gap-5">
        <h1 className="max-w-4xl text-4xl leading-tight font-semibold tracking-tight text-balance md:text-5xl lg:text-6xl">
          {quiz.title}
        </h1>
        {Boolean(quiz.description) && (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed md:text-lg">
            {quiz.description}
          </p>
        )}
      </header>

      <section className="ring-foreground/10 grid grid-cols-2 gap-px overflow-hidden rounded-3xl bg-gradient-to-br from-foreground/[0.08] to-foreground/[0.03] shadow-lg ring-1 md:grid-cols-4">
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

      <section className="flex flex-col gap-4">
        <Eyebrow>{t("org.session.quizzes.take.rules.eyebrow")}</Eyebrow>
        <ul className="ring-foreground/10 divide-foreground/5 flex max-w-2xl flex-col divide-y overflow-hidden rounded-2xl ring-1">
          <RuleRow icon={<ClockIcon />}>
            {t("org.session.quizzes.take.rules.timer", { minutes: quiz.duration_minutes ?? 0 })}
          </RuleRow>
          <RuleRow icon={quiz.no_back_navigation ? <LockKeyholeIcon /> : <ArrowLeftIcon />}>
            {quiz.no_back_navigation
              ? t("org.session.quizzes.take.rules.noBack")
              : t("org.session.quizzes.take.rules.backAllowed")}
          </RuleRow>
          {quiz.shuffle_questions && (
            <RuleRow icon={<ShuffleIcon />}>{t("org.session.quizzes.take.rules.shuffle")}</RuleRow>
          )}
          <RuleRow icon={<AlertTriangleIcon />}>
            {t("org.session.quizzes.take.rules.autoSubmit")}
          </RuleRow>
          {hasNegativeMarking && (
            <RuleRow icon={<AlertTriangleIcon />} destructive>
              {t("org.session.quizzes.take.penalty.explainer")}
            </RuleRow>
          )}
        </ul>
      </section>

      <div className="flex items-center gap-3">
        <Button size="lg" onClick={onBegin} disabled={starting || locating}>
          {starting || locating ? (
            <Spinner className="size-4" />
          ) : (
            <CheckCircle2Icon className="size-4" />
          )}
          {locating
            ? t("org.session.quizzes.take.requestingLocation")
            : t("org.session.quizzes.take.begin")}
        </Button>
        <Button variant="outline" render={<Link to={backHref} />}>
          {t("common.cancel")}
        </Button>
      </div>
    </div>
  )
}
