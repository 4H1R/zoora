import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { Trash2Icon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { useGetQuestionBanksIdQuestions } from "@/api/question-banks/question-banks"
import {
  getGetQuizzesIdRulesQueryKey,
  useDeleteQuizzesRulesRuleId,
  useGetQuizzesIdRules,
  usePostQuizzesIdRules,
} from "@/api/quizzes/quizzes"
import { BankPicker } from "@/components/admin/forms/BankPicker"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import { cn } from "@/lib/utils"

const NEGATIVE_MODES = ["none", "per_wrong", "accumulative"] as const
type NegativeMode = (typeof NEGATIVE_MODES)[number]

// Modes that require an explicit penalty value from the user.
const NEGATIVE_VALUE_MODES: NegativeMode[] = ["per_wrong", "accumulative"]

// The negative-marking default for a whole selection. "default" keeps each
// question's own setting; a mode becomes the rule-wide default sent to the
// backend, which derives per-question numbers from each option count.
type NegativeDefault = "default" | NegativeMode

interface QuizQuestionsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  quiz: Quiz | null
}

export function QuizQuestionsDialog({ open, onOpenChange, quiz }: QuizQuestionsDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const quizId = quiz?.id

  const [bankId, setBankId] = useState<string | undefined>()
  const [selected, setSelected] = useState<Record<string, boolean>>({})
  const [randomCount, setRandomCount] = useState<number>(5)
  const [negDefault, setNegDefault] = useState<NegativeDefault>("default")
  // Explicit penalty value for the rule-wide default (per_wrong / accumulative).
  // Held as a string so the field can be empty and validated inline.
  const [negValue, setNegValue] = useState<string>("")

  useEffect(() => {
    if (open) {
      setBankId(undefined)
      setSelected({})
      setRandomCount(5)
      setNegDefault("default")
      setNegValue("")
    }
  }, [open])

  // Switching mode resets the value: hidden modes clear it, and per_wrong ↔
  // accumulative carry different units so a stale number would mislead.
  const changeNegDefault = (value: NegativeDefault) => {
    setNegDefault(value)
    setNegValue("")
  }

  const showNegValue =
    negDefault !== "default" && NEGATIVE_VALUE_MODES.includes(negDefault)
  const negValueError: "required" | "positive" | "range" | null = (() => {
    if (!showNegValue) return null
    if (negValue.trim() === "") return "required"
    const n = Number(negValue)
    if (Number.isNaN(n) || n <= 0) return "positive"
    if (negDefault === "accumulative" && (!Number.isInteger(n) || n < 2 || n > 5))
      return "range"
    return null
  })()

  const { data: rulesData, isLoading: rulesLoading } = useGetQuizzesIdRules(
    quizId ?? "",
    undefined,
    { query: { enabled: open && !!quizId } }
  )
  const rules = (rulesData?.status === 200 && rulesData.data.data?.items) || []

  const { data: questionsData, isLoading: questionsLoading } = useGetQuestionBanksIdQuestions(
    bankId ?? "",
    {},
    { query: { enabled: open && !!bankId } }
  )
  const questions: Question[] =
    (questionsData?.status === 200 && questionsData.data.data?.items) || []

  const invalidateRules = () => {
    if (quizId) {
      queryClient.invalidateQueries({ queryKey: getGetQuizzesIdRulesQueryKey(quizId) })
    }
  }

  const createRule = usePostQuizzesIdRules({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.quizzes.questions.ruleAdded"))
        setSelected({})
        invalidateRules()
      },
    },
  })

  const deleteRule = useDeleteQuizzesRulesRuleId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.quizzes.questions.ruleDeleted"))
        invalidateRules()
      },
    },
  })

  const selectedIds = Object.entries(selected)
    .filter(([, v]) => v)
    .map(([k]) => k)

  const negativeDefaultMode = negDefault === "default" ? undefined : negDefault
  // The single value input maps to a different backend field per mode.
  const negativeDefaultValue =
    negDefault === "per_wrong" ? Number(negValue) : undefined
  const negativeDefaultWrongsPerPoint =
    negDefault === "accumulative" ? Number(negValue) : undefined

  const addManualRule = () => {
    if (!quizId || !bankId || selectedIds.length === 0 || negValueError) return
    createRule.mutate({
      id: quizId,
      data: {
        type: "manual",
        bank_id: bankId,
        question_ids: selectedIds,
        count: selectedIds.length,
        is_dynamic: false,
        negative_default_mode: negativeDefaultMode,
        negative_default_value: negativeDefaultValue,
        negative_default_wrongs_per_point: negativeDefaultWrongsPerPoint,
      },
    })
  }

  const addRandomRule = () => {
    if (!quizId || !bankId || randomCount <= 0 || negValueError) return
    createRule.mutate({
      id: quizId,
      data: {
        type: "random",
        bank_id: bankId,
        count: randomCount,
        is_dynamic: true,
        negative_default_mode: negativeDefaultMode,
        negative_default_value: negativeDefaultValue,
        negative_default_wrongs_per_point: negativeDefaultWrongsPerPoint,
      },
    })
  }

  const totalQuestions = rules.reduce((sum, r) => sum + (r.count ?? 0), 0)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {t("admin.quizzes.questions.title")}
            {quiz?.title && (
              <span className="text-muted-foreground ms-2 text-sm font-normal">
                · {quiz.title}
              </span>
            )}
          </DialogTitle>
          <DialogDescription>{t("admin.quizzes.questions.description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
            <span>{t("admin.quizzes.questions.currentRules")}</span>
            <Badge variant="secondary">
              {rules.length} · {totalQuestions} {t("admin.quizzes.questions.questions")}
            </Badge>
          </div>

          <ul className="divide-border max-h-48 divide-y overflow-y-auto rounded-md border">
            {rulesLoading && (
              <li className="text-muted-foreground px-3 py-2 text-center text-sm">…</li>
            )}
            {!rulesLoading && rules.length === 0 && (
              <li className="text-muted-foreground px-3 py-3 text-center text-xs">
                {t("admin.quizzes.questions.empty")}
              </li>
            )}
            {rules.map((rule) => (
              <li key={rule.id} className="flex items-center justify-between px-3 py-2">
                <div className="flex min-w-0 flex-col">
                  <div className="flex items-center gap-2">
                    <Badge variant={rule.type === "random" ? "secondary" : "default"}>
                      {t(`admin.quizzes.questions.ruleType.${rule.type ?? "manual"}`)}
                    </Badge>
                    <span className="text-xs tabular-nums">
                      {rule.count ?? 0} {t("admin.quizzes.questions.questions")}
                    </span>
                  </div>
                  {rule.bank?.name && (
                    <span className="text-muted-foreground truncate text-xs">
                      {rule.bank.name}
                    </span>
                  )}
                </div>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                  onClick={() => rule.id && deleteRule.mutate({ ruleId: rule.id })}
                  disabled={deleteRule.isPending}
                >
                  <Trash2Icon />
                </Button>
              </li>
            ))}
          </ul>

          <Tabs defaultValue="manual">
            <TabsList>
              <TabsTrigger value="manual">{t("admin.quizzes.questions.tabs.manual")}</TabsTrigger>
              <TabsTrigger value="random">{t("admin.quizzes.questions.tabs.random")}</TabsTrigger>
            </TabsList>

            <TabsContent value="manual">
              <FieldGroup>
                <Field>
                  <FieldLabel>{t("admin.quizzes.questions.bank")}</FieldLabel>
                  <BankPicker value={bankId} onChange={setBankId} />
                </Field>
                <NegativeDefaultField
                  value={negDefault}
                  onChange={changeNegDefault}
                  numberValue={negValue}
                  onNumberValueChange={setNegValue}
                  error={negValueError}
                />
                {bankId && (
                  <Field>
                    <FieldLabel className="flex items-center justify-between">
                      <span>{t("admin.quizzes.questions.selectQuestions")}</span>
                      <span className="text-muted-foreground text-xs">
                        {selectedIds.length} {t("admin.quizzes.questions.selected")}
                      </span>
                    </FieldLabel>
                    <ul className="divide-border max-h-64 divide-y overflow-y-auto rounded-md border">
                      {questionsLoading && (
                        <li className="text-muted-foreground px-3 py-2 text-center text-sm">…</li>
                      )}
                      {!questionsLoading && questions.length === 0 && (
                        <li className="text-muted-foreground px-3 py-3 text-center text-xs">
                          {t("admin.questions.noResults")}
                        </li>
                      )}
                      {questions.map((q) => {
                        const qid = q.id ?? ""
                        const isSelected = !!selected[qid]
                        return (
                          <li key={q.id} className="flex items-start gap-3 px-3 py-2">
                            <Checkbox
                              checked={isSelected}
                              onCheckedChange={(c) =>
                                setSelected((prev) => ({ ...prev, [qid]: !!c }))
                              }
                            />
                            <div className="min-w-0 flex-1">
                              <div className="line-clamp-2 text-sm">{q.text}</div>
                              <div className="text-muted-foreground text-xs">
                                {t(`admin.questions.types.${q.type ?? "descriptive"}`)}
                              </div>
                            </div>
                          </li>
                        )
                      })}
                    </ul>
                  </Field>
                )}
                <div className="flex justify-end">
                  <Button
                    type="button"
                    onClick={addManualRule}
                    disabled={
                      !bankId ||
                      selectedIds.length === 0 ||
                      !!negValueError ||
                      createRule.isPending
                    }
                  >
                    {t("admin.quizzes.questions.addManual")}
                  </Button>
                </div>
              </FieldGroup>
            </TabsContent>

            <TabsContent value="random">
              <FieldGroup>
                <Field>
                  <FieldLabel>{t("admin.quizzes.questions.bank")}</FieldLabel>
                  <BankPicker value={bankId} onChange={setBankId} />
                </Field>
                <NegativeDefaultField
                  value={negDefault}
                  onChange={changeNegDefault}
                  numberValue={negValue}
                  onNumberValueChange={setNegValue}
                  error={negValueError}
                />
                <Field>
                  <FieldLabel>{t("admin.quizzes.questions.count")}</FieldLabel>
                  <Input
                    type="number"
                    min={1}
                    value={randomCount}
                    onChange={(e) => setRandomCount(Number(e.target.value))}
                  />
                  <p className="text-muted-foreground text-xs">
                    {t("admin.quizzes.questions.randomHint")}
                  </p>
                </Field>
                <div className="flex justify-end">
                  <Button
                    type="button"
                    onClick={addRandomRule}
                    disabled={
                      !bankId || randomCount <= 0 || !!negValueError || createRule.isPending
                    }
                  >
                    {t("admin.quizzes.questions.addRandom")}
                  </Button>
                </div>
              </FieldGroup>
            </TabsContent>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  )
}

interface NegativeDefaultFieldProps {
  value: NegativeDefault
  onChange: (value: NegativeDefault) => void
  numberValue: string
  onNumberValueChange: (value: string) => void
  error: "required" | "positive" | "range" | null
}

// One negative-marking default for the whole selection. Applies to every
// multiple-choice question added from the bank — manual and random alike. For
// per_wrong / accumulative the user enters the penalty value; the field below
// collapses smoothly for the modes that don't take one.
function NegativeDefaultField({
  value,
  onChange,
  numberValue,
  onNumberValueChange,
  error,
}: NegativeDefaultFieldProps) {
  const { t } = useTranslation()
  const showValue = value === "per_wrong" || value === "accumulative"
  const isAccumulative = value === "accumulative"
  const valueLabel = isAccumulative
    ? t("admin.questions.form.negativeMark.wrongsPerPoint")
    : t("admin.questions.form.negativeMark.negativeValue")
  const valueHint = isAccumulative
    ? t("admin.questions.form.negativeMark.accumulativeHint")
    : t("admin.questions.form.negativeMark.perWrongHint")

  return (
    <Field>
      <FieldLabel>{t("admin.quizzes.questions.negativeDefault.label")}</FieldLabel>
      <Select value={value} onValueChange={(v) => onChange(v as NegativeDefault)}>
        <SelectTrigger>
          <SelectValue>
            {(v: NegativeDefault) =>
              v === "default"
                ? t("admin.quizzes.questions.negativeDefault.keepDefault")
                : t(`admin.questions.form.negativeMark.modes.${v}`)
            }
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="default">
            {t("admin.quizzes.questions.negativeDefault.keepDefault")}
          </SelectItem>
          {NEGATIVE_MODES.map((m) => (
            <SelectItem key={m} value={m}>
              {t(`admin.questions.form.negativeMark.modes.${m}`)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-muted-foreground text-xs">
        {t("admin.quizzes.questions.negativeDefault.hint")}
      </p>

      <div
        className={cn(
          "grid transition-all duration-200 ease-out",
          showValue
            ? "mt-1 grid-rows-[1fr] opacity-100"
            : "grid-rows-[0fr] opacity-0"
        )}
      >
        <div className="overflow-hidden">
          <FieldLabel htmlFor="neg-default-value">{valueLabel}</FieldLabel>
          <Input
            id="neg-default-value"
            type="number"
            inputMode="decimal"
            min={isAccumulative ? 2 : 0}
            max={isAccumulative ? 5 : undefined}
            step={isAccumulative ? 1 : "any"}
            value={numberValue}
            aria-invalid={!!error}
            onChange={(e) => onNumberValueChange(e.target.value)}
          />
          {error ? (
            <p className="text-destructive text-xs">
              {t(`admin.quizzes.questions.negativeDefault.errors.${error}`)}
            </p>
          ) : (
            <p className="text-muted-foreground text-xs">{valueHint}</p>
          )}
        </div>
      </div>
    </Field>
  )
}
