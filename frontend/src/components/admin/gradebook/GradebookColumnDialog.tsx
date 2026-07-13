import type { GithubCom4H1RZooraInternalDomainGradebookColumn as GradebookColumn } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { useDebounce } from "use-debounce"
import { z } from "zod"

import { useGetClassesIdSessions } from "@/api/classes/classes"
import {
  getGetClassesIdGradebookQueryKey,
  usePostClassesIdGradebookColumns,
  usePutClassesIdGradebookColumnsColumnId,
} from "@/api/gradebook/gradebook"
import { GithubCom4H1RZooraInternalDomainGradebookColumnType as ColumnType } from "@/api/model"
import { useGetPractices } from "@/api/practices/practices"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/utils"

const TYPES = [
  ColumnType.GradebookColumnAutoAttendance,
  ColumnType.GradebookColumnAutoPractice,
  ColumnType.GradebookColumnAutoQuiz,
  ColumnType.GradebookColumnManualGrade,
  ColumnType.GradebookColumnManualAttendance,
  ColumnType.GradebookColumnManualText,
] as const

const AUTO_TYPES = new Set<string>([
  ColumnType.GradebookColumnAutoAttendance,
  ColumnType.GradebookColumnAutoPractice,
  ColumnType.GradebookColumnAutoQuiz,
])

const maxScoreSchema = z.union([z.literal(""), z.coerce.number().positive()]).optional()

const createSchema = z
  .object({
    title: z.string().min(1),
    type: z.enum(TYPES),
    source_id: z.string().uuid().optional().or(z.literal("")),
    max_score: maxScoreSchema,
    order_index: z.coerce.number().int().min(0).default(0),
  })
  .refine(
    (v) => {
      if (AUTO_TYPES.has(v.type)) return !!v.source_id
      return true
    },
    { path: ["source_id"], params: { i18n: "validation.required" } }
  )

const editSchema = z.object({
  title: z.string().min(1),
  max_score: maxScoreSchema,
  order_index: z.coerce.number().int().min(0).default(0),
})

type CreateValues = z.input<typeof createSchema>
type EditValues = z.input<typeof editSchema>

interface GradebookColumnDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
  column?: GradebookColumn | null
}

export function GradebookColumnDialog({
  open,
  onOpenChange,
  classId,
  column,
}: GradebookColumnDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!column

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      title: "",
      type: ColumnType.GradebookColumnManualGrade,
      source_id: "",
      max_score: "",
      order_index: 0,
    },
  })

  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { title: "", max_score: "", order_index: 0 },
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && column) {
      editForm.reset({
        title: column.title ?? "",
        max_score: column.max_score ?? "",
        order_index: column.order_index ?? 0,
      })
    } else {
      createForm.reset({
        title: "",
        type: ColumnType.GradebookColumnManualGrade,
        source_id: "",
        max_score: "",
        order_index: 0,
      })
    }
  }, [open, column, isEdit])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetClassesIdGradebookQueryKey(classId) })
  }

  const createMutation = usePostClassesIdGradebookColumns({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.gradebook.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutClassesIdGradebookColumnsColumnId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.gradebook.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const selectedType = createForm.watch("type")
  const selectedSource = createForm.watch("source_id")
  const showSource = AUTO_TYPES.has(selectedType)

  const onSubmitCreate = createForm.handleSubmit((values) => {
    createMutation.mutate({
      id: classId,
      data: {
        title: values.title,
        type: values.type,
        source_id: values.source_id || undefined,
        max_score: values.max_score ? Number(values.max_score) : undefined,
        order_index: Number(values.order_index ?? 0),
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (!column?.id) return
    updateMutation.mutate({
      id: classId,
      columnId: column.id,
      data: {
        title: values.title,
        max_score: values.max_score ? Number(values.max_score) : undefined,
        order_index: Number(values.order_index ?? 0),
      },
    })
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={
        isEdit
          ? t("admin.gradebook.form.editTitle")
          : t("admin.gradebook.form.createTitle")
      }
      description={
        isEdit
          ? t("admin.gradebook.form.editDescription")
          : t("admin.gradebook.form.createDescription")
      }
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {isEdit ? (
          <>
            <Field data-invalid={!!editErrors.title || undefined}>
              <FieldLabel>{t("admin.gradebook.form.title")}</FieldLabel>
              <Input
                {...editForm.register("title")}
                placeholder={t("admin.gradebook.form.titlePlaceholder")}
              />
              <FieldError errors={[editErrors.title]} />
            </Field>
            <Field data-invalid={!!editErrors.max_score || undefined}>
              <FieldLabel>{t("admin.gradebook.form.maxScore")}</FieldLabel>
              <Input {...editForm.register("max_score")} type="number" min={0} step="any" />
              <FieldError errors={[editErrors.max_score]} />
            </Field>
            <Field data-invalid={!!editErrors.order_index || undefined}>
              <FieldLabel>{t("admin.gradebook.form.orderIndex")}</FieldLabel>
              <Input {...editForm.register("order_index")} type="number" min={0} />
              <FieldError errors={[editErrors.order_index]} />
            </Field>
          </>
        ) : (
          <>
            <Field data-invalid={!!createErrors.title || undefined}>
              <FieldLabel>{t("admin.gradebook.form.title")}</FieldLabel>
              <Input
                {...createForm.register("title")}
                placeholder={t("admin.gradebook.form.titlePlaceholder")}
              />
              <FieldError errors={[createErrors.title]} />
            </Field>
            <Field data-invalid={!!createErrors.type || undefined}>
              <FieldLabel>{t("admin.gradebook.form.type")}</FieldLabel>
              <Select
                value={selectedType}
                onValueChange={(v) => {
                  createForm.setValue("type", v as (typeof TYPES)[number], {
                    shouldValidate: true,
                  })
                  createForm.setValue("source_id", "", { shouldValidate: true })
                  createForm.setValue("max_score", "", { shouldValidate: true })
                }}
              >
                <SelectTrigger>
                  <SelectValue>{(v: (typeof TYPES)[number]) => t(`admin.gradebook.types.${v}`)}</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {TYPES.map((tp) => (
                    <SelectItem key={tp} value={tp}>
                      {t(`admin.gradebook.types.${tp}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldError errors={[createErrors.type]} />
            </Field>
            {showSource && (
              <Field data-invalid={!!createErrors.source_id || undefined}>
                <FieldLabel>{t(`admin.gradebook.form.source.${selectedType}`)}</FieldLabel>
                <SourcePicker
                  classId={classId}
                  type={selectedType}
                  value={selectedSource || undefined}
                  onChange={(id, maxScore) => {
                    createForm.setValue("source_id", id, { shouldValidate: true })
                    // Pre-fill max score from the quiz total / practice room max.
                    createForm.setValue("max_score", maxScore && maxScore > 0 ? maxScore : "", {
                      shouldValidate: true,
                    })
                  }}
                />
                <FieldError errors={[createErrors.source_id]} />
              </Field>
            )}
            <Field data-invalid={!!createErrors.max_score || undefined}>
              <FieldLabel>{t("admin.gradebook.form.maxScore")}</FieldLabel>
              <Input {...createForm.register("max_score")} type="number" min={0} step="any" />
              <FieldError errors={[createErrors.max_score]} />
            </Field>
            <Field data-invalid={!!createErrors.order_index || undefined}>
              <FieldLabel>{t("admin.gradebook.form.orderIndex")}</FieldLabel>
              <Input {...createForm.register("order_index")} type="number" min={0} />
              <FieldError errors={[createErrors.order_index]} />
            </Field>
          </>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

interface SourcePickerProps {
  classId: string
  type: (typeof TYPES)[number]
  value?: string
  onChange: (id: string, maxScore?: number) => void
}

function SourcePicker({ classId, type, value, onChange }: SourcePickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debounced] = useDebounce(search, 300)

  const sessionsQuery = useGetClassesIdSessions(
    classId,
    { search: debounced || undefined },
    { query: { enabled: type === ColumnType.GradebookColumnAutoAttendance } }
  )
  const practicesQuery = useGetPractices(
    { class_id: classId, search: debounced || undefined },
    { query: { enabled: type === ColumnType.GradebookColumnAutoPractice } }
  )
  const quizzesQuery = useGetQuizzes(
    { class_id: classId, search: debounced || undefined },
    { query: { enabled: type === ColumnType.GradebookColumnAutoQuiz } }
  )

  let items: { id: string; label: string; maxScore?: number }[] = []
  let selectedLabel: string | undefined

  if (type === ColumnType.GradebookColumnAutoAttendance) {
    const list = (sessionsQuery.data?.status === 200 && sessionsQuery.data.data.data?.items) || []
    items = list
      .filter((s): s is typeof s & { id: string } => !!s.id)
      .map((s) => ({ id: s.id, label: s.name ?? s.id }))
  } else if (type === ColumnType.GradebookColumnAutoPractice) {
    const list = (practicesQuery.data?.status === 200 && practicesQuery.data.data.data?.items) || []
    items = list
      .filter((p): p is typeof p & { id: string } => !!p.id)
      .map((p) => ({ id: p.id, label: p.title ?? p.id, maxScore: p.max_score }))
  } else if (type === ColumnType.GradebookColumnAutoQuiz) {
    const list = (quizzesQuery.data?.status === 200 && quizzesQuery.data.data.data?.items) || []
    items = list
      .filter((q): q is typeof q & { id: string } => !!q.id)
      .map((q) => ({ id: q.id, label: q.title ?? q.id, maxScore: q.total_score }))
  }

  selectedLabel = items.find((i) => i.id === value)?.label

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            className="w-full justify-between font-normal"
          />
        }
      >
        {selectedLabel ? (
          <span className="truncate">{selectedLabel}</span>
        ) : (
          <span className="text-muted-foreground">
            {t("admin.gradebook.form.sourcePlaceholder")}
          </span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={t("common.search")}
          />
          <CommandList>
            <CommandEmpty>{t("common.noResults")}</CommandEmpty>
            <CommandGroup>
              {items.map((item) => (
                <CommandItem
                  key={item.id}
                  value={item.label}
                  onSelect={() => {
                    onChange(item.id, item.maxScore)
                    setOpen(false)
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "me-2 size-4",
                      value === item.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  <span className="truncate text-sm">{item.label}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
