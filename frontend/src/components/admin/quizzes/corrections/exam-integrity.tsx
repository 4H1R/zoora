import type {
  GithubCom4H1RZooraInternalDomainSubmissionAntiCheatReport as AntiCheatReport,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"
import type { LucideIcon } from "lucide-react"

import { EyeOff, MapPin, MapPinOff, ShieldAlert, ShieldCheck, Zap } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"

// Mirror of domain.TabHiddenWarnCount — flag when a student left the tab more
// than this. Only a fallback for when the per-quiz report (which carries the
// server-computed tab_flagged) hasn't loaded.
const TAB_WARN_COUNT = 3

// A caution (amber) is a mild signal; a flag (danger/red) is strong enough to
// surface prominently. `level` rolls the three signals into one verdict.
type Tone = "warn" | "danger"
type Level = "clean" | "warn" | "danger"

// Shared tint classes so chips (table) and tiles (panel) never drift apart.
const TONE_CLASS: Record<Tone, string> = {
  danger: "border-destructive/40 bg-destructive/10 text-destructive",
  warn: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
}

const LEVEL_STYLE: Record<Level, { band: string; icon: LucideIcon; iconClass: string }> = {
  clean: { band: "bg-muted/20", icon: ShieldCheck, iconClass: "text-emerald-600 dark:text-emerald-400" },
  warn: {
    band: "border-amber-500/20 bg-amber-500/[0.03]",
    icon: ShieldAlert,
    iconClass: "text-amber-600 dark:text-amber-400",
  },
  danger: {
    band: "border-destructive/20 bg-destructive/[0.03]",
    icon: ShieldAlert,
    iconClass: "text-destructive",
  },
}

interface Integrity {
  tab: { count: number; seconds: number; flagged: boolean; active: boolean; tone: Tone }
  loc: { nearby: number; denied: boolean; active: boolean; tone: Tone }
  fast: { count: number; active: boolean; tone: Tone }
  anySignal: boolean
  level: Level
}

// Single source of truth for a submission's advisory anti-cheat signals:
// tab-focus loss, location (denied or same-place cluster), and too-fast answers.
// Tab and location live directly on the submission, so those two survive even
// without the per-quiz report; same-location and fast-answers need the report.
// None asserts guilt — the teacher reviews and decides.
export function resolveIntegrity(r?: AntiCheatReport, s?: QuizSubmission): Integrity {
  const tabCount = r?.tab_hidden_count ?? s?.tab_hidden_count ?? 0
  const tabFlagged = r?.tab_flagged ?? tabCount > TAB_WARN_COUNT
  const tab = {
    count: tabCount,
    seconds: r?.tab_hidden_seconds ?? s?.tab_hidden_seconds ?? 0,
    flagged: tabFlagged,
    active: tabCount > 0,
    tone: (tabFlagged ? "danger" : "warn") as Tone,
  }

  const nearby = r?.same_location_user_ids?.length ?? 0
  const loc = {
    nearby,
    denied: r?.gps_denied ?? s?.gps_denied ?? false,
    active: nearby > 0 || (r?.gps_denied ?? s?.gps_denied ?? false),
    tone: (nearby > 0 ? "danger" : "warn") as Tone,
  }

  const fastCount = r?.fast_answers?.length ?? 0
  const fast = { count: fastCount, active: fastCount > 0, tone: "warn" as Tone }

  const flagCount = (tab.flagged ? 1 : 0) + (loc.nearby > 0 ? 1 : 0) + (fast.active ? 1 : 0)
  const anySignal = tab.active || loc.active || fast.active
  const level: Level = flagCount > 0 ? "danger" : anySignal ? "warn" : "clean"

  return { tab, loc, fast, anySignal, level }
}

// question_ids answered faster than their min_seconds — used to badge the
// matching rows in the grading dialog.
export function fastQuestionIds(r?: AntiCheatReport): Set<string> {
  const out = new Set<string>()
  for (const f of r?.fast_answers ?? []) if (f.question_id) out.add(f.question_id)
  return out
}

// Compact per-row indicator for the corrections table. Icon chips for each
// active signal, a tooltip spelling them out, and a faint green shield when the
// submission is clean.
export function IntegrityCell({ report, submission }: { report?: AntiCheatReport; submission?: QuizSubmission }) {
  const { t } = useTranslation()

  if (!report && !submission) return <span className="text-muted-foreground text-xs">—</span>

  const { tab, loc, fast, anySignal } = resolveIntegrity(report, submission)

  if (!anySignal) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={
            <span className="text-muted-foreground/70 inline-flex items-center gap-1 text-xs">
              <ShieldCheck className="size-3.5 text-emerald-500/70" />
              {t("admin.corrections.integrity.clean")}
            </span>
          }
        />
        <TooltipContent>{t("admin.corrections.integrity.advisory")}</TooltipContent>
      </Tooltip>
    )
  }

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <span className="inline-flex items-center gap-1.5">
            {tab.active && (
              <Chip tone={tab.tone}>
                <EyeOff className="size-3.5" />
                <span className="tabular-nums">{tab.count}</span>
              </Chip>
            )}
            {loc.active && (
              <Chip tone={loc.tone}>
                {loc.nearby > 0 ? <MapPin className="size-3.5" /> : <MapPinOff className="size-3.5" />}
                {loc.nearby > 0 && <span className="tabular-nums">{loc.nearby}</span>}
              </Chip>
            )}
            {fast.active && (
              <Chip tone={fast.tone}>
                <Zap className="size-3.5" />
                <span className="tabular-nums">{fast.count}</span>
              </Chip>
            )}
          </span>
        }
      />
      <TooltipContent className="flex max-w-56 flex-col items-start gap-1 py-2 text-start">
        <span className="font-medium">{t("admin.corrections.integrity.title")}</span>
        {tab.active && (
          <span>{t("admin.corrections.integrity.tabDetail", { count: tab.count, seconds: tab.seconds })}</span>
        )}
        {loc.nearby > 0 && <span>{t("admin.corrections.integrity.sameLocation", { count: loc.nearby })}</span>}
        {loc.denied && <span>{t("admin.corrections.integrity.locationDenied")}</span>}
        {fast.active && <span>{t("admin.corrections.integrity.fastDetail", { count: fast.count })}</span>}
        <span className="text-background/70 mt-0.5">{t("admin.corrections.integrity.advisory")}</span>
      </TooltipContent>
    </Tooltip>
  )
}

// Full advisory panel for the grading dialog. Verdict header plus one tile per
// active signal; a reassuring green line when nothing is flagged. Renders from
// the submission alone when the per-quiz report isn't available.
export function ExamIntegrityPanel({ report, submission }: { report?: AntiCheatReport; submission?: QuizSubmission }) {
  const { t } = useTranslation()
  if (!report && !submission) return null

  const { tab, loc, fast, anySignal, level } = resolveIntegrity(report, submission)
  const style = LEVEL_STYLE[level]
  const ShieldIcon = style.icon

  return (
    <section className={cn("border-b px-6 py-3", style.band)}>
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        <div className="flex items-center gap-2">
          <ShieldIcon className={cn("size-4", style.iconClass)} />
          <span className="text-sm font-medium">{t("admin.corrections.integrity.title")}</span>
        </div>

        {!anySignal && <span className="text-muted-foreground text-sm">{t("admin.corrections.integrity.clean")}</span>}

        <div className="flex flex-wrap items-center gap-2">
          {tab.active && (
            <SignalTile
              tone={tab.tone}
              icon={<EyeOff className="size-3.5" />}
              label={t("admin.corrections.integrity.tab")}
              value={t("admin.corrections.integrity.tabValue", { count: tab.count, seconds: tab.seconds })}
            />
          )}
          {loc.active && (
            <SignalTile
              tone={loc.tone}
              icon={loc.nearby > 0 ? <MapPin className="size-3.5" /> : <MapPinOff className="size-3.5" />}
              label={t("admin.corrections.integrity.location")}
              value={
                loc.nearby > 0
                  ? t("admin.corrections.integrity.sameLocation", { count: loc.nearby })
                  : t("admin.corrections.integrity.locationDenied")
              }
            />
          )}
          {fast.active && (
            <SignalTile
              tone={fast.tone}
              icon={<Zap className="size-3.5" />}
              label={t("admin.corrections.integrity.fast")}
              value={t("admin.corrections.integrity.fastValue", { count: fast.count })}
            />
          )}
        </div>

        <span className="text-muted-foreground ms-auto text-xs">{t("admin.corrections.integrity.advisory")}</span>
      </div>
    </section>
  )
}

function SignalTile({ tone, icon, label, value }: { tone: Tone; icon: React.ReactNode; label: string; value: string }) {
  return (
    <span className={cn("inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-xs", TONE_CLASS[tone])}>
      {icon}
      <span className="text-muted-foreground/90 font-medium tracking-wide uppercase">{label}</span>
      <span className="tabular-nums">{value}</span>
    </span>
  )
}

function Chip({ tone, children }: { tone: Tone; children: React.ReactNode }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 rounded border px-1 py-0.5 text-xs font-medium",
        TONE_CLASS[tone]
      )}
    >
      {children}
    </span>
  )
}
