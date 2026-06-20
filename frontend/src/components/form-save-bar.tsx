import { RotateCcwIcon } from "lucide-react"
import type { FieldValues, UseFormReturn } from "react-hook-form"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

interface FormSaveBarProps<T extends FieldValues> {
  /** The react-hook-form instance. Drives visibility (isDirty) and pending (isSubmitting) state. */
  form: UseFormReturn<T>
  /** Submit handler — usually your `handleSubmit(...)` callback. */
  onSave: () => void
  /** Reset handler. Defaults to `form.reset()` (back to last default values). */
  onReset?: () => void
  /** Override the pending state instead of reading `formState.isSubmitting` (e.g. a mutation's isPending). */
  isPending?: boolean
  /** Override the visible state instead of reading `formState.isDirty`. */
  visible?: boolean
  /** Unsaved-changes message. Defaults to `common.unsavedChanges`. */
  message?: string
  /** Save button label. Defaults to `common.save`. */
  saveLabel?: string
}

/**
 * Floating "you have unsaved changes" bar with reset + save actions.
 * Slides in when the bound form is dirty. Drop it as a sibling of your form.
 */
export function FormSaveBar<T extends FieldValues>({
  form,
  onSave,
  onReset,
  isPending,
  visible,
  message,
  saveLabel,
}: FormSaveBarProps<T>) {
  const { t } = useTranslation()

  const isDirty = visible ?? form.formState.isDirty
  const pending = isPending ?? form.formState.isSubmitting
  const handleReset = onReset ?? (() => form.reset())

  return (
    <div
      className={cn(
        "pointer-events-none fixed inset-x-0 bottom-0 z-50 flex justify-center px-4 pb-6 transition-all duration-300",
        isDirty ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"
      )}
    >
      <div className="bg-popover/95 ring-foreground/10 pointer-events-auto flex w-full max-w-md items-center gap-3 rounded-full border py-2 ps-5 pe-2 shadow-xl ring-1 backdrop-blur">
        <span className="text-muted-foreground me-auto flex items-center gap-2 text-sm">
          <span className="bg-warning size-1.5 animate-pulse-dot rounded-full" />
          {message ?? t("common.unsavedChanges")}
        </span>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleReset}
          disabled={pending}
          className="rounded-full"
        >
          <RotateCcwIcon />
          {t("common.reset")}
        </Button>
        <Button
          type="button"
          size="sm"
          onClick={onSave}
          disabled={pending}
          className="rounded-full px-5 font-semibold"
        >
          {pending && <Spinner />}
          {pending ? t("common.saving") : (saveLabel ?? t("common.save"))}
        </Button>
      </div>
    </div>
  )
}
