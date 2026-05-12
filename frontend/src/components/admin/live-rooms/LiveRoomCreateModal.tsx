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
import { getGetAdminLiveRoomsQueryKey } from "@/api/admin-livesessions/admin-livesessions"
import { useGetClassesIdSessions } from "@/api/classes/classes"
import { usePostLiveRooms } from "@/api/live-sessions/live-sessions"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
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
import { cn } from "@/lib/utils"

const schema = z.object({
  class_id: z.string().uuid(),
  class_session_id: z.string().uuid(),
  max_participants: z.coerce.number().int().min(1).max(1000).optional(),
  auto_record: z.boolean().optional(),
  allow_mic_default: z.boolean().optional(),
  allow_camera_default: z.boolean().optional(),
  allow_screen_share_default: z.boolean().optional(),
})

type FormValues = z.infer<typeof schema>

interface LiveRoomCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const DEFAULTS: FormValues = {
  class_id: "",
  class_session_id: "",
  max_participants: 100,
  auto_record: false,
  allow_mic_default: true,
  allow_camera_default: true,
  allow_screen_share_default: false,
}

export function LiveRoomCreateModal({ open, onOpenChange }: LiveRoomCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  useEffect(() => {
    if (open) form.reset(DEFAULTS)
  }, [open])

  const createMutation = usePostLiveRooms({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.liveRooms.form.createSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminLiveRoomsQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const errors = form.formState.errors
  const selectedClassId = form.watch("class_id")
  const selectedSessionId = form.watch("class_session_id")
  const autoRecord = form.watch("auto_record")
  const allowMic = form.watch("allow_mic_default")
  const allowCamera = form.watch("allow_camera_default")
  const allowScreen = form.watch("allow_screen_share_default")

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        class_session_id: values.class_session_id,
        config: {
          max_participants: values.max_participants ?? 100,
          auto_record: !!values.auto_record,
          allow_mic_default: !!values.allow_mic_default,
          allow_camera_default: !!values.allow_camera_default,
          allow_screen_share_default: !!values.allow_screen_share_default,
        },
      },
    })
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("admin.liveRooms.form.createTitle")}
      description={t("admin.liveRooms.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={createMutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.class_id || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.class")}</FieldLabel>
          <ClassSelect
            value={selectedClassId || undefined}
            onChange={(id) => {
              form.setValue("class_id", id, { shouldValidate: true })
              form.setValue("class_session_id", "", { shouldValidate: true })
            }}
            placeholder={t("admin.liveRooms.form.classPlaceholder")}
          />
          <FieldError errors={[errors.class_id]} />
        </Field>

        <Field data-invalid={!!errors.class_session_id || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.session")}</FieldLabel>
          <SessionSelect
            classId={selectedClassId || undefined}
            value={selectedSessionId || undefined}
            onChange={(id) => form.setValue("class_session_id", id, { shouldValidate: true })}
            placeholder={t("admin.liveRooms.form.sessionPlaceholder")}
          />
          <FieldError errors={[errors.class_session_id]} />
        </Field>

        <Field data-invalid={!!errors.max_participants || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.maxParticipants")}</FieldLabel>
          <Input
            type="number"
            min={1}
            max={1000}
            {...form.register("max_participants")}
            placeholder={t("admin.liveRooms.form.maxParticipantsPlaceholder")}
          />
          <FieldError errors={[errors.max_participants]} />
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!allowMic}
            onCheckedChange={(c) => form.setValue("allow_mic_default", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.allowMicDefault")}
          </FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!allowCamera}
            onCheckedChange={(c) => form.setValue("allow_camera_default", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.allowCameraDefault")}
          </FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!allowScreen}
            onCheckedChange={(c) => form.setValue("allow_screen_share_default", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.allowScreenShareDefault")}
          </FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!autoRecord}
            onCheckedChange={(c) => form.setValue("auto_record", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.autoRecord")}
          </FieldLabel>
        </Field>
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
          <span className="text-muted-foreground">
            {placeholder ?? t("admin.liveRooms.form.classPlaceholder")}
          </span>
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
              ? (placeholder ?? t("admin.liveRooms.form.sessionPlaceholder"))
              : t("admin.liveRooms.form.sessionSelectClassFirst")}
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
