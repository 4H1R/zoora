import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRule as QuizRule,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { ChevronDownIcon, InboxIcon, ListChecksIcon, ShuffleIcon, Trash2Icon } from "lucide-react"
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
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
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
  const [advancedOpen, setAdvancedOpen] = useState(false)
  const [negDefault, setNegDefault] = useState<NegativeDefault>("default")
  // Explicit penalty value for the rule-wide default (per_wrong / accumulative).
  // Held as a string so the field can be empty and validated inline.
  const [negValue, setNegValue] = useState<string>("")

  useEffect(() => {
    if (open) {
      setBankId(undefined)
      setSelected({})
      setRandomCount(5)
      setAdvancedOpen(false)
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

  const showNegValue = negDefault !== "default" && NEGATIVE_VALUE_MODES.includes(negDefault)
  const negValueError: "required" | "positive" | "range" | null = (() => {
    if (!showNegValue) return null
    if (negValue.trim() === "") return "required"
    const n = Number(negValue)
    if (Number.isNaN(n) || n <= 0) return "positive"
    if (negDefault === "accumulative" && (!Number.isInteger(n) || n < 2 || n > 5)) return "range"
    return null
  })()

  const { data: rulesData, isLoading: rulesLoading } = useGetQuizzesIdRules(quizId ?? "", undefined, {
    query: { enabled: open && !!quizId },
  })
  const rules = (rulesData?.status === 200 && rulesData.data.data?.items) || []

  const { data: questionsData, isLoading: questionsLoading } = useGetQuestionBanksIdQuestions(
    bankId ?? "",
    { page_size: 200 },
    { query: { enabled: open && !!bankId } }
  )
  const questions: Question[] = (questionsData?.status === 200 && questionsData.data.data?.items) || []
  const bankTotal = questionsData?.status === 200 ? questionsData.data.data?.total : undefined

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
  const negativeDefaultValue = negDefault === "per_wrong" ? Number(negValue) : undefined
  const negativeDefaultWrongsPerPoint = negDefault === "accumulative" ? Number(negValue) : undefined

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

  const randomOverLimit = bankTotal != null && randomCount > bankTotal

  const addRandomRule = () => {
    if (!quizId || !bankId || randomCount <= 0 || randomOverLimit || negValueError) return
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

  const advancedSection = (
    <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen}>
      <CollapsibleTrigger className="text-muted-foreground hover:text-foreground flex items-center gap-1.5 text-xs font-medium transition-colors">
        <ChevronDownIcon
          className={cn("size-3.5 transition-transform", !advancedOpen && "-rotate-90 rtl:rotate-90")}
        />
        {t("admin.quizzes.questions.advancedToggle")}
        {negDefault !== "default" && (
          <Badge variant="secondary">{t(`admin.questions.form.negativeMark.modes.${negDefault}`)}</Badge>
        )}
      </CollapsibleTrigger>
      <CollapsibleContent className="pt-3">
        <NegativeDefaultField
          value={negDefault}
          onChange={changeNegDefault}
          numberValue={negValue}
          onNumberValueChange={setNegValue}
          error={negValueError}
        />
      </CollapsibleContent>
    </Collapsible>
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] flex-col sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {t("admin.quizzes.questions.title")}
            {quiz?.title && <span className="text-muted-foreground ms-2 text-sm font-normal">· {quiz.title}</span>}
          </DialogTitle>
          <DialogDescription>{t("admin.quizzes.questions.description")}</DialogDescription>
        </DialogHeader>

        <div className="-me-2 flex flex-col gap-5 overflow-y-auto pe-2">
          <section className="flex flex-col gap-2">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium">{t("admin.quizzes.questions.currentRules")}</h3>
              {rules.length > 0 && (
                <Badge variant="secondary" className="tabular-nums">
                  {t("admin.quizzes.questions.rulesSummary", {
                    questions: totalQuestions,
                    rules: rules.length,
                  })}
                </Badge>
              )}
            </div>
            <ul className="divide-border max-h-52 divide-y overflow-y-auto rounded-lg border">
              {rulesLoading && <li className="text-muted-foreground px-3 py-2 text-center text-sm">…</li>}
              {!rulesLoading && rules.length === 0 && (
                <li className="flex flex-col items-center gap-1 px-4 py-6 text-center">
                  <InboxIcon className="text-muted-foreground/60 mb-1 size-5" />
                  <span className="text-sm font-medium">{t("admin.quizzes.questions.empty")}</span>
                  <span className="text-muted-foreground text-xs">{t("admin.quizzes.questions.emptyHint")}</span>
                </li>
              )}
              {rules.map((rule) => (
                <RuleRow
                  key={rule.id}
                  rule={rule}
                  onDelete={() => rule.id && deleteRule.mutate({ ruleId: rule.id })}
                  deleting={deleteRule.isPending}
                />
              ))}
            </ul>
          </section>

          <section className="flex flex-col gap-4">
            <div className="flex items-center gap-3">
              <h3 className="shrink-0 text-sm font-medium">{t("admin.quizzes.questions.addSection")}</h3>
              <Separator className="flex-1" />
            </div>

            <Field>
              <FieldLabel className="flex items-center gap-2">
                <StepChip n={1} />
                {t("admin.quizzes.questions.stepBank")}
              </FieldLabel>
              <BankPicker value={bankId} onChange={setBankId} />
              {bankId && bankTotal != null && (
                <p className="text-muted-foreground text-xs tabular-nums">
                  {t("admin.quizzes.questions.bankCount", { count: bankTotal })}
                </p>
              )}
            </Field>

            <Field>
              <FieldLabel className="flex items-center gap-2">
                <StepChip n={2} />
                {t("admin.quizzes.questions.stepMethod")}
              </FieldLabel>
              <Tabs defaultValue="manual">
                <TabsList className="w-full">
                  <TabsTrigger value="manual">
                    <ListChecksIcon />
                    {t("admin.quizzes.questions.tabs.manual")}
                  </TabsTrigger>
                  <TabsTrigger value="random">
                    <ShuffleIcon />
                    {t("admin.quizzes.questions.tabs.random")}
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="manual">
                  <FieldGroup>
                    {!bankId ? (
                      <div className="text-muted-foreground rounded-lg border border-dashed px-4 py-8 text-center text-sm">
                        {t("admin.quizzes.questions.selectBankFirst")}
                      </div>
                    ) : (
                      <Field>
                        <FieldLabel className="flex items-center justify-between">
                          <span>{t("admin.quizzes.questions.selectQuestions")}</span>
                          <span className="text-muted-foreground text-xs tabular-nums">
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
                                  onCheckedChange={(c) => setSelected((prev) => ({ ...prev, [qid]: !!c }))}
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
                    {advancedSection}
                    <div className="flex justify-end">
                      <Button
                        type="button"
                        onClick={addManualRule}
                        disabled={!bankId || selectedIds.length === 0 || !!negValueError || createRule.isPending}
                      >
                        {selectedIds.length > 0
                          ? t("admin.quizzes.questions.addManualCount", { count: selectedIds.length })
                          : t("admin.quizzes.questions.addManual")}
                      </Button>
                    </div>
                  </FieldGroup>
                </TabsContent>

                <TabsContent value="random">
                  <FieldGroup>
                    <Field>
                      <FieldLabel>{t("admin.quizzes.questions.count")}</FieldLabel>
                      <Input
                        type="number"
                        min={1}
                        max={bankTotal}
                        value={randomCount}
                        aria-invalid={randomOverLimit}
                        onChange={(e) => setRandomCount(Number(e.target.value))}
                      />
                      {randomOverLimit ? (
                        <p className="text-destructive text-xs">
                          {t("admin.quizzes.questions.randomMax", { count: bankTotal })}
                        </p>
                      ) : (
                        <p className="text-muted-foreground text-xs">{t("admin.quizzes.questions.randomHint")}</p>
                      )}
                    </Field>
                    {advancedSection}
                    <div className="flex justify-end">
                      <Button
                        type="button"
                        onClick={addRandomRule}
                        disabled={
                          !bankId || randomCount <= 0 || randomOverLimit || !!negValueError || createRule.isPending
                        }
                      >
                        {randomCount > 0
                          ? t("admin.quizzes.questions.addRandomCount", { count: randomCount })
                          : t("admin.quizzes.questions.addRandom")}
                      </Button>
                    </div>
                  </FieldGroup>
                </TabsContent>
              </Tabs>
            </Field>
          </section>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// Small numbered marker for the two-step add flow — the bank must be picked
// before its questions can be listed, so the order carries real meaning.
function StepChip({ n }: { n: number }) {
  return (
    <span className="bg-primary/10 text-primary flex size-5 shrink-0 items-center justify-center rounded-full text-xs font-semibold tabular-nums">
      {n}
    </span>
  )
}

interface RuleRowProps {
  rule: QuizRule
  onDelete: () => void
  deleting: boolean
}

// One saved selection, described as a sentence ("7 manual questions from bank
// X") with its behavior and negative-marking default on the second line.
function RuleRow({ rule, onDelete, deleting }: RuleRowProps) {
  const { t } = useTranslation()
  const isRandom = rule.type === "random"

  const summary = t(
    isRandom ? "admin.quizzes.questions.summaryRandom" : "admin.quizzes.questions.summaryManual",
    { count: rule.count ?? 0, bank: rule.bank?.name ?? t("admin.quizzes.questions.deletedBank") }
  )

  const negSummary = (() => {
    const mode = rule.negative_default_mode
    if (!mode) return null
    if (mode === "none") return t("admin.quizzes.questions.neg.none")
    if (mode === "per_wrong") {
      return rule.negative_default_value != null
        ? t("admin.quizzes.questions.neg.perWrong", { value: rule.negative_default_value })
        : t("admin.quizzes.questions.neg.perWrongAuto")
    }
    return rule.negative_default_wrongs_per_point != null
      ? t("admin.quizzes.questions.neg.accumulative", { count: rule.negative_default_wrongs_per_point })
      : t("admin.quizzes.questions.neg.perWrongAuto")
  })()

  const details = [
    t(isRandom ? "admin.quizzes.questions.randomNote" : "admin.quizzes.questions.manualNote"),
    negSummary,
  ]
    .filter(Boolean)
    .join(" · ")

  return (
    <li className="flex items-center gap-3 px-3 py-2.5">
      <span
        className={cn(
          "flex size-8 shrink-0 items-center justify-center rounded-md",
          isRandom ? "bg-amber-500/10 text-amber-600 dark:text-amber-400" : "bg-primary/10 text-primary"
        )}
      >
        {isRandom ? <ShuffleIcon className="size-4" /> : <ListChecksIcon className="size-4" />}
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{summary}</p>
        <p className="text-muted-foreground truncate text-xs">{details}</p>
      </div>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={onDelete}
        disabled={deleting}
      >
        <Trash2Icon />
      </Button>
    </li>
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
// collapses smoothly for the modes that don't take one. The collapsible
// trigger above acts as this field's label.
function NegativeDefaultField({ value, onChange, numberValue, onNumberValueChange, error }: NegativeDefaultFieldProps) {
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
          <SelectItem value="default">{t("admin.quizzes.questions.negativeDefault.keepDefault")}</SelectItem>
          {NEGATIVE_MODES.map((m) => (
            <SelectItem key={m} value={m}>
              {t(`admin.questions.form.negativeMark.modes.${m}`)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-muted-foreground text-xs">{t("admin.quizzes.questions.negativeDefault.hint")}</p>

      <div
        className={cn(
          "grid transition-all duration-200 ease-out",
          showValue ? "mt-1 grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0"
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
            <p className="text-destructive text-xs">{t(`admin.quizzes.questions.negativeDefault.errors.${error}`)}</p>
          ) : (
            <p className="text-muted-foreground text-xs">{valueHint}</p>
          )}
        </div>
      </div>
    </Field>
  )
}
