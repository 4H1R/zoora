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
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"

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

  useEffect(() => {
    if (open) {
      setBankId(undefined)
      setSelected({})
      setRandomCount(5)
    }
  }, [open])

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
    createRule.mutate({
      id: quizId,
      data: {
        type: "manual",
        bank_id: bankId,
        question_ids: selectedIds,
        count: selectedIds.length,
        is_dynamic: false,
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
                      {questions.map((q) => (
                        <li key={q.id} className="flex items-start gap-3 px-3 py-2">
                          <Checkbox
                            checked={!!selected[q.id ?? ""]}
                            onCheckedChange={(c) =>
                              setSelected((prev) => ({
                                ...prev,
                                [q.id ?? ""]: !!c,
                              }))
                            }
                          />
                          <div className="min-w-0 flex-1">
                            <div className="line-clamp-2 text-sm">{q.text}</div>
                            <div className="text-muted-foreground text-xs">
                              {t(`admin.questions.types.${q.type ?? "descriptive"}`)}
                            </div>
                          </div>
                        </li>
                      ))}
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
