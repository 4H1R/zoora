import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { CalendarClockIcon, RadioIcon } from "lucide-react"
import { useEffect, useRef } from "react"
import { Controller, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetLiveRoomsQueryKey,
  usePostLiveRooms,
  usePostLiveRoomsIdStart,
} from "@/api/live-sessions/live-sessions"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useSessionTitle } from "@/lib/session-title"
import { cn } from "@/lib/utils"

const schema = z.object({
  name: z.string().min(1).max(255),
  mode: z.enum(["schedule", "now"]),
  scheduled_start_time: z.string().optional(),
})

type FormInput = z.input<typeof schema>
type FormValues = z.infer<typeof schema>

const DEFAULTS: FormValues = {
  name: "",
  mode: "schedule",
  scheduled_start_time: "",
}

interface LiveRoomFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classSessionId: string
}

export function LiveRoomFormDialog({ open, onOpenChange, classSessionId }: LiveRoomFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const genTitle = useSessionTitle()

  const form = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  // Tracks the last value we auto-filled so we only overwrite the name field
  // while the user hasn't typed their own title.
  const autoNameRef = useRef("")

  useEffect(() => {
    if (open) {
      form.reset(DEFAULTS)
      autoNameRef.current = ""
    }
  }, [open])

  const createMutation = usePostLiveRooms()
  const startMutation = usePostLiveRoomsIdStart()

  const errors = form.formState.errors
  const mode = form.watch("mode")
  const scheduledTime = form.watch("scheduled_start_time")

  // Default the title to a readable label derived from the start time
  // ("کلاس دوشنبه ۲۰ تیر ساعت ۱۱:۳۰"), but never clobber a title the user typed.
  useEffect(() => {
    if (!open) return
    const base = mode === "now" ? new Date() : scheduledTime ? new Date(scheduledTime) : null
    if (!base || Number.isNaN(base.getTime())) return
    const current = form.getValues("name") ?? ""
    if (current !== "" && current !== autoNameRef.current) return
    const next = genTitle(base)
    autoNameRef.current = next
    form.setValue("name", next, { shouldValidate: true })
  }, [open, mode, scheduledTime])

  const pending = createMutation.isPending || startMutation.isPending

  const onSubmit = form.handleSubmit(async (values) => {
    // Must be a real future timestamp; without it students get no countdown.
    let scheduledISO: string | undefined
    if (values.mode === "schedule") {
      const ts = values.scheduled_start_time ? new Date(values.scheduled_start_time).getTime() : NaN
      if (Number.isNaN(ts) || ts <= Date.now()) {
        form.setError("scheduled_start_time", { message: t("org.session.liveRooms.form.scheduledTimeRequired") })
        return
      }
      scheduledISO = new Date(ts).toISOString()
    }

    try {
      const result = await createMutation.mutateAsync({
        data: {
          class_session_id: classSessionId,
          name: values.name.trim(),
          scheduled_start_time: scheduledISO,
        },
      })
      queryClient.invalidateQueries({ queryKey: getGetLiveRoomsQueryKey() })

      const room = (result.status === 201 && result.data.data) || undefined
      if (values.mode === "now" && room?.id) {
        await startMutation.mutateAsync({ id: room.id })
        onOpenChange(false)
        navigate({ to: "/live/$liveId", params: { liveId: room.id } })
        return
      }

      toast.success(t("org.session.liveRooms.form.createSuccess"))
      onOpenChange(false)
    } catch {
      toast.error(t("org.session.liveRooms.form.createError"))
    }
  })

  const submitLabel = mode === "now" ? t("org.session.liveRooms.form.createAndStart") : t("common.create")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("org.session.liveRooms.form.createTitle")}
      description={t("org.session.liveRooms.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={pending}
      submitLabel={submitLabel}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("org.session.liveRooms.form.name")}</FieldLabel>
          <Input {...form.register("name")} placeholder={t("org.session.liveRooms.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>

        <Field>
          <FieldLabel>{t("org.session.liveRooms.form.mode")}</FieldLabel>
          <div className="bg-muted/60 grid grid-cols-2 gap-1 rounded-xl p-1">
            <ModeButton
              active={mode === "schedule"}
              icon={<CalendarClockIcon className="size-4" />}
              label={t("org.session.liveRooms.form.scheduleLater")}
              onClick={() => form.setValue("mode", "schedule")}
            />
            <ModeButton
              active={mode === "now"}
              icon={<RadioIcon className="size-4" />}
              label={t("org.session.liveRooms.form.startNow")}
              onClick={() => form.setValue("mode", "now")}
            />
          </div>
        </Field>

        {mode === "schedule" && (
          <Field data-invalid={!!errors.scheduled_start_time || undefined}>
            <FieldLabel>{t("org.session.liveRooms.form.scheduledTime")}</FieldLabel>
            <Controller
              control={form.control}
              name="scheduled_start_time"
              render={({ field, fieldState }) => (
                <DateTimePicker
                  value={field.value || undefined}
                  onChange={(v) => field.onChange(v ?? "")}
                  invalid={fieldState.invalid}
                  minDate={new Date()}
                />
              )}
            />
            <FieldError errors={[errors.scheduled_start_time]} />
            <p className="text-muted-foreground text-xs">{t("org.session.liveRooms.form.scheduledTimeHint")}</p>
          </Field>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

function ModeButton({
  active,
  icon,
  label,
  onClick,
}: {
  active: boolean
  icon: React.ReactNode
  label: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
        active ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
      )}
    >
      {icon}
      {label}
    </button>
  )
}
