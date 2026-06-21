import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { CalendarClockIcon, RadioIcon } from "lucide-react"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetLiveRoomsQueryKey,
  usePostLiveRooms,
  usePostLiveRoomsIdStart,
} from "@/api/live-sessions/live-sessions"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

const schema = z.object({
  name: z.string().max(255).optional(),
  mode: z.enum(["schedule", "now"]),
  scheduled_start_time: z.string().optional(),
  max_participants: z.coerce.number().int().min(1).max(1000).optional(),
  auto_record: z.boolean().optional(),
  allow_mic_default: z.boolean().optional(),
  allow_camera_default: z.boolean().optional(),
  allow_screen_share_default: z.boolean().optional(),
})

type FormInput = z.input<typeof schema>
type FormValues = z.infer<typeof schema>

const DEFAULTS: FormValues = {
  name: "",
  mode: "schedule",
  scheduled_start_time: "",
  max_participants: 100,
  auto_record: false,
  allow_mic_default: true,
  allow_camera_default: true,
  allow_screen_share_default: false,
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

  const form = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  useEffect(() => {
    if (open) form.reset(DEFAULTS)
  }, [open])

  const createMutation = usePostLiveRooms()
  const startMutation = usePostLiveRoomsIdStart()

  const errors = form.formState.errors
  const mode = form.watch("mode")
  const autoRecord = form.watch("auto_record")
  const allowMic = form.watch("allow_mic_default")
  const allowCamera = form.watch("allow_camera_default")
  const allowScreen = form.watch("allow_screen_share_default")

  const pending = createMutation.isPending || startMutation.isPending

  const onSubmit = form.handleSubmit(async (values) => {
    // Schedule mode must carry a real future timestamp — otherwise the room is
    // "scheduled" in name only (the old behaviour) and students get no countdown.
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
          name: values.name?.trim() || undefined,
          scheduled_start_time: scheduledISO,
          config: {
            max_participants: values.max_participants ?? 100,
            auto_record: !!values.auto_record,
            allow_mic_default: !!values.allow_mic_default,
            allow_camera_default: !!values.allow_camera_default,
            allow_screen_share_default: !!values.allow_screen_share_default,
          },
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

        {mode === "schedule" ? (
          <Field data-invalid={!!errors.scheduled_start_time || undefined}>
            <FieldLabel>{t("org.session.liveRooms.form.scheduledTime")}</FieldLabel>
            <Input type="datetime-local" {...form.register("scheduled_start_time")} />
            <FieldError errors={[errors.scheduled_start_time]} />
            <p className="text-muted-foreground text-xs">{t("org.session.liveRooms.form.scheduledTimeHint")}</p>
          </Field>
        ) : null}

        <Field data-invalid={!!errors.max_participants || undefined}>
          <FieldLabel>{t("org.session.liveRooms.form.maxParticipants")}</FieldLabel>
          <Input type="number" min={1} max={1000} {...form.register("max_participants")} placeholder="100" />
          <FieldError errors={[errors.max_participants]} />
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!allowMic} onCheckedChange={(c) => form.setValue("allow_mic_default", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.allowMic")}</FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!allowCamera} onCheckedChange={(c) => form.setValue("allow_camera_default", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.allowCamera")}</FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!allowScreen} onCheckedChange={(c) => form.setValue("allow_screen_share_default", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.allowScreen")}</FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!autoRecord} onCheckedChange={(c) => form.setValue("auto_record", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.autoRecord")}</FieldLabel>
        </Field>
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
