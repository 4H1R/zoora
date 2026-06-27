import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizQuestionNegativeOverride as NegativeOverride,
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

const NEGATIVE_MODES = ["none", "per_wrong", "accumulative"] as const
type NegativeMode = (typeof NEGATIVE_MODES)[number]

function clampInt(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}

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
  // Per-question negative-marking override keyed by question id. Absent =>
  // keep the question default. mode "none" means "no penalty for this question".
  const [overrides, setOverrides] = useState<Record<string, NegativeOverride>>({})

  useEffect(() => {
    if (open) {
      setBankId(undefined)
      setSelected({})
      setRandomCount(5)
      setOverrides({})
    }
  }, [open])

  const setOverride = (questionId: string, next: NegativeOverride | null) => {
    setOverrides((prev) => {
      if (!next) {
        const { [questionId]: _omit, ...rest } = prev
        return rest
      }
      return { ...prev, [questionId]: next }
    })
  }

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

  const addManualRule = () => {
    if (!quizId || !bankId || selectedIds.length === 0) return
    const negativeOverrides = selectedIds
      .map((id) => overrides[id])
      .filter((o): o is NegativeOverride => !!o)
    createRule.mutate({
      id: quizId,
      data: {
        type: "manual",
        bank_id: bankId,
        question_ids: selectedIds,
        count: selectedIds.length,
        is_dynamic: false,
        negative_overrides: negativeOverrides,
      },
    })
  }

  const addRandomRule = () => {
    if (!quizId || !bankId || randomCount <= 0) return
    createRule.mutate({
      id: quizId,
      data: {
        type: "random",
        bank_id: bankId,
        count: randomCount,
        is_dynamic: true,
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
                        const isChoice = q.type === "choice"
                        const override = overrides[qid]
                        return (
                          <li key={q.id} className="flex flex-col gap-2 px-3 py-2">
                            <div className="flex items-start gap-3">
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
                            </div>
                            {isSelected && isChoice && (
                              <QuestionOverrideControl
                                optionCount={q.options?.length ?? 0}
                                override={override}
                                onChange={(next) => setOverride(qid, next ? { ...next, question_id: qid } : null)}
                              />
                            )}
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
                    disabled={!bankId || selectedIds.length === 0 || createRule.isPending}
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
                    disabled={!bankId || randomCount <= 0 || createRule.isPending}
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

interface QuestionOverrideControlProps {
  optionCount: number
  override?: NegativeOverride
  onChange: (next: NegativeOverride | null) => void
}

// "default" keeps the question's own negative-marking; choosing a mode produces
// a per-question override sent in the rule's negative_overrides.
function QuestionOverrideControl({ optionCount, override, onChange }: QuestionOverrideControlProps) {
  const { t } = useTranslation()
  const selection: "default" | NegativeMode = override ? (override.mode as NegativeMode) : "default"

  const handleSelect = (value: string | null) => {
    if (!value || value === "default") {
      onChange(null)
      return
    }
    const mode = value as NegativeMode
    if (mode === "per_wrong") {
      onChange({ mode, negative_value: fractionForCount(optionCount), wrongs_per_point: 0 })
    } else if (mode === "accumulative") {
      onChange({ mode, negative_value: 0, wrongs_per_point: clampInt(optionCount, 2, 5) })
    } else {
      onChange({ mode, negative_value: 0, wrongs_per_point: 0 })
    }
  }

  return (
    <div className="ms-7 flex flex-col gap-2 rounded-md border border-dashed p-2">
      <span className="text-muted-foreground text-xs">
        {t("admin.quizzes.questions.override.label")}
      </span>
      <Select value={selection} onValueChange={handleSelect}>
        <SelectTrigger className="h-8">
          <SelectValue>
            {(value: "default" | NegativeMode) =>
              value === "default"
                ? t("admin.quizzes.questions.override.keepDefault")
                : t(`admin.questions.form.negativeMark.modes.${value}`)
            }
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="default">
            {t("admin.quizzes.questions.override.keepDefault")}
          </SelectItem>
          {NEGATIVE_MODES.map((m) => (
            <SelectItem key={m} value={m}>
              {t(`admin.questions.form.negativeMark.modes.${m}`)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {override?.mode === "per_wrong" && (
        <Input
          type="number"
          step="any"
          min={0}
          className="h-8"
          value={override.negative_value ?? 0}
          onChange={(e) =>
            onChange({ ...override, negative_value: Number(e.target.value), wrongs_per_point: 0 })
          }
        />
      )}
      {override?.mode === "accumulative" && (
        <Input
          type="number"
          min={2}
          max={5}
          step={1}
          className="h-8"
          value={override.wrongs_per_point ?? 0}
          onChange={(e) =>
            onChange({ ...override, wrongs_per_point: Number(e.target.value), negative_value: 0 })
          }
        />
      )}
    </div>
  )
}

// fractionForCount mirrors backend domain.FractionFor for prefill suggestions.
function fractionForCount(optionCount: number): number {
  switch (optionCount) {
    case 2:
      return 0.5
    case 3:
      return 0.33
    case 4:
      return 0.25
    case 5:
      return 0.2
  }
  if (optionCount <= 0) return 0
  return 1 / optionCount
}
