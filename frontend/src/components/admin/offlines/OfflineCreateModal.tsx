import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { useDebounce } from "use-debounce"
import { z } from "zod"

import { useGetAdminClasses } from "@/api/admin-classes/admin-classes"
import { getGetAdminOfflinesQueryKey } from "@/api/admin-offlines/admin-offlines"
import { useGetClassesIdSessions } from "@/api/classes/classes"
import { usePostOfflines, usePutOfflinesId } from "@/api/offlines/offlines"
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
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"

const createSchema = z.object({
  class_id: z.string().uuid(),
  class_session_id: z.string().uuid(),
  title: z.string().min(2),
  description: z.string().optional(),
  published_at: z.string().optional(),
})

const editSchema = z.object({
  title: z.string().min(2),
  description: z.string().optional(),
  published_at: z.string().optional(),
})

type CreateValues = z.infer<typeof createSchema>
type EditValues = z.infer<typeof editSchema>

interface OfflineCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  room?: OfflineRoom | null
}

function isoToLocalInput(iso?: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ""
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function localInputToISO(value?: string): string | undefined {
  if (!value) return undefined
  const d = new Date(value)
  if (isNaN(d.getTime())) return undefined
  return d.toISOString()
}

export function OfflineCreateModal({ open, onOpenChange, room }: OfflineCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!room

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: { class_id: "", class_session_id: "", title: "", description: "", published_at: "" },
  })

  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { title: "", description: "", published_at: "" },
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && room) {
      editForm.reset({
        title: room.title ?? "",
        description: room.description ?? "",
        published_at: isoToLocalInput(room.published_at),
      })
    } else {
      createForm.reset({ class_id: "", class_session_id: "", title: "", description: "", published_at: "" })
    }
  }, [open, room, isEdit])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminOfflinesQueryKey() })
  }

  const createMutation = usePostOfflines({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.offlines.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutOfflinesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.offlines.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const selectedClassId = createForm.watch("class_id")
  const selectedSessionId = createForm.watch("class_session_id")

  const onSubmitCreate = createForm.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        class_session_id: values.class_session_id,
        title: values.title,
        description: values.description,
        published_at: localInputToISO(values.published_at),
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (!room?.id) return
    updateMutation.mutate({
      id: room.id,
      data: {
        title: values.title,
        description: values.description,
        published_at: localInputToISO(values.published_at),
      },
    })
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.offlines.form.editTitle") : t("admin.offlines.form.createTitle")}
      description={
        isEdit ? t("admin.offlines.form.editDescription") : t("admin.offlines.form.createDescription")
      }
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {isEdit ? (
          <>
            <Field data-invalid={!!editErrors.title || undefined}>
              <FieldLabel>{t("admin.offlines.form.title")}</FieldLabel>
              <Input
                {...editForm.register("title")}
                placeholder={t("admin.offlines.form.titlePlaceholder")}
              />
              <FieldError errors={[editErrors.title]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.description")}</FieldLabel>
              <Textarea
                {...editForm.register("description")}
                placeholder={t("admin.offlines.form.descriptionPlaceholder")}
                rows={3}
              />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.publishedAt")}</FieldLabel>
              <Input type="datetime-local" {...editForm.register("published_at")} />
            </Field>
          </>
        ) : (
          <>
            <Field data-invalid={!!createErrors.class_id || undefined}>
              <FieldLabel>{t("admin.offlines.form.class")}</FieldLabel>
              <ClassSelect
                value={selectedClassId || undefined}
                onChange={(id) => {
                  createForm.setValue("class_id", id, { shouldValidate: true })
                  createForm.setValue("class_session_id", "", { shouldValidate: true })
                }}
                placeholder={t("admin.offlines.form.classPlaceholder")}
              />
              <FieldError errors={[createErrors.class_id]} />
            </Field>
            <Field data-invalid={!!createErrors.class_session_id || undefined}>
              <FieldLabel>{t("admin.offlines.form.session")}</FieldLabel>
              <SessionSelect
                classId={selectedClassId || undefined}
                value={selectedSessionId || undefined}
                onChange={(id) => createForm.setValue("class_session_id", id, { shouldValidate: true })}
                placeholder={t("admin.offlines.form.sessionPlaceholder")}
              />
              <FieldError errors={[createErrors.class_session_id]} />
            </Field>
            <Field data-invalid={!!createErrors.title || undefined}>
              <FieldLabel>{t("admin.offlines.form.title")}</FieldLabel>
              <Input
                {...createForm.register("title")}
                placeholder={t("admin.offlines.form.titlePlaceholder")}
              />
              <FieldError errors={[createErrors.title]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.description")}</FieldLabel>
              <Textarea
                {...createForm.register("description")}
                placeholder={t("admin.offlines.form.descriptionPlaceholder")}
                rows={3}
              />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.publishedAt")}</FieldLabel>
              <Input type="datetime-local" {...createForm.register("published_at")} />
            </Field>
          </>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

interface ClassSelectProps {
  value?: string
  onChange: (id: string) => void
  placeholder?: string
}

function ClassSelect({ value, onChange, placeholder }: ClassSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetAdminClasses({ search: debouncedSearch || undefined })
  const classes = (data?.status === 200 && data.data.data?.items) || []
  const selected = classes.find((c) => c.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button variant="outline" role="combobox" className="w-full justify-between font-normal" />
        }
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">{placeholder ?? t("admin.offlines.form.classPlaceholder")}</span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={t("admin.classes.searchPlaceholder")}
          />
          <CommandList>
            <CommandEmpty>{t("admin.classes.noResults")}</CommandEmpty>
            <CommandGroup>
              {classes.map((cls) => (
                <CommandItem
                  key={cls.id}
                  value={cls.name ?? ""}
                  onSelect={() => {
                    if (cls.id) {
                      onChange(cls.id)
                      setOpen(false)
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4", value === cls.id ? "opacity-100" : "opacity-0")} />
                  <div className="min-w-0">
                    <div className="truncate text-sm">{cls.name}</div>
                    {cls.user?.name && (
                      <div className="text-muted-foreground truncate text-xs">{cls.user.name}</div>
                    )}
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

interface SessionSelectProps {
  classId?: string
  value?: string
  onChange: (id: string) => void
  placeholder?: string
}

function SessionSelect({ classId, value, onChange, placeholder }: SessionSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetClassesIdSessions(
    classId ?? "",
    { search: debouncedSearch || undefined },
    { query: { enabled: !!classId } }
  )
  const sessions = (data?.status === 200 && data.data.data?.items) || []
  const selected = sessions.find((s) => s.id === value)

  return (
    <Popover open={open} onOpenChange={(o) => classId && setOpen(o)}>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            disabled={!classId}
            className="w-full justify-between font-normal"
          />
        }
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">
            {classId
              ? (placeholder ?? t("admin.offlines.form.sessionPlaceholder"))
              : t("admin.offlines.form.sessionSelectClassFirst")}
          </span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={t("admin.sessions.searchPlaceholder")}
          />
          <CommandList>
            <CommandEmpty>{t("admin.sessions.noResults")}</CommandEmpty>
            <CommandGroup>
              {sessions.map((s) => (
                <CommandItem
                  key={s.id}
                  value={s.name ?? ""}
                  onSelect={() => {
                    if (s.id) {
                      onChange(s.id)
                      setOpen(false)
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4", value === s.id ? "opacity-100" : "opacity-0")} />
                  <span className="truncate text-sm">{s.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
