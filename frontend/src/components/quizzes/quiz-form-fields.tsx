import type { ReactNode } from "react"
import type {
  Control,
  FieldError as RHFFieldError,
  FieldErrors,
  UseFormRegister,
} from "react-hook-form"

import {
  ClipboardX,
  Dices,
  Eye,
  Lock,
  MapPin,
  MousePointerClick,
  Shuffle,
  ShieldCheck,
  Trophy,
} from "lucide-react"
import { Controller } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { BooleanFieldRow } from "@/components/form/boolean-field-row"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

export interface QuizCoreValues {
  title: string
  description?: string
  // `unknown` because callers feed in the zod *input* type, where
  // `z.coerce.number()` fields are typed `unknown` before coercion.
  duration_minutes: unknown
  no_back_navigation: boolean
  shuffle_questions: boolean
}

export interface QuizScheduleValues {
  started_at: string
  ended_at: string
}

interface QuizCoreFieldsProps {
  // Loose register so callers with stricter generics can pass it without casts.
  register: UseFormRegister<any>
  errors: FieldErrors<QuizCoreValues>
  prefix: string
}

export function QuizCoreFields({ register, errors, prefix }: QuizCoreFieldsProps) {
  const { t } = useTranslation()
  return (
    <>
      <Field data-invalid={!!errors.title || undefined}>
        <FieldLabel>{t(`${prefix}.title`)}</FieldLabel>
        <Input {...register("title")} placeholder={t(`${prefix}.titlePlaceholder`)} />
        <FieldError errors={[errors.title as RHFFieldError | undefined]} />
      </Field>
      <Field>
        <FieldLabel>{t(`${prefix}.description`)}</FieldLabel>
        <Textarea
          {...register("description")}
          placeholder={t(`${prefix}.descriptionPlaceholder`)}
          rows={3}
        />
      </Field>
      <Field data-invalid={!!errors.duration_minutes || undefined}>
        <FieldLabel>{t(`${prefix}.duration`)}</FieldLabel>
        <Input type="number" min={1} {...register("duration_minutes")} />
        <FieldError errors={[errors.duration_minutes as RHFFieldError | undefined]} />
      </Field>
    </>
  )
}

interface QuizScheduleFieldsProps {
  // Loose control so callers with stricter generics can pass it without casts.
  control: Control<any, any, any>
  errors: FieldErrors<QuizScheduleValues>
  prefix: string
}

export function QuizScheduleFields({ control, errors, prefix }: QuizScheduleFieldsProps) {
  const { t } = useTranslation()
  const endedAtError = errors.ended_at as RHFFieldError | undefined
  return (
    <div className="grid gap-3 sm:grid-cols-2">
      <Field data-invalid={!!errors.started_at || undefined}>
        <FieldLabel>{t(`${prefix}.startedAt`)}</FieldLabel>
        <Controller
          control={control}
          name="started_at"
          render={({ field, fieldState }) => (
            <DateTimePicker
              value={field.value || undefined}
              onChange={(v) => field.onChange(v ?? "")}
              invalid={fieldState.invalid}
            />
          )}
        />
        <p className="text-muted-foreground text-xs">{t(`${prefix}.startedAtHint`)}</p>
        <FieldError errors={[errors.started_at as RHFFieldError | undefined]} />
      </Field>
      <Field data-invalid={!!endedAtError || undefined}>
        <FieldLabel>{t(`${prefix}.endedAt`)}</FieldLabel>
        <Controller
          control={control}
          name="ended_at"
          render={({ field, fieldState }) => (
            <DateTimePicker
              value={field.value || undefined}
              onChange={(v) => field.onChange(v ?? "")}
              invalid={fieldState.invalid}
            />
          )}
        />
        <p className="text-muted-foreground text-xs">{t(`${prefix}.endedAtHint`)}</p>
        <FieldError errors={[endedAtError]} />
      </Field>
    </div>
  )
}

/** All quiz anti-cheat toggles, mirroring the backend Create/Update DTO booleans. */
export interface AntiCheatValues {
  no_back_navigation: boolean
  shuffle_questions: boolean
  shuffle_options: boolean
  track_tab_switches: boolean
  require_gps: boolean
  disable_copy_paste: boolean
  disable_right_click_shortcuts: boolean
  show_results: boolean
}

export type AntiCheatKey = keyof AntiCheatValues

/** Zod shape for all anti-cheat booleans — spread into quiz create/edit schemas. */
export const antiCheatSchemaShape = {
  no_back_navigation: z.boolean(),
  shuffle_questions: z.boolean(),
  shuffle_options: z.boolean(),
  track_tab_switches: z.boolean(),
  require_gps: z.boolean(),
  disable_copy_paste: z.boolean(),
  disable_right_click_shortcuts: z.boolean(),
  show_results: z.boolean(),
}

/** All anti-cheat toggles off — the default state for a new quiz. */
export const antiCheatDefaults: AntiCheatValues = {
  no_back_navigation: false,
  shuffle_questions: false,
  shuffle_options: false,
  track_tab_switches: false,
  require_gps: false,
  disable_copy_paste: false,
  disable_right_click_shortcuts: false,
  show_results: false,
}

/** Pulls the anti-cheat subset out of a quiz object, defaulting missing flags to false. */
export function antiCheatFromQuiz(
  quiz: Partial<Record<AntiCheatKey, boolean | undefined>>
): AntiCheatValues {
  return {
    no_back_navigation: quiz.no_back_navigation ?? false,
    shuffle_questions: quiz.shuffle_questions ?? false,
    shuffle_options: quiz.shuffle_options ?? false,
    track_tab_switches: quiz.track_tab_switches ?? false,
    require_gps: quiz.require_gps ?? false,
    disable_copy_paste: quiz.disable_copy_paste ?? false,
    disable_right_click_shortcuts: quiz.disable_right_click_shortcuts ?? false,
    show_results: quiz.show_results ?? false,
  }
}

interface AntiCheatToggle {
  key: AntiCheatKey
  icon: ReactNode
}

// Toggles grouped by how the backend treats them, so teachers understand what
// each control actually does. See internal/domain/quiz.go anti-cheat classes.
const ANTI_CHEAT_GROUPS: { id: string; toggles: AntiCheatToggle[] }[] = [
  {
    id: "enforced",
    toggles: [
      { key: "shuffle_questions", icon: <Shuffle /> },
      { key: "shuffle_options", icon: <Dices /> },
      { key: "no_back_navigation", icon: <Lock /> },
    ],
  },
  {
    id: "monitoring",
    toggles: [
      { key: "track_tab_switches", icon: <Eye /> },
      { key: "require_gps", icon: <MapPin /> },
    ],
  },
  {
    id: "restrictions",
    toggles: [
      { key: "disable_copy_paste", icon: <ClipboardX /> },
      { key: "disable_right_click_shortcuts", icon: <MousePointerClick /> },
    ],
  },
  {
    id: "results",
    toggles: [{ key: "show_results", icon: <Trophy /> }],
  },
]

// Maps snake_case DTO keys to camelCase i18n suffixes under `quizAntiCheat.toggles`.
const TOGGLE_I18N: Record<AntiCheatKey, string> = {
  no_back_navigation: "noBackNavigation",
  shuffle_questions: "shuffleQuestions",
  shuffle_options: "shuffleOptions",
  track_tab_switches: "trackTabSwitches",
  require_gps: "requireGps",
  disable_copy_paste: "disableCopyPaste",
  disable_right_click_shortcuts: "disableRightClickShortcuts",
  show_results: "showResults",
}

interface QuizFlagsFieldsProps {
  values: AntiCheatValues
  onChange: (key: AntiCheatKey, value: boolean) => void
}

export function QuizFlagsFields({ values, onChange }: QuizFlagsFieldsProps) {
  const { t } = useTranslation()
  const activeCount = ANTI_CHEAT_GROUPS.reduce(
    (n, g) => n + g.toggles.filter((tg) => values[tg.key]).length,
    0
  )

  return (
    <section className="border-foreground/10 bg-muted/20 flex flex-col gap-4 rounded-lg border p-4">
      <header className="flex items-start gap-3">
        <span className="bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-md [&_svg]:size-5">
          <ShieldCheck />
        </span>
        <div className="flex min-w-0 flex-1 flex-col gap-0.5">
          <div className="flex items-center gap-2">
            <h3 className="text-sm leading-snug font-semibold">{t("quizAntiCheat.title")}</h3>
            {activeCount > 0 && (
              <span className="bg-primary/10 text-primary rounded-full px-2 py-0.5 text-xs font-medium tabular-nums">
                {t("quizAntiCheat.activeCount", { count: activeCount })}
              </span>
            )}
          </div>
          <p className="text-muted-foreground text-xs leading-relaxed">
            {t("quizAntiCheat.subtitle")}
          </p>
        </div>
      </header>

      <div className="flex flex-col gap-4">
        {ANTI_CHEAT_GROUPS.map((group) => (
          <div key={group.id} className="flex flex-col gap-2">
            <span className="text-muted-foreground/80 text-[0.7rem] font-medium tracking-wide uppercase">
              {t(`quizAntiCheat.groups.${group.id}`)}
            </span>
            {group.toggles.map((toggle) => (
              <BooleanFieldRow
                key={toggle.key}
                icon={toggle.icon}
                label={t(`quizAntiCheat.toggles.${TOGGLE_I18N[toggle.key]}.label`)}
                hint={t(`quizAntiCheat.toggles.${TOGGLE_I18N[toggle.key]}.hint`)}
                checked={values[toggle.key]}
                onCheckedChange={(v) => onChange(toggle.key, v)}
              />
            ))}
          </div>
        ))}
      </div>
    </section>
  )
}
