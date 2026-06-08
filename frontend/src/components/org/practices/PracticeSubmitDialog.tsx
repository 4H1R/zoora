import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetPracticesQueryKey,
  usePostPracticesIdSubmissions,
} from "@/api/practices/practices"
import {
  MediaAttachmentUploader,
  type PendingAttachment,
} from "@/components/media/MediaAttachmentUploader"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"

const schema = z.object({ content: z.string().optional() })
type Values = z.infer<typeof schema>

interface PracticeSubmitDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  practice: PracticeRoom | null
}

export function PracticeSubmitDialog({ open, onOpenChange, practice }: PracticeSubmitDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [attachments, setAttachments] = useState<PendingAttachment[]>([])

  const form = useForm<Values>({ resolver: zodResolver(schema), defaultValues: { content: "" } })

  useEffect(() => {
    if (open) {
      form.reset({ content: "" })
      setAttachments([])
    }
  }, [open])

  const submitMutation = usePostPracticesIdSubmissions({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.practices.submit.success"))
        queryClient.invalidateQueries({ queryKey: getGetPracticesQueryKey() })
        onOpenChange(false)
      },
      onError: () => toast.error(t("org.session.practices.submit.failed")),
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    if (!practice?.id) return
    submitMutation.mutate({
      id: practice.id,
      data: { content: values.content, attachments: attachments.map((a) => a.media_id) },
    })
  })

  const hasAnswer = !!form.watch("content")?.trim() || attachments.length > 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("org.session.practices.submit.title")}</DialogTitle>
          <DialogDescription>
            {practice?.title
              ? t("org.session.practices.submit.descriptionNamed", { title: practice.title })
              : t("org.session.practices.submit.description")}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={onSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel>{t("org.session.practices.submit.answer")}</FieldLabel>
              <Textarea
                {...form.register("content")}
                placeholder={t("org.session.practices.submit.placeholder")}
                rows={7}
              />
              <FieldError errors={[form.formState.errors.content]} />
            </Field>
            <Field>
              <FieldLabel>{t("org.session.practices.submit.attachments")}</FieldLabel>
              <MediaAttachmentUploader
                value={attachments}
                onChange={setAttachments}
                modelType="practice_submission"
              />
            </Field>
          </FieldGroup>

          <DialogFooter className="mt-6">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={submitMutation.isPending}
            >
              {t("common.cancel")}
            </Button>
            <Button type="submit" disabled={submitMutation.isPending || !hasAnswer}>
              {t("org.session.practices.submit.submit")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
