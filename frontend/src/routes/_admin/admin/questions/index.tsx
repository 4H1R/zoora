import type { GithubCom4H1RZooraInternalDomainQuestion as Question } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { HelpCircleIcon, LibraryIcon, PlusIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminQuestions } from "@/api/admin-questionbanks/admin-questionbanks"
import { BankPicker } from "@/components/admin/forms/BankPicker"
import { BankCreateModal } from "@/components/admin/questions/BankCreateModal"
import { QuestionCreateModal } from "@/components/admin/questions/QuestionCreateModal"
import { QuestionTable } from "@/components/admin/questions/QuestionTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/questions/")({
  head: () => adminHead("admin.questions.title"),
  validateSearch: adminSearchSchema,
  component: QuestionsPage,
})

const TYPE_VALUES = ["descriptive", "short_answer", "choice"] as const

function QuestionsPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const [bankId, setBankId] = useState<string | undefined>(undefined)
  const [typeFilter, setTypeFilter] = useState<string>("all")

  const [formOpen, setFormOpen] = useState(false)
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null)
  const [bankFormOpen, setBankFormOpen] = useState(false)

  const handleEdit = (q: Question) => {
    setEditingQuestion(q)
    setFormOpen(true)
  }

  const handleCreate = () => {
    setEditingQuestion(null)
    setFormOpen(true)
  }

  const handleFormOpenChange = (open: boolean) => {
    setFormOpen(open)
    if (!open) setEditingQuestion(null)
  }

  const handleClearFilters = () => {
    setBankId(undefined)
    setTypeFilter("all")
  }

  const { data, isLoading } = useGetAdminQuestions({
    bank_id: bankId || undefined,
    type: typeFilter === "all" ? undefined : typeFilter,
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const questionsData = (data?.status === 200 && data.data.data) || undefined
  const questions = questionsData?.items ?? []
  const total = questionsData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <HelpCircleIcon />,
      label: t("admin.questions.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.questions.title")}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => setBankFormOpen(true)}>
              <LibraryIcon data-icon="inline-start" />
              {t("admin.questionBanks.newBank")}
            </Button>
            <Button size="sm" onClick={handleCreate} disabled={!bankId}>
              <PlusIcon data-icon="inline-start" />
              {t("admin.questions.newQuestion")}
            </Button>
          </div>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.questions.filter.bank")}</label>
          <BankPicker value={bankId} onChange={(id) => setBankId(id || undefined)} />
        </div>
        <div className="w-full sm:w-48">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.questions.filter.type")}</label>
          <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v ?? "all")}>
            <SelectTrigger>
              <SelectValue>
                {(v: string) => (v === "all" ? t("admin.questions.filter.allTypes") : t(`admin.questions.types.${v}`))}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t("admin.questions.filter.allTypes")}</SelectItem>
              {TYPE_VALUES.map((v) => (
                <SelectItem key={v} value={v}>
                  {t(`admin.questions.types.${v}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {(bankId || typeFilter !== "all") && (
          <Button variant="outline" size="sm" onClick={handleClearFilters}>
            <XIcon data-icon="inline-start" />
            {t("admin.questions.filter.clear")}
          </Button>
        )}
      </Card>
      <QuestionTable questions={questions} total={total} isLoading={isLoading} sorting={sorting} onEdit={handleEdit} />
      <QuestionCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        question={editingQuestion}
        defaultBankId={bankId}
      />
      <BankCreateModal open={bankFormOpen} onOpenChange={setBankFormOpen} onCreated={(id) => setBankId(id)} />
    </div>
  )
}
